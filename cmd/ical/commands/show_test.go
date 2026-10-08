package commands

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/ical/internal/ui"
)

const (
	storeUUID = "00000000-1111-2222-3333-444444444444"
	idA       = storeUUID + ":AAAAAAAA-0000-0000-0000-00000000000A"
	idB       = storeUUID + ":BBBBBBBB-0000-0000-0000-00000000000B"
)

type fakeLookup struct {
	byID        map[string]calendar.Event
	all         []calendar.Event
	lenient     *calendar.Event
	lookupErr   error
	searchErr   error
	searchCalls int
}

func (f *fakeLookup) Event(id string) (*calendar.Event, error) {
	if f.lookupErr != nil {
		return nil, f.lookupErr
	}
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
	f.searchCalls++
	return f.all, f.searchErr
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

func TestFindEventByID(t *testing.T) {
	failure := errors.New("access denied")
	tests := []struct {
		name    string
		input   string
		client  *fakeLookup
		wantID  string
		wantErr error
	}{
		{"exact", idA, twoEventStore(nil), idA, nil},
		{"partial", storeUUID[:13], twoEventStore(&calendar.Event{ID: idB}), "", calendar.ErrNotFound},
		{"stale", "stale-id", twoEventStore(&calendar.Event{ID: idB}), "", calendar.ErrNotFound},
		{"unknown", "unknown", twoEventStore(nil), "", calendar.ErrNotFound},
		{"empty", "", twoEventStore(nil), "", calendar.ErrNotFound},
		{"lookup failure", idA, &fakeLookup{lookupErr: failure}, "", failure},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := findEventByID(tt.client, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error=%v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if event != nil {
					t.Fatal("failed lookup returned an event")
				}
				return
			}
			if event == nil || event.ID != tt.wantID {
				t.Errorf("event=%v, want ID %s", event, tt.wantID)
			}
		})
	}
}

func TestFindEventByPrefixFailures(t *testing.T) {
	failure := errors.New("access denied")
	tests := []struct {
		name         string
		input        string
		client       *fakeLookup
		cachedID     string
		wantErr      error
		wantSearches int
	}{
		{"empty", "", twoEventStore(nil), "", calendar.ErrNotFound, 0},
		{"lookup failure", idA, &fakeLookup{lookupErr: failure}, "", failure, 0},
		{"search failure", "prefix", &fakeLookup{searchErr: failure}, "", failure, 1},
		{"stale row", "1", twoEventStore(&calendar.Event{ID: idB}), "stale-id", calendar.ErrNotFound, 0},
		{"row lookup failure", "1", &fakeLookup{lookupErr: failure}, idA, failure, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			if tt.cachedID != "" {
				ui.SaveLastList([]calendar.Event{{ID: tt.cachedID}})
			}
			event, err := findEventByPrefix(tt.client, tt.input)
			if event != nil || !errors.Is(err, tt.wantErr) {
				t.Errorf("event=%v error=%v, want nil and %v", event, err, tt.wantErr)
			}
			if tt.client.searchCalls != tt.wantSearches {
				t.Errorf("search calls=%d, want %d", tt.client.searchCalls, tt.wantSearches)
			}
		})
	}
}

func TestFindEventByPrefixCachedRow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ui.SaveLastList([]calendar.Event{{ID: idA}})
	client := twoEventStore(nil)
	event, err := findEventByPrefix(client, "1")
	if err != nil || event == nil || event.ID != idA {
		t.Fatalf("cached row event=%v error=%v", event, err)
	}
	if client.searchCalls != 0 {
		t.Error("cached row searched prefixes")
	}
}
