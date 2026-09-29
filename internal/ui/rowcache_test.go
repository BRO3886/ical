package ui

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestRowCacheSessionKey(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"none", nil, ""},
		{"tmux pane", map[string]string{"TMUX_PANE": "%3"}, "TMUX_PANE=%3"},
		{"terminal session", map[string]string{"TERM_SESSION_ID": "w0t1p0:ABC"}, "TERM_SESSION_ID=w0t1p0:ABC"},
		{"claude session beats tmux pane", map[string]string{"TMUX_PANE": "%3", "CLAUDE_CODE_SESSION_ID": "s1"}, "CLAUDE_CODE_SESSION_ID=s1"},
		{"explicit override beats everything", map[string]string{"ICAL_SESSION": "mine", "CLAUDE_CODE_SESSION_ID": "s1"}, "ICAL_SESSION=mine"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rowCacheSessionKey(func(k string) string { return tt.env[k] })
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRowCachePath(t *testing.T) {
	home := "/home/u"
	if got, want := rowCachePath(home, ""), filepath.Join(home, ".ical-last-list"); got != want {
		t.Errorf("no session: got %q, want legacy %q", got, want)
	}
	a := rowCachePath(home, "TMUX_PANE=%1")
	b := rowCachePath(home, "TMUX_PANE=%2")
	if a == b {
		t.Errorf("different sessions share a cache file: %q", a)
	}
	if a != rowCachePath(home, "TMUX_PANE=%1") {
		t.Error("session cache path is not stable")
	}
	if filepath.Dir(a) != filepath.Join(home, ".cache", "ical", "rows") {
		t.Errorf("unexpected cache dir %q", filepath.Dir(a))
	}
}

// Two sessions listing different events must each resolve row 1 to their own.
func TestRowNumbersAreScopedPerSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, k := range rowCacheEnvKeys {
		t.Setenv(k, "")
	}

	t.Setenv("TMUX_PANE", "%1")
	SaveLastList([]calendar.Event{{ID: "pane-one"}})
	t.Setenv("TMUX_PANE", "%2")
	SaveLastList([]calendar.Event{{ID: "pane-two"}})

	if got, _ := LookupRow(1); got.ID != "pane-two" {
		t.Errorf("pane 2 row 1 = %q, want pane-two", got.ID)
	}
	t.Setenv("TMUX_PANE", "%1")
	if got, _ := LookupRow(1); got.ID != "pane-one" {
		t.Errorf("pane 1 row 1 = %q, want pane-one", got.ID)
	}
}

// Every occurrence of a series shares one ID, so a row must also remember
// which occurrence it showed.
func TestRowRemembersOccurrence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ICAL_SESSION", "occ-test")

	occ := time.Date(2026, 12, 16, 14, 0, 0, 0, time.UTC)
	SaveLastList([]calendar.Event{
		{ID: "single"},
		{ID: "series", Recurring: true, OccurrenceDate: &occ},
	})

	one, ok := LookupRow(1)
	if !ok || one.ID != "single" || one.Occurrence != nil {
		t.Errorf("row 1 = %+v, %v", one, ok)
	}
	two, ok := LookupRow(2)
	if !ok || two.ID != "series" || two.Occurrence == nil || !two.Occurrence.Equal(occ) {
		t.Errorf("row 2 = %+v, %v", two, ok)
	}
	if _, ok := LookupRow(3); ok {
		t.Error("row 3 should not exist")
	}
}

// Rows written by older versions (bare IDs) still resolve.
func TestParseRowLegacyLine(t *testing.T) {
	r := parseRow("ABC:DEF")
	if r.ID != "ABC:DEF" || r.Occurrence != nil {
		t.Errorf("got %+v", r)
	}
}
