package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/travelgd/server-go/internal/agent"
	"github.com/travelgd/server-go/internal/ai"
)

type fakeWeatherProvider struct {
	enabled bool
	weather *ai.Weather
	err     error
}

func (f fakeWeatherProvider) Enabled() bool { return f.enabled }

func (f fakeWeatherProvider) Forecast(ctx context.Context, adcode string) (*ai.Weather, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.weather, nil
}

func executeWeatherTool(t *testing.T, tools *AgentTools, args string) map[string]any {
	t.Helper()
	result, err := tools.getWeather().Execute(
		context.Background(), &agent.ExecutionContext{Language: "zh"},
		json.RawMessage(args))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(result.Content), &payload); err != nil {
		t.Fatalf("unmarshal result %q: %v", result.Content, err)
	}
	return payload
}

func TestAgentToolsRegistersWeather(t *testing.T) {
	tools := (&AgentTools{}).All()
	if len(tools) != 9 {
		t.Fatalf("got %d tools, want 9", len(tools))
	}
	for _, tool := range tools {
		if tool.Name() == "get_weather" {
			return
		}
	}
	t.Fatal("get_weather is not registered")
}

func TestGetWeatherToolSuccess(t *testing.T) {
	tools := &AgentTools{weather: fakeWeatherProvider{
		enabled: true,
		weather: &ai.Weather{
			City:       "广州市",
			Province:   "广东",
			Adcode:     "440100",
			ReportTime: "2026-11-06 10:30:35",
			Casts: []ai.WeatherCast{{
				Date: "2026-11-06", WeekLabel: "周五",
				DayWeather: "小雨", NightWeather: "多云",
				DayTemp: "31", NightTemp: "21",
			}},
		},
	}}
	payload := executeWeatherTool(t, tools, `{"city":"广州"}`)
	if payload["available"] != true {
		t.Fatalf("available = %v, want true", payload["available"])
	}
	if payload["city"] != "广州市" || payload["adcode"] != "440100" {
		t.Fatalf("unexpected payload: %v", payload)
	}
	casts, ok := payload["casts"].([]any)
	if !ok || len(casts) != 1 {
		t.Fatalf("casts = %v, want 1 cast", payload["casts"])
	}
}

func TestGetWeatherToolUnsupportedCity(t *testing.T) {
	tools := &AgentTools{}
	payload := executeWeatherTool(t, tools, `{"city":"上海"}`)
	if payload["available"] != false || payload["reason"] != "unsupported city" {
		t.Fatalf("unexpected payload: %v", payload)
	}
	if cities, ok := payload["supportedCities"].([]any); !ok || len(cities) != 21 {
		t.Fatalf("supportedCities = %v, want 21", payload["supportedCities"])
	}
}

func TestGetWeatherToolNotConfigured(t *testing.T) {
	tools := &AgentTools{}
	payload := executeWeatherTool(t, tools, `{"city":"广州"}`)
	if payload["available"] != false || payload["reason"] != "weather service not configured" {
		t.Fatalf("unexpected payload: %v", payload)
	}
}

func TestGetWeatherToolUpstreamError(t *testing.T) {
	tools := &AgentTools{weather: fakeWeatherProvider{enabled: true, err: errors.New("boom")}}
	payload := executeWeatherTool(t, tools, `{"city":"广州"}`)
	if payload["available"] != false || payload["reason"] != "weather service unavailable" {
		t.Fatalf("unexpected payload: %v", payload)
	}
	if _, leaked := payload["error"]; leaked {
		t.Fatalf("upstream error leaked to model: %v", payload)
	}
}

func TestGetWeatherToolEmptyCity(t *testing.T) {
	tools := &AgentTools{}
	payload := executeWeatherTool(t, tools, `{"city":"  "}`)
	if payload["available"] != false || payload["reason"] != "city is required" {
		t.Fatalf("unexpected payload: %v", payload)
	}
}
