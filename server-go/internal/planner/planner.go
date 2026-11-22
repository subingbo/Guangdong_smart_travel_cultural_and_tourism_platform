package planner

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

// Plan solves one itinerary request. It is deterministic for a fixed
// provider and returns at most three solutions ordered by score.
func Plan(ctx context.Context, req Request) (Outcome, error) {
	if req.Travel == nil {
		return Outcome{}, errors.New("planner: travel provider is required")
	}
	if !req.WindowEnd.After(req.WindowStart) {
		return noSolution("时间窗口无效"), nil
	}
	if req.WindowEnd.Sub(req.WindowStart) > MaxWindowHours*time.Hour {
		return noSolution("时间窗口不能超过 24 小时"), nil
	}
	mode, ok := NormalizeMode(req.TransportMode)
	if !ok {
		return noSolution("交通方式仅支持 WALK / RIDE / DRIVE / TAXI / TRANSIT"), nil
	}
	if len(req.Candidates) == 0 {
		return noSolution("至少需要一个候选地点"), nil
	}

	candidates := append([]Place(nil), req.Candidates...)
	if len(candidates) > MaxCandidates {
		candidates = candidates[:MaxCandidates]
	}
	for i := range candidates {
		if candidates[i].StayMinutes <= 0 {
			candidates[i].StayMinutes = DefaultStay
		}
	}
	buffer := req.BufferAfterMinutes
	if buffer < 0 {
		buffer = 0
	}
	headBuffer := req.HeadBufferMinutes
	if headBuffer < 0 {
		headBuffer = 0
	}

	stayFloor := 0
	for _, p := range candidates {
		stayFloor += p.StayMinutes
	}
	need := stayFloor + buffer + headBuffer
	windows := FindFree(req.Busy, req.WindowStart, req.WindowEnd, need)
	if len(windows) == 0 {
		return noSolution(fmt.Sprintf(
			"所选时间范围内没有 ≥%d 分钟的连续空闲（最少需要：停留 %d + 收尾缓冲 %d + 首段衔接 %d）",
			need, stayFloor, buffer, headBuffer)), nil
	}

	perms := permutations(candidates)
	if len(perms) == 0 {
		return noSolution("候选地点组合超出求解能力（可减少地点数量）"), nil
	}

	solutions := make([]Solution, 0, MaxSolutions*4)
	windowsExamined := 0
	for _, window := range windows {
		windowsExamined++
		for _, order := range perms {
			if err := ctx.Err(); err != nil {
				return Outcome{}, err
			}
			sol, ok, err := solve(ctx, window, order, mode, buffer, headBuffer,
				req.Busy, req.Travel, req.Origin, req.ReturnToStart)
			if err != nil {
				return Outcome{}, err
			}
			if ok {
				solutions = append(solutions, sol)
			}
			if len(solutions) >= MaxPermutations*2 {
				break
			}
		}
		if len(solutions) >= MaxPermutations*2 {
			break
		}
	}
	if len(solutions) == 0 {
		return noSolution(fmt.Sprintf(
			"窗口时长可容纳停留（%d 分钟），但加入实际交通耗时或首尾衔接缓冲后超出窗口；可尝试更近的地点、更快的交通方式、缩短停留或减少地点",
			stayFloor)), nil
	}

	sort.SliceStable(solutions, func(i, j int) bool {
		return solutions[i].Score > solutions[j].Score
	})
	if len(solutions) > MaxSolutions {
		solutions = solutions[:MaxSolutions]
	}
	for i := range solutions {
		solutions[i].SolutionID = fmt.Sprintf("sol-%d", i+1)
	}
	return Outcome{Solutions: solutions, WindowsExamined: windowsExamined}, nil
}

func noSolution(reason string) Outcome {
	return Outcome{Suggestions: []string{reason}, WindowsExamined: 0}
}

func solve(ctx context.Context,
	window Interval,
	order []Place,
	mode string,
	bufferAfter, headBuffer int,
	busy []Interval,
	travel TravelProvider,
	origin *Origin,
	returnToStart bool,
) (Solution, bool, error) {
	cursor := window.Start
	for _, b := range busy {
		if b.End.After(window.Start) {
			continue
		}
		earliest := b.End.Add(time.Duration(headBuffer) * time.Minute)
		if earliest.After(cursor) {
			cursor = earliest
		}
	}

	type node struct {
		place  Place
		origin bool
	}
	seq := make([]node, 0, len(order)+2)
	if origin != nil {
		seq = append(seq, node{
			place:  Place{Name: origin.Name, Lat: origin.Lat, Lng: origin.Lng, Source: "USER_LOCATION"},
			origin: true,
		})
	}
	for _, p := range order {
		seq = append(seq, node{place: p})
	}
	if origin != nil && returnToStart {
		seq = append(seq, node{
			place:  Place{Name: origin.Name, Lat: origin.Lat, Lng: origin.Lng, Source: "USER_LOCATION"},
			origin: true,
		})
	}

	stops := make([]Stop, 0, len(seq))
	segments := make([]Segment, 0, len(seq)-1)
	totalTravel := 0
	totalDistance := 0
	anyEstimated := false
	stayTotal := 0

	for i, n := range seq {
		if i > 0 {
			prev := seq[i-1].place
			info, err := travel.Route(ctx, mode, Coordinate{Lat: prev.Lat, Lng: prev.Lng}, Coordinate{Lat: n.place.Lat, Lng: n.place.Lng})
			if err != nil {
				return Solution{}, false, err
			}
			if info.DurationMinutes < 0 {
				info.DurationMinutes = 0
			}
			totalTravel += info.DurationMinutes
			totalDistance += info.DistanceMeters
			anyEstimated = anyEstimated || info.Estimated
			segments = append(segments, Segment{
				FromIndex:      i - 1,
				ToIndex:        i,
				TransportMode:  mode,
				DistanceMeters: info.DistanceMeters,
				DurationMin:    info.DurationMinutes,
				Estimated:      info.Estimated,
				Polyline:       info.Polyline,
			})
			cursor = cursor.Add(time.Duration(info.DurationMinutes) * time.Minute)
		}

		if n.origin {
			stops = append(stops, Stop{
				Name:          n.place.Name,
				Lat:           n.place.Lat,
				Lng:           n.place.Lng,
				Source:        "USER_LOCATION",
				ArrivalTime:   cursor,
				LeaveTime:     cursor,
				StayMinutes:   0,
				TransportMode: mode,
			})
			continue
		}

		stay := n.place.StayMinutes
		if stay <= 0 {
			stay = DefaultStay
		}
		stayTotal += stay
		leave := cursor.Add(time.Duration(stay) * time.Minute)
		stops = append(stops, Stop{
			PlaceID:       n.place.PlaceID,
			Name:          n.place.Name,
			NameEn:        n.place.NameEn,
			Lat:           n.place.Lat,
			Lng:           n.place.Lng,
			Source:        n.place.Source,
			ArrivalTime:   cursor,
			LeaveTime:     leave,
			StayMinutes:   stay,
			TransportMode: mode,
		})
		cursor = leave
	}

	tailEnd := cursor.Add(time.Duration(bufferAfter) * time.Minute)
	if tailEnd.After(window.End) {
		return Solution{}, false, nil
	}
	remaining := window.End.Sub(tailEnd)

	ranks := make([]int, 0, len(order))
	rankSum := 0
	favorites := 0
	for _, p := range order {
		ranks = append(ranks, p.Rank)
		rankSum += p.Rank
		if p.Source == "FAVORITE" {
			favorites++
		}
	}
	denom := len(order) - 1
	if denom < 1 {
		denom = 1
	}
	breakdown := ScoreBreakdown{
		RankScore:      clamp01(1 - float64(rankSum)/float64(len(order))/float64(denom)),
		RemainingScore: clamp01(float64(remaining) / float64(window.End.Sub(window.Start))),
		FavoriteScore:  clamp01(float64(favorites) / float64(len(order))),
		TravelScore:    clamp01(1 - float64(totalTravel)/float64(totalTravel+stayTotal+1)),
	}
	confidence := ConfidenceNone
	if len(segments) > 0 {
		confidence = ConfidenceAMAP
		if anyEstimated {
			confidence = ConfidenceEstimated
		}
	}

	return Solution{
		Stops:                  stops,
		Segments:               segments,
		TotalTravelMinutes:     totalTravel,
		TotalDistanceMeters:    totalDistance,
		BufferRemainingMinutes: int(remaining / time.Minute),
		Score:                  breakdown.Total(),
		ScoreBreakdown:         breakdown,
		RouteConfidence:        confidence,
		WindowStart:            window.Start,
		WindowEnd:              window.End,
		CandidateRanks:         ranks,
	}, true, nil
}

func permutations(items []Place) [][]Place {
	out := make([][]Place, 0, MaxPermutations)
	work := append([]Place(nil), items...)
	var walk func(k int)
	walk = func(k int) {
		if len(out) >= MaxPermutations {
			return
		}
		if k == len(work) {
			out = append(out, append([]Place(nil), work...))
			return
		}
		for i := k; i < len(work); i++ {
			work[k], work[i] = work[i], work[k]
			walk(k + 1)
			work[k], work[i] = work[i], work[k]
			if len(out) >= MaxPermutations {
				return
			}
		}
	}
	walk(0)
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// NewDraft turns a solution into the server-owned plan snapshot.
func NewDraft(sol Solution, title, summary string, version int, now time.Time) Draft {
	return Draft{
		Version:             version,
		PlannerVersion:      Version,
		Title:               title,
		Summary:             summary,
		WindowStart:         sol.WindowStart,
		WindowEnd:           sol.WindowEnd,
		TransportMode:       transportOf(sol),
		RouteConfidence:     sol.RouteConfidence,
		Stops:               sol.Stops,
		Segments:            sol.Segments,
		TotalTravelMinutes:  sol.TotalTravelMinutes,
		TotalDistanceMeters: sol.TotalDistanceMeters,
		Score:               sol.Score,
		ScoreBreakdown:      sol.ScoreBreakdown,
		CreatedAt:           now,
		Status:              "draft",
	}
}

func transportOf(sol Solution) string {
	for _, s := range sol.Segments {
		if s.TransportMode != "" {
			return s.TransportMode
		}
	}
	return ""
}
