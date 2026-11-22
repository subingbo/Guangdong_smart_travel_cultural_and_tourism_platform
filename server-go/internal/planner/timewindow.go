package planner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Window is the resolved planning window in the caller's location.
type Window struct {
	Start time.Time
	End   time.Time
}

var (
	weekdayRe = regexp.MustCompile(`(下)?(?:周|星期|礼拜)([一二三四五六日天])`)
	dateRe    = regexp.MustCompile(`(\d{4})-(\d{1,2})-(\d{1,2})`)
	cnDateRe  = regexp.MustCompile(`(\d{1,2})月(\d{1,2})日`)
	clockRe   = regexp.MustCompile(`(\d{1,2})\s*[:：点]\s*(\d{1,2})?`)
)

// ParseWindow resolves a natural-language window such as "周六下午" or
// "明天下午两点到五点" into absolute times. Relative dates are resolved
// against now, never against the model's own clock.
func ParseWindow(when string, durationMinutes int, now time.Time) (Window, error) {
	raw := strings.TrimSpace(when)
	if raw == "" {
		return Window{}, fmt.Errorf("请提供规划时间，例如“明天下午”或“2026-11-07 14:00-18:00”")
	}
	loc := now.Location()

	// Parse ISO before lower-casing: RFC3339 needs its uppercase T, and the
	// timezone offset (+08:00) must not be mistaken for a clock time.
	if w, ok := parseISORange(raw, loc); ok {
		return validateWindow(w, now, durationMinutes)
	}

	text := normalizeClockWords(strings.ToLower(raw))

	day, weekdayExpr, err := resolveDay(text, now)
	if err != nil {
		return Window{}, err
	}

	periodStart, periodEnd := periodOf(text)
	times := clockRe.FindAllStringSubmatch(text, -1)
	var start, end time.Time
	switch len(times) {
	case 0:
		start = time.Date(day.Year(), day.Month(), day.Day(), periodStart, 0, 0, 0, loc)
		end = time.Date(day.Year(), day.Month(), day.Day(), periodEnd, 0, 0, 0, loc)
		if durationMinutes > 0 {
			want := start.Add(time.Duration(durationMinutes) * time.Minute)
			if want.Before(end) {
				end = want
			}
		}
	case 1:
		h, _ := strconv.Atoi(times[0][1])
		h = applyPeriod(h, periodStart)
		m := 0
		if strings.Contains(times[0][0], "半") {
			m = 30
		} else if times[0][2] != "" {
			m, _ = strconv.Atoi(times[0][2])
		}
		start = time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, loc)
		if durationMinutes > 0 {
			end = start.Add(time.Duration(durationMinutes) * time.Minute)
		} else {
			end = time.Date(day.Year(), day.Month(), day.Day(), periodEnd, 0, 0, 0, loc)
			if !end.After(start) {
				end = start.Add(4 * time.Hour)
			}
		}
	default:
		h1, _ := strconv.Atoi(times[0][1])
		h1 = applyPeriod(h1, periodStart)
		m1 := 0
		if times[0][2] != "" {
			m1, _ = strconv.Atoi(times[0][2])
		}
		h2, _ := strconv.Atoi(times[1][1])
		h2 = applyPeriod(h2, periodStart)
		m2 := 0
		if times[1][2] != "" {
			m2, _ = strconv.Atoi(times[1][2])
		}
		start = time.Date(day.Year(), day.Month(), day.Day(), h1, m1, 0, 0, loc)
		end = time.Date(day.Year(), day.Month(), day.Day(), h2, m2, 0, 0, loc)
	}

	// A bare weekday that has already started today means the next week.
	if weekdayExpr && !start.After(now) {
		start = start.AddDate(0, 0, 7)
		end = end.AddDate(0, 0, 7)
	}
	return validateWindow(Window{Start: start, End: end}, now, durationMinutes)
}

func parseISORange(text string, loc *time.Location) (Window, bool) {
	parts := strings.Split(text, "/")
	if len(parts) != 2 {
		parts = strings.Split(text, " ~ ")
	}
	if len(parts) == 2 {
		start, err1 := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
		end, err2 := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
		if err1 == nil && err2 == nil {
			return Window{Start: start.In(loc), End: end.In(loc)}, true
		}
	}
	if start, err := time.Parse(time.RFC3339, text); err == nil {
		return Window{Start: start.In(loc), End: start.In(loc).Add(4 * time.Hour)}, true
	}
	return Window{}, false
}

func resolveDay(text string, now time.Time) (time.Time, bool, error) {
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	switch {
	case strings.Contains(text, "大后天"):
		return today.AddDate(0, 0, 3), false, nil
	case strings.Contains(text, "后天") || strings.Contains(text, "day after tomorrow"):
		return today.AddDate(0, 0, 2), false, nil
	case strings.Contains(text, "明天") || strings.Contains(text, "tomorrow"):
		return today.AddDate(0, 0, 1), false, nil
	case strings.Contains(text, "今天") || strings.Contains(text, "today"):
		return today, false, nil
	}

	if m := dateRe.FindStringSubmatch(text); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		return time.Date(y, time.Month(mo), d, 0, 0, 0, 0, loc), false, nil
	}
	if m := cnDateRe.FindStringSubmatch(text); len(m) == 3 {
		mo, _ := strconv.Atoi(m[1])
		d, _ := strconv.Atoi(m[2])
		day := time.Date(now.Year(), time.Month(mo), d, 0, 0, 0, 0, loc)
		if day.Before(today) {
			day = day.AddDate(1, 0, 0)
		}
		return day, false, nil
	}

	if m := weekdayRe.FindStringSubmatch(text); len(m) == 3 {
		target := weekdayFromCN(m[2])
		nextWeek := m[1] == "下"
		delta := int(target - now.Weekday())
		if delta < 0 {
			delta += 7
		}
		if nextWeek {
			delta += 7
		}
		return today.AddDate(0, 0, delta), true, nil
	}

	switch {
	case strings.Contains(text, "monday"):
		return nextEnglishWeekday(today, now, time.Monday, text), true, nil
	case strings.Contains(text, "tuesday"):
		return nextEnglishWeekday(today, now, time.Tuesday, text), true, nil
	case strings.Contains(text, "wednesday"):
		return nextEnglishWeekday(today, now, time.Wednesday, text), true, nil
	case strings.Contains(text, "thursday"):
		return nextEnglishWeekday(today, now, time.Thursday, text), true, nil
	case strings.Contains(text, "friday"):
		return nextEnglishWeekday(today, now, time.Friday, text), true, nil
	case strings.Contains(text, "saturday"):
		return nextEnglishWeekday(today, now, time.Saturday, text), true, nil
	case strings.Contains(text, "sunday"):
		return nextEnglishWeekday(today, now, time.Sunday, text), true, nil
	}

	return today, false, fmt.Errorf("无法识别日期“%s”，请改为“明天下午”或“2026-11-07 14:00-18:00”", text)
}

func weekdayFromCN(s string) time.Weekday {
	switch s {
	case "一":
		return time.Monday
	case "二":
		return time.Tuesday
	case "三":
		return time.Wednesday
	case "四":
		return time.Thursday
	case "五":
		return time.Friday
	case "六":
		return time.Saturday
	default:
		return time.Sunday
	}
}

func nextEnglishWeekday(today, now time.Time, target time.Weekday, text string) time.Time {
	delta := int(target - now.Weekday())
	if delta < 0 {
		delta += 7
	}
	if strings.Contains(text, "next") {
		delta += 7
	}
	return today.AddDate(0, 0, delta)
}

func periodOf(text string) (int, int) {
	switch {
	case strings.Contains(text, "凌晨") || strings.Contains(text, "dawn"):
		return 5, 8
	case strings.Contains(text, "早上") || strings.Contains(text, "早晨") || strings.Contains(text, "上午") || strings.Contains(text, "morning"):
		return 8, 12
	case strings.Contains(text, "中午") || strings.Contains(text, "noon"):
		return 11, 14
	case strings.Contains(text, "下午") || strings.Contains(text, "afternoon"):
		return 13, 18
	case strings.Contains(text, "傍晚"):
		return 17, 19
	case strings.Contains(text, "晚上") || strings.Contains(text, "evening") || strings.Contains(text, "night"):
		return 18, 22
	default:
		return 9, 18
	}
}

// normalizeClockWords lets "下午两点" reach the numeric clock regex.
func normalizeClockWords(text string) string {
	replacer := strings.NewReplacer(
		"十二点", "12点", "十一点", "11点", "十点", "10点",
		"九点", "9点", "八点", "8点", "七点", "7点",
		"六点", "6点", "五点", "5点", "四点", "4点",
		"三点", "3点", "两点", "2点", "一点", "1点",
	)
	return replacer.Replace(text)
}

// applyPeriod turns "下午两点" into 14:00 while leaving "上午十点" alone.
func applyPeriod(hour, periodStart int) int {
	if periodStart >= 12 && hour < 12 {
		return hour + 12
	}
	return hour
}

func validateWindow(w Window, now time.Time, durationMinutes int) (Window, error) {
	if !w.End.After(w.Start) {
		return Window{}, fmt.Errorf("结束时间必须晚于开始时间")
	}
	if w.End.Sub(w.Start) > MaxWindowHours*time.Hour {
		return Window{}, fmt.Errorf("时间窗口不能超过 %d 小时", MaxWindowHours)
	}
	if w.Start.Before(now.Add(-MinStartLeadTime)) {
		return Window{}, fmt.Errorf("开始时间已过去，请改为未来的时间")
	}
	if w.Start.After(now.Add(MaxPlanHorizon)) {
		return Window{}, fmt.Errorf("规划时间不能超过 %d 天", int(MaxPlanHorizon.Hours()/24))
	}
	if durationMinutes > 0 && w.End.Sub(w.Start) < time.Duration(durationMinutes)*time.Minute {
		return Window{}, fmt.Errorf("时间窗口短于所需时长 %d 分钟", durationMinutes)
	}
	return w, nil
}
