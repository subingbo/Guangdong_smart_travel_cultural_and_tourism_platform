package ai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestResolveAdcode(t *testing.T) {
	cases := []struct {
		city string
		want string
		ok   bool
	}{
		{"广州", "440100", true},
		{"广州市", "440100", true},
		{"广东广州", "440100", true},
		{"广东省广州市", "440100", true},
		{"Guangzhou", "440100", true},
		{"guangzhou city", "440100", true},
		{"深圳 ", "440300", true},
		{"汕头市", "440500", true},
		{"440100", "440100", true},
		{"上海", "", false},
		{"Hong Kong", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := ResolveAdcode(tc.city)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("ResolveAdcode(%q) = (%q, %v), want (%q, %v)", tc.city, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSupportedWeatherCities(t *testing.T) {
	cities := SupportedWeatherCities()
	if len(cities) != 21 {
		t.Fatalf("got %d cities, want 21", len(cities))
	}
	seen := make(map[string]bool, len(cities))
	for _, city := range cities {
		if seen[city] {
			t.Fatalf("duplicate city %q", city)
		}
		seen[city] = true
		if _, ok := guangdongCityAdcodes[city]; !ok {
			t.Fatalf("supported city %q has no adcode", city)
		}
	}
}

const weatherFixture = `{
  "status":"1",
  "count":"1",
  "info":"OK",
  "infocode":"10000",
  "forecasts":[{
    "city":"广州市",
    "province":"广东",
    "adcode":"440100",
    "reporttime":"2026-11-06 10:30:35",
    "casts":[
      {"date":"2026-11-06","week":"5","dayweather":"小雨","nightweather":"多云","daytemp":"31","nighttemp":"21","daywind":"北","nightwind":"北","daypower":"1-3","nightpower":"1-3"},
      {"date":"2026-11-07","week":"6","dayweather":"多云","nightweather":"晴","daytemp":"30","nighttemp":"22","daywind":"北","nightwind":"北","daypower":"1-3","nightpower":"1-3"},
      {"date":"2026-11-08","week":"7","dayweather":"多云","nightweather":"多云","daytemp":"31","nighttemp":"18","daywind":"北","nightwind":"北","daypower":"1-3","nightpower":"1-3"},
      {"date":"2026-11-09","week":"1","dayweather":"多云","nightweather":"多云","daytemp":"25","nighttemp":"15","daywind":"北","nightwind":"北","daypower":"1-3","nightpower":"1-3"},
      {"date":"2026-11-10","week":"2","dayweather":"晴","nightweather":"晴","daytemp":"26","nighttemp":"16","daywind":"北","nightwind":"北","daypower":"1-3","nightpower":"1-3"}
    ]
  }]
}`

func TestWeatherClientForecastAndCache(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if got := r.URL.Query().Get("city"); got != "440100" {
			t.Errorf("city = %q, want 440100", got)
		}
		if got := r.URL.Query().Get("extensions"); got != "all" {
			t.Errorf("extensions = %q, want all", got)
		}
		if got := r.URL.Query().Get("key"); got != "test-key" {
			t.Errorf("key = %q, want test-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, weatherFixture)
	}))
	defer server.Close()

	client := NewWeatherClient("test-key", server.URL)
	now := time.Date(2026, 11, 6, 10, 30, 0, 0, time.UTC)
	client.now = func() time.Time { return now }

	got, err := client.Forecast(context.Background(), "440100")
	if err != nil {
		t.Fatalf("Forecast: %v", err)
	}
	if got.City != "广州市" || got.Adcode != "440100" || got.ReportTime == "" {
		t.Fatalf("unexpected weather: %+v", got)
	}
	if len(got.Casts) != 4 {
		t.Fatalf("got %d casts, want 4", len(got.Casts))
	}
	if got.Casts[0].DayWeather != "小雨" || got.Casts[0].WeekLabel != "周五" {
		t.Fatalf("unexpected first cast: %+v", got.Casts[0])
	}

	if _, err := client.Forecast(context.Background(), "440100"); err != nil {
		t.Fatalf("cached Forecast: %v", err)
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Fatalf("server calls = %d, want 1 (cache hit)", n)
	}

	now = now.Add(11 * time.Minute)
	if _, err := client.Forecast(context.Background(), "440100"); err != nil {
		t.Fatalf("expired Forecast: %v", err)
	}
	if n := atomic.LoadInt32(&calls); n != 2 {
		t.Fatalf("server calls = %d, want 2 after cache expiry", n)
	}
}

func TestWeatherClientUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"0","info":"INVALID_USER_KEY","infocode":"10001"}`)
	}))
	defer server.Close()

	client := NewWeatherClient("bad-key", server.URL)
	_, err := client.Forecast(context.Background(), "440100")
	if err == nil || !strings.Contains(err.Error(), "10001") {
		t.Fatalf("error = %v, want infocode 10001", err)
	}
}

func TestWeatherClientHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewWeatherClient("test-key", server.URL)
	if _, err := client.Forecast(context.Background(), "440100"); err == nil {
		t.Fatal("want error for HTTP 500")
	}
}

func TestWeatherClientEmptyCasts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"1","info":"OK","infocode":"10000","forecasts":[{"city":"广州市","adcode":"440100","casts":[]}]}`)
	}))
	defer server.Close()

	client := NewWeatherClient("test-key", server.URL)
	if _, err := client.Forecast(context.Background(), "440100"); err == nil {
		t.Fatal("want error for empty casts")
	}
}

func TestWeatherClientDisabled(t *testing.T) {
	client := NewWeatherClient("", "")
	if client.Enabled() {
		t.Fatal("client with empty key should be disabled")
	}
	if _, err := client.Forecast(context.Background(), "440100"); err == nil {
		t.Fatal("want error for disabled client")
	}
}
