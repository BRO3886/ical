package export

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
)

// ICS exports events in iCalendar format (RFC 5545).
func ICS(events []calendar.Event, w io.Writer) error {
	fmt.Fprintln(w, "BEGIN:VCALENDAR")
	fmt.Fprintln(w, "VERSION:2.0")
	fmt.Fprintln(w, "PRODID:-//ical CLI//EN")
	fmt.Fprintln(w, "CALSCALE:GREGORIAN")

	for _, e := range events {
		fmt.Fprintln(w, "BEGIN:VEVENT")
		fmt.Fprintf(w, "UID:%s\n", e.ID)

		if e.AllDay {
			fmt.Fprintf(w, "DTSTART;VALUE=DATE:%s\n", e.StartDate.Format("20060102"))
			fmt.Fprintf(w, "DTEND;VALUE=DATE:%s\n", e.EndDate.Format("20060102"))
		} else {
			fmt.Fprintf(w, "DTSTART:%s\n", e.StartDate.UTC().Format("20060102T150405Z"))
			fmt.Fprintf(w, "DTEND:%s\n", e.EndDate.UTC().Format("20060102T150405Z"))
		}

		fmt.Fprintf(w, "SUMMARY:%s\n", escapeICS(e.Title))

		if e.Location != "" {
			fmt.Fprintf(w, "LOCATION:%s\n", escapeICS(e.Location))
		}
		if e.Notes != "" {
			fmt.Fprintf(w, "DESCRIPTION:%s\n", escapeICS(e.Notes))
		}
		if e.URL != "" {
			fmt.Fprintf(w, "URL:%s\n", e.URL)
		}

		for _, rule := range e.RecurrenceRules {
			fmt.Fprintf(w, "RRULE:%s\n", formatRRule(rule))
		}

		for _, alert := range e.Alerts {
			fmt.Fprintln(w, "BEGIN:VALARM")
			fmt.Fprintln(w, "ACTION:DISPLAY")
			fmt.Fprintf(w, "DESCRIPTION:%s\n", escapeICS(e.Title))
			d := alert.RelativeOffset
			if d < 0 {
				d = -d
			}
			fmt.Fprintf(w, "TRIGGER:-PT%dM\n", int(d.Minutes()))
			fmt.Fprintln(w, "END:VALARM")
		}

		fmt.Fprintf(w, "CREATED:%s\n", e.CreatedAt.UTC().Format("20060102T150405Z"))
		fmt.Fprintf(w, "LAST-MODIFIED:%s\n", e.ModifiedAt.UTC().Format("20060102T150405Z"))

		fmt.Fprintln(w, "END:VEVENT")
	}

	fmt.Fprintln(w, "END:VCALENDAR")
	return nil
}

func escapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func formatRRule(rule eventkit.RecurrenceRule) string {
	var parts []string

	switch rule.Frequency {
	case eventkit.FrequencyDaily:
		parts = append(parts, "FREQ=DAILY")
	case eventkit.FrequencyWeekly:
		parts = append(parts, "FREQ=WEEKLY")
	case eventkit.FrequencyMonthly:
		parts = append(parts, "FREQ=MONTHLY")
	case eventkit.FrequencyYearly:
		parts = append(parts, "FREQ=YEARLY")
	}

	if rule.Interval > 1 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", rule.Interval))
	}

	if len(rule.DaysOfTheWeek) > 0 {
		dayAbbrs := map[eventkit.Weekday]string{
			eventkit.Sunday:    "SU",
			eventkit.Monday:    "MO",
			eventkit.Tuesday:   "TU",
			eventkit.Wednesday: "WE",
			eventkit.Thursday:  "TH",
			eventkit.Friday:    "FR",
			eventkit.Saturday:  "SA",
		}
		days := make([]string, len(rule.DaysOfTheWeek))
		for i, d := range rule.DaysOfTheWeek {
			abbr := dayAbbrs[d.DayOfTheWeek]
			if d.WeekNumber != 0 {
				days[i] = fmt.Sprintf("%d%s", d.WeekNumber, abbr)
			} else {
				days[i] = abbr
			}
		}
		parts = append(parts, "BYDAY="+strings.Join(days, ","))
	}

	if len(rule.DaysOfTheMonth) > 0 {
		days := make([]string, len(rule.DaysOfTheMonth))
		for i, d := range rule.DaysOfTheMonth {
			days[i] = fmt.Sprintf("%d", d)
		}
		parts = append(parts, "BYMONTHDAY="+strings.Join(days, ","))
	}

	if rule.End != nil {
		if rule.End.EndDate != nil {
			parts = append(parts, "UNTIL="+rule.End.EndDate.UTC().Format("20060102T150405Z"))
		}
		if rule.End.OccurrenceCount > 0 {
			parts = append(parts, fmt.Sprintf("COUNT=%d", rule.End.OccurrenceCount))
		}
	}

	return strings.Join(parts, ";")
}

// ParseICS reads an ICS (iCalendar RFC 5545) file and returns CreateEventInput slice.
func ParseICS(r io.Reader) ([]calendar.CreateEventInput, error) {
	lines := unfoldICS(r)

	var inputs []calendar.CreateEventInput
	var cur *icsEvent
	inAlarm := false

	for _, line := range lines {
		switch {
		case line == "BEGIN:VEVENT":
			cur = &icsEvent{}
			inAlarm = false
		case line == "END:VEVENT":
			if cur == nil {
				continue
			}
			input, err := cur.toInput()
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, input)
			cur = nil
		case line == "BEGIN:VALARM":
			inAlarm = true
		case line == "END:VALARM":
			inAlarm = false
		default:
			if cur == nil {
				continue
			}
			if inAlarm {
				if key, val, ok := splitICSLine(line); ok && key == "TRIGGER" {
					if d, err := parseTrigger(val); err == nil {
						cur.alerts = append(cur.alerts, d)
					}
				}
				continue
			}
			key, val, ok := splitICSLine(line)
			if !ok {
				continue
			}
			switch {
			case key == "SUMMARY":
				cur.title = unescapeICS(val)
			case key == "LOCATION":
				cur.location = unescapeICS(val)
			case key == "DESCRIPTION":
				cur.notes = unescapeICS(val)
			case key == "URL":
				cur.url = val
			case key == "RRULE":
				if rule, err := parseRRule(val); err == nil {
					cur.rrules = append(cur.rrules, rule)
				}
			case key == "DTSTART" || strings.HasPrefix(key, "DTSTART;"):
				cur.dtstart = val
				cur.dtstartAllDay = strings.Contains(key, "VALUE=DATE")
			case key == "DTEND" || strings.HasPrefix(key, "DTEND;"):
				cur.dtend = val
				cur.dtendAllDay = strings.Contains(key, "VALUE=DATE")
			}
		}
	}

	return inputs, nil
}

// icsEvent holds parsed VEVENT properties before conversion.
type icsEvent struct {
	title        string
	dtstart      string
	dtend        string
	dtstartAllDay bool
	dtendAllDay  bool
	location     string
	notes        string
	url          string
	rrules       []eventkit.RecurrenceRule
	alerts       []time.Duration
}

func (e *icsEvent) toInput() (calendar.CreateEventInput, error) {
	if e.title == "" {
		return calendar.CreateEventInput{}, fmt.Errorf("VEVENT missing SUMMARY")
	}

	allDay := e.dtstartAllDay
	start, err := parseICSDateTime(e.dtstart, allDay)
	if err != nil {
		return calendar.CreateEventInput{}, fmt.Errorf("invalid DTSTART %q: %w", e.dtstart, err)
	}

	// DTEND is optional in ICS; default to start + 1 hour (or +1 day for all-day)
	var end time.Time
	if e.dtend != "" {
		end, err = parseICSDateTime(e.dtend, e.dtendAllDay)
		if err != nil {
			return calendar.CreateEventInput{}, fmt.Errorf("invalid DTEND %q: %w", e.dtend, err)
		}
	} else if allDay {
		end = start.AddDate(0, 0, 1)
	} else {
		end = start.Add(time.Hour)
	}

	var alerts []calendar.Alert
	for _, d := range e.alerts {
		alerts = append(alerts, calendar.Alert{RelativeOffset: d})
	}

	return calendar.CreateEventInput{
		Title:           e.title,
		StartDate:       start,
		EndDate:         end,
		AllDay:          allDay,
		Location:        e.location,
		Notes:           e.notes,
		URL:             e.url,
		Alerts:          alerts,
		RecurrenceRules: e.rrules,
	}, nil
}

// unfoldICS reads an ICS stream and unfolds continuation lines (RFC 5545 §3.1).
// Lines starting with a space or tab are appended to the previous line.
func unfoldICS(r io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if len(lines) > 0 {
				lines[len(lines)-1] += line[1:]
			}
		} else {
			lines = append(lines, line)
		}
	}
	return lines
}

// splitICSLine splits "KEY:VALUE" or "KEY;PARAMS:VALUE" returning the full key
// (including params) and the value. Returns false if the line has no colon.
func splitICSLine(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	return line[:idx], line[idx+1:], true
}

func unescapeICS(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\N", "\n")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// parseICSDateTime parses an ICS date or datetime string.
// For all-day events (VALUE=DATE), format is "20060102".
// For timed events, format is "20060102T150405Z" (UTC) or "20060102T150405".
func parseICSDateTime(s string, dateOnly bool) (time.Time, error) {
	if dateOnly {
		return time.Parse("20060102", s)
	}
	if strings.HasSuffix(s, "Z") {
		return time.Parse("20060102T150405Z", s)
	}
	return time.Parse("20060102T150405", s)
}

// parseRRule parses an RRULE value like "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE".
func parseRRule(val string) (eventkit.RecurrenceRule, error) {
	rule := eventkit.RecurrenceRule{Interval: 1}
	hasFreq := false

	for _, part := range strings.Split(val, ";") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]

		switch key {
		case "FREQ":
			hasFreq = true
			switch value {
			case "DAILY":
				rule.Frequency = eventkit.FrequencyDaily
			case "WEEKLY":
				rule.Frequency = eventkit.FrequencyWeekly
			case "MONTHLY":
				rule.Frequency = eventkit.FrequencyMonthly
			case "YEARLY":
				rule.Frequency = eventkit.FrequencyYearly
			default:
				return rule, fmt.Errorf("unknown FREQ %q", value)
			}
		case "INTERVAL":
			n, err := strconv.Atoi(value)
			if err != nil {
				return rule, fmt.Errorf("invalid INTERVAL %q: %w", value, err)
			}
			rule.Interval = n
		case "BYDAY":
			for _, dayStr := range strings.Split(value, ",") {
				dow, err := parseBYDAY(dayStr)
				if err != nil {
					return rule, err
				}
				rule.DaysOfTheWeek = append(rule.DaysOfTheWeek, dow)
			}
		case "BYMONTHDAY":
			for _, ds := range strings.Split(value, ",") {
				n, err := strconv.Atoi(ds)
				if err != nil {
					return rule, fmt.Errorf("invalid BYMONTHDAY %q: %w", ds, err)
				}
				rule.DaysOfTheMonth = append(rule.DaysOfTheMonth, n)
			}
		case "UNTIL":
			t, err := parseICSDateTime(value, len(value) == 8)
			if err != nil {
				return rule, fmt.Errorf("invalid UNTIL %q: %w", value, err)
			}
			rule.End = &eventkit.RecurrenceEnd{EndDate: &t}
		case "COUNT":
			n, err := strconv.Atoi(value)
			if err != nil {
				return rule, fmt.Errorf("invalid COUNT %q: %w", value, err)
			}
			rule.End = &eventkit.RecurrenceEnd{OccurrenceCount: n}
		}
	}

	if !hasFreq {
		return rule, fmt.Errorf("RRULE missing FREQ")
	}
	return rule, nil
}

// parseBYDAY parses a BYDAY value like "MO", "2TU", "-1FR".
func parseBYDAY(s string) (eventkit.RecurrenceDayOfWeek, error) {
	dayAbbrs := map[string]eventkit.Weekday{
		"SU": eventkit.Sunday,
		"MO": eventkit.Monday,
		"TU": eventkit.Tuesday,
		"WE": eventkit.Wednesday,
		"TH": eventkit.Thursday,
		"FR": eventkit.Friday,
		"SA": eventkit.Saturday,
	}

	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return eventkit.RecurrenceDayOfWeek{}, fmt.Errorf("invalid BYDAY %q", s)
	}

	abbr := s[len(s)-2:]
	day, ok := dayAbbrs[abbr]
	if !ok {
		return eventkit.RecurrenceDayOfWeek{}, fmt.Errorf("unknown day %q in BYDAY", abbr)
	}

	var weekNum int
	if len(s) > 2 {
		n, err := strconv.Atoi(s[:len(s)-2])
		if err != nil {
			return eventkit.RecurrenceDayOfWeek{}, fmt.Errorf("invalid week number in BYDAY %q: %w", s, err)
		}
		weekNum = n
	}

	return eventkit.RecurrenceDayOfWeek{
		DayOfTheWeek: day,
		WeekNumber:   weekNum,
	}, nil
}

// parseTrigger parses an ICS TRIGGER value like "-PT15M", "-PT1H", "-P1D", "-P1W".
func parseTrigger(val string) (time.Duration, error) {
	s := val
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}

	if !strings.HasPrefix(s, "P") {
		return 0, fmt.Errorf("invalid trigger %q: missing P", val)
	}
	s = s[1:]

	var total time.Duration

	// Handle week duration: P1W
	if strings.HasSuffix(s, "W") {
		n, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid trigger weeks %q: %w", val, err)
		}
		total = time.Duration(n) * 7 * 24 * time.Hour
		if neg {
			total = -total
		}
		return total, nil
	}

	// Handle day portion: P1D, P1DT2H30M, etc.
	if idx := strings.Index(s, "D"); idx >= 0 {
		n, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, fmt.Errorf("invalid trigger days %q: %w", val, err)
		}
		total += time.Duration(n) * 24 * time.Hour
		s = s[idx+1:]
	}

	// Handle time portion after "T"
	if strings.HasPrefix(s, "T") {
		s = s[1:]
	}

	// Parse hours
	if idx := strings.Index(s, "H"); idx >= 0 {
		n, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, fmt.Errorf("invalid trigger hours %q: %w", val, err)
		}
		total += time.Duration(n) * time.Hour
		s = s[idx+1:]
	}

	// Parse minutes
	if idx := strings.Index(s, "M"); idx >= 0 {
		n, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, fmt.Errorf("invalid trigger minutes %q: %w", val, err)
		}
		total += time.Duration(n) * time.Minute
		s = s[idx+1:]
	}

	// Parse seconds
	if idx := strings.Index(s, "S"); idx >= 0 {
		n, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, fmt.Errorf("invalid trigger seconds %q: %w", val, err)
		}
		total += time.Duration(n) * time.Second
	}

	if neg {
		total = -total
	}
	return total, nil
}

// FormatRRuleToHuman converts a recurrence rule to human-readable text.
func FormatRRuleToHuman(rule eventkit.RecurrenceRule) string {
	var b strings.Builder

	switch rule.Frequency {
	case eventkit.FrequencyDaily:
		if rule.Interval == 1 {
			b.WriteString("Every day")
		} else {
			fmt.Fprintf(&b, "Every %d days", rule.Interval)
		}
	case eventkit.FrequencyWeekly:
		if rule.Interval == 1 {
			b.WriteString("Every week")
		} else {
			fmt.Fprintf(&b, "Every %d weeks", rule.Interval)
		}
	case eventkit.FrequencyMonthly:
		if rule.Interval == 1 {
			b.WriteString("Every month")
		} else {
			fmt.Fprintf(&b, "Every %d months", rule.Interval)
		}
	case eventkit.FrequencyYearly:
		if rule.Interval == 1 {
			b.WriteString("Every year")
		} else {
			fmt.Fprintf(&b, "Every %d years", rule.Interval)
		}
	}

	if rule.End != nil {
		if rule.End.EndDate != nil {
			fmt.Fprintf(&b, " until %s", rule.End.EndDate.Format(time.DateOnly))
		}
		if rule.End.OccurrenceCount > 0 {
			fmt.Fprintf(&b, " for %d occurrences", rule.End.OccurrenceCount)
		}
	}

	return b.String()
}
