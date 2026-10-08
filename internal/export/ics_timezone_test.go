package export

import (
	"strings"
	"testing"
	"time"
)

func TestICSImportTZID(t *testing.T) {
	ics := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nSUMMARY:TZ repro\nDTSTART;TZID=America/New_York:20260819T153500\nDTEND;TZID=America/New_York:20260819T164000\nEND:VEVENT\nEND:VCALENDAR\n"
	inputs, err := ParseICS(strings.NewReader(ics))
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 1 {
		t.Fatalf("got %d events, want 1", len(inputs))
	}
	input := inputs[0]
	if got := input.StartDate.UTC().Format("2006-01-02T15:04:05Z"); got != "2026-08-19T19:35:00Z" {
		t.Errorf("start = %s, want 2026-08-19T19:35:00Z", got)
	}
	if got := input.EndDate.UTC().Format("2006-01-02T15:04:05Z"); got != "2026-08-19T20:40:00Z" {
		t.Errorf("end = %s, want 2026-08-19T20:40:00Z", got)
	}
	if input.TimeZone != "America/New_York" {
		t.Errorf("timezone = %q, want America/New_York", input.TimeZone)
	}
}

func TestICSImportDateForms(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	old := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = old })
	for _, tc := range []struct{ name, property, want string }{
		{"winter", "DTSTART;TZID=America/New_York:20260119T153500", "2026-01-19T20:35:00Z"},
		{"quoted zone", "DTSTART;TZID=\"America/New_York\":20260819T153500", "2026-08-19T19:35:00Z"},
		{"UTC", "DTSTART:20260819T153500Z", "2026-08-19T15:35:00Z"},
		{"floating", "DTSTART:20260819T153500", "2026-08-19T19:35:00Z"},
		{"fall overlap", "DTSTART;TZID=America/New_York:20071104T013000", "2007-11-04T05:30:00Z"},
		{"spring gap", "DTSTART;TZID=America/New_York:20070311T023000", "2007-03-11T07:30:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inputs, err := ParseICS(strings.NewReader("BEGIN:VCALENDAR\nBEGIN:VEVENT\nSUMMARY:Date form\n" + tc.property + "\nEND:VEVENT\nEND:VCALENDAR\n"))
			if err != nil {
				t.Fatal(err)
			}
			if got := inputs[0].StartDate.UTC().Format(time.RFC3339); got != tc.want {
				t.Errorf("start = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestICSImportUnknownTZID(t *testing.T) {
	for _, property := range []string{"DTSTART", "DTEND"} {
		t.Run(property, func(t *testing.T) {
			ics := "BEGIN:VCALENDAR\nBEGIN:VEVENT\nSUMMARY:Invalid zone\nDTSTART:20260819T153500Z\n" + property + ";TZID=Unknown/Zone:20260819T164000\nEND:VEVENT\nEND:VCALENDAR\n"
			inputs, err := ParseICS(strings.NewReader(ics))
			if err == nil || !strings.Contains(err.Error(), "unsupported TZID") {
				t.Fatalf("inputs = %v, error = %v, want unsupported TZID", inputs, err)
			}
			if inputs != nil {
				t.Fatal("invalid file returned usable events")
			}
		})
	}
}
