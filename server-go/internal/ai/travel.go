package ai

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/planner"
)

// TravelService adapts AMap to the planner's TravelProvider contract.
// It memoizes legs for the lifetime of a run, deduplicates concurrent
// requests, caps upstream concurrency, and never fails a plan because a map
// provider is down: it degrades to a clearly-marked straight-line estimate.
type TravelService struct {
	amap  *AmapClient
	sem   chan struct{}
	group singleflight.Group

	mu    sync.RWMutex
	cache map[string]cacheEntry
	now   func() time.Time
}

type cacheEntry struct {
	info    planner.TravelInfo
	expires time.Time
}

var estimateSpeedKMH = map[string]float64{
	planner.ModeWalk:    4.5,
	planner.ModeRide:    15,
	planner.ModeDrive:   30,
	planner.ModeTaxi:    30,
	planner.ModeTransit: 15,
}

func NewTravelService(amap *AmapClient, maxConcurrency int) *TravelService {
	if maxConcurrency < 1 {
		maxConcurrency = 4
	}
	return &TravelService{
		amap:  amap,
		sem:   make(chan struct{}, maxConcurrency),
		cache: make(map[string]cacheEntry),
		now:   time.Now,
	}
}

func (s *TravelService) Enabled() bool { return s.amap != nil && s.amap.Enabled() }

// Route returns a single leg, preferring a real AMap response and falling
// back to an estimate on any provider error.
func (s *TravelService) Route(ctx context.Context, mode string, from, to planner.Coordinate) (planner.TravelInfo, error) {
	if err := ctx.Err(); err != nil {
		return planner.TravelInfo{}, err
	}
	normalized, ok := planner.NormalizeMode(mode)
	if !ok {
		normalized = planner.ModeDrive
	}
	key := legKey(normalized, from, to)

	if info, ok := s.get(key); ok {
		return info, nil
	}

	v, _, _ := s.group.Do(key, func() (interface{}, error) {
		if info, ok := s.get(key); ok {
			return info, nil
		}
		info := s.fetch(ctx, normalized, from, to)
		s.put(key, info)
		return info, nil
	})
	info, _ := v.(planner.TravelInfo)
	return info, nil
}

func (s *TravelService) fetch(ctx context.Context, mode string, from, to planner.Coordinate) planner.TravelInfo {
	if !s.Enabled() {
		return estimate(mode, from, to)
	}
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return estimate(mode, from, to)
	}

	legCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	amapMode := mapToAmapMode(mode)
	leg, err := s.amap.LegRoute(legCtx, amapMode, toModelCoord(from), toModelCoord(to))
	if err != nil || leg == nil {
		return estimate(mode, from, to)
	}
	points := make([]planner.Coordinate, 0, len(leg.Points))
	for _, p := range leg.Points {
		points = append(points, planner.Coordinate{Lat: p.Lat, Lng: p.Lng})
	}
	return planner.TravelInfo{
		DurationMinutes: leg.DurationMinutes,
		DistanceMeters:  leg.DistanceMeters,
		Estimated:       false,
		Polyline:        points,
	}
}

// Prefetch warms the cache for a set of candidate points (origin included)
// so the deterministic solver does not serialize every AMap call.
func (s *TravelService) Prefetch(ctx context.Context, mode string, points []planner.Coordinate, maxParallel int) {
	if len(points) < 2 {
		return
	}
	if maxParallel < 1 {
		maxParallel = 4
	}
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxParallel)
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			from, to := points[i], points[j]
			g.Go(func() error {
				_, _ = s.Route(gctx, mode, from, to)
				return nil
			})
		}
	}
	_ = g.Wait()
}

func (s *TravelService) get(key string) (planner.TravelInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.cache[key]
	if !ok || s.now().After(entry.expires) {
		return planner.TravelInfo{}, false
	}
	return entry.info, true
}

func (s *TravelService) put(key string, info planner.TravelInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.cache) > 2048 {
		s.cache = make(map[string]cacheEntry)
	}
	s.cache[key] = cacheEntry{info: info, expires: s.now().Add(10 * time.Minute)}
}

func estimate(mode string, from, to planner.Coordinate) planner.TravelInfo {
	km := haversineKM(from, to) * 1.3
	speed := estimateSpeedKMH[mode]
	if speed <= 0 {
		speed = 30
	}
	minutes := int(math.Ceil(km / speed * 60))
	if minutes < 1 {
		minutes = 1
	}
	return planner.TravelInfo{
		DurationMinutes: minutes,
		DistanceMeters:  int(math.Round(km * 1000)),
		Estimated:       true,
		Polyline:        []planner.Coordinate{from, to},
	}
}

func haversineKM(a, b planner.Coordinate) float64 {
	const earthRadiusKM = 6371
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(a.Lat*math.Pi/180)*math.Cos(b.Lat*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusKM * math.Asin(math.Sqrt(h))
}

func legKey(mode string, from, to planner.Coordinate) string {
	return fmt.Sprintf("%s|%.5f,%.5f|%.5f,%.5f", mode, from.Lat, from.Lng, to.Lat, to.Lng)
}

func mapToAmapMode(mode string) string {
	switch mode {
	case planner.ModeWalk:
		return "walking"
	case planner.ModeRide:
		return "bicycling"
	case planner.ModeTransit:
		return "transit"
	default:
		return "driving"
	}
}

func toModelCoord(c planner.Coordinate) model.Coordinate {
	return model.Coordinate{Lat: c.Lat, Lng: c.Lng}
}
