package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDate parses a natural language or formatted date string into a time.Time.
// Adapted from github.com/BRO3886/rem parser with calendar-specific extensions.
func ParseDate(input string) (time.Time, error) {
	return ParseDateRelativeTo(input, time.Now())
}

// ParseDateRelativeTo parses a date string relative to the given time.
// This allows testable date parsing without depending on wall clock.
//
// Supported inputs:
//   - Keywords: "today", "tomorrow", "yesterday", "now"
//   - End-of-period: "eod"/"end of day" (today 5pm), "eow"/"end of week" (Fri 5pm),
//     "this week" (Sun 23:59), "next week" (next Mon), "next month" (1st of next month)
//   - Relative: "in 3 hours", "in 2 weeks", "5 days ago"
//   - Weekdays: "next friday", "monday 2pm", "friday at 3:30pm"
//   - Month-day: "mar 15", "december 31 11:59pm"
//   - Time only: "5pm", "17:00", "3:30pm"
//   - Standard formats: ISO 8601, RFC 3339, US date, etc.
func ParseDateRelativeTo(input string, now time.Time) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try standard formats first
	if t, err := tryStandardFormats(input, now.Location()); err == nil {
		return t, nil
	}

	lower := strings.ToLower(input)

	// Handle "today", "tomorrow", "yesterday", "now"
	switch lower {
	case "today":
		return todayAt(now, 0, 0), nil
	case "tomorrow":
		return todayAt(now.AddDate(0, 0, 1), 0, 0), nil
	case "yesterday":
		return todayAt(now.AddDate(0, 0, -1), 0, 0), nil
	case "now":
		return now, nil
	case "eod", "end of day":
		return todayAt(now, 17, 0), nil
	case "eow", "end of week":
		// Work week end: Friday 5pm. If already Friday, returns today 5pm.
		daysUntilFriday := (5 - int(now.Weekday()) + 7) % 7
		if daysUntilFriday == 0 {
			return todayAt(now, 17, 0), nil
		}
		return todayAt(now.AddDate(0, 0, daysUntilFriday), 17, 0), nil
	case "this week":
		// Calendar week end: Sunday 23:59. If already Sunday, returns today 23:59.
		// Distinct from "eow" which targets the work week (Friday 5pm).
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		sunday := now.AddDate(0, 0, daysUntilSunday)
		return todayAt(sunday, 23, 59), nil
	case "next week":
		return nextWeekdayAt(now, time.Monday, 0, 0), nil
	case "next month":
		y, m, _ := now.Date()
		return time.Date(y, m+1, 1, 0, 0, 0, 0, now.Location()), nil
	}

	// Handle "in X hours/minutes/days/weeks/months"
	if t, err := parseRelative(lower, now); err == nil {
		return t, nil
	}

	// Handle "X hours/days/... ago"
	if t, err := parseAgo(lower, now); err == nil {
		return t, nil
	}

	// Handle "next monday", "next tuesday at 2pm", etc.
	if t, err := parseNextWeekday(lower, now); err == nil {
		return t, nil
	}

	// Handle "today at 5pm", "tomorrow at 3:30pm", "today 5pm", "tomorrow 3:30pm"
	if t, err := parseDateWithTime(lower, now); err == nil {
		return t, nil
	}

	// Handle "<weekday> <time>" e.g., "friday 2pm", "monday 10:00"
	if t, err := parseWeekdayWithTime(lower, now); err == nil {
		return t, nil
	}

	// Handle "<month> <day>" e.g., "mar 15", "march 15", "mar 15 2pm"
	if t, err := parseMonthDay(lower, now); err == nil {
		return t, nil
	}

	// Handle standalone weekday "monday", "friday"
	if wd, ok := weekdays[lower]; ok {
		return nextWeekdayAt(now, wd, 0, 0), nil
	}

	// Handle standalone time like "5pm", "17:00", "3:30pm"
	if t, err := parseTimeOnly(lower, now); err == nil {
		return t, nil
	}

	// Handle "<date> <time>" where date is ISO-like
	if t, err := parseDateTimeParts(lower, now); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("could not parse date: %q. Try: 'today', 'tomorrow 2pm', 'this week', 'eow', 'next friday', 'in 3 hours', or '2026-03-15 14:00'", input)
}

func tryStandardFormats(input string, loc *time.Location) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 3:04PM",
		"2006-01-02 3:04pm",
		"2006-01-02",
		"01/02/2006",
		"01/02/2006 15:04",
		"01/02/2006 3:04PM",
		"Jan 2, 2006",
		"Jan 2, 2006 3:04PM",
		"Jan 2, 2006 15:04",
		"January 2, 2006",
		"January 2, 2006 3:04PM",
		"January 2, 2006 15:04",
		"2 Jan 2006",
		"02 Jan 2006",
	}

	for _, f := range formats {
		if t, err := time.ParseInLocation(f, input, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("no standard format matched")
}

var relativePattern = regexp.MustCompile(`^in\s+(\d+)\s+(minute|minutes|min|mins|hour|hours|hr|hrs|day|days|week|weeks|month|months)$`)

func parseRelative(input string, now time.Time) (time.Time, error) {
	matches := relativePattern.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, fmt.Errorf("not a relative date")
	}

	amount, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch {
	case strings.HasPrefix(unit, "min"):
		return now.Add(time.Duration(amount) * time.Minute), nil
	case strings.HasPrefix(unit, "hour"), strings.HasPrefix(unit, "hr"):
		return now.Add(time.Duration(amount) * time.Hour), nil
	case strings.HasPrefix(unit, "day"):
		return now.AddDate(0, 0, amount), nil
	case strings.HasPrefix(unit, "week"):
		return now.AddDate(0, 0, amount*7), nil
	case strings.HasPrefix(unit, "month"):
		return now.AddDate(0, amount, 0), nil
	}
	return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
}

var agoPattern = regexp.MustCompile(`^(\d+)\s+(minute|minutes|min|mins|hour|hours|hr|hrs|day|days|week|weeks|month|months)\s+ago$`)

func parseAgo(input string, now time.Time) (time.Time, error) {
	matches := agoPattern.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, fmt.Errorf("not an ago date")
	}

	amount, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch {
	case strings.HasPrefix(unit, "min"):
		return now.Add(-time.Duration(amount) * time.Minute), nil
	case strings.HasPrefix(unit, "hour"), strings.HasPrefix(unit, "hr"):
		return now.Add(-time.Duration(amount) * time.Hour), nil
	case strings.HasPrefix(unit, "day"):
		return now.AddDate(0, 0, -amount), nil
	case strings.HasPrefix(unit, "week"):
		return now.AddDate(0, 0, -amount*7), nil
	case strings.HasPrefix(unit, "month"):
		return now.AddDate(0, -amount, 0), nil
	}
	return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
}

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday,
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
}

func parseNextWeekday(input string, now time.Time) (time.Time, error) {
	parts := strings.Fields(input)
	if len(parts) < 2 || parts[0] != "next" {
		return time.Time{}, fmt.Errorf("not a next weekday expression")
	}

	dayName := parts[1]
	targetDay, ok := weekdays[dayName]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown weekday: %s", dayName)
	}

	result := nextWeekdayAt(now, targetDay, 0, 0)

	// Check for "at" time specification: "next monday at 2pm"
	if len(parts) >= 4 && parts[2] == "at" {
		timeStr := strings.Join(parts[3:], " ")
		if hour, min, err := parseTimeStr(timeStr); err == nil {
			result = todayAt(result, hour, min)
		}
	} else if len(parts) >= 3 {
		// "next monday 2pm"
		timeStr := strings.Join(parts[2:], " ")
		if hour, min, err := parseTimeStr(timeStr); err == nil {
			result = todayAt(result, hour, min)
		}
	}

	return result, nil
}

func parseDateWithTime(input string, now time.Time) (time.Time, error) {
	// Try "today at 5pm" / "tomorrow at 3:30pm" first
	if parts := strings.SplitN(input, " at ", 2); len(parts) == 2 {
		datePart := strings.TrimSpace(parts[0])
		timePart := strings.TrimSpace(parts[1])

		baseDate, err := resolveBaseDate(datePart, now)
		if err != nil {
			return time.Time{}, err
		}
		hour, min, err := parseTimeStr(timePart)
		if err != nil {
			return time.Time{}, err
		}
		return todayAt(baseDate, hour, min), nil
	}

	// Try "today 5pm" / "tomorrow 3:30pm"
	parts := strings.Fields(input)
	if len(parts) >= 2 {
		datePart := parts[0]
		timePart := strings.Join(parts[1:], " ")
		baseDate, err := resolveBaseDate(datePart, now)
		if err != nil {
			return time.Time{}, err
		}
		hour, min, err := parseTimeStr(timePart)
		if err != nil {
			return time.Time{}, err
		}
		return todayAt(baseDate, hour, min), nil
	}

	return time.Time{}, fmt.Errorf("not a date with time")
}

func resolveBaseDate(datePart string, now time.Time) (time.Time, error) {
	switch datePart {
	case "today":
		return now, nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	}
	return time.Time{}, fmt.Errorf("unknown date base: %s", datePart)
}

func parseWeekdayWithTime(input string, now time.Time) (time.Time, error) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("not a weekday with time")
	}

	wd, ok := weekdays[parts[0]]
	if !ok {
		return time.Time{}, fmt.Errorf("not a weekday")
	}

	timeStr := strings.Join(parts[1:], " ")
	// Strip optional "at"
	timeStr = strings.TrimPrefix(timeStr, "at ")
	hour, min, err := parseTimeStr(timeStr)
	if err != nil {
		return time.Time{}, err
	}

	target := nextWeekdayAt(now, wd, hour, min)
	return target, nil
}

var months = map[string]time.Month{
	"jan": time.January, "january": time.January,
	"feb": time.February, "february": time.February,
	"mar": time.March, "march": time.March,
	"apr": time.April, "april": time.April,
	"may": time.May,
	"jun": time.June, "june": time.June,
	"jul": time.July, "july": time.July,
	"aug": time.August, "august": time.August,
	"sep": time.September, "september": time.September,
	"oct": time.October, "october": time.October,
	"nov": time.November, "november": time.November,
	"dec": time.December, "december": time.December,
}

var reMonthDay = regexp.MustCompile(`^([a-z]+)\s+(\d{1,2})(?:\s+(.+))?$`)

func parseMonthDay(input string, now time.Time) (time.Time, error) {
	m := reMonthDay.FindStringSubmatch(input)
	if m == nil {
		return time.Time{}, fmt.Errorf("not a month-day expression")
	}

	mon, ok := months[m[1]]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown month: %s", m[1])
	}

	day, _ := strconv.Atoi(m[2])
	if day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("invalid day: %d", day)
	}

	y := now.Year()
	result := time.Date(y, mon, day, 0, 0, 0, 0, now.Location())

	// If there's a time part
	if m[3] != "" {
		timeStr := strings.TrimSpace(m[3])
		timeStr = strings.TrimPrefix(timeStr, "at ")
		if hour, min, err := parseTimeStr(timeStr); err == nil {
			result = time.Date(y, mon, day, hour, min, 0, 0, now.Location())
		}
	}

	return result, nil
}

func parseTimeOnly(input string, now time.Time) (time.Time, error) {
	hour, min, err := parseTimeStr(input)
	if err != nil {
		return time.Time{}, err
	}
	return todayAt(now, hour, min), nil
}

func parseDateTimeParts(input string, now time.Time) (time.Time, error) {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("not a date-time pair")
	}

	d, err := tryStandardFormats(parts[0], now.Location())
	if err != nil {
		return time.Time{}, err
	}

	hour, min, err := parseTimeStr(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	return todayAt(d, hour, min), nil
}

// Time parsing

var timePatterns = []struct {
	re     *regexp.Regexp
	parser func([]string) (int, int, error)
}{
	{
		re: regexp.MustCompile(`^(\d{1,2})\s*(am|pm)$`),
		parser: func(m []string) (int, int, error) {
			h, _ := strconv.Atoi(m[1])
			return convertTo24(h, 0, m[2])
		},
	},
	{
		re: regexp.MustCompile(`^(\d{1,2}):(\d{2})\s*(am|pm)$`),
		parser: func(m []string) (int, int, error) {
			h, _ := strconv.Atoi(m[1])
			min, _ := strconv.Atoi(m[2])
			return convertTo24(h, min, m[3])
		},
	},
	{
		re: regexp.MustCompile(`^(\d{1,2}):(\d{2})$`),
		parser: func(m []string) (int, int, error) {
			h, _ := strconv.Atoi(m[1])
			min, _ := strconv.Atoi(m[2])
			if h > 23 || min > 59 {
				return 0, 0, fmt.Errorf("invalid time")
			}
			return h, min, nil
		},
	},
}

func parseTimeStr(s string) (int, int, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	for _, p := range timePatterns {
		matches := p.re.FindStringSubmatch(s)
		if matches != nil {
			return p.parser(matches)
		}
	}
	return 0, 0, fmt.Errorf("unable to parse time: %q", s)
}

func convertTo24(hour, min int, period string) (int, int, error) {
	if hour < 1 || hour > 12 {
		return 0, 0, fmt.Errorf("invalid hour: %d", hour)
	}
	if period == "am" {
		if hour == 12 {
			hour = 0
		}
	} else {
		if hour != 12 {
			hour += 12
		}
	}
	return hour, min, nil
}

func todayAt(base time.Time, hour, min int) time.Time {
	return time.Date(base.Year(), base.Month(), base.Day(), hour, min, 0, 0, base.Location())
}

func nextWeekdayAt(now time.Time, target time.Weekday, hour, min int) time.Time {
	daysAhead := int(target) - int(now.Weekday())
	if daysAhead <= 0 {
		daysAhead += 7
	}
	d := now.AddDate(0, 0, daysAhead)
	return todayAt(d, hour, min)
}

// FormatDuration returns a human-readable duration like "1h 30m", "All Day", "3 days".
func FormatDuration(start, end time.Time, allDay bool) string {
	if allDay {
		days := int(end.Sub(start).Hours()/24 + 0.5)
		if days <= 1 {
			return "All Day"
		}
		return fmt.Sprintf("%d days", days)
	}
	d := end.Sub(start)
	if d < time.Minute {
		return "0m"
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if hours > 0 && mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", mins)
}

// FormatTimeRange returns a human-readable time range for table display.
func FormatTimeRange(start, end time.Time, allDay bool) string {
	if allDay {
		if todayAt(start, 0, 0).Equal(todayAt(end, 0, 0)) || end.Sub(start) <= 24*time.Hour {
			return "All Day"
		}
		return fmt.Sprintf("%s - %s", start.Format("Jan 02"), end.Format("Jan 02"))
	}
	if todayAt(start, 0, 0).Equal(todayAt(end, 0, 0)) {
		return fmt.Sprintf("%s - %s", start.Format("15:04"), end.Format("15:04"))
	}
	return fmt.Sprintf("%s - %s", start.Format("Jan 02 15:04"), end.Format("Jan 02 15:04"))
}

// ParseAlertDuration parses an alert offset string like "15m", "1h", "1d" into a time.Duration.
func ParseAlertDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, fmt.Errorf("empty alert duration")
	}

	re := regexp.MustCompile(`^(\d+)(m|min|mins|minutes?|h|hours?|d|days?)$`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid alert duration: %q (use e.g. 15m, 1h, 1d)", s)
	}

	n, _ := strconv.Atoi(m[1])
	unit := m[2]
	switch {
	case strings.HasPrefix(unit, "m"):
		return time.Duration(n) * time.Minute, nil
	case strings.HasPrefix(unit, "h"):
		return time.Duration(n) * time.Hour, nil
	case strings.HasPrefix(unit, "d"):
		return time.Duration(n) * 24 * time.Hour, nil
	}
	return 0, fmt.Errorf("invalid alert duration: %q", s)
}
