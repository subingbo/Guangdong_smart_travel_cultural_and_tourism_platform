package planner

import (
	"context"
	"testing"
	"time"
)

var cst = time.FixedZone("CST", 8*3600)

type fakeTravel struct {
	minutes   int
	estimated bool
	err       error
}

func (f fakeTravel) Route(ctx context.Context, mode string, from, to Coordinate) (TravelInfo, error) {
	if f.err != nil {
		return TravelInfo{}, f.err
	}
	return TravelInfo{
		DurationMinutes: f.minutes,
		DistanceMeters:  1000,
		Estimated:       f.estimated,
		Polyline:        []Coordinate{from, to},
	}, nil
}

func TestFindFreeMergesAndSubtractsBusy(t *testing.T) {
	base := time.Date(2026, 11, 7, 14, 0, 0, 0, cst)
	busy := []Interval{
		{Start: base.Add(60 * time.Minute), End: base.Add(120 * time.Minute)},
		{Start: base.Add(60 * time.Minute), End: base.Add(90 * time.Minute)},
	}
	free := FindFree(busy, base, base.Add(6*time.Hour), 60)
	if len(free) != 2 {
		t.Fatalf("expected 2 free windows, got %d: %+v", len(free), free)
	}
	if !free[0].Start.Equal(base) || !free[0].End.Equal(base.Add(60*time.Minute)) {
		t.Fatalf("unexpected first window: %+v", free[0])
	}
	if !free[1].Start.Equal(base.Add(120*time.Minute)) || !free[1].End.Equal(base.Add(6*time.Hour)) {
		t.Fatalf("unexpected second window: %+v", free[1])
	}
}

func TestParseWindowRelativeChinese(t *testing.T) {
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, cst) // Thursday

	w, err := ParseWindow("明天下午", 120, now)
	if err != nil {
		t.Fatal(err)
	}
	if w.Start.Format("2006-01-02 15:04") != "2026-11-06 13:00" {
		t.Fatalf("unexpected start: %v", w.Start)
	}
	if w.End.Format("15:04") != "15:00" {
		t.Fatalf("unexpected end: %v", w.End)
	}

	w, err = ParseWindow("周六下午两点到五点", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if w.Start.Format("2006-01-02 15:04") != "2026-11-07 14:00" {
		t.Fatalf("unexpected Saturday start: %v", w.Start)
	}
	if w.End.Format("15:04") != "17:00" {
		t.Fatalf("unexpected Saturday end: %v", w.End)
	}
}

func TestParseWindowISOAndPast(t *testing.T) {
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, cst)
	w, err := ParseWindow("2026-11-07T14:00:00+08:00/2026-11-07T18:30:00+08:00", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if w.Start.Hour() != 14 || w.End.Hour() != 18 || w.End.Minute() != 30 {
		t.Fatalf("unexpected ISO window: %+v", w)
	}

	if _, err := ParseWindow("今天上午", 0, now); err == nil {
		t.Fatal("expected an error for a past window")
	}
	if _, err := ParseWindow("明天下午", 360, now); err == nil {
		t.Fatal("expected an error when the duration exceeds the period")
	}
}

func TestPlanBuildsFeasibleTimeline(t *testing.T) {
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, cst)
	start := time.Date(2026, 11, 7, 14, 0, 0, 0, cst)
	candidates := []Place{
		{PlaceID: 1, Name: "A", Lat: 23.10, Lng: 113.30, StayMinutes: 60, Source: "FAVORITE", Rank: 0},
		{PlaceID: 2, Name: "B", Lat: 23.11, Lng: 113.31, StayMinutes: 60, Source: "SYSTEM", Rank: 1},
		{PlaceID: 3, Name: "C", Lat: 23.12, Lng: 113.32, StayMinutes: 60, Source: "SYSTEM", Rank: 2},
	}
	out, err := Plan(context.Background(), Request{
		WindowStart:        start,
		WindowEnd:          start.Add(5 * time.Hour),
		Candidates:         candidates,
		TransportMode:      ModeWalk,
		BufferAfterMinutes: 10,
		HeadBufferMinutes:  0,
		Origin:             &Origin{Name: "酒店", Lat: 23.09, Lng: 113.29},
		ReturnToStart:      true,
		Travel:             fakeTravel{minutes: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Solutions) == 0 {
		t.Fatalf("expected solutions, got suggestions: %v", out.Suggestions)
	}
	sol := out.Solutions[0]
	if len(sol.Stops) != 5 {
		t.Fatalf("expected origin + 3 places + origin, got %d stops", len(sol.Stops))
	}
	if sol.Stops[0].Source != "USER_LOCATION" || sol.Stops[4].Source != "USER_LOCATION" {
		t.Fatalf("origin stops are not marked: %+v", sol.Stops)
	}
	if sol.TotalTravelMinutes != 40 {
		t.Fatalf("expected 40 travel minutes, got %d", sol.TotalTravelMinutes)
	}
	if sol.RouteConfidence != ConfidenceAMAP {
		t.Fatalf("expected AMAP confidence, got %s", sol.RouteConfidence)
	}
	if sol.BufferRemainingMinutes <= 0 {
		t.Fatalf("expected remaining buffer, got %d", sol.BufferRemainingMinutes)
	}
	if sol.Stops[1].ArrivalTime.Before(start) || sol.Stops[3].LeaveTime.After(start.Add(5*time.Hour)) {
		t.Fatalf("stops outside the window: %+v", sol.Stops)
	}
	if len(sol.CandidateRanks) != 3 {
		t.Fatalf("expected candidate ranks, got %v", sol.CandidateRanks)
	}

	draft := NewDraft(sol, "周六下午", "三个地点", 1, now)
	if err := ValidateDraft(&draft, now); err != nil {
		t.Fatalf("draft should be valid: %v", err)
	}
}

func TestPlanReportsInfeasibleWindow(t *testing.T) {
	start := time.Date(2026, 11, 7, 14, 0, 0, 0, cst)
	out, err := Plan(context.Background(), Request{
		WindowStart:        start,
		WindowEnd:          start.Add(2 * time.Hour),
		Candidates:         []Place{{PlaceID: 1, Name: "A", Lat: 23.1, Lng: 113.3, StayMinutes: 180, Rank: 0}},
		TransportMode:      ModeWalk,
		BufferAfterMinutes: 10,
		Travel:             fakeTravel{minutes: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Solutions) != 0 {
		t.Fatal("expected no solutions")
	}
	if len(out.Suggestions) == 0 {
		t.Fatal("expected suggestions for an infeasible window")
	}
}

func TestPlanHonorsCancellation(t *testing.T) {
	start := time.Date(2026, 11, 7, 14, 0, 0, 0, cst)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Plan(ctx, Request{
		WindowStart:   start,
		WindowEnd:     start.Add(4 * time.Hour),
		Candidates:    []Place{{PlaceID: 1, Name: "A", Lat: 23.1, Lng: 113.3, StayMinutes: 60, Rank: 0}},
		TransportMode: ModeWalk,
		Travel:        fakeTravel{minutes: 10},
	})
	if err == nil {
		t.Fatal("expected context cancellation")
	}
}

func TestValidateDraftRejectsTamperedTimes(t *testing.T) {
	now := time.Date(2026, 11, 5, 10, 0, 0, 0, cst)
	start := time.Date(2026, 11, 7, 14, 0, 0, 0, cst)
	out, err := Plan(context.Background(), Request{
		WindowStart:   start,
		WindowEnd:     start.Add(4 * time.Hour),
		Candidates:    []Place{{PlaceID: 1, Name: "A", Lat: 23.1, Lng: 113.3, StayMinutes: 60, Rank: 0}},
		TransportMode: ModeWalk,
		Travel:        fakeTravel{minutes: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	draft := NewDraft(out.Solutions[0], "bad", "", 1, now)
	draft.Stops[0].ArrivalTime = start.Add(-time.Hour)
	if err := ValidateDraft(&draft, now); err == nil {
		t.Fatal("expected a validation error")
	}
}
