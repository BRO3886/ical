package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

// TestBuildAlertSummary verifies the alert summary builder.
func TestBuildAlertSummary(t *testing.T) {
	tests := []struct {
		name   string
		alerts []calendar.Alert
		want   string
	}{
		{"none", nil, "none"},
		{"empty", []calendar.Alert{}, "none"},
		{
			"15m before",
			[]calendar.Alert{{RelativeOffset: -15 * time.Minute}},
			"15m before",
		},
		{
			"1h before",
			[]calendar.Alert{{RelativeOffset: -1 * time.Hour}},
			"1h before",
		},
		{
			"1d before",
			[]calendar.Alert{{RelativeOffset: -24 * time.Hour}},
			"1d before",
		},
		{
			"multiple",
			[]calendar.Alert{
				{RelativeOffset: -15 * time.Minute},
				{RelativeOffset: -1 * time.Hour},
			},
			"15m before, 1h before",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAlertSummary(tt.alerts)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFormatAlertDurationShort verifies compact alert duration formatting.
func TestFormatAlertDurationShort(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"minutes", 15 * time.Minute, "15m"},
		{"hours", 2 * time.Hour, "2h"},
		{"days", 48 * time.Hour, "2d"},
		{"0 minutes", 0, "0m"},
		{"1 minute", time.Minute, "1m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAlertDurationShort(tt.d)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLocalizeEventTime verifies timezone handling for event display.
func TestLocalizeEventTime(t *testing.T) {
	utcTime := time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)

	t.Run("empty timezone uses local", func(t *testing.T) {
		got := localizeEventTime(utcTime, "")
		if got.Location() != time.Local {
			t.Errorf("expected local timezone, got %v", got.Location())
		}
	})

	t.Run("valid timezone converts", func(t *testing.T) {
		got := localizeEventTime(utcTime, "America/New_York")
		ny, _ := time.LoadLocation("America/New_York")
		want := utcTime.In(ny)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid timezone falls back to local", func(t *testing.T) {
		got := localizeEventTime(utcTime, "Invalid/Zone")
		if got.Location() != time.Local {
			t.Errorf("expected local timezone, got %v", got.Location())
		}
	})
}

// TestStartOfDay verifies the start-of-day helper.
func TestStartOfDay(t *testing.T) {
	input := time.Date(2026, 3, 15, 14, 30, 45, 123, time.Local)
	want := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	got := startOfDay(input)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
