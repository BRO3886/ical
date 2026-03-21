package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/dateparser"
)

// Fixed reference time for deterministic tests: Wednesday, Feb 11, 2026 10:30:00 IST
func refTime(loc *time.Location) time.Time {
	return time.Date(2026, 2, 11, 10, 30, 0, 0, loc)
}

func mustLoadLoc(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("failed to load location %q: %v", name, err)
	}
	return loc
}

// TestDateparserKeywords verifies the shared dateparser package handles all
// keyword inputs that ical relies on (previously covered by internal/parser tests).
func TestDateparserKeywords(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"today", "today", time.Date(2026, 2, 11, 0, 0, 0, 0, ist)},
		{"tomorrow", "tomorrow", time.Date(2026, 2, 12, 0, 0, 0, 0, ist)},
		{"yesterday", "yesterday", time.Date(2026, 2, 10, 0, 0, 0, 0, ist)},
		{"now", "now", now},
		{"eod", "eod", time.Date(2026, 2, 11, 17, 0, 0, 0, ist)},
		{"end of day", "end of day", time.Date(2026, 2, 11, 17, 0, 0, 0, ist)},
		{"next week", "next week", time.Date(2026, 2, 16, 0, 0, 0, 0, ist)},
		{"next month", "next month", time.Date(2026, 3, 1, 0, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserEOW verifies end-of-week behavior.
func TestDateparserEOW(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			"wednesday to friday",
			time.Date(2026, 2, 11, 10, 0, 0, 0, ist),
			time.Date(2026, 2, 13, 17, 0, 0, 0, ist),
		},
		{
			"friday returns same friday",
			time.Date(2026, 2, 13, 10, 0, 0, 0, ist),
			time.Date(2026, 2, 13, 17, 0, 0, 0, ist),
		},
		{
			"saturday to next friday",
			time.Date(2026, 2, 14, 10, 0, 0, 0, ist),
			time.Date(2026, 2, 20, 17, 0, 0, 0, ist),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo("eow", tt.now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserRelative verifies "in X units" and "X units ago" patterns.
func TestDateparserRelative(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"in 3 hours", "in 3 hours", now.Add(3 * time.Hour)},
		{"in 30 minutes", "in 30 minutes", now.Add(30 * time.Minute)},
		{"in 5 days", "in 5 days", now.AddDate(0, 0, 5)},
		{"in 2 weeks", "in 2 weeks", now.AddDate(0, 0, 14)},
		{"in 1 month", "in 1 month", now.AddDate(0, 1, 0)},
		{"2 hours ago", "2 hours ago", now.Add(-2 * time.Hour)},
		{"5 days ago", "5 days ago", now.AddDate(0, 0, -5)},
		{"1 month ago", "1 month ago", now.AddDate(0, -1, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserWeekdays verifies weekday-based parsing.
func TestDateparserWeekdays(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist) // Wednesday

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"next monday", "next monday", time.Date(2026, 2, 16, 0, 0, 0, 0, ist)},
		{"next friday", "next friday", time.Date(2026, 2, 13, 0, 0, 0, 0, ist)},
		{"next friday at 2pm", "next friday at 2pm", time.Date(2026, 2, 13, 14, 0, 0, 0, ist)},
		{"friday 2pm", "friday 2pm", time.Date(2026, 2, 13, 14, 0, 0, 0, ist)},
		{"monday", "monday", time.Date(2026, 2, 16, 0, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserMonthDay verifies month-day parsing (both orderings).
func TestDateparserMonthDay(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"mar 15", "mar 15", time.Date(2026, 3, 15, 0, 0, 0, 0, ist)},
		{"march 15", "march 15", time.Date(2026, 3, 15, 0, 0, 0, 0, ist)},
		{"dec 25", "dec 25", time.Date(2026, 12, 25, 0, 0, 0, 0, ist)},
		{"21 mar", "21 mar", time.Date(2026, 3, 21, 0, 0, 0, 0, ist)},
		{"21 march 2pm", "21 march 2pm", time.Date(2026, 3, 21, 14, 0, 0, 0, ist)},
		{"21 march 2026", "21 march 2026", time.Date(2026, 3, 21, 0, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserTimeOnly verifies standalone time parsing.
func TestDateparserTimeOnly(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"5pm", "5pm", time.Date(2026, 2, 11, 17, 0, 0, 0, ist)},
		{"17:00", "17:00", time.Date(2026, 2, 11, 17, 0, 0, 0, ist)},
		{"3:30pm", "3:30pm", time.Date(2026, 2, 11, 15, 30, 0, 0, ist)},
		{"9am", "9am", time.Date(2026, 2, 11, 9, 0, 0, 0, ist)},
		{"12pm", "12pm", time.Date(2026, 2, 11, 12, 0, 0, 0, ist)},
		{"12am", "12am", time.Date(2026, 2, 11, 0, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserStandardFormats verifies ISO 8601 and other standard formats.
func TestDateparserStandardFormats(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"ISO date", "2026-03-15", time.Date(2026, 3, 15, 0, 0, 0, 0, ist)},
		{"ISO datetime", "2026-03-15 14:00", time.Date(2026, 3, 15, 14, 0, 0, 0, ist)},
		{"ISO datetime T", "2026-03-15T14:00", time.Date(2026, 3, 15, 14, 0, 0, 0, ist)},
		{"US date", "01/15/2026", time.Date(2026, 1, 15, 0, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserDateWithTime verifies "today/tomorrow at Xpm" patterns.
func TestDateparserDateWithTime(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")
	now := refTime(ist)

	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"today at 5pm", "today at 5pm", time.Date(2026, 2, 11, 17, 0, 0, 0, ist)},
		{"tomorrow at 3:30pm", "tomorrow at 3:30pm", time.Date(2026, 2, 12, 15, 30, 0, 0, ist)},
		{"today 5pm", "today 5pm", time.Date(2026, 2, 11, 17, 0, 0, 0, ist)},
		{"tomorrow 3pm", "tomorrow 3pm", time.Date(2026, 2, 12, 15, 0, 0, 0, ist)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserErrors verifies proper error handling.
func TestDateparserErrors(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"gibberish", "not a date at all"},
		{"timezone abbreviation", "2026-06-17 2pm CDT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dateparser.ParseDateRelativeTo(tt.input, now)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}

// TestDateparserThisWeek verifies "this week" resolves to Sunday 23:59.
func TestDateparserThisWeek(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			"wednesday",
			time.Date(2026, 2, 11, 10, 0, 0, 0, ist),
			time.Date(2026, 2, 15, 23, 59, 0, 0, ist),
		},
		{
			"sunday returns same sunday",
			time.Date(2026, 2, 15, 10, 0, 0, 0, ist),
			time.Date(2026, 2, 15, 23, 59, 0, 0, ist),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseDateRelativeTo("this week", tt.now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatDuration verifies the shared FormatDuration function.
func TestFormatDuration(t *testing.T) {
	base := time.Date(2026, 3, 15, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name   string
		start  time.Time
		end    time.Time
		allDay bool
		want   string
	}{
		{"1 hour", base, base.Add(time.Hour), false, "1h"},
		{"30 minutes", base, base.Add(30 * time.Minute), false, "30m"},
		{"1h 30m", base, base.Add(90 * time.Minute), false, "1h 30m"},
		{"all day single", base, base.Add(24 * time.Hour), true, "All Day"},
		{"all day multi", base, base.AddDate(0, 0, 3), true, "3 days"},
		{"0 minutes", base, base.Add(30 * time.Second), false, "0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dateparser.FormatDuration(tt.start, tt.end, tt.allDay)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFormatTimeRange verifies the shared FormatTimeRange function.
func TestFormatTimeRange(t *testing.T) {
	base := time.Date(2026, 3, 15, 10, 0, 0, 0, time.Local)

	tests := []struct {
		name   string
		start  time.Time
		end    time.Time
		allDay bool
		want   string
	}{
		{"same day", base, base.Add(2 * time.Hour), false, "10:00 - 12:00"},
		{"all day single", base, base.Add(24 * time.Hour), true, "All Day"},
		{"all day multi", base, base.AddDate(0, 0, 3), true, "Mar 15 - Mar 18"},
		{"cross day", base, base.AddDate(0, 0, 1).Add(2 * time.Hour), false, "Mar 15 10:00 - Mar 16 12:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dateparser.FormatTimeRange(tt.start, tt.end, tt.allDay)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestParseAlertDuration verifies alert duration parsing via the shared package.
func TestParseAlertDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"15m", "15m", 15 * time.Minute, false},
		{"1h", "1h", time.Hour, false},
		{"1d", "1d", 24 * time.Hour, false},
		{"30min", "30min", 30 * time.Minute, false},
		{"2hours", "2hours", 2 * time.Hour, false},
		{"empty", "", 0, true},
		{"invalid", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dateparser.ParseAlertDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEndOfDayIfMidnight verifies the --to date bumping logic.
func TestEndOfDayIfMidnight(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			"midnight gets bumped",
			time.Date(2026, 2, 12, 0, 0, 0, 0, time.Local),
			time.Date(2026, 2, 12, 23, 59, 59, 0, time.Local),
		},
		{
			"non-midnight unchanged",
			time.Date(2026, 2, 12, 14, 30, 0, 0, time.Local),
			time.Date(2026, 2, 12, 14, 30, 0, 0, time.Local),
		},
		{
			"1am unchanged",
			time.Date(2026, 2, 12, 1, 0, 0, 0, time.Local),
			time.Date(2026, 2, 12, 1, 0, 0, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := endOfDayIfMidnight(tt.in)
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateparserWithOptions verifies that go-eventkit v0.4.0 options work.
func TestDateparserWithOptions(t *testing.T) {
	ist := mustLoadLoc(t, "Asia/Kolkata")

	t.Run("WithDefaultHour sets default for bare dates", func(t *testing.T) {
		now := time.Date(2026, 2, 11, 10, 0, 0, 0, ist)
		got, err := dateparser.ParseDateRelativeTo("mar 15", now, dateparser.WithDefaultHour(9))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 3, 15, 9, 0, 0, 0, ist)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("WithEOWSkipToday on Friday skips to next Friday", func(t *testing.T) {
		friday := time.Date(2026, 2, 13, 10, 0, 0, 0, ist) // Friday
		got, err := dateparser.ParseDateRelativeTo("eow", friday, dateparser.WithEOWSkipToday())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 2, 20, 17, 0, 0, 0, ist) // next Friday 5pm
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
