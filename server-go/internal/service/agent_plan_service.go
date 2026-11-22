package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/planner"
	"github.com/travelgd/server-go/internal/repo"
)

type AgentPlanService struct {
	runs     *repo.AgentRepo
	guide    *repo.GuideRepo
	places   *repo.PlaceRepo
	routes   *repo.RouteRepo
	now      func() time.Time
	location *time.Location
}

func NewAgentPlanService(runs *repo.AgentRepo, guide *repo.GuideRepo, places *repo.PlaceRepo, routes *repo.RouteRepo) *AgentPlanService {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return &AgentPlanService{runs: runs, guide: guide, places: places, routes: routes, now: time.Now, location: loc}
}

// SaveDraft stores the server-owned solution as the next draft version.
func (s *AgentPlanService) SaveDraft(ctx context.Context, runID, conversationID int64, solution planner.Solution, title, summary string, baseVersion int) (int, []byte, error) {
	version := baseVersion + 1
	draft := planner.NewDraft(solution, title, summary, version, s.now().In(s.location))
	if err := planner.ValidateDraft(&draft, s.now().In(s.location)); err != nil {
		return 0, nil, common.BadRequest("plan validation failed: " + err.Error())
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		return 0, nil, common.Internal("could not encode plan")
	}
	if err := s.runs.SaveDraft(ctx, runID, conversationID, raw, version); err != nil {
		return 0, nil, err
	}
	return version, raw, nil
}

type ConfirmResult struct {
	RouteID         int64  `json:"routeId"`
	RouteName       string `json:"routeName"`
	TotalDistance   string `json:"totalDistance"`
	RouteConfidence string `json:"routeConfidence"`
}

// Confirm revalidates a draft against current data and writes it to routes.
func (s *AgentPlanService) Confirm(ctx context.Context, userID, runID int64, version int, idempotencyKey string) (*ConfirmResult, error) {
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 64 {
		return nil, common.BadRequest("idempotencyKey must be 8-64 characters")
	}
	run, err := s.runs.FindRun(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, common.NotFound("Run not found")
	}

	if run.RouteID != nil && *run.RouteID > 0 {
		rt, err := s.routes.FindByID(ctx, *run.RouteID, userID)
		if err != nil {
			return nil, err
		}
		if rt == nil {
			return nil, common.NotFound("Saved route not found")
		}
		confidence := ""
		var draft planner.Draft
		if len(run.PlanJSON) > 0 && json.Unmarshal(run.PlanJSON, &draft) == nil {
			confidence = draft.RouteConfidence
		}
		return &ConfirmResult{
			RouteID: rt.ID, RouteName: rt.RouteName,
			TotalDistance: rt.TotalDistance, RouteConfidence: confidence,
		}, nil
	}

	if run.Status != model.RunSucceeded {
		return nil, common.Conflict("Run is not finished successfully")
	}
	if len(run.PlanJSON) == 0 {
		return nil, common.Conflict("Run has no plan to confirm")
	}
	if run.PlanVersion != version {
		return nil, common.Conflict(fmt.Sprintf("Plan version changed (current %d)", run.PlanVersion))
	}

	var draft planner.Draft
	if err := json.Unmarshal(run.PlanJSON, &draft); err != nil {
		return nil, common.Internal("Stored plan is not valid JSON")
	}
	now := s.now().In(s.location)
	if err := planner.ValidateDraft(&draft, now); err != nil {
		return nil, common.Conflict("Plan is no longer valid: " + err.Error())
	}

	placeIDs := make([]int64, 0, len(draft.Stops))
	for _, stop := range draft.Stops {
		if stop.PlaceID > 0 {
			placeIDs = append(placeIDs, stop.PlaceID)
		}
	}
	places, err := s.places.FindByIDs(ctx, placeIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]model.Place, len(places))
	for _, p := range places {
		byID[p.ID] = p
	}
	for _, stop := range draft.Stops {
		if stop.PlaceID <= 0 {
			continue
		}
		p, ok := byID[stop.PlaceID]
		if !ok {
			return nil, common.Conflict(fmt.Sprintf("Place %d no longer exists", stop.PlaceID))
		}
		if abs(p.Lat-stop.Lat) > 0.001 || abs(p.Lng-stop.Lng) > 0.001 {
			return nil, common.Conflict(fmt.Sprintf("Place %d moved; please replan", stop.PlaceID))
		}
	}

	draft.Status = "confirmed"
	confirmedJSON, err := json.Marshal(draft)
	if err != nil {
		return nil, common.Internal("could not encode plan")
	}
	route := s.buildRoute(userID, draft, byID)
	routeID, err := s.runs.ConfirmPlan(ctx, repo.ConfirmPlanParams{
		UserID:         userID,
		RunID:          runID,
		ConversationID: run.ConversationID,
		IdempotencyKey: idempotencyKey,
		PlanJSON:       confirmedJSON,
		PlanVersion:    version,
		Route:          route,
	})
	if err != nil {
		return nil, err
	}
	return &ConfirmResult{
		RouteID:         routeID,
		RouteName:       route.RouteName,
		TotalDistance:   route.TotalDistance,
		RouteConfidence: draft.RouteConfidence,
	}, nil
}

func (s *AgentPlanService) buildRoute(userID int64, draft planner.Draft, places map[int64]model.Place) *model.Route {
	waypoints := make([]model.Waypoint, 0, len(draft.Stops))
	interests := make([]string, 0, 8)
	seenInterest := map[string]bool{}
	for _, stop := range draft.Stops {
		if stop.PlaceID <= 0 {
			continue
		}
		name := stop.Name
		nameEn := stop.NameEn
		description := draft.Summary
		if p, ok := places[stop.PlaceID]; ok {
			if name == "" {
				name = p.NameCn
			}
			if nameEn == "" {
				nameEn = p.NameEn
			}
			if description == "" {
				description = truncateRunes(p.DescriptionCn, 200)
			}
			for _, tag := range p.Tags {
				if !seenInterest[tag] && len(interests) < 8 {
					seenInterest[tag] = true
					interests = append(interests, tag)
				}
			}
		}
		waypoints = append(waypoints, model.Waypoint{
			Name:        name,
			NameEn:      nameEn,
			Lat:         stop.Lat,
			Lng:         stop.Lng,
			StayTime:    fmt.Sprintf("%d 分钟", stop.StayMinutes),
			Description: description,
		})
	}

	polyline := make([]model.Coordinate, 0, len(draft.Stops)*4)
	for _, seg := range draft.Segments {
		for _, point := range seg.Polyline {
			if len(polyline) > 0 {
				last := polyline[len(polyline)-1]
				if last.Lat == point.Lat && last.Lng == point.Lng {
					continue
				}
			}
			polyline = append(polyline, model.Coordinate{Lat: point.Lat, Lng: point.Lng})
		}
	}
	if len(polyline) < 2 {
		for _, wp := range waypoints {
			polyline = append(polyline, model.Coordinate{Lat: wp.Lat, Lng: wp.Lng})
		}
	}

	distance := 0
	for _, seg := range draft.Segments {
		distance += seg.DistanceMeters
	}
	totalDistance := formatDistance(distance)
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		title = "AI 广东行程"
	}
	return &model.Route{
		UserID:        userID,
		RouteName:     truncateRunes(title, 200),
		DurationDays:  1,
		TotalDistance: totalDistance,
		Mode:          planner.MapMode(draft.TransportMode),
		Pace:          "balanced",
		Interests:     interests,
		Waypoints:     waypoints,
		Polyline:      polyline,
	}
}

func formatDistance(meters int) string {
	if meters >= 1000 {
		return fmt.Sprintf("%.1fkm", float64(meters)/1000)
	}
	return fmt.Sprintf("%dm", meters)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
