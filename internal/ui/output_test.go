package ui

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit"
)

func TestShortID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdefgh-1234-5678", "abcdefgh-1234"},
		{"short", "short"},
		{"1234567890123", "1234567890123"},
		{"", ""},
		{"1234567", "1234567"},
		{"12345678901234", "1234567890123"},
		{"577B8983-DF44-4665-966E-58129A363B3A:20250212", "577B8983-DF44"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ShortID(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatRecurrenceRule(t *testing.T) {
	until := time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		rule eventkit.RecurrenceRule
		want string
	}{
		{
			"daily",
			eventkit.Daily(1),
			"Every day",
		},
		{
			"every 3 days",
			eventkit.Daily(3),
			"Every 3 days",
		},
		{
			"weekly",
			eventkit.Weekly(1),
			"Every week",
		},
		{
			"every 2 weeks on Mon, Wed",
			eventkit.Weekly(2, eventkit.Monday, eventkit.Wednesday),
			"Every 2 weeks on mon, wed",
		},
		{
			"monthly",
			eventkit.Monthly(1),
			"Every month",
		},
		{
			"monthly on 15th",
			eventkit.Monthly(1, 15),
			"Every month on the 15th",
		},
		{
			"every 2 months",
			eventkit.Monthly(2),
			"Every 2 months",
		},
		{
			"yearly",
			eventkit.Yearly(1),
			"Every year",
		},
		{
			"every 2 years",
			eventkit.Yearly(2),
			"Every 2 years",
		},
		{
			"daily until date",
			eventkit.Daily(1).Until(until),
			"Every day until Mar 15, 2027",
		},
		{
			"weekly for 10 occurrences",
			eventkit.Weekly(1).Count(10),
			"Every week for 10 occurrences",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRecurrenceRule(tt.rule)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocalizeTime(t *testing.T) {
	utcTime := time.Date(2026, 2, 11, 4, 30, 0, 0, time.UTC)

	t.Run("with valid timezone", func(t *testing.T) {
		got := localizeTime(utcTime, "Asia/Kolkata")
		if got.Hour() != 10 || got.Minute() != 0 {
			t.Errorf("expected 10:00 IST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("with empty timezone falls back to local", func(t *testing.T) {
		got := localizeTime(utcTime, "")
		// Should be in local time
		if got.Location() != time.Local {
			t.Errorf("expected local timezone, got %v", got.Location())
		}
	})

	t.Run("with invalid timezone falls back to local", func(t *testing.T) {
		got := localizeTime(utcTime, "Invalid/Timezone")
		if got.Location() != time.Local {
			t.Errorf("expected local timezone, got %v", got.Location())
		}
	})

	t.Run("with UTC timezone", func(t *testing.T) {
		got := localizeTime(utcTime, "UTC")
		if got.Hour() != 4 || got.Minute() != 30 {
			t.Errorf("expected 04:30 UTC, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("with negative offset timezone", func(t *testing.T) {
		got := localizeTime(utcTime, "America/New_York")
		// UTC 04:30 = EST 23:30 (previous day) in February (EST, no DST)
		if got.Hour() != 23 {
			t.Errorf("expected 23:xx EST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a long string", 10, "this is..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}
