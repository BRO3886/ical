//go:build darwin && integration

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestImportEventKit(t *testing.T) {
	source := os.Getenv("ICAL_TEST_CALENDAR_SOURCE")
	if source == "" {
		t.Skip("set ICAL_TEST_CALENDAR_SOURCE to create a temporary test calendar")
	}
	client, err := calendar.New()
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("ical-issue-57-%d", time.Now().UnixNano())
	cal, err := client.CreateCalendar(calendar.CreateCalendarInput{Title: name, Source: source})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.DeleteCalendar(cal.ID); err != nil {
			t.Errorf("delete temporary calendar: %v", err)
		}
		importCalendar = ""
		importForce = false
		importNoAlert = false
		rootCmd.SetArgs(nil)
	})
	path := filepath.Join(t.TempDir(), "event.ics")
	ics := "BEGIN:VCALENDAR\nVERSION:2.0\nBEGIN:VEVENT\nUID:issue-57-test\nSUMMARY:Timezone integration test\nDTSTART;TZID=America/New_York:20260819T153500\nDTEND;TZID=America/New_York:20260819T164000\nBEGIN:VALARM\nACTION:DISPLAY\nTRIGGER:-PT30M\nEND:VALARM\nEND:VEVENT\nEND:VCALENDAR\n"
	if err := os.WriteFile(path, []byte(ics), 0600); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"import", path, "--calendar", name, "--force", "--no-alert"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)
	events, err := client.Events(start, start.Add(24*time.Hour), calendar.WithCalendarID(cal.ID))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d stored events, want 1", len(events))
	}
	event, err := client.Event(events[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := event.StartDate.UTC().Format(time.RFC3339); got != "2026-08-19T19:35:00Z" {
		t.Errorf("stored start = %s", got)
	}
	if got := event.EndDate.UTC().Format(time.RFC3339); got != "2026-08-19T20:40:00Z" {
		t.Errorf("stored end = %s", got)
	}
	if event.TimeZone != "America/New_York" {
		t.Errorf("stored zone = %q", event.TimeZone)
	}
	if len(event.Alerts) != 0 {
		t.Errorf("stored alerts = %d, want 0", len(event.Alerts))
	}
}
