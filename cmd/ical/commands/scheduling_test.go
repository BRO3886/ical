package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func timeAtHour(h int) time.Time {
	return time.Date(2026, 6, 11, h, 0, 0, 0, time.UTC)
}

func TestParseAttendees(t *testing.T) {
	tests := []struct {
		name      string
		invites   []string
		want      []calendar.AttendeeInput
		wantError bool
	}{
		{"nil", nil, nil, false},
		{"bare email", []string{"a@x.com"}, []calendar.AttendeeInput{{Email: "a@x.com"}}, false},
		{"named", []string{"Alice <a@x.com>"}, []calendar.AttendeeInput{{Name: "Alice", Email: "a@x.com"}}, false},
		{"named with spaces", []string{"  Bob Smith  <b@y.com> "}, []calendar.AttendeeInput{{Name: "Bob Smith", Email: "b@y.com"}}, false},
		{
			"multiple mixed",
			[]string{"a@x.com", "Carol <c@z.com>"},
			[]calendar.AttendeeInput{{Email: "a@x.com"}, {Name: "Carol", Email: "c@z.com"}},
			false,
		},
		{"no at sign", []string{"notanemail"}, nil, true},
		{"empty angle brackets", []string{"Name <>"}, nil, true},
		{"name without email no at", []string{"Just A Name"}, nil, true},
		// Extreme / commonly-overlooked cases:
		{"plus addressing", []string{"sidd+cal@x.com"}, []calendar.AttendeeInput{{Email: "sidd+cal@x.com"}}, false},
		{"unicode display name", []string{"Сидд <s@x.com>"}, []calendar.AttendeeInput{{Name: "Сидд", Email: "s@x.com"}}, false},
		{"subdomain multi-tld", []string{"Team <team@sub.x.co.uk>"}, []calendar.AttendeeInput{{Name: "Team", Email: "team@sub.x.co.uk"}}, false},
		{"leading/trailing whitespace bare", []string{"   a@x.com   "}, []calendar.AttendeeInput{{Email: "a@x.com"}}, false},
		{"bracket-only no name (paste format)", []string{"<a@x.com>"}, []calendar.AttendeeInput{{Email: "a@x.com"}}, false},
		{"empty string entry", []string{""}, nil, true},
		{"whitespace-only entry", []string{"   "}, nil, true},
		{"name with @ but empty brackets still errors", []string{"a@b.com <>"}, nil, true},
		{"second entry invalid fails whole batch", []string{"a@x.com", "bogus"}, nil, true},
		// The overlooked footgun: comma/space-packed addresses in one value
		// must fail loudly, not silently become one malformed attendee.
		{"comma-packed addresses rejected", []string{"a@x.com,b@y.com"}, nil, true},
		{"space-packed addresses rejected", []string{"a@x.com b@y.com"}, nil, true},
		{"semicolon-packed rejected", []string{"a@x.com;b@y.com"}, nil, true},
		{"double at rejected", []string{"a@@x.com"}, nil, true},
		{"no dot in domain rejected", []string{"a@localhost"}, nil, true},
		{"trailing dot domain rejected", []string{"a@x."}, nil, true},
		{"leading dot domain rejected", []string{"a@.com"}, nil, true},
		{"empty local part rejected", []string{"@x.com"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAttendees(tt.invites)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d attendees, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("attendee %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitNameEmail(t *testing.T) {
	tests := []struct {
		raw       string
		wantName  string
		wantEmail string
	}{
		{"a@x.com", "", "a@x.com"},
		{"Alice <a@x.com>", "Alice", "a@x.com"},
		{"  Bob  <b@y.com>  ", "Bob", "b@y.com"},
		{"weird <name> <e@z.com>", "weird <name>", "e@z.com"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			name, email := splitNameEmail(tt.raw)
			if name != tt.wantName || email != tt.wantEmail {
				t.Errorf("got (%q, %q), want (%q, %q)", name, email, tt.wantName, tt.wantEmail)
			}
		})
	}
}

func TestParseRSVPStatus(t *testing.T) {
	tests := []struct {
		in        string
		want      calendar.ParticipantStatus
		wantError bool
	}{
		{"accepted", calendar.ParticipantStatusAccepted, false},
		{"accept", calendar.ParticipantStatusAccepted, false},
		{"yes", calendar.ParticipantStatusAccepted, false},
		{"Y", calendar.ParticipantStatusAccepted, false},
		{"declined", calendar.ParticipantStatusDeclined, false},
		{"no", calendar.ParticipantStatusDeclined, false},
		{"tentative", calendar.ParticipantStatusTentative, false},
		{"maybe", calendar.ParticipantStatusTentative, false},
		{"  ACCEPTED  ", calendar.ParticipantStatusAccepted, false},
		{"garbage", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseRSVPStatus(tt.in)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for %q", tt.in)
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

func TestFilterBusySpans(t *testing.T) {
	mk := func(typ calendar.AvailabilityType, hour int) calendar.AvailabilitySpan {
		return calendar.AvailabilitySpan{Type: typ, Start: timeAtHour(hour), End: timeAtHour(hour + 1)}
	}
	t.Run("drops free, keeps busy/tentative/unavailable", func(t *testing.T) {
		spans := []calendar.AvailabilitySpan{
			mk(calendar.AvailabilityTypeFree, 9),
			mk(calendar.AvailabilityTypeBusy, 10),
			mk(calendar.AvailabilityTypeTentative, 11),
			mk(calendar.AvailabilityTypeUnavailable, 12),
		}
		got := filterBusySpans(spans)
		if len(got) != 3 {
			t.Fatalf("got %d, want 3", len(got))
		}
	})
	t.Run("sorts by start", func(t *testing.T) {
		spans := []calendar.AvailabilitySpan{
			mk(calendar.AvailabilityTypeBusy, 14),
			mk(calendar.AvailabilityTypeBusy, 9),
			mk(calendar.AvailabilityTypeBusy, 11),
		}
		got := filterBusySpans(spans)
		if !got[0].Start.Before(got[1].Start) || !got[1].Start.Before(got[2].Start) {
			t.Error("not sorted ascending by start")
		}
	})
	t.Run("all free yields empty", func(t *testing.T) {
		got := filterBusySpans([]calendar.AvailabilitySpan{mk(calendar.AvailabilityTypeFree, 9)})
		if len(got) != 0 {
			t.Errorf("got %d, want 0", len(got))
		}
	})
	t.Run("nil input", func(t *testing.T) {
		if got := filterBusySpans(nil); len(got) != 0 {
			t.Errorf("got %d, want 0", len(got))
		}
	})
	t.Run("identical starts keep stable order", func(t *testing.T) {
		a := calendar.AvailabilitySpan{Type: calendar.AvailabilityTypeBusy, Start: timeAtHour(9), End: timeAtHour(10)}
		b := calendar.AvailabilitySpan{Type: calendar.AvailabilityTypeTentative, Start: timeAtHour(9), End: timeAtHour(11)}
		got := filterBusySpans([]calendar.AvailabilitySpan{a, b})
		if got[0].Type != calendar.AvailabilityTypeBusy || got[1].Type != calendar.AvailabilityTypeTentative {
			t.Error("stable order on equal starts not preserved")
		}
	})
}
