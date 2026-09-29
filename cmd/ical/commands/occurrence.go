package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/go-eventkit/dateparser"
)

// Every occurrence of a recurring series shares one event ID, and looking an
// ID up always yields the series' first occurrence. Writes aimed at a later
// occurrence (a row from a listing, a picker choice, or --occurrence) must
// therefore carry that occurrence's original start through to EventKit.

// seriesID strips the "/RID=<unix time>" suffix EventKit gives a detached
// occurrence, leaving the ID shared by the whole series.
func seriesID(id string) string {
	base, _, _ := strings.Cut(id, "/RID=")
	return base
}

// occurrenceTarget returns the occurrence a write to e should target, or nil
// to act on the event as its ID resolves (the series' first occurrence).
// --span all always acts on the series as a whole.
func occurrenceTarget(e *calendar.Event, spanFlag string) *time.Time {
	if !(e.Recurring || e.IsDetached) || e.OccurrenceDate == nil {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(spanFlag), "all") {
		return nil
	}
	return e.OccurrenceDate
}

// findOccurrence picks the occurrence of series id whose original start is
// at, or with wholeDay, the only one whose original start falls on at's
// local day.
func findOccurrence(events []calendar.Event, id string, at time.Time, wholeDay bool) (*calendar.Event, error) {
	dayStart := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, at.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	series := seriesID(id)
	var matches []*calendar.Event
	for i := range events {
		e := &events[i]
		if seriesID(e.ID) != series {
			continue
		}
		occ := e.StartDate
		if e.OccurrenceDate != nil {
			occ = *e.OccurrenceDate
		}
		if wholeDay {
			if !occ.Before(dayStart) && occ.Before(dayEnd) {
				matches = append(matches, e)
			}
		} else if occ.Sub(at).Abs() < time.Second {
			matches = append(matches, e)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return nil, fmt.Errorf("no occurrence of this event at %s", at.Format("2006-01-02 15:04"))
	default:
		return nil, fmt.Errorf("%d occurrences on %s; pass a time too, e.g. --occurrence \"%s\"",
			len(matches), at.Format("2006-01-02"), matches[0].OccurrenceDate.In(at.Location()).Format("2006-01-02 15:04"))
	}
}

// fetchOccurrence loads the occurrence of series id at the given time. Moved
// occurrences are found up to a week from their original start.
func fetchOccurrence(client *calendar.Client, id string, at time.Time, wholeDay bool) (*calendar.Event, error) {
	events, err := client.Events(at.AddDate(0, 0, -7), at.AddDate(0, 0, 8))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch occurrences: %w", err)
	}
	return findOccurrence(events, id, at, wholeDay)
}

// applyOccurrenceFlag narrows event to the occurrence named by --occurrence
// (a date for the only occurrence that day, or a date and time).
func applyOccurrenceFlag(client *calendar.Client, event *calendar.Event, value string) (*calendar.Event, error) {
	if !(event.Recurring || event.IsDetached) {
		return nil, fmt.Errorf("--occurrence only applies to recurring events")
	}
	at, err := dateparser.ParseDate(value)
	if err != nil {
		return nil, fmt.Errorf("invalid --occurrence: %w", err)
	}
	wholeDay := at.Hour() == 0 && at.Minute() == 0 && at.Second() == 0
	return fetchOccurrence(client, event.ID, at, wholeDay)
}

// updateEvent writes input to event, targeting its occurrence when it has one.
func updateEvent(client *calendar.Client, event *calendar.Event, input calendar.UpdateEventInput, spanFlag string) (*calendar.Event, error) {
	span, err := spanFromFlag(spanFlag)
	if err != nil {
		return nil, err
	}
	if occ := occurrenceTarget(event, spanFlag); occ != nil {
		return client.UpdateEventOccurrence(event.ID, *occ, input, span)
	}
	return client.UpdateEvent(seriesID(event.ID), input, span)
}

// deleteEvent removes event, targeting its occurrence when it has one.
func deleteEvent(client *calendar.Client, event *calendar.Event, spanFlag string) error {
	span, err := spanFromFlag(spanFlag)
	if err != nil {
		return err
	}
	if occ := occurrenceTarget(event, spanFlag); occ != nil {
		return client.DeleteEventOccurrence(event.ID, *occ, span)
	}
	return client.DeleteEvent(seriesID(event.ID), span)
}
