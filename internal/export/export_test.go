package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
)

func sampleEvents() []calendar.Event {
	return []calendar.Event{
		{
			ID:        "event-1",
			Title:     "Team Standup",
			StartDate: time.Date(2026, 2, 11, 9, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 2, 11, 9, 30, 0, 0, time.UTC),
			AllDay:    false,
			Calendar:  "Work",
			Location:  "Room A",
			Notes:     "Daily sync",
			Status:    1, // Confirmed
			TimeZone:  "Asia/Kolkata",
		},
		{
			ID:        "event-2",
			Title:     "Company Holiday",
			StartDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
			AllDay:    true,
			Calendar:  "Work",
			Status:    1,
		},
	}
}

func TestJSON_Export(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()

	err := JSON(events, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []eventExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 events, got %d", len(result))
	}
	if result[0].Title != "Team Standup" {
		t.Errorf("expected title 'Team Standup', got %q", result[0].Title)
	}
	if result[0].Location != "Room A" {
		t.Errorf("expected location 'Room A', got %q", result[0].Location)
	}
	if result[1].AllDay != true {
		t.Error("expected all-day to be true for second event")
	}
}

func TestJSON_ExportEmpty(t *testing.T) {
	var buf bytes.Buffer
	err := JSON([]calendar.Event{}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "[]") {
		t.Errorf("expected empty array, got %q", buf.String())
	}
}

func TestJSON_Roundtrip(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()

	err := JSON(events, &buf)
	if err != nil {
		t.Fatalf("export error: %v", err)
	}

	inputs, err := ParseJSON(&buf)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(inputs))
	}
	if inputs[0].Title != "Team Standup" {
		t.Errorf("title mismatch: got %q", inputs[0].Title)
	}
	if inputs[0].Location != "Room A" {
		t.Errorf("location mismatch: got %q", inputs[0].Location)
	}
	if inputs[0].Calendar != "Work" {
		t.Errorf("calendar mismatch: got %q", inputs[0].Calendar)
	}
	if inputs[1].AllDay != true {
		t.Error("all-day should be preserved")
	}
}

func TestCSV_Export(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()

	err := CSV(events, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) != 3 { // header + 2 events
		t.Errorf("expected 3 records (header + 2), got %d", len(records))
	}

	header := records[0]
	expected := []string{"ID", "Title", "Start", "End", "AllDay", "Calendar", "Location", "Notes", "URL", "Status", "Recurring", "Timezone"}
	for i, h := range expected {
		if header[i] != h {
			t.Errorf("header[%d]: expected %q, got %q", i, h, header[i])
		}
	}

	// Check first data row
	if records[1][1] != "Team Standup" {
		t.Errorf("expected title 'Team Standup', got %q", records[1][1])
	}
}

func TestCSV_Roundtrip(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()

	err := CSV(events, &buf)
	if err != nil {
		t.Fatalf("export error: %v", err)
	}

	inputs, err := ParseCSV(&buf)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(inputs))
	}
	if inputs[0].Title != "Team Standup" {
		t.Errorf("title mismatch: got %q", inputs[0].Title)
	}
	if inputs[0].Location != "Room A" {
		t.Errorf("location mismatch: got %q", inputs[0].Location)
	}
}

func TestCSV_ParseEmpty(t *testing.T) {
	r := strings.NewReader("ID,Title,Start,End,AllDay\n")
	_, err := ParseCSV(r)
	if err == nil {
		t.Error("expected error for CSV with header only")
	}
}

func TestCSV_ParseInvalidDate(t *testing.T) {
	r := strings.NewReader("ID,Title,Start,End,AllDay\n1,Test,invalid,invalid,false\n")
	_, err := ParseCSV(r)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestICS_Export(t *testing.T) {
	var buf bytes.Buffer
	events := sampleEvents()

	err := ICS(events, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check structure
	if !strings.Contains(output, "BEGIN:VCALENDAR") {
		t.Error("missing BEGIN:VCALENDAR")
	}
	if !strings.Contains(output, "END:VCALENDAR") {
		t.Error("missing END:VCALENDAR")
	}
	if !strings.Contains(output, "BEGIN:VEVENT") {
		t.Error("missing BEGIN:VEVENT")
	}
	if !strings.Contains(output, "SUMMARY:Team Standup") {
		t.Error("missing event summary")
	}
	if !strings.Contains(output, "LOCATION:Room A") {
		t.Error("missing location")
	}
	if !strings.Contains(output, "DESCRIPTION:Daily sync") {
		t.Error("missing description")
	}

	// All-day event should use VALUE=DATE
	if !strings.Contains(output, "DTSTART;VALUE=DATE:20260315") {
		t.Error("all-day event should use VALUE=DATE format")
	}

	// Non-all-day event should use UTC format
	if !strings.Contains(output, "DTSTART:20260211T090000Z") {
		t.Error("timed event should use UTC format")
	}
}

func TestICS_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	events := []calendar.Event{
		{
			ID:        "event-special",
			Title:     "Meeting; with, special\\chars",
			Notes:     "Line 1\nLine 2",
			StartDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC),
		},
	}

	err := ICS(events, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Semicolons should be escaped
	if !strings.Contains(output, "\\;") {
		t.Error("semicolons should be escaped in ICS")
	}
	// Commas should be escaped
	if !strings.Contains(output, "\\,") {
		t.Error("commas should be escaped in ICS")
	}
	// Newlines should be escaped
	if !strings.Contains(output, "\\n") {
		t.Error("newlines should be escaped in ICS")
	}
}

func TestICS_WithAlerts(t *testing.T) {
	var buf bytes.Buffer
	events := []calendar.Event{
		{
			ID:        "event-alert",
			Title:     "Meeting",
			StartDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC),
			Alerts: []calendar.Alert{
				{RelativeOffset: -15 * time.Minute},
				{RelativeOffset: -1 * time.Hour},
			},
		},
	}

	err := ICS(events, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "BEGIN:VALARM") {
		t.Error("missing VALARM")
	}
	if !strings.Contains(output, "TRIGGER:-PT15M") {
		t.Error("missing 15-minute trigger")
	}
	if !strings.Contains(output, "TRIGGER:-PT60M") {
		t.Error("missing 60-minute trigger")
	}
}

func TestJSON_ParseInvalid(t *testing.T) {
	r := strings.NewReader("not json")
	_, err := ParseJSON(r)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCSV_ExportWithSpecialChars(t *testing.T) {
	var buf bytes.Buffer
	events := []calendar.Event{
		{
			ID:        "event-csv",
			Title:     "Meeting, with commas",
			Notes:     "Line 1\nLine 2",
			StartDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC),
			Calendar:  "Work",
		},
	}

	err := CSV(events, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CSV should properly quote fields with commas
	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}
	if records[1][1] != "Meeting, with commas" {
		t.Errorf("commas not properly handled: got %q", records[1][1])
	}
}

// --- ICS Import Tests ---

func TestICS_ParseBasic(t *testing.T) {
	ics := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
SUMMARY:Team Standup
DTSTART:20260211T090000Z
DTEND:20260211T093000Z
LOCATION:Room A
DESCRIPTION:Daily sync
URL:https://meet.example.com
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(inputs))
	}

	e := inputs[0]
	if e.Title != "Team Standup" {
		t.Errorf("title: got %q", e.Title)
	}
	if e.Location != "Room A" {
		t.Errorf("location: got %q", e.Location)
	}
	if e.Notes != "Daily sync" {
		t.Errorf("notes: got %q", e.Notes)
	}
	if e.URL != "https://meet.example.com" {
		t.Errorf("url: got %q", e.URL)
	}
	if !e.StartDate.Equal(time.Date(2026, 2, 11, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("start: got %v", e.StartDate)
	}
	if !e.EndDate.Equal(time.Date(2026, 2, 11, 9, 30, 0, 0, time.UTC)) {
		t.Errorf("end: got %v", e.EndDate)
	}
	if e.AllDay {
		t.Error("should not be all-day")
	}
}

func TestICS_ParseAllDay(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Company Holiday
DTSTART;VALUE=DATE:20260315
DTEND;VALUE=DATE:20260316
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(inputs))
	}

	e := inputs[0]
	if !e.AllDay {
		t.Error("should be all-day")
	}
	if !e.StartDate.Equal(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start: got %v", e.StartDate)
	}
	if !e.EndDate.Equal(time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end: got %v", e.EndDate)
	}
}

func TestICS_ParseMultipleEvents(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Event 1
DTSTART:20260101T100000Z
DTEND:20260101T110000Z
END:VEVENT
BEGIN:VEVENT
SUMMARY:Event 2
DTSTART:20260102T140000Z
DTEND:20260102T150000Z
END:VEVENT
BEGIN:VEVENT
SUMMARY:Event 3
DTSTART;VALUE=DATE:20260103
DTEND;VALUE=DATE:20260104
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 3 {
		t.Fatalf("expected 3 events, got %d", len(inputs))
	}
	if inputs[0].Title != "Event 1" {
		t.Errorf("event 1 title: got %q", inputs[0].Title)
	}
	if inputs[1].Title != "Event 2" {
		t.Errorf("event 2 title: got %q", inputs[1].Title)
	}
	if inputs[2].Title != "Event 3" {
		t.Errorf("event 3 title: got %q", inputs[2].Title)
	}
	if !inputs[2].AllDay {
		t.Error("event 3 should be all-day")
	}
}

func TestICS_ParseSpecialCharacters(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Meeting\; with\, special\\chars
DESCRIPTION:Line 1\nLine 2
DTSTART:20260101T100000Z
DTEND:20260101T110000Z
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := inputs[0]
	if e.Title != "Meeting; with, special\\chars" {
		t.Errorf("title unescape: got %q", e.Title)
	}
	if e.Notes != "Line 1\nLine 2" {
		t.Errorf("notes unescape: got %q", e.Notes)
	}
}

func TestICS_ParseLineFolding(t *testing.T) {
	// RFC 5545: long lines are folded with CRLF + space
	ics := "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nSUMMARY:This is a very long\r\n  event title that was folded\r\nDTSTART:20260101T100000Z\r\nDTEND:20260101T110000Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inputs[0].Title != "This is a very long event title that was folded" {
		t.Errorf("folded title: got %q", inputs[0].Title)
	}
}

func TestICS_ParseWithAlerts(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Meeting
DTSTART:20260101T100000Z
DTEND:20260101T110000Z
BEGIN:VALARM
ACTION:DISPLAY
DESCRIPTION:Meeting
TRIGGER:-PT15M
END:VALARM
BEGIN:VALARM
ACTION:DISPLAY
DESCRIPTION:Meeting
TRIGGER:-PT1H
END:VALARM
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := inputs[0]
	if len(e.Alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(e.Alerts))
	}
	if e.Alerts[0].RelativeOffset != -15*time.Minute {
		t.Errorf("alert 0: got %v", e.Alerts[0].RelativeOffset)
	}
	if e.Alerts[1].RelativeOffset != -1*time.Hour {
		t.Errorf("alert 1: got %v", e.Alerts[1].RelativeOffset)
	}
}

func TestICS_ParseWithRRule(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Weekly Standup
DTSTART:20260211T090000Z
DTEND:20260211T093000Z
RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,WE
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := inputs[0]
	if len(e.RecurrenceRules) != 1 {
		t.Fatalf("expected 1 rrule, got %d", len(e.RecurrenceRules))
	}
	rule := e.RecurrenceRules[0]
	if rule.Frequency != eventkit.FrequencyWeekly {
		t.Errorf("frequency: got %v", rule.Frequency)
	}
	if rule.Interval != 2 {
		t.Errorf("interval: got %d", rule.Interval)
	}
	if len(rule.DaysOfTheWeek) != 2 {
		t.Fatalf("expected 2 days, got %d", len(rule.DaysOfTheWeek))
	}
	if rule.DaysOfTheWeek[0].DayOfTheWeek != eventkit.Monday {
		t.Errorf("day 0: got %v", rule.DaysOfTheWeek[0].DayOfTheWeek)
	}
	if rule.DaysOfTheWeek[1].DayOfTheWeek != eventkit.Wednesday {
		t.Errorf("day 1: got %v", rule.DaysOfTheWeek[1].DayOfTheWeek)
	}
}

func TestICS_ParseRRuleWithCount(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Daily
DTSTART:20260101T090000Z
DTEND:20260101T100000Z
RRULE:FREQ=DAILY;COUNT=30
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := inputs[0].RecurrenceRules[0]
	if rule.Frequency != eventkit.FrequencyDaily {
		t.Errorf("frequency: got %v", rule.Frequency)
	}
	if rule.End == nil || rule.End.OccurrenceCount != 30 {
		t.Errorf("expected COUNT=30, got %+v", rule.End)
	}
}

func TestICS_ParseRRuleWithUntil(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Monthly
DTSTART:20260101T090000Z
DTEND:20260101T100000Z
RRULE:FREQ=MONTHLY;BYMONTHDAY=1,15;UNTIL=20261231T235959Z
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := inputs[0].RecurrenceRules[0]
	if rule.Frequency != eventkit.FrequencyMonthly {
		t.Errorf("frequency: got %v", rule.Frequency)
	}
	if len(rule.DaysOfTheMonth) != 2 || rule.DaysOfTheMonth[0] != 1 || rule.DaysOfTheMonth[1] != 15 {
		t.Errorf("days of month: got %v", rule.DaysOfTheMonth)
	}
	if rule.End == nil || rule.End.EndDate == nil {
		t.Fatal("expected UNTIL end date")
	}
	expected := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	if !rule.End.EndDate.Equal(expected) {
		t.Errorf("until: got %v, want %v", rule.End.EndDate, expected)
	}
}

func TestICS_ParseRRuleWithWeekNumber(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Second Tuesday
DTSTART:20260101T090000Z
DTEND:20260101T100000Z
RRULE:FREQ=MONTHLY;BYDAY=2TU
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rule := inputs[0].RecurrenceRules[0]
	if len(rule.DaysOfTheWeek) != 1 {
		t.Fatalf("expected 1 day, got %d", len(rule.DaysOfTheWeek))
	}
	dow := rule.DaysOfTheWeek[0]
	if dow.DayOfTheWeek != eventkit.Tuesday {
		t.Errorf("day: got %v", dow.DayOfTheWeek)
	}
	if dow.WeekNumber != 2 {
		t.Errorf("week number: got %d", dow.WeekNumber)
	}
}

func TestICS_ParseNoDTEND(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:No End Time
DTSTART:20260101T100000Z
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := inputs[0]
	// Should default to start + 1 hour
	expected := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
	if !e.EndDate.Equal(expected) {
		t.Errorf("end: got %v, want %v", e.EndDate, expected)
	}
}

func TestICS_ParseNoDTENDAllDay(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:All Day No End
DTSTART;VALUE=DATE:20260315
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := inputs[0]
	// Should default to start + 1 day
	expected := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	if !e.EndDate.Equal(expected) {
		t.Errorf("end: got %v, want %v", e.EndDate, expected)
	}
}

func TestICS_ParseEmpty(t *testing.T) {
	ics := `BEGIN:VCALENDAR
VERSION:2.0
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 0 {
		t.Errorf("expected 0 events, got %d", len(inputs))
	}
}

func TestICS_ParseMissingSummary(t *testing.T) {
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
DTSTART:20260101T100000Z
DTEND:20260101T110000Z
END:VEVENT
END:VCALENDAR
`
	_, err := ParseICS(strings.NewReader(ics))
	if err == nil {
		t.Error("expected error for missing SUMMARY")
	}
}

func TestICS_Roundtrip(t *testing.T) {
	events := []calendar.Event{
		{
			ID:        "event-rt-1",
			Title:     "Team Standup",
			StartDate: time.Date(2026, 2, 11, 9, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 2, 11, 9, 30, 0, 0, time.UTC),
			AllDay:    false,
			Location:  "Room A",
			Notes:     "Daily sync",
			URL:       "https://meet.example.com",
			Alerts: []calendar.Alert{
				{RelativeOffset: -15 * time.Minute},
			},
		},
		{
			ID:        "event-rt-2",
			Title:     "Company Holiday",
			StartDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
			AllDay:    true,
		},
	}

	// Export
	var buf bytes.Buffer
	err := ICS(events, &buf)
	if err != nil {
		t.Fatalf("export error: %v", err)
	}

	// Import
	inputs, err := ParseICS(&buf)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(inputs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(inputs))
	}

	// Verify first event
	if inputs[0].Title != "Team Standup" {
		t.Errorf("title: got %q", inputs[0].Title)
	}
	if inputs[0].Location != "Room A" {
		t.Errorf("location: got %q", inputs[0].Location)
	}
	if inputs[0].Notes != "Daily sync" {
		t.Errorf("notes: got %q", inputs[0].Notes)
	}
	if inputs[0].URL != "https://meet.example.com" {
		t.Errorf("url: got %q", inputs[0].URL)
	}
	if !inputs[0].StartDate.Equal(events[0].StartDate) {
		t.Errorf("start mismatch: got %v", inputs[0].StartDate)
	}
	if !inputs[0].EndDate.Equal(events[0].EndDate) {
		t.Errorf("end mismatch: got %v", inputs[0].EndDate)
	}
	if inputs[0].AllDay {
		t.Error("event 1 should not be all-day")
	}
	if len(inputs[0].Alerts) != 1 || inputs[0].Alerts[0].RelativeOffset != -15*time.Minute {
		t.Errorf("alerts mismatch: got %v", inputs[0].Alerts)
	}

	// Verify second event (all-day)
	if inputs[1].Title != "Company Holiday" {
		t.Errorf("title: got %q", inputs[1].Title)
	}
	if !inputs[1].AllDay {
		t.Error("event 2 should be all-day")
	}
}

func TestICS_RoundtripSpecialChars(t *testing.T) {
	events := []calendar.Event{
		{
			ID:        "event-sc",
			Title:     "Meeting; with, special\\chars",
			Notes:     "Line 1\nLine 2\nLine 3",
			StartDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	if err := ICS(events, &buf); err != nil {
		t.Fatalf("export error: %v", err)
	}

	inputs, err := ParseICS(&buf)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if inputs[0].Title != events[0].Title {
		t.Errorf("title roundtrip: got %q, want %q", inputs[0].Title, events[0].Title)
	}
	if inputs[0].Notes != events[0].Notes {
		t.Errorf("notes roundtrip: got %q, want %q", inputs[0].Notes, events[0].Notes)
	}
}

func TestICS_RoundtripRRule(t *testing.T) {
	events := []calendar.Event{
		{
			ID:        "event-rr",
			Title:     "Biweekly",
			StartDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC),
			RecurrenceRules: []eventkit.RecurrenceRule{
				{
					Frequency: eventkit.FrequencyWeekly,
					Interval:  2,
					DaysOfTheWeek: []eventkit.RecurrenceDayOfWeek{
						{DayOfTheWeek: eventkit.Monday},
						{DayOfTheWeek: eventkit.Friday},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := ICS(events, &buf); err != nil {
		t.Fatalf("export error: %v", err)
	}

	inputs, err := ParseICS(&buf)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(inputs[0].RecurrenceRules) != 1 {
		t.Fatalf("expected 1 rrule, got %d", len(inputs[0].RecurrenceRules))
	}
	rule := inputs[0].RecurrenceRules[0]
	if rule.Frequency != eventkit.FrequencyWeekly {
		t.Errorf("frequency: got %v", rule.Frequency)
	}
	if rule.Interval != 2 {
		t.Errorf("interval: got %d", rule.Interval)
	}
	if len(rule.DaysOfTheWeek) != 2 {
		t.Fatalf("expected 2 days, got %d", len(rule.DaysOfTheWeek))
	}
	if rule.DaysOfTheWeek[0].DayOfTheWeek != eventkit.Monday {
		t.Errorf("day 0: got %v", rule.DaysOfTheWeek[0].DayOfTheWeek)
	}
	if rule.DaysOfTheWeek[1].DayOfTheWeek != eventkit.Friday {
		t.Errorf("day 1: got %v", rule.DaysOfTheWeek[1].DayOfTheWeek)
	}
}

func TestParseTrigger(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"-PT15M", -15 * time.Minute},
		{"-PT1H", -1 * time.Hour},
		{"-PT1H30M", -(1*time.Hour + 30*time.Minute)},
		{"-P1D", -24 * time.Hour},
		{"-P1DT2H", -(26 * time.Hour)},
		{"-P1W", -7 * 24 * time.Hour},
		{"-PT0M", 0},
		{"-PT30S", -30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTrigger(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRRule(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, r eventkit.RecurrenceRule)
	}{
		{
			name:  "daily",
			input: "FREQ=DAILY",
			check: func(t *testing.T, r eventkit.RecurrenceRule) {
				if r.Frequency != eventkit.FrequencyDaily {
					t.Errorf("got %v", r.Frequency)
				}
				if r.Interval != 1 {
					t.Errorf("interval: got %d", r.Interval)
				}
			},
		},
		{
			name:  "yearly",
			input: "FREQ=YEARLY;INTERVAL=1",
			check: func(t *testing.T, r eventkit.RecurrenceRule) {
				if r.Frequency != eventkit.FrequencyYearly {
					t.Errorf("got %v", r.Frequency)
				}
			},
		},
		{
			name:  "weekly with negative week number",
			input: "FREQ=MONTHLY;BYDAY=-1FR",
			check: func(t *testing.T, r eventkit.RecurrenceRule) {
				if len(r.DaysOfTheWeek) != 1 {
					t.Fatalf("expected 1 day, got %d", len(r.DaysOfTheWeek))
				}
				if r.DaysOfTheWeek[0].WeekNumber != -1 {
					t.Errorf("week number: got %d", r.DaysOfTheWeek[0].WeekNumber)
				}
				if r.DaysOfTheWeek[0].DayOfTheWeek != eventkit.Friday {
					t.Errorf("day: got %v", r.DaysOfTheWeek[0].DayOfTheWeek)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := parseRRule(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, rule)
		})
	}
}

func TestParseRRule_InvalidFreq(t *testing.T) {
	_, err := parseRRule("FREQ=SECONDLY")
	if err == nil {
		t.Error("expected error for unknown FREQ")
	}
}

func TestParseRRule_MissingFreq(t *testing.T) {
	_, err := parseRRule("INTERVAL=2;BYDAY=MO")
	if err == nil {
		t.Error("expected error for missing FREQ")
	}
}

func TestICS_ParseCRLF(t *testing.T) {
	// Real-world ICS files use CRLF line endings
	ics := "BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\nSUMMARY:CRLF Test\r\nDTSTART:20260101T100000Z\r\nDTEND:20260101T110000Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(inputs))
	}
	if inputs[0].Title != "CRLF Test" {
		t.Errorf("title: got %q", inputs[0].Title)
	}
}

func TestICS_ParseLocalDateTime(t *testing.T) {
	// Some ICS files use local time without Z suffix
	ics := `BEGIN:VCALENDAR
BEGIN:VEVENT
SUMMARY:Local Time
DTSTART:20260101T100000
DTEND:20260101T110000
END:VEVENT
END:VCALENDAR
`
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inputs[0].StartDate.Hour() != 10 {
		t.Errorf("start hour: got %d", inputs[0].StartDate.Hour())
	}
}
