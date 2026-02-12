package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BRO3886/ical/internal/parser"
	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// lastListPath returns the path to the cached event ID list.
func lastListPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ical-last-list")
}

// SaveLastList writes event IDs to cache so row numbers can be used later.
func SaveLastList(events []calendar.Event) {
	ids := make([]string, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	_ = os.WriteFile(lastListPath(), []byte(strings.Join(ids, "\n")+"\n"), 0644)
}

// LookupRowNumber returns the full event ID for a 1-based row number
// from the last listing cache. Returns "" if not found.
func LookupRowNumber(n int) string {
	data, err := os.ReadFile(lastListPath())
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if n < 1 || n > len(lines) {
		return ""
	}
	return lines[n-1]
}

// PrintEvents prints events in the specified format and caches event IDs
// for row-number-based lookup by show/update/delete.
func PrintEvents(events []calendar.Event, format string) {
	SaveLastList(events)
	switch format {
	case "json":
		printEventsJSON(events, os.Stdout)
	case "plain":
		printEventsPlain(events, os.Stdout)
	default:
		printEventsTable(events, os.Stdout)
	}
}

// PrintCalendars prints calendars in the specified format.
func PrintCalendars(calendars []calendar.Calendar, format string) {
	switch format {
	case "json":
		printCalendarsJSON(calendars, os.Stdout)
	case "plain":
		printCalendarsPlain(calendars, os.Stdout)
	default:
		printCalendarsTable(calendars, os.Stdout)
	}
}

// PrintEventDetail prints a single event with full details.
func PrintEventDetail(event *calendar.Event, format string) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(event, "", "  ")
		fmt.Println(string(data))
	case "plain":
		printEventDetailPlain(event, os.Stdout)
	default:
		printEventDetailTable(event, os.Stdout)
	}
}

// Events — Table

func printEventsTable(events []calendar.Event, w io.Writer) {
	if len(events) == 0 {
		fmt.Fprintln(w, "No events found.")
		return
	}

	t := tablewriter.NewTable(w,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
			Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
		}),
	)
	t.Header("#", "Time", "Title", "Calendar", "Location", "Duration")

	for i, e := range events {
		start := localizeTime(e.StartDate, e.TimeZone)
		end := localizeTime(e.EndDate, e.TimeZone)
		timeStr := parser.FormatTimeRange(start, end, e.AllDay)
		title := truncate(e.Title, 40)
		loc := truncate(e.Location, 25)
		dur := parser.FormatDuration(start, end, e.AllDay)
		calName := e.Calendar

		if e.AllDay {
			title = color.HiYellowString(title)
		}
		if e.Recurring {
			title = title + " " + color.HiCyanString("↻")
		}

		t.Append(fmt.Sprintf("%d", i+1), timeStr, title, calName, loc, dur)
	}

	t.Render()
}

// Events — JSON

type eventJSON struct {
	ID                 string                        `json:"id"`
	Title              string                        `json:"title"`
	StartDate          time.Time                     `json:"start_date"`
	EndDate            time.Time                     `json:"end_date"`
	AllDay             bool                          `json:"all_day"`
	Calendar           string                        `json:"calendar"`
	CalendarID         string                        `json:"calendar_id"`
	Location           string                        `json:"location,omitempty"`
	StructuredLocation *eventkit.StructuredLocation   `json:"structured_location,omitempty"`
	Notes              string                        `json:"notes,omitempty"`
	URL                string                        `json:"url,omitempty"`
	Status             string                        `json:"status"`
	Availability       string                        `json:"availability"`
	Organizer          string                        `json:"organizer,omitempty"`
	Attendees          []calendar.Attendee           `json:"attendees,omitempty"`
	Recurring          bool                          `json:"recurring"`
	RecurrenceRules    []eventkit.RecurrenceRule     `json:"recurrence_rules,omitempty"`
	Alerts             []calendar.Alert              `json:"alerts,omitempty"`
	TimeZone           string                        `json:"timezone,omitempty"`
	CreatedAt          time.Time                     `json:"created_at"`
	ModifiedAt         time.Time                     `json:"modified_at"`
}

func toEventJSON(e calendar.Event) eventJSON {
	return eventJSON{
		ID:                 e.ID,
		Title:              e.Title,
		StartDate:          e.StartDate,
		EndDate:            e.EndDate,
		AllDay:             e.AllDay,
		Calendar:           e.Calendar,
		CalendarID:         e.CalendarID,
		Location:           e.Location,
		StructuredLocation: e.StructuredLocation,
		Notes:              e.Notes,
		URL:                e.URL,
		Status:             e.Status.String(),
		Availability:       e.Availability.String(),
		Organizer:          e.Organizer,
		Attendees:          e.Attendees,
		Recurring:          e.Recurring,
		RecurrenceRules:    e.RecurrenceRules,
		Alerts:             e.Alerts,
		TimeZone:           e.TimeZone,
		CreatedAt:          e.CreatedAt,
		ModifiedAt:         e.ModifiedAt,
	}
}

func printEventsJSON(events []calendar.Event, w io.Writer) {
	out := make([]eventJSON, len(events))
	for i, e := range events {
		out[i] = toEventJSON(e)
	}
	data, _ := json.Marshal(out)
	fmt.Fprintln(w, string(data))
}

// Events — Plain

func printEventsPlain(events []calendar.Event, w io.Writer) {
	for i, e := range events {
		start := localizeTime(e.StartDate, e.TimeZone)
		end := localizeTime(e.EndDate, e.TimeZone)
		if e.AllDay {
			loc := ""
			if e.Location != "" {
				loc = " @ " + e.Location
			}
			fmt.Fprintf(w, "#%d [All Day] %s (%s)%s\n", i+1, e.Title, e.Calendar, loc)
		} else {
			loc := ""
			if e.Location != "" {
				loc = " @ " + e.Location
			}
			fmt.Fprintf(w, "#%d [%s-%s] %s (%s)%s\n",
				i+1,
				start.Format("15:04"),
				end.Format("15:04"),
				e.Title,
				e.Calendar,
				loc,
			)
		}
	}
}

// Calendars — Table

func printCalendarsTable(calendars []calendar.Calendar, w io.Writer) {
	if len(calendars) == 0 {
		fmt.Fprintln(w, "No calendars found.")
		return
	}

	t := tablewriter.NewTable(w,
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignCenter}},
			Row:    tw.CellConfig{Formatting: tw.CellFormatting{Alignment: tw.AlignLeft}},
		}),
	)
	t.Header("Name", "Source", "Type", "Color", "ReadOnly")

	for _, c := range calendars {
		readOnly := ""
		if c.ReadOnly {
			readOnly = "yes"
		}
		t.Append(c.Title, c.Source, c.Type.String(), c.Color, readOnly)
	}

	t.Render()
}

// Calendars — JSON

func printCalendarsJSON(calendars []calendar.Calendar, w io.Writer) {
	data, _ := json.Marshal(calendars)
	fmt.Fprintln(w, string(data))
}

// Calendars — Plain

func printCalendarsPlain(calendars []calendar.Calendar, w io.Writer) {
	for _, c := range calendars {
		ro := ""
		if c.ReadOnly {
			ro = " (read-only)"
		}
		fmt.Fprintf(w, "%s [%s] %s%s\n", c.Title, c.Source, c.Type.String(), ro)
	}
}

// Event Detail

func printEventDetailTable(e *calendar.Event, w io.Writer) {
	bold := color.New(color.Bold)

	start := localizeTime(e.StartDate, e.TimeZone)
	end := localizeTime(e.EndDate, e.TimeZone)

	bold.Fprintf(w, "Title:        ")
	fmt.Fprintln(w, e.Title)

	bold.Fprintf(w, "Calendar:     ")
	fmt.Fprintln(w, e.Calendar)

	bold.Fprintf(w, "Status:       ")
	fmt.Fprintln(w, e.Status.String())

	bold.Fprintf(w, "Start:        ")
	fmt.Fprintln(w, start.Format("Mon, 02 Jan 2006 15:04 MST"))

	bold.Fprintf(w, "End:          ")
	fmt.Fprintln(w, end.Format("Mon, 02 Jan 2006 15:04 MST"))

	bold.Fprintf(w, "Duration:     ")
	fmt.Fprintln(w, parser.FormatDuration(start, end, e.AllDay))

	if e.AllDay {
		bold.Fprintf(w, "All Day:      ")
		fmt.Fprintln(w, "Yes")
	}

	if e.Location != "" {
		bold.Fprintf(w, "Location:     ")
		fmt.Fprintln(w, e.Location)
	}

	if e.StructuredLocation != nil {
		bold.Fprintf(w, "Coordinates:  ")
		fmt.Fprintf(w, "%.4f, %.4f\n", e.StructuredLocation.Latitude, e.StructuredLocation.Longitude)
	}

	if e.URL != "" {
		bold.Fprintf(w, "URL:          ")
		fmt.Fprintln(w, e.URL)
	}

	if e.Notes != "" {
		bold.Fprintf(w, "Notes:        ")
		fmt.Fprintln(w, e.Notes)
	}

	if e.Recurring {
		bold.Fprintf(w, "Recurrence:   ")
		for i, rule := range e.RecurrenceRules {
			if i > 0 {
				fmt.Fprint(w, "              ")
			}
			fmt.Fprintln(w, FormatRecurrenceRule(rule))
		}
	}

	if len(e.Alerts) > 0 {
		bold.Fprintf(w, "Alerts:       ")
		for i, alert := range e.Alerts {
			if i > 0 {
				fmt.Fprint(w, "              ")
			}
			d := alert.RelativeOffset
			if d < 0 {
				d = -d
			}
			fmt.Fprintf(w, "%s before\n", formatAlertDuration(d))
		}
	}

	if len(e.Attendees) > 0 {
		bold.Fprintf(w, "Attendees:    ")
		for i, att := range e.Attendees {
			if i > 0 {
				fmt.Fprint(w, "              ")
			}
			fmt.Fprintf(w, "%s <%s> [%s]\n", att.Name, att.Email, att.Status.String())
		}
	}

	if e.Organizer != "" {
		bold.Fprintf(w, "Organizer:    ")
		fmt.Fprintln(w, e.Organizer)
	}

	if e.TimeZone != "" {
		bold.Fprintf(w, "Timezone:     ")
		fmt.Fprintln(w, e.TimeZone)
	}

	bold.Fprintf(w, "ID:           ")
	fmt.Fprintln(w, e.ID)

	bold.Fprintf(w, "Created:      ")
	fmt.Fprintln(w, e.CreatedAt.In(time.Local).Format("Mon, 02 Jan 2006 15:04 MST"))

	bold.Fprintf(w, "Modified:     ")
	fmt.Fprintln(w, e.ModifiedAt.In(time.Local).Format("Mon, 02 Jan 2006 15:04 MST"))
}

func printEventDetailPlain(e *calendar.Event, w io.Writer) {
	start := localizeTime(e.StartDate, e.TimeZone)
	end := localizeTime(e.EndDate, e.TimeZone)
	fmt.Fprintf(w, "Title: %s\n", e.Title)
	fmt.Fprintf(w, "Calendar: %s\n", e.Calendar)
	fmt.Fprintf(w, "Start: %s\n", start.Format(time.RFC3339))
	fmt.Fprintf(w, "End: %s\n", end.Format(time.RFC3339))
	if e.AllDay {
		fmt.Fprintln(w, "All Day: Yes")
	}
	if e.Location != "" {
		fmt.Fprintf(w, "Location: %s\n", e.Location)
	}
	if e.Notes != "" {
		fmt.Fprintf(w, "Notes: %s\n", e.Notes)
	}
	fmt.Fprintf(w, "ID: %s\n", e.ID)
}

// FormatRecurrenceRule returns a human-readable recurrence description.
func FormatRecurrenceRule(rule eventkit.RecurrenceRule) string {
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
		if len(rule.DaysOfTheWeek) > 0 {
			days := make([]string, len(rule.DaysOfTheWeek))
			for i, d := range rule.DaysOfTheWeek {
				days[i] = d.DayOfTheWeek.String()[:3]
			}
			fmt.Fprintf(&b, " on %s", strings.Join(days, ", "))
		}
	case eventkit.FrequencyMonthly:
		if rule.Interval == 1 {
			b.WriteString("Every month")
		} else {
			fmt.Fprintf(&b, "Every %d months", rule.Interval)
		}
		if len(rule.DaysOfTheMonth) > 0 {
			days := make([]string, len(rule.DaysOfTheMonth))
			for i, d := range rule.DaysOfTheMonth {
				days[i] = ordinal(d)
			}
			fmt.Fprintf(&b, " on the %s", strings.Join(days, ", "))
		}
		if len(rule.DaysOfTheWeek) > 0 {
			for _, d := range rule.DaysOfTheWeek {
				prefix := ""
				if d.WeekNumber == -1 {
					prefix = "last "
				} else if d.WeekNumber > 0 {
					prefix = ordinal(d.WeekNumber) + " "
				}
				fmt.Fprintf(&b, " on the %s%s", prefix, d.DayOfTheWeek.String()[:3])
			}
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
			fmt.Fprintf(&b, " until %s", rule.End.EndDate.Format("Jan 2, 2006"))
		}
		if rule.End.OccurrenceCount > 0 {
			fmt.Fprintf(&b, " for %d occurrences", rule.End.OccurrenceCount)
		}
	}

	return b.String()
}

func ordinal(n int) string {
	if n < 0 {
		return fmt.Sprintf("%d", n)
	}
	suffix := "th"
	switch n % 10 {
	case 1:
		if n%100 != 11 {
			suffix = "st"
		}
	case 2:
		if n%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if n%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

func formatAlertDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	if d >= time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh %dm", hours, mins)
		}
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	mins := int(d.Minutes())
	if mins == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", mins)
}

// localizeTime converts a time to the event's timezone, falling back to local time.
func localizeTime(t time.Time, tz string) time.Time {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return t.In(loc)
		}
	}
	return t.In(time.Local)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// ShortID returns the first 13 chars of an event ID.
// This covers two UUID segments (e.g. "577B8983-DF44") which is
// enough to disambiguate events from the same source.
func ShortID(id string) string {
	if len(id) <= 13 {
		return id
	}
	return id[:13]
}

// PrintCreatedEvent prints summary info for a newly created event.
func PrintCreatedEvent(e *calendar.Event) {
	start := localizeTime(e.StartDate, e.TimeZone)
	end := localizeTime(e.EndDate, e.TimeZone)
	green := color.New(color.FgGreen, color.Bold)
	green.Print("Created: ")
	fmt.Printf("%s\n", e.Title)
	fmt.Printf("  Calendar: %s\n", e.Calendar)
	fmt.Printf("  When:     %s\n", parser.FormatTimeRange(start, end, e.AllDay))
	fmt.Printf("  ID:       %s\n", ShortID(e.ID))
}

// PrintCreatedCalendar prints summary info for a newly created calendar.
func PrintCreatedCalendar(c *calendar.Calendar) {
	green := color.New(color.FgGreen, color.Bold)
	green.Print("Created: ")
	fmt.Printf("%s\n", c.Title)
	fmt.Printf("  Source: %s\n", c.Source)
	if c.Color != "" {
		fmt.Printf("  Color:  %s\n", c.Color)
	}
	fmt.Printf("  ID:     %s\n", c.ID)
}

// PrintUpdatedCalendar prints summary info for an updated calendar.
func PrintUpdatedCalendar(c *calendar.Calendar) {
	green := color.New(color.FgGreen, color.Bold)
	green.Print("Updated: ")
	fmt.Printf("%s\n", c.Title)
	fmt.Printf("  Source: %s\n", c.Source)
	if c.Color != "" {
		fmt.Printf("  Color:  %s\n", c.Color)
	}
	fmt.Printf("  ID:     %s\n", c.ID)
}

// PrintUpdatedEvent prints summary info for an updated event.
func PrintUpdatedEvent(e *calendar.Event) {
	start := localizeTime(e.StartDate, e.TimeZone)
	end := localizeTime(e.EndDate, e.TimeZone)
	green := color.New(color.FgGreen, color.Bold)
	green.Print("Updated: ")
	fmt.Printf("%s\n", e.Title)
	fmt.Printf("  Calendar: %s\n", e.Calendar)
	fmt.Printf("  When:     %s\n", parser.FormatTimeRange(start, end, e.AllDay))
	fmt.Printf("  ID:       %s\n", ShortID(e.ID))
}
