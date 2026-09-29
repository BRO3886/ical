package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestOccurrenceTarget(t *testing.T) {
	occ := time.Date(2026, 12, 16, 14, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		event calendar.Event
		span  string
		want  *time.Time
	}{
		{"non-recurring", calendar.Event{OccurrenceDate: &occ}, "this", nil},
		{"recurring this", calendar.Event{Recurring: true, OccurrenceDate: &occ}, "this", &occ},
		{"recurring default span", calendar.Event{Recurring: true, OccurrenceDate: &occ}, "", &occ},
		{"recurring future", calendar.Event{Recurring: true, OccurrenceDate: &occ}, "future", &occ},
		{"detached occurrence", calendar.Event{IsDetached: true, OccurrenceDate: &occ}, "this", &occ},
		// "all" acts on the series from its first occurrence, which is what an
		// ID lookup resolves to.
		{"recurring all", calendar.Event{Recurring: true, OccurrenceDate: &occ}, "ALL", nil},
		{"recurring without occurrence date", calendar.Event{Recurring: true}, "this", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := occurrenceTarget(&tt.event, tt.span)
			switch {
			case tt.want == nil && got != nil:
				t.Errorf("got %v, want nil", *got)
			case tt.want != nil && (got == nil || !got.Equal(*tt.want)):
				t.Errorf("got %v, want %v", got, *tt.want)
			}
		})
	}
}

func TestSeriesID(t *testing.T) {
	for in, want := range map[string]string{
		"CAL:EVT":               "CAL:EVT",
		"CAL:EVT/RID=812984400": "CAL:EVT",
		"":                      "",
	} {
		if got := seriesID(in); got != want {
			t.Errorf("seriesID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFindOccurrence(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	at := func(d, h int) time.Time { return time.Date(2026, 12, d, h, 0, 0, 0, ny) }
	occ := func(id string, d, h int, title string) calendar.Event {
		o := at(d, h)
		return calendar.Event{ID: id, Title: title, Recurring: true, StartDate: o, OccurrenceDate: &o}
	}
	// A detached occurrence's ID is the series ID plus "/RID=<unix time>".
	moved := occ("S/RID=1796824800", 9, 9, "moved")
	moved.IsDetached = true
	moved.StartDate = at(10, 15) // moved; occurrence date stays Dec 9
	events := []calendar.Event{
		occ("S", 2, 9, "first"),
		moved,
		occ("S", 16, 9, "third"),
		occ("OTHER", 16, 9, "other series"),
		occ("T", 23, 9, "twice a day am"),
		occ("T", 23, 17, "twice a day pm"),
	}

	tests := []struct {
		name     string
		id       string
		at       time.Time
		wholeDay bool
		want     string
		wantErr  bool
	}{
		{"exact time", "S", at(16, 9), false, "third", false},
		{"exact time in another zone", "S", at(16, 9).UTC(), false, "third", false},
		{"whole day", "S", at(16, 0), true, "third", false},
		{"detached, moved occurrence by series ID and original date", "S", at(9, 9), false, "moved", false},
		{"sibling occurrence by a detached occurrence's ID", "S/RID=1796824800", at(16, 0), true, "third", false},
		{"no occurrence at that time", "S", at(16, 10), false, "", true},
		{"no occurrence that day", "S", at(17, 0), true, "", true},
		{"two occurrences on one day are ambiguous", "T", at(23, 0), true, "", true},
		{"exact time disambiguates", "T", at(23, 17), false, "twice a day pm", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findOccurrence(events, tt.id, tt.at, tt.wholeDay)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %q", got.Title)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Title != tt.want {
				t.Errorf("got %q, want %q", got.Title, tt.want)
			}
		})
	}
}
