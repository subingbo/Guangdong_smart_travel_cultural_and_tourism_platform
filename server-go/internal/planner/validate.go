package planner

import (
	"fmt"
	"math"
	"time"
)

// ValidateDraft checks the shape of a server-built plan before it is shown
// or saved. It intentionally does not query the database; callers revalidate
// that places still exist at confirm time.
func ValidateDraft(d *Draft, now time.Time) error {
	if d == nil {
		return fmt.Errorf("plan is empty")
	}
	if d.Version <= 0 {
		return fmt.Errorf("plan version must be positive")
	}
	if d.PlannerVersion == "" {
		return fmt.Errorf("planner version is missing")
	}
	if !d.WindowEnd.After(d.WindowStart) {
		return fmt.Errorf("plan window is invalid")
	}
	if d.WindowEnd.Sub(d.WindowStart) > MaxWindowHours*time.Hour {
		return fmt.Errorf("plan window exceeds %d hours", MaxWindowHours)
	}
	if d.WindowStart.Before(now.Add(-MinStartLeadTime)) {
		return fmt.Errorf("plan window is in the past")
	}
	if _, ok := NormalizeMode(d.TransportMode); !ok {
		// A single-stop plan may have no segments and therefore no mode.
		if len(d.Segments) > 0 {
			return fmt.Errorf("unsupported transport mode %q", d.TransportMode)
		}
	}
	if d.RouteConfidence != ConfidenceAMAP && d.RouteConfidence != ConfidenceEstimated && d.RouteConfidence != ConfidenceNone {
		return fmt.Errorf("unsupported route confidence %q", d.RouteConfidence)
	}
	if math.IsNaN(d.Score) || math.IsInf(d.Score, 0) || d.Score < 0 || d.Score > 1.000001 {
		return fmt.Errorf("plan score is out of range")
	}

	places := 0
	prevLeave := d.WindowStart
	for i, s := range d.Stops {
		if s.PlaceID <= 0 {
			if s.Source != "USER_LOCATION" {
				return fmt.Errorf("stop %d has no place id", i)
			}
		} else {
			places++
		}
		if s.StayMinutes < 0 {
			return fmt.Errorf("stop %d has a negative stay", i)
		}
		if !validCoord(s.Lat, s.Lng) {
			return fmt.Errorf("stop %d has invalid coordinates", i)
		}
		if s.LeaveTime.Before(s.ArrivalTime) {
			return fmt.Errorf("stop %d leaves before it arrives", i)
		}
		if s.ArrivalTime.Before(prevLeave) {
			return fmt.Errorf("stop %d overlaps the previous stop", i)
		}
		if s.ArrivalTime.Before(d.WindowStart) || s.LeaveTime.After(d.WindowEnd) {
			return fmt.Errorf("stop %d is outside the plan window", i)
		}
		prevLeave = s.LeaveTime
	}
	if places == 0 {
		return fmt.Errorf("plan has no place stops")
	}
	if places > MaxDraftStops {
		return fmt.Errorf("plan has too many stops (max %d)", MaxDraftStops)
	}

	travel := 0
	distance := 0
	for i, seg := range d.Segments {
		if seg.FromIndex < 0 || seg.ToIndex < 0 ||
			seg.FromIndex >= len(d.Stops) || seg.ToIndex >= len(d.Stops) ||
			seg.FromIndex == seg.ToIndex {
			return fmt.Errorf("segment %d has invalid stop indexes", i)
		}
		if seg.DurationMin < 0 || seg.DistanceMeters < 0 {
			return fmt.Errorf("segment %d has a negative value", i)
		}
		travel += seg.DurationMin
		distance += seg.DistanceMeters
	}
	if d.TotalTravelMinutes != travel && len(d.Segments) > 0 {
		return fmt.Errorf("total travel minutes do not match segments")
	}
	if d.TotalDistanceMeters != distance && len(d.Segments) > 0 {
		return fmt.Errorf("total distance does not match segments")
	}
	return nil
}

// ValidateSolution checks a solver output before it is cached for submit_plan.
func ValidateSolution(s *Solution) error {
	if s == nil {
		return fmt.Errorf("solution is empty")
	}
	if len(s.Stops) == 0 {
		return fmt.Errorf("solution has no stops")
	}
	if math.IsNaN(s.Score) || math.IsInf(s.Score, 0) {
		return fmt.Errorf("solution score is invalid")
	}
	return nil
}

// validCoord keeps drafts inside mainland China. Guangdong-specific bboxes
// are too narrow once users plan cross-city trips.
func validCoord(lat, lng float64) bool {
	if math.IsNaN(lat) || math.IsNaN(lng) || math.IsInf(lat, 0) || math.IsInf(lng, 0) {
		return false
	}
	return lat >= 18 && lat <= 54 && lng >= 73 && lng <= 135
}
