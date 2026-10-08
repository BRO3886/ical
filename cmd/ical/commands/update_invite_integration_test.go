//go:build darwin && integration

package commands

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestUpdateInviteEventKit(t *testing.T) {
	source := os.Getenv("ICAL_TEST_CALENDAR_SOURCE")
	invite := os.Getenv("ICAL_TEST_INVITE_EMAIL")
	if source == "" || invite == "" {
		t.Skip("set ICAL_TEST_CALENDAR_SOURCE and ICAL_TEST_INVITE_EMAIL; this test sends an invitation")
	}
	if _, err := parseAttendees([]string{invite}); err != nil {
		t.Fatal(err)
	}
	client, err := calendar.New()
	if err != nil {
		t.Fatal(err)
	}
	if !client.AttendeeWritesSupported() {
		t.Fatal("attendee writes are unavailable on this macOS")
	}
	name := fmt.Sprintf("ical-issue-60-%d", time.Now().UnixNano())
	cal, err := client.CreateCalendar(calendar.CreateCalendarInput{Title: name, Source: source})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.DeleteCalendar(cal.ID); err != nil {
			t.Errorf("delete temporary calendar: %v", err)
		}
		updateInvite = nil
		updateID = ""
		updateTitle = ""
		for _, name := range []string{"invite", "id", "title"} {
			updateCmd.Flags().Lookup(name).Changed = false
		}
	})
	start := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	event, err := client.CreateEvent(calendar.CreateEventInput{
		Title: "Invite integration test", Calendar: name,
		StartDate: start, EndDate: start.Add(time.Hour),
		SuppressDefaultAlarms: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.ParseFlags([]string{"--id", event.ID, "--invite", "Test guest <" + invite + ">"}); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.RunE(updateCmd, nil); err != nil {
		t.Fatal(err)
	}
	updated, err := client.Event(event.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, attendee := range updated.Attendees {
		if strings.EqualFold(attendee.Email, invite) {
			found = true
		}
	}
	if !found {
		t.Error("saved event does not contain the invited guest")
	}
	if updated.ID != event.ID || updated.Title != event.Title || !updated.StartDate.Equal(start) || !updated.EndDate.Equal(start.Add(time.Hour)) {
		t.Error("invitation changed the event ID, title, or time")
	}
	updateInvite = nil
	updateCmd.Flags().Lookup("invite").Changed = false
	if err := updateCmd.ParseFlags([]string{"--id", event.ID, "--title", "Updated integration test"}); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.RunE(updateCmd, nil); err != nil {
		t.Fatal(err)
	}
	after, err := client.Event(event.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Title != "Updated integration test" {
		t.Error("title update was not saved")
	}
	for _, before := range updated.Attendees {
		preserved := false
		for _, attendee := range after.Attendees {
			if attendee.Email == before.Email && attendee.Status == before.Status {
				preserved = true
			}
		}
		if !preserved {
			t.Error("an existing attendee or RSVP changed during a title update")
		}
	}
}
