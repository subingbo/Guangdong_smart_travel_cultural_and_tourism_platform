package ai

import (
	"context"

	"github.com/travelgd/server-go/internal/model"
)

// RouteProvider returns a real road geometry for a list of waypoints.
// Implementations fail soft: they return nil so the caller can fall back to
// straight lines between waypoints.
type RouteProvider interface {
	FetchPolyline(ctx context.Context, waypoints []model.Waypoint, mode string) *RouteResult
}

// RouteResult is the geometry plus the provider-reported road distance.
type RouteResult struct {
	Points         []model.Coordinate
	DistanceMeters int
}
