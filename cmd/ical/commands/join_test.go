package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

func joinEvt(title, conf string, start, end time.Time) calendar.Event {
	return calendar.Event{Title: title, ConferenceURL: conf, StartDate: start, EndDate: end}
}

func TestNextJoinableEvent(t *testing.T) {
	now := time.Date(2026, 6, 11, 14, 0, 0, 0, time.UTC)
	h := func(d time.Duration) time.Time { return now.Add(d) }

	tests := []struct {
		name   string
		events []calendar.Event
		want   string // title of expected pick; "" means nil
	}{
		{"empty slice", nil, ""},
		{
			"no events have links",
			[]calendar.Event{joinEvt("a", "", h(-time.Hour), h(time.Hour))},
			"",
		},
		{
			"ongoing beats upcoming",
			[]calendar.Event{
				joinEvt("later", "https://zoom.us/j/2", h(30*time.Minute), h(time.Hour)),
				joinEvt("now", "https://zoom.us/j/1", h(-30*time.Minute), h(30*time.Minute)),
			},
			"now",
		},
		{
			"most recently started ongoing wins",
			[]calendar.Event{
				joinEvt("long standup", "https://zoom.us/j/1", h(-2*time.Hour), h(time.Hour)),
				joinEvt("just started", "https://zoom.us/j/2", h(-5*time.Minute), h(time.Hour)),
			},
			"just started",
		},
		{
			"ongoing without link is skipped in favor of upcoming with link",
			[]calendar.Event{
				joinEvt("now no link", "", h(-time.Hour), h(time.Hour)),
				joinEvt("next with link", "https://meet.google.com/abc-defg-hij", h(time.Hour), h(2*time.Hour)),
			},
			"next with link",
		},
		{
			"soonest upcoming wins",
			[]calendar.Event{
				joinEvt("tomorrow", "https://zoom.us/j/2", h(24*time.Hour), h(25*time.Hour)),
				joinEvt("in 10m", "https://zoom.us/j/1", h(10*time.Minute), h(time.Hour)),
			},
			"in 10m",
		},
		{
			"ended events are ignored",
			[]calendar.Event{
				joinEvt("over", "https://zoom.us/j/1", h(-2*time.Hour), h(-time.Hour)),
			},
			"",
		},
		{
			"event ending exactly now is not ongoing",
			[]calendar.Event{
				joinEvt("ends now", "https://zoom.us/j/1", h(-time.Hour), now),
			},
			"",
		},
		{
			"event starting exactly now is ongoing",
			[]calendar.Event{
				joinEvt("starts now", "https://zoom.us/j/1", now, h(time.Hour)),
			},
			"starts now",
		},
		{
			"stable tie on identical upcoming starts",
			[]calendar.Event{
				joinEvt("first listed", "https://zoom.us/j/1", h(time.Hour), h(2*time.Hour)),
				joinEvt("second listed", "https://zoom.us/j/2", h(time.Hour), h(2*time.Hour)),
			},
			"first listed",
		},
		{
			"all-day ongoing event with link is joinable",
			[]calendar.Event{
				{Title: "workshop", ConferenceURL: "https://zoom.us/j/1", StartDate: h(-6 * time.Hour), EndDate: h(18 * time.Hour), AllDay: true},
			},
			"workshop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextJoinableEvent(tt.events, now)
			switch {
			case tt.want == "" && got != nil:
				t.Errorf("got %q, want nil", got.Title)
			case tt.want != "" && got == nil:
				t.Errorf("got nil, want %q", tt.want)
			case tt.want != "" && got.Title != tt.want:
				t.Errorf("got %q, want %q", got.Title, tt.want)
			}
		})
	}
}

func TestNextJoinableEvent_DoesNotMutateInput(t *testing.T) {
	now := time.Date(2026, 6, 11, 14, 0, 0, 0, time.UTC)
	events := []calendar.Event{
		joinEvt("b", "https://zoom.us/j/2", now.Add(2*time.Hour), now.Add(3*time.Hour)),
		joinEvt("a", "https://zoom.us/j/1", now.Add(time.Hour), now.Add(2*time.Hour)),
	}
	_ = nextJoinableEvent(events, now)
	if events[0].Title != "b" || events[1].Title != "a" {
		t.Error("input slice order was mutated")
	}
}
