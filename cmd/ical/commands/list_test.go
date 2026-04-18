package commands

import (
	"testing"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestNormalizeCalendarName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Work", "work"},
		{"already lowercase", "work", "work"},
		{"leading space", "  Work", "work"},
		{"trailing space", "Work   ", "work"},
		{"both sides padded", "  Work  ", "work"},
		{"tab padded", "\tWork\t", "work"},
		{"internal spaces preserved", "Holidays in India", "holidays in india"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCalendarName(tt.in)
			if got != tt.want {
				t.Errorf("normalizeCalendarName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFilterExcludedCalendars(t *testing.T) {
	events := []calendar.Event{
		{Title: "A", Calendar: "Work"},
		{Title: "B", Calendar: "Personal"},
		{Title: "C", Calendar: ""},
		{Title: "D", Calendar: "  Holidays  "},
	}

	titles := func(evs []calendar.Event) []string {
		out := make([]string, len(evs))
		for i, e := range evs {
			out[i] = e.Title
		}
		return out
	}

	eq := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	tests := []struct {
		name    string
		exclude []string
		want    []string
	}{
		{"nil exclude keeps everything", nil, []string{"A", "B", "C", "D"}},
		{"empty slice keeps everything", []string{}, []string{"A", "B", "C", "D"}},
		{"exact name match", []string{"Work"}, []string{"B", "C", "D"}},
		{"case insensitive match", []string{"PERSONAL"}, []string{"A", "C", "D"}},
		{"padded flag matches padded calendar name", []string{"Holidays"}, []string{"A", "B", "C"}},
		{"multiple excludes", []string{"Work", "Personal"}, []string{"C", "D"}},
		{"whitespace-only exclude does not drop empty-calendar events", []string{"   "}, []string{"A", "B", "C", "D"}},
		{"whitespace-only mixed with valid exclude applies only the valid one", []string{"   ", "Work"}, []string{"B", "C", "D"}},
		{"unknown exclude is a no-op", []string{"Nonexistent"}, []string{"A", "B", "C", "D"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := titles(filterExcludedCalendars(events, tt.exclude))
			if !eq(got, tt.want) {
				t.Errorf("titles = %v, want %v", got, tt.want)
			}
		})
	}
}
