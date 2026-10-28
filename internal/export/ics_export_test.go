package export

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
)

func TestICSExportAllDayUsesLocalDate(t *testing.T) {
	madrid, _ := time.LoadLocation("Europe/Madrid")
	start := time.Date(2030, 1, 16, 0, 0, 0, 0, madrid).UTC()
	e := calendar.Event{ID: "a", Title: "x", AllDay: true, TimeZone: "Europe/Madrid", StartDate: start, EndDate: start.AddDate(0, 0, 1)}
	var b bytes.Buffer
	if err := ICS([]calendar.Event{e}, &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "DTSTART;VALUE=DATE:20300116\n") || !strings.Contains(b.String(), "DTEND;VALUE=DATE:20300117\n") {
		t.Errorf("wrong all-day dates:\n%s", b.String())
	}
}

func TestICSExportRecurringOncePerSeries(t *testing.T) {
	rule := eventkit.RecurrenceRule{Frequency: eventkit.FrequencyWeekly, Interval: 1, End: &eventkit.RecurrenceEnd{OccurrenceCount: 3}}
	d := time.Date(2030, 1, 14, 9, 0, 0, 0, time.UTC)
	var evs []calendar.Event
	for i := 0; i < 3; i++ {
		s := d.AddDate(0, 0, 7*i)
		evs = append(evs, calendar.Event{ID: "s", Title: "w", Recurring: true, RecurrenceRules: []eventkit.RecurrenceRule{rule}, StartDate: s, EndDate: s.Add(time.Hour), OccurrenceDate: &s})
	}
	// Move the second occurrence: detached, id carries /RID=.
	moved := evs[1].StartDate.Add(2 * time.Hour)
	occ := evs[1].StartDate
	evs[1] = calendar.Event{ID: "s/RID=1", Title: "moved", Recurring: true, IsDetached: true, StartDate: moved, EndDate: moved.Add(time.Hour), OccurrenceDate: &occ}

	var b bytes.Buffer
	if err := ICS(evs, &b); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if n := strings.Count(out, "BEGIN:VEVENT"); n != 2 {
		t.Errorf("VEVENT count = %d, want 2\n%s", n, out)
	}
	if n := strings.Count(out, "RRULE:"); n != 1 {
		t.Errorf("RRULE count = %d, want 1", n)
	}
	if strings.Contains(out, "/RID=") || strings.Count(out, "UID:s\n") != 2 {
		t.Errorf("UID not normalized:\n%s", out)
	}
	if !strings.Contains(out, "RECURRENCE-ID:20300121T090000Z\n") {
		t.Errorf("missing RECURRENCE-ID:\n%s", out)
	}
}
