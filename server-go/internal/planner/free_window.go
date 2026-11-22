package planner

import (
	"sort"
	"time"
)

// FindFree subtracts merged busy intervals from [windowStart, windowEnd) and
// returns the gaps that are at least needMinutes long.
func FindFree(busy []Interval, windowStart, windowEnd time.Time, needMinutes int) []Interval {
	if !windowEnd.After(windowStart) || needMinutes < 0 {
		return nil
	}

	clipped := make([]Interval, 0, len(busy))
	for _, b := range busy {
		start := b.Start
		if start.Before(windowStart) {
			start = windowStart
		}
		end := b.End
		if end.After(windowEnd) {
			end = windowEnd
		}
		if end.After(start) {
			clipped = append(clipped, Interval{Start: start, End: end})
		}
	}
	sort.Slice(clipped, func(i, j int) bool {
		return clipped[i].Start.Before(clipped[j].Start)
	})

	merged := make([]Interval, 0, len(clipped))
	for _, b := range clipped {
		if len(merged) == 0 || b.Start.After(merged[len(merged)-1].End) {
			merged = append(merged, b)
			continue
		}
		if b.End.After(merged[len(merged)-1].End) {
			merged[len(merged)-1].End = b.End
		}
	}

	gaps := make([]Interval, 0, len(merged)+1)
	cursor := windowStart
	for _, b := range merged {
		if b.Start.After(cursor) {
			gaps = append(gaps, Interval{Start: cursor, End: b.Start})
		}
		if b.End.After(cursor) {
			cursor = b.End
		}
	}
	if windowEnd.After(cursor) {
		gaps = append(gaps, Interval{Start: cursor, End: windowEnd})
	}

	need := time.Duration(needMinutes) * time.Minute
	out := make([]Interval, 0, len(gaps))
	for _, gap := range gaps {
		if gap.End.Sub(gap.Start) >= need {
			out = append(out, gap)
		}
	}
	return out
}
