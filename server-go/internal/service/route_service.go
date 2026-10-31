package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type RouteService struct {
	routes *repo.RouteRepo
	qwen   *ai.QwenClient
	mapAPI ai.RouteProvider
}

func NewRouteService(routes *repo.RouteRepo, qwen *ai.QwenClient, mapAPI ai.RouteProvider) *RouteService {
	return &RouteService{routes: routes, qwen: qwen, mapAPI: mapAPI}
}

type RouteGenRequest struct {
	Days      int      `json:"days"`
	Interests []string `json:"interests"`
	Pace      string   `json:"pace"`
	Mode      string   `json:"mode"`
}

func (s *RouteService) Generate(ctx context.Context, userID int64, req RouteGenRequest) (*model.Route, error) {
	if req.Days < 1 || req.Days > 7 {
		return nil, common.BadRequest("days must be between 1 and 7")
	}
	interests := req.Interests
	if len(interests) == 0 {
		interests = []string{"sightseeing"}
	}
	pace := normalizeChoice(req.Pace, []string{"relaxed", "balanced", "packed"}, "balanced")
	mode := normalizeChoice(req.Mode, []string{"driving", "walking", "bicycling", "transit"}, "driving")

	systemPrompt := fmt.Sprintf(`You are an expert travel planner for Guangzhou. Generate a detailed %d-day itinerary with a %s pace, focusing on these interests: %s.
You MUST return a VALID JSON object matching the structure: { "routeName": string, "duration": string, "totalDistance": string, "waypoints": [{"day": number, "name": string, "nameEn": string, "latitude": number, "longitude": number, "description": string, "stayTime": string}] }.
Ensure coordinates are accurate for Guangzhou. Do NOT output markdown.`,
		req.Days, pace, strings.Join(interests, ", "))

	raw, err := s.qwen.ChatDefault(ctx, []ai.QwenMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "Plan my trip in Guangzhou."},
	})
	if err != nil {
		return nil, err
	}

	route, err := parseRoute(raw, req.Days, interests, pace, mode)
	if err != nil {
		return nil, err
	}

	// Enrich with a real road polyline when a map provider is configured; otherwise straight lines.
	geometry := s.mapAPI.FetchPolyline(ctx, route.Waypoints, mode)
	if geometry != nil && len(geometry.Points) > 1 {
		route.Polyline = geometry.Points
	} else {
		route.Polyline = straightLines(route.Waypoints)
	}
	if route.TotalDistance == "" {
		if geometry != nil && geometry.DistanceMeters > 0 {
			route.TotalDistance = formatMeters(geometry.DistanceMeters)
		} else {
			route.TotalDistance = estimateDistance(route.Polyline)
		}
	}

	route.UserID = userID
	if err := s.routes.Insert(ctx, route); err != nil {
		return nil, err
	}
	return s.routes.FindByID(ctx, route.ID, userID)
}

func parseRoute(raw string, days int, interests []string, pace, mode string) (*model.Route, error) {
	var node struct {
		RouteName     string `json:"routeName"`
		Duration      string `json:"duration"`
		TotalDistance string `json:"totalDistance"`
		Waypoints     []struct {
			Day         int     `json:"day"`
			Name        string  `json:"name"`
			NameEn      string  `json:"nameEn"`
			Latitude    float64 `json:"latitude"`
			Longitude   float64 `json:"longitude"`
			StayTime    string  `json:"stayTime"`
			Description string  `json:"description"`
		} `json:"waypoints"`
	}
	if err := json.Unmarshal([]byte(ai.StripFences(raw)), &node); err != nil || len(node.Waypoints) == 0 {
		return nil, common.BadGateway("Could not parse AI route response")
	}
	waypoints := make([]model.Waypoint, 0, len(node.Waypoints))
	for _, w := range node.Waypoints {
		nameEn := w.NameEn
		if nameEn == "" {
			nameEn = w.Name
		}
		waypoints = append(waypoints, model.Waypoint{
			Name: w.Name, NameEn: nameEn, Lat: w.Latitude, Lng: w.Longitude,
			StayTime: w.StayTime, Description: w.Description,
		})
	}
	return &model.Route{
		RouteName:     orDefault(node.RouteName, "Guangzhou Trip"),
		DurationDays:  days,
		TotalDistance: node.TotalDistance,
		Mode:          mode,
		Pace:          pace,
		Interests:     interests,
		Waypoints:     waypoints,
	}, nil
}

func straightLines(waypoints []model.Waypoint) []model.Coordinate {
	points := make([]model.Coordinate, 0, len(waypoints))
	for _, wp := range waypoints {
		points = append(points, model.Coordinate{Lat: wp.Lat, Lng: wp.Lng})
	}
	return points
}

func estimateDistance(polyline []model.Coordinate) string {
	total := 0.0
	for i := 1; i < len(polyline); i++ {
		total += haversine(polyline[i-1], polyline[i])
	}
	if total >= 1000 {
		return fmt.Sprintf("%.1fkm", total/1000)
	}
	return fmt.Sprintf("%.0fm", total)
}

func formatMeters(meters int) string {
	if meters >= 1000 {
		return fmt.Sprintf("%.1fkm", float64(meters)/1000)
	}
	return fmt.Sprintf("%dm", meters)
}

func haversine(a, b model.Coordinate) float64 {
	const R = 6371000
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(a.Lat*math.Pi/180)*math.Cos(b.Lat*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * R * math.Asin(math.Sqrt(h))
}

func normalizeChoice(value string, allowed []string, fallback string) string {
	for _, a := range allowed {
		if value == a {
			return a
		}
	}
	return fallback
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func (s *RouteService) ListMine(ctx context.Context, userID int64) ([]model.Route, error) {
	return s.routes.FindByUser(ctx, userID)
}

func (s *RouteService) GetMine(ctx context.Context, userID, id int64) (*model.Route, error) {
	rt, err := s.routes.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, common.NotFound("Route not found")
	}
	return rt, nil
}
