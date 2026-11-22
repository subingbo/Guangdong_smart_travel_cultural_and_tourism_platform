package ai

import (
	"context"
	"os"
	"testing"

	"github.com/travelgd/server-go/internal/model"
)

// Live check, skipped unless AMAP_WEB_KEY is exported. No credential is
// stored in this file.
func TestAmapLegRouteLive(t *testing.T) {
	key := os.Getenv("AMAP_WEB_KEY")
	if key == "" {
		t.Skip("AMAP_WEB_KEY not set")
	}
	client := NewAmapClient(key, "", "广州")
	leg, err := client.LegRoute(context.Background(), "driving",
		model.Coordinate{Lat: 23.1387, Lng: 113.2644},
		model.Coordinate{Lat: 23.1065, Lng: 113.3247})
	if err != nil {
		t.Fatal(err)
	}
	if leg.DurationMinutes <= 0 || leg.DistanceMeters <= 0 {
		t.Fatalf("unexpected leg: %+v", leg)
	}
	t.Logf("duration=%d min distance=%d m points=%d",
		leg.DurationMinutes, leg.DistanceMeters, len(leg.Points))
}
