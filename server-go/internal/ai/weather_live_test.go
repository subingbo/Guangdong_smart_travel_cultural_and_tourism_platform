package ai

import (
	"context"
	"os"
	"testing"
)

// Live check, skipped unless AMAP_WEB_KEY is exported. No credential is
// stored in this file.
func TestAmapWeatherLive(t *testing.T) {
	key := os.Getenv("AMAP_WEB_KEY")
	if key == "" {
		t.Skip("AMAP_WEB_KEY not set")
	}
	client := NewWeatherClient(key, "")
	weather, err := client.Forecast(context.Background(), "440100")
	if err != nil {
		t.Fatal(err)
	}
	if weather.City == "" || len(weather.Casts) == 0 {
		t.Fatalf("unexpected weather: %+v", weather)
	}
	t.Logf("city=%s reportTime=%s casts=%d first=%s %s/%s %s-%s",
		weather.City, weather.ReportTime, len(weather.Casts),
		weather.Casts[0].Date, weather.Casts[0].DayWeather, weather.Casts[0].NightWeather,
		weather.Casts[0].DayTemp, weather.Casts[0].NightTemp)
}
