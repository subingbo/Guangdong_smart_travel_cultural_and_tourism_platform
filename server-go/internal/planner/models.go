// Package planner is a deterministic spatiotemporal itinerary solver.
// It never calls an LLM and never touches the database: callers provide
// candidates, a time window and a TravelProvider.
package planner

import (
	"context"
	"time"
)

const (
	// Version is stamped on drafts and traces so algorithm changes are visible.
	Version = "deterministic-v1"

	MaxCandidates    = 4
	MaxPermutations  = 24
	MaxSolutions     = 3
	DefaultStay      = 60
	DefaultBuffer    = 10
	MaxDraftStops    = 8
	MaxWindowHours   = 24
	MaxPlanHorizon   = 30 * 24 * time.Hour
	MinStartLeadTime = 5 * time.Minute
)

const (
	ModeWalk    = "WALK"
	ModeRide    = "RIDE"
	ModeDrive   = "DRIVE"
	ModeTaxi    = "TAXI"
	ModeTransit = "TRANSIT"
)

const (
	ConfidenceAMAP      = "AMAP"
	ConfidenceEstimated = "ESTIMATED"
	ConfidenceNone      = "NONE"
)

// Coordinate is a GCJ-02 point, matching the seeded places data.
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Interval is a half-open busy interval [Start, End).
type Interval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Place is a candidate the solver may schedule.
type Place struct {
	PlaceID     int64   `json:"placeId"`
	Name        string  `json:"name"`
	NameEn      string  `json:"nameEn,omitempty"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	StayMinutes int     `json:"stayMinutes"`
	Source      string  `json:"source"` // SYSTEM | FAVORITE | USER_LOCATION
	Rank        int     `json:"rank"`   // 0 is the most relevant candidate
}

// Origin is an optional starting point (user location or a hotel).
type Origin struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

type Stop struct {
	PlaceID       int64     `json:"placeId"`
	Name          string    `json:"name"`
	NameEn        string    `json:"nameEn,omitempty"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	Source        string    `json:"source"`
	ArrivalTime   time.Time `json:"arrivalTime"`
	LeaveTime     time.Time `json:"leaveTime"`
	StayMinutes   int       `json:"stayMinutes"`
	TransportMode string    `json:"transportMode"`
}

type Segment struct {
	FromIndex      int          `json:"fromIndex"`
	ToIndex        int          `json:"toIndex"`
	TransportMode  string       `json:"transportMode"`
	DistanceMeters int          `json:"distanceMeters"`
	DurationMin    int          `json:"durationMinutes"`
	Estimated      bool         `json:"estimated"`
	Polyline       []Coordinate `json:"polyline,omitempty"`
}

type ScoreBreakdown struct {
	RankScore      float64 `json:"rankScore"`
	RemainingScore float64 `json:"remainingScore"`
	FavoriteScore  float64 `json:"favoriteScore"`
	TravelScore    float64 `json:"travelScore"`
}

func (s ScoreBreakdown) Total() float64 {
	return 0.35*s.RankScore + 0.25*s.RemainingScore + 0.20*s.FavoriteScore + 0.20*s.TravelScore
}

type Solution struct {
	SolutionID             string         `json:"solutionId"`
	Stops                  []Stop         `json:"stops"`
	Segments               []Segment      `json:"segments"`
	TotalTravelMinutes     int            `json:"totalTravelMinutes"`
	TotalDistanceMeters    int            `json:"totalDistanceMeters"`
	BufferRemainingMinutes int            `json:"bufferRemainingMinutes"`
	Score                  float64        `json:"score"`
	ScoreBreakdown         ScoreBreakdown `json:"scoreBreakdown"`
	RouteConfidence        string         `json:"routeConfidence"`
	WindowStart            time.Time      `json:"windowStart"`
	WindowEnd              time.Time      `json:"windowEnd"`
	CandidateRanks         []int          `json:"candidateRanks"`
}

type Outcome struct {
	Solutions       []Solution `json:"solutions"`
	Suggestions     []string   `json:"suggestions,omitempty"`
	WindowsExamined int        `json:"windowsExamined"`
}

// TravelInfo is a single travel leg. Estimated is true when the provider
// degraded to a straight-line estimate; callers must surface that.
type TravelInfo struct {
	DurationMinutes int
	DistanceMeters  int
	Estimated       bool
	Polyline        []Coordinate
}

type TravelProvider interface {
	Route(ctx context.Context, mode string, from, to Coordinate) (TravelInfo, error)
}

// Request is the solver input. Busy may be empty when the domain has no
// calendar module; sharing one still lets a calendar be plugged in later.
type Request struct {
	Busy               []Interval
	WindowStart        time.Time
	WindowEnd          time.Time
	Candidates         []Place
	TransportMode      string
	BufferAfterMinutes int
	HeadBufferMinutes  int
	Origin             *Origin
	ReturnToStart      bool
	Travel             TravelProvider
}

// Draft is the server-owned plan snapshot stored in ai_runs.plan_json and
// guide_sessions.plan_json. All times come from the solver, never the model.
type Draft struct {
	Version             int            `json:"version"`
	PlannerVersion      string         `json:"plannerVersion"`
	Title               string         `json:"title"`
	Summary             string         `json:"summary,omitempty"`
	WindowStart         time.Time      `json:"windowStart"`
	WindowEnd           time.Time      `json:"windowEnd"`
	TransportMode       string         `json:"transportMode"`
	RouteConfidence     string         `json:"routeConfidence"`
	Stops               []Stop         `json:"stops"`
	Segments            []Segment      `json:"segments"`
	TotalTravelMinutes  int            `json:"totalTravelMinutes"`
	TotalDistanceMeters int            `json:"totalDistanceMeters"`
	Score               float64        `json:"score"`
	ScoreBreakdown      ScoreBreakdown `json:"scoreBreakdown"`
	Suggestions         []string       `json:"suggestions,omitempty"`
	CreatedAt           time.Time      `json:"createdAt"`
	Status              string         `json:"status"` // draft | confirmed
}

// NormalizeMode maps loose user/model spellings to a solver mode.
func NormalizeMode(raw string) (string, bool) {
	switch raw {
	case ModeWalk, "walk", "walking", "WALKING", "步行":
		return ModeWalk, true
	case ModeRide, "ride", "bike", "bicycling", "BICYCLING", "骑行":
		return ModeRide, true
	case ModeDrive, "drive", "driving", "DRIVING", "驾车", "自驾":
		return ModeDrive, true
	case ModeTaxi, "taxi", "打车":
		return ModeTaxi, true
	case ModeTransit, "transit", "public", "公交", "地铁":
		return ModeTransit, true
	default:
		return "", false
	}
}

// MapMode converts a solver mode to the routes.mode column vocabulary.
func MapMode(mode string) string {
	switch mode {
	case ModeWalk:
		return "walking"
	case ModeRide:
		return "bicycling"
	case ModeTransit:
		return "transit"
	default:
		return "driving"
	}
}
