package ui

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestFormatTravelTime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Minute, "30m"},
		{time.Hour, "1h"},
		{90 * time.Minute, "1h30m"},
		{2 * time.Hour, "2h"},
		{45*time.Minute + 30*time.Second, "46m"}, // rounds to nearest minute
		{5 * time.Minute, "5m"},
		// Overlooked boundaries:
		{30 * time.Second, "1m"},          // rounds up to 1m
		{29 * time.Second, "0m"},          // rounds down — sub-30s shows 0m (documented)
		{25 * time.Hour, "25h"},           // beyond Apple's presets, still formats
		{time.Hour + time.Minute, "1h1m"}, // singular-minute tail
		{23*time.Hour + 59*time.Minute, "23h59m"},
		{59 * time.Minute, "59m"}, // just under an hour stays in minutes
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := formatTravelTime(tt.d); got != tt.want {
				t.Errorf("formatTravelTime(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestTravelTimeJSON(t *testing.T) {
	if got := travelTimeJSON(0); got != "" {
		t.Errorf("zero should be empty, got %q", got)
	}
	if got := travelTimeJSON(-time.Minute); got != "" {
		t.Errorf("negative should be empty, got %q", got)
	}
	if got := travelTimeJSON(30 * time.Minute); got != "30m" {
		t.Errorf("got %q, want 30m", got)
	}
}

func TestSelfStatusJSON(t *testing.T) {
	if got := selfStatusJSON(calendar.ParticipantStatusUnknown); got != "" {
		t.Errorf("unknown should be empty, got %q", got)
	}
	if got := selfStatusJSON(calendar.ParticipantStatusAccepted); got != "accepted" {
		t.Errorf("got %q, want accepted", got)
	}
}
