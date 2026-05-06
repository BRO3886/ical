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

func TestFilterRecurring(t *testing.T) {
	t.Run("drops only recurring entries", func(t *testing.T) {
		events := []calendar.Event{
			{Title: "A", Recurring: false},
			{Title: "B", Recurring: true},
			{Title: "C", Recurring: false},
			{Title: "D", Recurring: true},
		}
		got := filterRecurring(events)
		if len(got) != 2 || got[0].Title != "A" || got[1].Title != "C" {
			t.Errorf("wrong events retained, got %+v", got)
		}
	})

	t.Run("nil input returns nil without panicking", func(t *testing.T) {
		if got := filterRecurring(nil); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("empty slice returns same empty slice", func(t *testing.T) {
		in := []calendar.Event{}
		got := filterRecurring(in)
		if len(got) != 0 {
			t.Errorf("expected empty, got %+v", got)
		}
	})

	t.Run("all-recurring input yields empty result", func(t *testing.T) {
		in := []calendar.Event{{Title: "X", Recurring: true}, {Title: "Y", Recurring: true}}
		if got := filterRecurring(in); len(got) != 0 {
			t.Errorf("expected empty, got %+v", got)
		}
	})

	t.Run("no-recurring input passes through as the same slice", func(t *testing.T) {
		in := []calendar.Event{{Title: "Y", Recurring: false}, {Title: "Z", Recurring: false}}
		got := filterRecurring(in)
		if len(got) != len(in) {
			t.Fatalf("expected pass-through of %d events, got %d", len(in), len(got))
		}
		// Fast path: should return the same backing slice, not a copy.
		if &got[0] != &in[0] {
			t.Error("expected pass-through to return the input slice, got a copy")
		}
	})
}

func TestFilterIncludedCalendars(t *testing.T) {
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
		include []string
		want    []string
	}{
		{"nil include keeps everything", nil, []string{"A", "B", "C", "D"}},
		{"empty slice keeps everything", []string{}, []string{"A", "B", "C", "D"}},
		{"single exact match", []string{"Work"}, []string{"A"}},
		{"case insensitive match", []string{"personal"}, []string{"B"}},
		{"padded flag matches padded calendar name", []string{"Holidays"}, []string{"D"}},
		{"multiple includes", []string{"Work", "Personal"}, []string{"A", "B"}},
		{"unknown include yields empty result", []string{"Nonexistent"}, []string{}},
		{"whitespace-only include is a no-op (keeps everything)", []string{"   "}, []string{"A", "B", "C", "D"}},
		{"whitespace-only mixed with valid include applies only the valid one", []string{"   ", "Work"}, []string{"A"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := titles(filterIncludedCalendars(events, tt.include))
			if !eq(got, tt.want) {
				t.Errorf("titles = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeCalendarNames(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"nil returns nil", nil, nil},
		{"empty returns nil", []string{}, nil},
		{"single name", []string{"Work"}, []string{"work"}},
		{"multiple names", []string{"Work", "Personal"}, []string{"work", "personal"}},
		{"trims whitespace", []string{"  Work  ", " Personal"}, []string{"work", "personal"}},
		{"drops whitespace-only entries", []string{"   ", "Work"}, []string{"work"}},
		{"all whitespace-only returns nil", []string{"   ", "  "}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCalendarNames(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
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
