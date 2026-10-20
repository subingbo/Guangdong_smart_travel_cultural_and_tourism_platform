package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/model"
)

/**
 * Optional Tencent Maps route planner. When the key is absent the caller
 * falls back to straight lines between waypoints, matching the cloud function.
 */
type TencentRouteClient struct {
	apiKey string
	apiURL string
	client *http.Client
}

func NewTencentRouteClient(apiKey, apiURL string) *TencentRouteClient {
	return &TencentRouteClient{
		apiKey: apiKey,
		apiURL: apiURL,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *TencentRouteClient) Enabled() bool { return t.apiKey != "" }

type RouteResult struct {
	Points []model.Coordinate
}

func (t *TencentRouteClient) FetchPolyline(ctx context.Context, waypoints []model.Waypoint, mode string) *RouteResult {
	if !t.Enabled() || len(waypoints) < 2 {
		return nil
	}
	from := fmt.Sprintf("%f,%f", waypoints[0].Lat, waypoints[0].Lng)
	to := fmt.Sprintf("%f,%f", waypoints[len(waypoints)-1].Lat, waypoints[len(waypoints)-1].Lng)
	var viaParts []string
	for _, wp := range waypoints[1 : len(waypoints)-1] {
		viaParts = append(viaParts, fmt.Sprintf("%f,%f", wp.Lat, wp.Lng))
	}
	via := strings.Join(viaParts, ";")

	endpoint := t.apiURL + "/driving/"
	switch mode {
	case "walking":
		endpoint = t.apiURL + "/walking/"
	case "bicycling":
		endpoint = t.apiURL + "/bicycling/"
	case "transit":
		endpoint = t.apiURL + "/transit/"
	}
	url := fmt.Sprintf("%s?from=%s&to=%s&key=%s", endpoint, from, to, t.apiKey)
	if via != "" {
		url += "&waypoints=" + via
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var out struct {
		Result struct {
			Routes []struct {
				Polyline []float64 `json:"polyline"`
			} `json:"routes"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &out); err != nil || len(out.Result.Routes) == 0 {
		return nil
	}
	polyline := out.Result.Routes[0].Polyline
	if len(polyline) < 4 {
		return nil
	}
	return &RouteResult{Points: decodePolyline(polyline)}
}

/** Decodes the compressed polyline format (first point absolute, rest deltas * 1e6). */
func decodePolyline(pl []float64) []model.Coordinate {
	if len(pl) < 2 {
		return nil
	}
	if math.Abs(pl[0]) > 180 {
		pl[0] /= 1000000
		pl[1] /= 1000000
	}
	for i := 2; i < len(pl); i++ {
		pl[i] = math.Round((pl[i-2]+pl[i]/1000000)*1000000) / 1000000
	}
	points := make([]model.Coordinate, 0, len(pl)/2)
	for i := 0; i+1 < len(pl); i += 2 {
		points = append(points, model.Coordinate{Lat: pl[i], Lng: pl[i+1]})
	}
	return points
}
