package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	defaultWeatherURL  = "https://restapi.amap.com/v3/weather/weatherInfo"
	weatherCacheTTL    = 10 * time.Minute
	weatherCacheMax    = 256
	weatherMaxForecast = 4
)

// Weather is the normalized forecast returned to the agent. AMap reports
// temperatures, wind and humidity as strings, so the model gets the same
// values without a lossy numeric conversion.
type Weather struct {
	City       string        `json:"city"`
	Province   string        `json:"province,omitempty"`
	Adcode     string        `json:"adcode"`
	ReportTime string        `json:"reportTime"`
	Casts      []WeatherCast `json:"casts"`
}

// WeatherCast is one day of the AMap forecast.
type WeatherCast struct {
	Date         string `json:"date"`
	Week         string `json:"week"`
	WeekLabel    string `json:"weekLabel,omitempty"`
	DayWeather   string `json:"dayWeather"`
	NightWeather string `json:"nightWeather"`
	DayTemp      string `json:"dayTemp"`
	NightTemp    string `json:"nightTemp"`
	DayWind      string `json:"dayWind,omitempty"`
	NightWind    string `json:"nightWind,omitempty"`
	DayPower     string `json:"dayPower,omitempty"`
	NightPower   string `json:"nightPower,omitempty"`
}

// WeatherClient talks to the AMap Web service weather endpoint. It keeps a
// small in-process cache because the planner may ask for the same city many
// times in one conversation, and deduplicates concurrent identical requests.
type WeatherClient struct {
	apiKey string
	apiURL string
	client *http.Client

	mu    sync.Mutex
	cache map[string]weatherCacheEntry
	group singleflight.Group
	now   func() time.Time
}

type weatherCacheEntry struct {
	weather *Weather
	expires time.Time
}

func NewWeatherClient(apiKey, apiURL string) *WeatherClient {
	if apiURL == "" {
		apiURL = defaultWeatherURL
	}
	return &WeatherClient{
		apiKey: apiKey,
		apiURL: strings.TrimRight(apiURL, "/"),
		client: &http.Client{Timeout: 8 * time.Second},
		cache:  make(map[string]weatherCacheEntry),
		now:    time.Now,
	}
}

func (w *WeatherClient) Enabled() bool { return w != nil && w.apiKey != "" }

// Forecast returns today's and the next three days' forecast for an adcode.
// Errors are not cached: a transient AMap failure should be retried on the
// next run instead of pinning an empty answer for ten minutes.
func (w *WeatherClient) Forecast(ctx context.Context, adcode string) (*Weather, error) {
	if !w.Enabled() {
		return nil, fmt.Errorf("weather key is not configured")
	}
	adcode = strings.TrimSpace(adcode)
	if adcode == "" {
		return nil, fmt.Errorf("adcode is required")
	}
	if cached := w.cached(adcode); cached != nil {
		return cached, nil
	}
	value, err, _ := w.group.Do(adcode, func() (any, error) {
		if cached := w.cached(adcode); cached != nil {
			return cached, nil
		}
		weather, err := w.fetch(ctx, adcode)
		if err != nil {
			return nil, err
		}
		w.store(adcode, weather)
		return weather, nil
	})
	if err != nil {
		return nil, err
	}
	weather, ok := value.(*Weather)
	if !ok || weather == nil {
		return nil, fmt.Errorf("amap weather: empty result")
	}
	return weather, nil
}

func (w *WeatherClient) cached(adcode string) *Weather {
	w.mu.Lock()
	defer w.mu.Unlock()
	entry, ok := w.cache[adcode]
	if !ok {
		return nil
	}
	if w.now().After(entry.expires) {
		delete(w.cache, adcode)
		return nil
	}
	return entry.weather
}

func (w *WeatherClient) store(adcode string, weather *Weather) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.cache) >= weatherCacheMax {
		for key := range w.cache {
			delete(w.cache, key)
		}
	}
	w.cache[adcode] = weatherCacheEntry{weather: weather, expires: w.now().Add(weatherCacheTTL)}
}

func (w *WeatherClient) fetch(ctx context.Context, adcode string) (*Weather, error) {
	params := url.Values{}
	params.Set("key", w.apiKey)
	params.Set("city", adcode)
	params.Set("extensions", "all")

	var resp amapWeatherResponse
	if err := getJSON(ctx, w.client, w.apiURL, params, &resp); err != nil {
		return nil, fmt.Errorf("amap weather request: %w", err)
	}
	if resp.Status != "1" {
		return nil, fmt.Errorf("amap weather: status=%s infocode=%s info=%s",
			resp.Status, resp.Infocode, resp.Info)
	}
	if len(resp.Forecasts) == 0 || len(resp.Forecasts[0].Casts) == 0 {
		return nil, fmt.Errorf("amap weather: empty forecast")
	}

	forecast := resp.Forecasts[0]
	casts := make([]WeatherCast, 0, len(forecast.Casts))
	for i, cast := range forecast.Casts {
		if i >= weatherMaxForecast {
			break
		}
		casts = append(casts, WeatherCast{
			Date:         cast.Date,
			Week:         cast.Week,
			WeekLabel:    weatherWeekLabel(cast.Week),
			DayWeather:   cast.DayWeather,
			NightWeather: cast.NightWeather,
			DayTemp:      cast.DayTemp,
			NightTemp:    cast.NightTemp,
			DayWind:      cast.DayWind,
			NightWind:    cast.NightWind,
			DayPower:     cast.DayPower,
			NightPower:   cast.NightPower,
		})
	}
	return &Weather{
		City:       forecast.City,
		Province:   forecast.Province,
		Adcode:     forecast.Adcode,
		ReportTime: forecast.ReportTime,
		Casts:      casts,
	}, nil
}

func weatherWeekLabel(raw string) string {
	labels := [...]string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 1 || n > len(labels) {
		return ""
	}
	return labels[n-1]
}

// guangdongCityAdcodes covers all 21 prefecture-level cities in Guangdong.
// Keys are lower-case Chinese (without the 市 suffix) and pinyin.
var guangdongCityAdcodes = map[string]string{
	"广州": "440100", "guangzhou": "440100",
	"韶关": "440200", "shaoguan": "440200",
	"深圳": "440300", "shenzhen": "440300",
	"珠海": "440400", "zhuhai": "440400",
	"汕头": "440500", "shantou": "440500",
	"佛山": "440600", "foshan": "440600",
	"江门": "440700", "jiangmen": "440700",
	"湛江": "440800", "zhanjiang": "440800",
	"茂名": "440900", "maoming": "440900",
	"肇庆": "441200", "zhaoqing": "441200",
	"惠州": "441300", "huizhou": "441300",
	"梅州": "441400", "meizhou": "441400",
	"汕尾": "441500", "shanwei": "441500",
	"河源": "441600", "heyuan": "441600",
	"阳江": "441700", "yangjiang": "441700",
	"清远": "441800", "qingyuan": "441800",
	"东莞": "441900", "dongguan": "441900",
	"中山": "442000", "zhongshan": "442000",
	"潮州": "445100", "chaozhou": "445100",
	"揭阳": "445200", "jieyang": "445200",
	"云浮": "445300", "yunfu": "445300",
}

// ResolveAdcode maps a city name or a raw six-digit adcode to an AMap adcode.
// It accepts 广州 / 广州市 / 广东广州 / Guangzhou / Guangzhou City and returns
// false for anything outside the 21 Guangdong prefecture-level cities.
func ResolveAdcode(city string) (string, bool) {
	raw := strings.TrimSpace(city)
	if raw == "" {
		return "", false
	}
	if isSixDigitAdcode(raw) {
		return raw, true
	}
	key := strings.ToLower(raw)
	key = strings.ReplaceAll(key, " ", "")
	key = strings.TrimPrefix(key, "广东省")
	key = strings.TrimPrefix(key, "广东")
	key = strings.TrimSuffix(key, "city")
	key = strings.TrimSuffix(key, "市")
	code, ok := guangdongCityAdcodes[key]
	return code, ok
}

// SupportedWeatherCities is returned to the model when a city is unsupported,
// so it can ask the user instead of guessing an adcode.
func SupportedWeatherCities() []string {
	return []string{
		"广州", "深圳", "珠海", "汕头", "佛山", "韶关", "湛江", "肇庆",
		"江门", "茂名", "惠州", "梅州", "汕尾", "河源", "阳江", "清远",
		"东莞", "中山", "潮州", "揭阳", "云浮",
	}
}

func isSixDigitAdcode(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// --- AMap response shape ---

type amapWeatherResponse struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Infocode  string `json:"infocode"`
	Forecasts []struct {
		City       string `json:"city"`
		Province   string `json:"province"`
		Adcode     string `json:"adcode"`
		ReportTime string `json:"reporttime"`
		Casts      []struct {
			Date         string `json:"date"`
			Week         string `json:"week"`
			DayWeather   string `json:"dayweather"`
			NightWeather string `json:"nightweather"`
			DayTemp      string `json:"daytemp"`
			NightTemp    string `json:"nighttemp"`
			DayWind      string `json:"daywind"`
			NightWind    string `json:"nightwind"`
			DayPower     string `json:"daypower"`
			NightPower   string `json:"nightpower"`
		} `json:"casts"`
	} `json:"forecasts"`
}
