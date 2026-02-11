package export

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
)

// ICS exports events in iCalendar format (RFC 5545).
func ICS(events []calendar.Event, w io.Writer) error {
	fmt.Fprintln(w, "BEGIN:VCALENDAR")
	fmt.Fprintln(w, "VERSION:2.0")
	fmt.Fprintln(w, "PRODID:-//cal CLI//EN")
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
