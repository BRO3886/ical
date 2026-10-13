package commands

import (
	"testing"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestParseAvailability(t *testing.T) {
	cases := map[string]calendar.Availability{
		"busy":        calendar.AvailabilityBusy,
		"free":        calendar.AvailabilityFree,
		"Free":        calendar.AvailabilityFree,
		" tentative ": calendar.AvailabilityTentative,
		"UNAVAILABLE": calendar.AvailabilityUnavailable,
	}
	for in, want := range cases {
		got, err := parseAvailability(in)
		if err != nil {
			t.Errorf("parseAvailability(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parseAvailability(%q) = %v, want %v", in, got, want)
		}
	}
	for _, bad := range []string{"", "maybe", "notSupported", "1"} {
		if _, err := parseAvailability(bad); err == nil {
			t.Errorf("parseAvailability(%q) expected error", bad)
		}
	}
}
