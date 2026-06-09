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

// TestSpanFromFlag verifies --span flag parsing, including the new "all" value.
func TestSpanFromFlag(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    calendar.Span
		wantErr bool
	}{
		{"empty defaults to this", "", calendar.SpanThisEvent, false},
		{"this", "this", calendar.SpanThisEvent, false},
		{"future", "future", calendar.SpanFutureEvents, false},
		{"all maps to future", "all", calendar.SpanFutureEvents, false},
		{"case-insensitive", "ALL", calendar.SpanFutureEvents, false},
		{"trims whitespace", " future ", calendar.SpanFutureEvents, false},
		{"unknown errors", "everything", calendar.SpanThisEvent, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := spanFromFlag(tt.in)
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

// TestDeletedMessage verifies span-aware confirmation wording for recurring
// events versus the plain message for single events (issue #40).
func TestDeletedMessage(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		span      string
		recurring bool
		want      string
	}{
		{"non-recurring", "Lunch", "this", false, "Deleted: Lunch"},
		{"non-recurring with all span", "Lunch", "all", false, "Deleted: Lunch"},
		{"recurring this", "Test club", "this", true, `Deleted 1 occurrence of recurring series "Test club"`},
		{"recurring default span", "Test club", "", true, `Deleted 1 occurrence of recurring series "Test club"`},
		{"recurring future", "Test club", "future", true, `Deleted recurring series "Test club" (this and future occurrences)`},
		{"recurring all", "Test club", "all", true, `Deleted recurring series "Test club" (all occurrences)`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deletedMessage(tt.title, tt.span, tt.recurring)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
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
