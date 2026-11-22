package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/agent"
	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/planner"
	"github.com/travelgd/server-go/internal/repo"
)

// WeatherProvider keeps get_weather testable without a live AMap key.
type WeatherProvider interface {
	Enabled() bool
	Forecast(ctx context.Context, adcode string) (*ai.Weather, error)
}

// AgentTools implements the explicit tool allowlist. Identity always comes
// from ExecutionContext, never from model-supplied arguments.
type AgentTools struct {
	places    *repo.PlaceRepo
	favorites *repo.FavoriteRepo
	knowledge *repo.KnowledgeRepo
	routes    *repo.RouteRepo
	travel    *ai.TravelService
	plans     *AgentPlanService
	weather   WeatherProvider
	now       func() time.Time
	loc       *time.Location
}

func NewAgentTools(
	places *repo.PlaceRepo,
	favorites *repo.FavoriteRepo,
	knowledge *repo.KnowledgeRepo,
	routes *repo.RouteRepo,
	travel *ai.TravelService,
	plans *AgentPlanService,
	weather WeatherProvider,
) *AgentTools {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	if travel == nil {
		travel = ai.NewTravelService(nil, 4)
	}
	return &AgentTools{
		places: places, favorites: favorites, knowledge: knowledge,
		routes: routes, travel: travel, plans: plans, weather: weather,
		now: time.Now, loc: loc,
	}
}

func (t *AgentTools) All() []agent.Tool {
	return []agent.Tool{
		t.searchPlaces(),
		t.getPlace(),
		t.listFavorites(),
		t.searchKnowledge(),
		t.getWeather(),
		t.routeLeg(),
		t.planItinerary(),
		t.submitPlan(),
		t.listMyRoutes(),
	}
}

func (t *AgentTools) searchPlaces() agent.Tool {
	return agent.NewTool("search_places",
		"Search Guangdong attractions/places by keyword, tag, or indoor preference. Returns place ids and basic facts. Use it before plan_itinerary.",
		json.RawMessage(`{
			"type":"object",
			"properties":{
				"query":{"type":"string","description":"Keyword in Chinese or English, e.g. 陈家祠 or museum"},
				"tag":{"type":"string","description":"Optional tag such as 历史, 美食, 公园, 室内"},
				"indoor":{"type":"boolean","description":"Optional indoor filter"},
				"limit":{"type":"integer","description":"1-20, default 10"}
			},
			"required":["query"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				Query  string `json:"query"`
				Tag    string `json:"tag"`
				Indoor *bool  `json:"indoor"`
				Limit  int    `json:"limit"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			if args.Limit <= 0 || args.Limit > 20 {
				args.Limit = 10
			}
			list, err := t.places.Search(ctx, args.Query, args.Tag, args.Indoor, args.Limit)
			if err != nil {
				return agent.ToolResult{}, err
			}
			views := make([]map[string]any, 0, len(list))
			for _, p := range list {
				views = append(views, t.placeSummary(p, ec.Language, 300))
			}
			return agent.ToolResult{Content: toJSON(map[string]any{
				"places": views,
				"count":  len(views),
			})}, nil
		})
}

func (t *AgentTools) getPlace() agent.Tool {
	return agent.NewTool("get_place",
		"Get one place's full details by its numeric placeId.",
		json.RawMessage(`{
			"type":"object",
			"properties":{"placeId":{"type":"integer"}},
			"required":["placeId"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				PlaceID int64 `json:"placeId"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			if args.PlaceID <= 0 {
				return agent.ToolResult{}, fmt.Errorf("placeId must be positive")
			}
			p, err := t.places.FindByID(ctx, args.PlaceID)
			if err != nil {
				return agent.ToolResult{}, err
			}
			if p == nil {
				return agent.ToolResult{}, fmt.Errorf("place %d not found", args.PlaceID)
			}
			return agent.ToolResult{Content: toJSON(map[string]any{
				"place": t.placeSummary(*p, ec.Language, 1000),
			})}, nil
		})
}

func (t *AgentTools) listFavorites() agent.Tool {
	return agent.NewTool("list_favorites",
		"List the current user's favorite places. Use this when the user asks to prefer places they saved.",
		json.RawMessage(`{"type":"object","properties":{}}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			list, err := t.favorites.FindPlacesByUser(ctx, ec.UserID)
			if err != nil {
				return agent.ToolResult{}, err
			}
			if len(list) > 20 {
				list = list[:20]
			}
			views := make([]map[string]any, 0, len(list))
			for _, p := range list {
				view := t.placeSummary(p, ec.Language, 200)
				view["favorite"] = true
				views = append(views, view)
			}
			return agent.ToolResult{Content: toJSON(map[string]any{
				"favorites": views,
				"count":     len(views),
			})}, nil
		})
}

func (t *AgentTools) searchKnowledge() agent.Tool {
	return agent.NewTool("search_knowledge",
		"Search the Lingnan culture knowledge base for background, history, and local customs.",
		json.RawMessage(`{
			"type":"object",
			"properties":{
				"query":{"type":"string"},
				"limit":{"type":"integer","description":"1-5, default 3"}
			},
			"required":["query"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				Query string `json:"query"`
				Limit int    `json:"limit"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			if strings.TrimSpace(args.Query) == "" {
				return agent.ToolResult{}, fmt.Errorf("query is required")
			}
			if args.Limit <= 0 || args.Limit > 5 {
				args.Limit = 3
			}
			list, err := t.knowledge.SearchByKeyword(ctx, truncateRunes(args.Query, 60), args.Limit)
			if err != nil {
				return agent.ToolResult{}, err
			}
			entries := make([]map[string]string, 0, len(list))
			for _, k := range list {
				if isEnglish(ec.Language) {
					entries = append(entries, map[string]string{"title": k.TitleEn, "content": k.ContentEn})
				} else {
					entries = append(entries, map[string]string{"title": k.TitleCn, "content": k.ContentCn})
				}
			}
			return agent.ToolResult{Content: toJSON(map[string]any{"entries": entries})}, nil
		})
}

func (t *AgentTools) getWeather() agent.Tool {
	return agent.NewTool("get_weather",
		"Get today's and next 3 days' weather forecast for a Guangdong city. Use it when the user asks about weather or before planning outdoor activities.",
		json.RawMessage(`{
			"type":"object",
			"properties":{
				"city":{"type":"string","description":"Guangdong city name in Chinese or pinyin, e.g. 广州, 深圳, Guangzhou"}
			},
			"required":["city"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				City string `json:"city"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			city := strings.TrimSpace(args.City)
			if city == "" {
				return weatherUnavailable("city is required", nil), nil
			}
			adcode, ok := ai.ResolveAdcode(city)
			if !ok {
				return weatherUnavailable("unsupported city", ai.SupportedWeatherCities()), nil
			}
			if t.weather == nil || !t.weather.Enabled() {
				return weatherUnavailable("weather service not configured", nil), nil
			}
			forecast, err := t.weather.Forecast(ctx, adcode)
			if err != nil {
				// Do not leak the upstream body or the API key to the model.
				log.Printf("get_weather %s: %v", adcode, err)
				return weatherUnavailable("weather service unavailable", nil), nil
			}
			return agent.ToolResult{Content: toJSON(map[string]any{
				"available":  true,
				"city":       forecast.City,
				"province":   forecast.Province,
				"adcode":     forecast.Adcode,
				"reportTime": forecast.ReportTime,
				"casts":      forecast.Casts,
			})}, nil
		})
}

func weatherUnavailable(reason string, supported []string) agent.ToolResult {
	payload := map[string]any{"available": false, "reason": reason}
	if len(supported) > 0 {
		payload["supportedCities"] = supported
	}
	return agent.ToolResult{Content: toJSON(payload)}
}

func (t *AgentTools) routeLeg() agent.Tool {
	return agent.NewTool("route_leg",
		"Estimate one travel leg between two GCJ-02 coordinates. Returns duration, distance and whether the provider fell back to an estimate.",
		json.RawMessage(`{
			"type":"object",
			"properties":{
				"from":{"type":"object","properties":{"lat":{"type":"number"},"lng":{"type":"number"}},"required":["lat","lng"]},
				"to":{"type":"object","properties":{"lat":{"type":"number"},"lng":{"type":"number"}},"required":["lat","lng"]},
				"mode":{"type":"string","enum":["WALK","RIDE","DRIVE","TAXI","TRANSIT"]}
			},
			"required":["from","to","mode"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				From planner.Coordinate `json:"from"`
				To   planner.Coordinate `json:"to"`
				Mode string             `json:"mode"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			if !validCoordinate(args.From.Lat, args.From.Lng) || !validCoordinate(args.To.Lat, args.To.Lng) {
				return agent.ToolResult{}, fmt.Errorf("coordinates must be within China (GCJ-02)")
			}
			mode, ok := planner.NormalizeMode(args.Mode)
			if !ok {
				return agent.ToolResult{}, fmt.Errorf("unsupported mode %q", args.Mode)
			}
			info, err := t.travel.Route(ctx, mode, args.From, args.To)
			if err != nil {
				return agent.ToolResult{}, err
			}
			return agent.ToolResult{Content: toJSON(map[string]any{
				"mode":            mode,
				"durationMinutes": info.DurationMinutes,
				"distanceMeters":  info.DistanceMeters,
				"estimated":       info.Estimated,
			})}, nil
		})
}

func (t *AgentTools) planItinerary() agent.Tool {
	return agent.NewTool("plan_itinerary",
		"Compute up to 3 deterministic itineraries from 1-4 candidate places. Use when you have a time window. Returns solutionIds; pass one to submit_plan. Never compute times yourself.",
		json.RawMessage(`{
			"type":"object",
			"properties":{
				"placeIds":{"type":"array","items":{"type":"integer"},"description":"1-4 place ids, most relevant first"},
				"when":{"type":"string","description":"e.g. 明天下午, 周六下午两点到五点, 2026-11-07T14:00:00+08:00/2026-11-07T18:00:00+08:00"},
				"durationMinutes":{"type":"integer","description":"Optional total available minutes"},
				"mode":{"type":"string","enum":["WALK","RIDE","DRIVE","TAXI","TRANSIT"]},
				"bufferMinutes":{"type":"integer","description":"Tail buffer, default 10"},
				"origin":{"type":"object","properties":{"name":{"type":"string"},"lat":{"type":"number"},"lng":{"type":"number"}},"required":["lat","lng"]},
				"returnToStart":{"type":"boolean"},
				"stays":{"type":"object","description":"Optional placeId -> stay minutes, e.g. {\"12\":90}"}
			},
			"required":["placeIds","when","mode"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				PlaceIDs        []int64         `json:"placeIds"`
				When            string          `json:"when"`
				DurationMinutes int             `json:"durationMinutes"`
				Mode            string          `json:"mode"`
				BufferMinutes   int             `json:"bufferMinutes"`
				Origin          *planner.Origin `json:"origin"`
				ReturnToStart   bool            `json:"returnToStart"`
				Stays           map[string]int  `json:"stays"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			ids := dedupeIDs(args.PlaceIDs)
			if len(ids) == 0 {
				return agent.ToolResult{}, fmt.Errorf("placeIds is required (1-4 places)")
			}
			if len(ids) > planner.MaxCandidates {
				ids = ids[:planner.MaxCandidates]
			}
			mode, ok := planner.NormalizeMode(args.Mode)
			if !ok {
				return agent.ToolResult{}, fmt.Errorf("unsupported mode %q", args.Mode)
			}
			if args.DurationMinutes < 0 || args.DurationMinutes > 720 {
				return agent.ToolResult{}, fmt.Errorf("durationMinutes must be 0-720")
			}
			if args.BufferMinutes < 0 {
				args.BufferMinutes = 0
			}
			if args.BufferMinutes > 60 {
				args.BufferMinutes = 60
			}
			if args.BufferMinutes == 0 {
				args.BufferMinutes = planner.DefaultBuffer
			}
			if args.Origin != nil {
				if !validCoordinate(args.Origin.Lat, args.Origin.Lng) {
					return agent.ToolResult{}, fmt.Errorf("origin coordinates must be within China")
				}
				if args.Origin.Name == "" {
					args.Origin.Name = "出发点"
				}
			}
			window, err := planner.ParseWindow(args.When, args.DurationMinutes, t.now().In(t.loc))
			if err != nil {
				return agent.ToolResult{}, err
			}

			places, err := t.places.FindByIDs(ctx, ids)
			if err != nil {
				return agent.ToolResult{}, err
			}
			byID := make(map[int64]model.Place, len(places))
			for _, p := range places {
				byID[p.ID] = p
			}
			favorites, err := t.favorites.FindPlacesByUser(ctx, ec.UserID)
			if err != nil {
				return agent.ToolResult{}, err
			}
			favSet := make(map[int64]bool, len(favorites))
			for _, p := range favorites {
				favSet[p.ID] = true
			}

			candidates := make([]planner.Place, 0, len(ids))
			for rank, id := range ids {
				p, ok := byID[id]
				if !ok {
					return agent.ToolResult{}, fmt.Errorf("place %d not found", id)
				}
				stay := p.Duration
				if stay <= 0 {
					stay = planner.DefaultStay
				}
				if override, ok := args.Stays[strconv.FormatInt(id, 10)]; ok && override > 0 {
					stay = override
				}
				source := "SYSTEM"
				if favSet[id] {
					source = "FAVORITE"
				}
				candidates = append(candidates, planner.Place{
					PlaceID: id, Name: p.NameCn, NameEn: p.NameEn,
					Lat: p.Lat, Lng: p.Lng, StayMinutes: stay,
					Source: source, Rank: rank,
				})
			}

			points := make([]planner.Coordinate, 0, len(candidates)+1)
			if args.Origin != nil {
				points = append(points, planner.Coordinate{Lat: args.Origin.Lat, Lng: args.Origin.Lng})
			}
			for _, c := range candidates {
				points = append(points, planner.Coordinate{Lat: c.Lat, Lng: c.Lng})
			}
			t.travel.Prefetch(ctx, mode, points, 4)

			outcome, err := planner.Plan(ctx, planner.Request{
				WindowStart:        window.Start,
				WindowEnd:          window.End,
				Candidates:         candidates,
				TransportMode:      mode,
				BufferAfterMinutes: args.BufferMinutes,
				HeadBufferMinutes:  args.BufferMinutes,
				Origin:             args.Origin,
				ReturnToStart:      args.ReturnToStart,
				Travel:             t.travel,
			})
			if err != nil {
				return agent.ToolResult{}, err
			}
			if len(outcome.Solutions) == 0 {
				return agent.ToolResult{Content: toJSON(map[string]any{
					"solutions":   []any{},
					"suggestions": outcome.Suggestions,
				})}, nil
			}

			callNo := ec.State.Incr("planner.calls")
			store, _ := ec.State.Get("planner.solutions")
			solutions, _ := store.(map[string]planner.Solution)
			if solutions == nil {
				solutions = make(map[string]planner.Solution)
			}
			views := make([]map[string]any, 0, len(outcome.Solutions))
			for i := range outcome.Solutions {
				id := fmt.Sprintf("p%d-%s", callNo, outcome.Solutions[i].SolutionID)
				outcome.Solutions[i].SolutionID = id
				solutions[id] = outcome.Solutions[i]
				views = append(views, solutionView(outcome.Solutions[i]))
			}
			ec.State.Set("planner.solutions", solutions)

			return agent.ToolResult{Content: toJSON(map[string]any{
				"solutions":   views,
				"suggestions": outcome.Suggestions,
				"window": map[string]string{
					"start": window.Start.Format(time.RFC3339),
					"end":   window.End.Format(time.RFC3339),
				},
			})}, nil
		})
}

func (t *AgentTools) submitPlan() agent.Tool {
	return agent.NewTool("submit_plan",
		"Save one planner solution as a draft for the user to confirm. Only pass a solutionId returned by plan_itinerary; the server fills in every stop and time.",
		json.RawMessage(`{
			"type":"object",
			"properties":{
				"solutionId":{"type":"string"},
				"title":{"type":"string","description":"Short plan title"},
				"summary":{"type":"string","description":"One or two sentences about the plan"}
			},
			"required":["solutionId"]
		}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				SolutionID string `json:"solutionId"`
				Title      string `json:"title"`
				Summary    string `json:"summary"`
			}
			if err := decodeArgs(raw, &args); err != nil {
				return agent.ToolResult{}, err
			}
			store, _ := ec.State.Get("planner.solutions")
			solutions, _ := store.(map[string]planner.Solution)
			solution, ok := solutions[args.SolutionID]
			if !ok {
				return agent.ToolResult{}, fmt.Errorf("unknown solutionId %q; call plan_itinerary first and use one of its ids", args.SolutionID)
			}
			title := strings.TrimSpace(args.Title)
			if title == "" {
				title = defaultPlanTitle(solution)
			}
			base := ec.PlanVersion
			if v, ok := ec.State.Get("plan.version"); ok {
				if n, ok := v.(int); ok {
					base = n
				}
			}
			version, planJSON, err := t.plans.SaveDraft(ctx, ec.RunID, ec.ConversationID, solution, title, args.Summary, base)
			if err != nil {
				return agent.ToolResult{}, err
			}
			ec.State.Set("plan.version", version)
			return agent.ToolResult{
				Content: toJSON(map[string]any{
					"status":     "draft_saved",
					"version":    version,
					"solutionId": args.SolutionID,
					"title":      title,
				}),
				Plan: &agent.PlanDraft{
					Version: version,
					Title:   title,
					Summary: args.Summary,
					Plan:    planJSON,
				},
			}, nil
		})
}

func (t *AgentTools) listMyRoutes() agent.Tool {
	return agent.NewTool("list_my_routes",
		"List the current user's recently saved routes for reference.",
		json.RawMessage(`{"type":"object","properties":{}}`),
		func(ctx context.Context, ec *agent.ExecutionContext, raw json.RawMessage) (agent.ToolResult, error) {
			list, err := t.routes.FindByUser(ctx, ec.UserID)
			if err != nil {
				return agent.ToolResult{}, err
			}
			if len(list) > 10 {
				list = list[:10]
			}
			views := make([]map[string]any, 0, len(list))
			for _, rt := range list {
				views = append(views, map[string]any{
					"id":            rt.ID,
					"routeName":     rt.RouteName,
					"durationDays":  rt.DurationDays,
					"mode":          rt.Mode,
					"totalDistance": rt.TotalDistance,
					"createdAt":     rt.CreatedAt.Format(time.RFC3339),
				})
			}
			return agent.ToolResult{Content: toJSON(map[string]any{"routes": views})}, nil
		})
}

func (t *AgentTools) placeSummary(p model.Place, language string, descriptionLimit int) map[string]any {
	name := p.NameCn
	description := p.DescriptionCn
	if isEnglish(language) {
		name = p.NameEn
		description = p.DescriptionEn
	}
	return map[string]any{
		"placeId":  p.ID,
		"name":     name,
		"nameCn":   p.NameCn,
		"nameEn":   p.NameEn,
		"tags":     p.Tags,
		"indoor":   p.Indoor,
		"duration": p.Duration,
		"lat":      p.Lat,
		"lng":      p.Lng,
		"address":  "",
		"summary":  truncateRunes(description, descriptionLimit),
		"favorite": false,
	}
}

func solutionView(sol planner.Solution) map[string]any {
	stops := make([]map[string]any, 0, len(sol.Stops))
	for _, s := range sol.Stops {
		stops = append(stops, map[string]any{
			"placeId":     s.PlaceID,
			"name":        s.Name,
			"arrivalTime": s.ArrivalTime.Format("2006-01-02 15:04"),
			"leaveTime":   s.LeaveTime.Format("2006-01-02 15:04"),
			"stayMinutes": s.StayMinutes,
			"source":      s.Source,
		})
	}
	segments := make([]map[string]any, 0, len(sol.Segments))
	for _, seg := range sol.Segments {
		segments = append(segments, map[string]any{
			"fromIndex":       seg.FromIndex,
			"toIndex":         seg.ToIndex,
			"mode":            seg.TransportMode,
			"durationMinutes": seg.DurationMin,
			"distanceMeters":  seg.DistanceMeters,
			"estimated":       seg.Estimated,
		})
	}
	return map[string]any{
		"solutionId":             sol.SolutionID,
		"score":                  math.Round(sol.Score*1000) / 1000,
		"routeConfidence":        sol.RouteConfidence,
		"windowStart":            sol.WindowStart.Format("2006-01-02 15:04"),
		"windowEnd":              sol.WindowEnd.Format("2006-01-02 15:04"),
		"totalTravelMinutes":     sol.TotalTravelMinutes,
		"totalDistanceMeters":    sol.TotalDistanceMeters,
		"bufferRemainingMinutes": sol.BufferRemainingMinutes,
		"stops":                  stops,
		"segments":               segments,
	}
}

func defaultPlanTitle(sol planner.Solution) string {
	names := make([]string, 0, 3)
	for _, s := range sol.Stops {
		if s.PlaceID > 0 {
			names = append(names, s.Name)
		}
		if len(names) == 3 {
			break
		}
	}
	if len(names) == 0 {
		return "AI 广东行程"
	}
	return strings.Join(names, " · ")
}

func decodeArgs(raw json.RawMessage, target any) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		raw = []byte(`{}`)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return nil
}

func toJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return `{"error":"could not encode result"}`
	}
	return string(raw)
}

func dedupeIDs(ids []int64) []int64 {
	seen := make(map[int64]bool, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func validCoordinate(lat, lng float64) bool {
	if math.IsNaN(lat) || math.IsNaN(lng) || math.IsInf(lat, 0) || math.IsInf(lng, 0) {
		return false
	}
	return lat >= 18 && lat <= 54 && lng >= 73 && lng <= 135
}

func isEnglish(language string) bool {
	return strings.HasPrefix(strings.ToLower(language), "en")
}
