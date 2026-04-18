package commands

import (
	"testing"

	"github.com/BRO3886/go-eventkit/calendar"
)

func TestAttendeeMatches(t *testing.T) {
	tests := []struct {
		name  string
		event calendar.Event
		query string
		want  bool
	}{
		{
			name: "name match case insensitive",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Sarah Smith", Email: "sarah@example.com"},
				},
			},
			query: "sarah",
			want:  true,
		},
		{
			name: "email match",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Bob Jones", Email: "bob@example.com"},
				},
			},
			query: "bob@example",
			want:  true,
		},
		{
			name: "organizer match",
			event: calendar.Event{
				Organizer: "Alice Johnson",
			},
			query: "alice",
			want:  true,
		},
		{
			name: "substring match",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Sarah Smith", Email: "sarah@example.com"},
				},
			},
			query: "smi",
			want:  true,
		},
		{
			name:  "no attendees returns false",
			event: calendar.Event{},
			query: "anyone",
			want:  false,
		},
		{
			name: "no match returns false",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Sarah Smith", Email: "sarah@example.com"},
				},
				Organizer: "Bob Jones",
			},
			query: "charlie",
			want:  false,
		},
		{
			name: "multiple attendees one matches",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Alice Brown", Email: "alice@example.com"},
					{Name: "Bob Green", Email: "bob@example.com"},
					{Name: "Carol White", Email: "carol@example.com"},
				},
			},
			query: "bob",
			want:  true,
		},
		{
			name: "uppercase query matches lowercase name",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "sarah smith", Email: "sarah@example.com"},
				},
			},
			query: "SARAH",
			want:  true,
		},
		{
			name: "padded query is trimmed before matching",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Sarah Smith", Email: "sarah@example.com"},
				},
			},
			query: "  sarah  ",
			want:  true,
		},
		{
			name: "whitespace-only query does not match anything",
			event: calendar.Event{
				Attendees: []calendar.Attendee{
					{Name: "Sarah Smith", Email: "sarah@example.com"},
				},
				Organizer: "Bob Jones",
			},
			query: "   ",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attendeeMatches(tt.event, tt.query)
			if got != tt.want {
				t.Errorf("attendeeMatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
