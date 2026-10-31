package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/model"
)

const (
	// AMap driving accepts at most 16 intermediate waypoints per request.
	amapMaxDrivingVias = 16
	// Keep the persisted JSON column bounded even for long multi-day routes.
	amapMaxPolylinePoints = 5000
)

// AmapClient talks to the AMap (Gaode) Web service API.
// All input/output coordinates are GCJ-02, matching the seeded places data.
type AmapClient struct {
	apiKey      string
	apiURL      string
	transitCity string
	client      *http.Client
}

func NewAmapClient(apiKey, apiURL, transitCity string) *AmapClient {
	if apiURL == "" {
		apiURL = "https://restapi.amap.com/v3/direction"
	}
	if transitCity == "" {
		transitCity = "广州"
	}
	return &AmapClient{
		apiKey:      apiKey,
		apiURL:      strings.TrimRight(apiURL, "/"),
		transitCity: transitCity,
		client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func (a *AmapClient) Enabled() bool { return a.apiKey != "" }

// FetchPolyline resolves a road polyline for the requested travel mode.
// It returns nil on any provider error so the caller falls back to straight lines.
func (a *AmapClient) FetchPolyline(ctx context.Context, waypoints []model.Waypoint, mode string) *RouteResult {
	if !a.Enabled() || len(waypoints) < 2 {
		return nil
	}
	switch mode {
	case "walking":
		return a.fetchPerLeg(ctx, waypoints, "walking")
	case "bicycling":
		return a.fetchPerLeg(ctx, waypoints, "bicycling")
	case "transit":
		return a.fetchTransit(ctx, waypoints)
	default:
		return a.fetchDriving(ctx, waypoints)
	}
}

func (a *AmapClient) fetchDriving(ctx context.Context, waypoints []model.Waypoint) *RouteResult {
	var points []model.Coordinate
	distance := 0
	for start := 0; start < len(waypoints)-1; {
		// origin + up to 16 vias + destination
		end := start + amapMaxDrivingVias + 2
		if end > len(waypoints) {
			end = len(waypoints)
		}
		seg := waypoints[start:end]
		params := url.Values{}
		params.Set("key", a.apiKey)
		params.Set("origin", amapLngLat(seg[0]))
		params.Set("destination", amapLngLat(seg[len(seg)-1]))
		params.Set("extensions", "base")
		if len(seg) > 2 {
			vias := make([]string, 0, len(seg)-2)
			for _, wp := range seg[1 : len(seg)-1] {
				vias = append(vias, amapLngLat(wp))
			}
			params.Set("waypoints", strings.Join(vias, ";"))
		}

		var out amapDriveResp
		if err := a.getJSON(ctx, a.apiURL+"/driving", params, &out); err != nil {
			log.Printf("amap driving request failed: %v", err)
			return nil
		}
		if out.Status != "1" || len(out.Route.Paths) == 0 {
			log.Printf("amap driving error: status=%s info=%s infocode=%s", out.Status, out.Info, out.Infocode)
			return nil
		}
		for _, p := range out.Route.Paths {
			for _, st := range p.Steps {
				points = append(points, parseAmapPolyline(st.Polyline)...)
			}
			distance += amapDistance(p.Distance)
		}
		start = end - 1
	}
	return finishRoute(points, distance)
}

// Walking and bicycling do not support intermediate waypoints, so each leg is
// requested separately and the segments are concatenated.
func (a *AmapClient) fetchPerLeg(ctx context.Context, waypoints []model.Waypoint, mode string) *RouteResult {
	var points []model.Coordinate
	distance := 0
	for i := 0; i+1 < len(waypoints); i++ {
		params := url.Values{}
		params.Set("key", a.apiKey)
		params.Set("origin", amapLngLat(waypoints[i]))
		params.Set("destination", amapLngLat(waypoints[i+1]))

		if mode == "bicycling" {
			var out amapBikeResp
			if err := a.getJSON(ctx, a.bicyclingURL(), params, &out); err != nil {
				log.Printf("amap bicycling request failed: %v", err)
				return nil
			}
			if out.Errcode != 0 || len(out.Data.Paths) == 0 {
				log.Printf("amap bicycling error: errcode=%d errmsg=%s", out.Errcode, out.Errmsg)
				return nil
			}
			for _, p := range out.Data.Paths {
				for _, st := range p.Steps {
					points = append(points, parseAmapPolyline(st.Polyline)...)
				}
				distance += amapDistance(p.Distance)
			}
			continue
		}

		var out amapWalkResp
		if err := a.getJSON(ctx, a.apiURL+"/walking", params, &out); err != nil {
			log.Printf("amap walking request failed: %v", err)
			return nil
		}
		if out.Status != "1" || len(out.Route.Paths) == 0 {
			log.Printf("amap walking error: status=%s info=%s infocode=%s", out.Status, out.Info, out.Infocode)
			return nil
		}
		for _, p := range out.Route.Paths {
			for _, st := range p.Steps {
				points = append(points, parseAmapPolyline(st.Polyline)...)
			}
			distance += amapDistance(p.Distance)
		}
	}
	return finishRoute(points, distance)
}

func (a *AmapClient) fetchTransit(ctx context.Context, waypoints []model.Waypoint) *RouteResult {
	params := url.Values{}
	params.Set("key", a.apiKey)
	params.Set("origin", amapLngLat(waypoints[0]))
	params.Set("destination", amapLngLat(waypoints[len(waypoints)-1]))
	params.Set("city", a.transitCity)
	params.Set("extensions", "base")

	var out amapTransitResp
	if err := a.getJSON(ctx, a.apiURL+"/transit/integrated", params, &out); err != nil {
		log.Printf("amap transit request failed: %v", err)
		return nil
	}
	if out.Status != "1" || len(out.Route.Transits) == 0 {
		log.Printf("amap transit error: status=%s info=%s infocode=%s", out.Status, out.Info, out.Infocode)
		return nil
	}

	var points []model.Coordinate
	distance := 0
	for _, seg := range out.Route.Transits[0].Segments {
		for _, st := range seg.Walking.Steps {
			points = append(points, parseAmapPolyline(st.Polyline)...)
		}
		for _, line := range seg.Bus.Buslines {
			points = append(points, parseAmapPolyline(line.Polyline)...)
		}
		if seg.Railway.Name != "" {
			for _, st := range seg.Railway.Steps {
				points = append(points, parseAmapPolyline(st.Polyline)...)
			}
		}
	}
	return finishRoute(points, distance)
}

func (a *AmapClient) bicyclingURL() string {
	// Bicycling lives on v4 while the rest of the route API is v3.
	if base := strings.Replace(a.apiURL, "/v3/direction", "/v4/direction", 1); base != a.apiURL {
		return base + "/bicycling"
	}
	return a.apiURL + "/bicycling"
}

func (a *AmapClient) getJSON(ctx context.Context, endpoint string, params url.Values, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, out)
}

// amapLngLat formats a waypoint in AMap's expected "lng,lat" order.
func amapLngLat(w model.Waypoint) string {
	return fmt.Sprintf("%.6f,%.6f", w.Lng, w.Lat)
}

// parseAmapPolyline parses "lng,lat;lng,lat;..." and swaps into {Lat, Lng}.
func parseAmapPolyline(raw string) []model.Coordinate {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	points := make([]model.Coordinate, 0, len(parts))
	for _, pair := range parts {
		latLng := strings.Split(strings.TrimSpace(pair), ",")
		if len(latLng) != 2 {
			continue
		}
		lng, errLng := strconv.ParseFloat(strings.TrimSpace(latLng[0]), 64)
		lat, errLat := strconv.ParseFloat(strings.TrimSpace(latLng[1]), 64)
		if errLng != nil || errLat != nil || lat == 0 || lng == 0 {
			continue
		}
		points = append(points, model.Coordinate{Lat: lat, Lng: lng})
	}
	return points
}

func finishRoute(points []model.Coordinate, distanceMeters int) *RouteResult {
	points = dedupeAdjacent(points)
	if len(points) < 2 {
		return nil
	}
	if len(points) > amapMaxPolylinePoints {
		points = decimate(points, amapMaxPolylinePoints)
	}
	return &RouteResult{Points: points, DistanceMeters: distanceMeters}
}

func dedupeAdjacent(points []model.Coordinate) []model.Coordinate {
	if len(points) < 2 {
		return points
	}
	out := make([]model.Coordinate, 0, len(points))
	out = append(out, points[0])
	for _, p := range points[1:] {
		last := out[len(out)-1]
		if p.Lat != last.Lat || p.Lng != last.Lng {
			out = append(out, p)
		}
	}
	return out
}

func decimate(points []model.Coordinate, max int) []model.Coordinate {
	if len(points) <= max {
		return points
	}
	step := float64(len(points)-1) / float64(max-1)
	out := make([]model.Coordinate, 0, max)
	for i := 0; i < max; i++ {
		out = append(out, points[int(float64(i)*step+0.5)])
	}
	return out
}

// --- AMap response shapes ---

type amapStep struct {
	Polyline string `json:"polyline"`
}

type amapPath struct {
	Distance json.Number `json:"distance"`
	Steps    []amapStep  `json:"steps"`
}

type amapDriveResp struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Route    struct {
		Paths []amapPath `json:"paths"`
	} `json:"route"`
}

type amapWalkResp struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Route    struct {
		Paths []amapPath `json:"paths"`
	} `json:"route"`
}

type amapBikeResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Data    struct {
		Paths []amapPath `json:"paths"`
	} `json:"data"`
}

type amapTransitResp struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Route    struct {
		Transits []struct {
			Segments []struct {
				Walking struct {
					Steps []amapStep `json:"steps"`
				} `json:"walking"`
				Bus struct {
					Buslines []struct {
						Polyline string `json:"polyline"`
					} `json:"buslines"`
				} `json:"bus"`
				Railway struct {
					Name  string     `json:"name"`
					Steps []amapStep `json:"steps"`
				} `json:"railway"`
			} `json:"segments"`
		} `json:"transits"`
	} `json:"route"`
}

// amapDistance accepts both v3 string distances ("1234") and v4 numeric ones.
func amapDistance(n json.Number) int {
	if n == "" {
		return 0
	}
	f, err := strconv.ParseFloat(string(n), 64)
	if err != nil {
		return 0
	}
	return int(f)
}
