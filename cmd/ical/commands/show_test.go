package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/ical/internal/ui"
	"github.com/BRO3886/go-eventkit/calendar"
)

// macOS event identifiers are "<store UUID>:<event UUID>", so every event in a
// store shares a long leading run of characters. Anything derived from that run
// alone - a truncated ID, a stale cache entry - identifies no single event.
const (
	storeUUID = "00000000-1111-2222-3333-444444444444"
	idA       = storeUUID + ":AAAAAAAA-0000-0000-0000-00000000000A"
	idB       = storeUUID + ":BBBBBBBB-0000-0000-0000-00000000000B"
)

// fakeLookup stands in for *calendar.Client. Its lenient field reproduces the
// EventKit behaviour that makes the round-trip check necessary: an identifier
// that matches nothing resolves to some unrelated event rather than failing.
type fakeLookup struct {
	byID    map[string]calendar.Event
	all     []calendar.Event
	lenient *calendar.Event
}

func (f *fakeLookup) Event(id string) (*calendar.Event, error) {
	if e, ok := f.byID[id]; ok {
		return &e, nil
	}
	if f.lenient != nil {
		e := *f.lenient
		return &e, nil
	}
	return nil, calendar.ErrNotFound
}

func (f *fakeLookup) Events(start, end time.Time, opts ...calendar.ListOption) ([]calendar.Event, error) {
	return f.all, nil
}

func twoEventStore(lenient *calendar.Event) *fakeLookup {
	a := calendar.Event{ID: idA, Title: "Interview - round 2"}
	b := calendar.Event{ID: idB, Title: "Dentist"}
	return &fakeLookup{
		byID:    map[string]calendar.Event{idA: a, idB: b},
		all:     []calendar.Event{a, b},
		lenient: lenient,
	}
}

// A truncated ID names no event. The resolver must say so rather than hand back
// whatever EventKit resolved it to - delete acts on what this returns.
func TestFindEventByPrefixRejectsLenientMatch(t *testing.T) {
	stranger := calendar.Event{ID: storeUUID + ":CCCCCCCC-0000-0000-0000-00000000000C", Title: "Cancel broadband"}
	client := twoEventStore(&stranger)

	event, err := findEventByPrefix(client, storeUUID[:13])

	if event != nil {
		t.Fatalf("resolved an ambiguous ID to %q (%s); nothing should have matched", event.Title, event.ID)
	}
	if err == nil {
		t.Fatal("expected an error for an ambiguous ID, got nil")
	}
	if !strings.Contains(err.Error(), "Multiple events match") {
		t.Errorf("expected an ambiguity error, got %q", err)
	}
}

// A stale row number points at an ID that no longer resolves. The lenient
// lookup must not be allowed to substitute a different event for it.
func TestFindEventByPrefixRejectsStaleRowNumber(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ui.SaveLastList([]calendar.Event{{ID: "DEAD0000-0000-0000-0000-000000000000:1"}})

	stranger := calendar.Event{ID: idB, Title: "Dentist"}
	client := twoEventStore(&stranger)

	event, err := findEventByPrefix(client, "1")

	if event != nil {
		t.Fatalf("stale row number resolved to %q (%s)", event.Title, event.ID)
	}
	if err == nil {
		t.Fatal("expected an error for a stale row number, got nil")
	}
}

func TestFindEventByPrefixExactID(t *testing.T) {
	client := twoEventStore(nil)

	event, err := findEventByPrefix(client, idA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != idA {
		t.Errorf("got %s, want %s", event.ID, idA)
	}
}

// A prefix that reaches into the event-specific half is unambiguous, so prefix
// matching still works - the fix narrows the resolver, it does not disable it.
func TestFindEventByPrefixUniquePrefix(t *testing.T) {
	client := twoEventStore(nil)

	event, err := findEventByPrefix(client, storeUUID+":AAAA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.ID != idA {
		t.Errorf("got %s, want %s", event.ID, idA)
	}
}

func TestFindEventByPrefixNoMatch(t *testing.T) {
	stranger := calendar.Event{ID: idB, Title: "Dentist"}
	client := twoEventStore(&stranger)

	event, err := findEventByPrefix(client, "no-such-event")

	if event != nil {
		t.Fatalf("unknown ID resolved to %q (%s)", event.Title, event.ID)
	}
	if err == nil || !strings.Contains(err.Error(), "no event found") {
		t.Errorf("expected a not-found error, got %v", err)
	}
}
