package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

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
