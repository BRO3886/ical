package export

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

type eventExport struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	AllDay      bool      `json:"all_day"`
	Calendar    string    `json:"calendar"`
	CalendarID  string    `json:"calendar_id"`
	Location    string    `json:"location,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	URL         string    `json:"url,omitempty"`
	Status      string    `json:"status"`
	Recurring   bool      `json:"recurring"`
	TimeZone    string    `json:"timezone,omitempty"`
}

// JSON exports events as a JSON array.
func JSON(events []calendar.Event, w io.Writer) error {
	out := make([]eventExport, len(events))
	for i, e := range events {
		out[i] = eventExport{
			ID:         e.ID,
			Title:      e.Title,
			StartDate:  e.StartDate,
			EndDate:    e.EndDate,
			AllDay:     e.AllDay,
			Calendar:   e.Calendar,
			CalendarID: e.CalendarID,
			Location:   e.Location,
			Notes:      e.Notes,
			URL:        e.URL,
			Status:     e.Status.String(),
			Recurring:  e.Recurring,
			TimeZone:   e.TimeZone,
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	_, err = fmt.Fprintln(w)
	return err
}

// ParseJSON reads a JSON file and returns CreateEventInput slice.
func ParseJSON(r io.Reader) ([]calendar.CreateEventInput, error) {
	var events []eventExport
	if err := json.NewDecoder(r).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	inputs := make([]calendar.CreateEventInput, len(events))
	for i, e := range events {
		inputs[i] = calendar.CreateEventInput{
			Title:     e.Title,
			StartDate: e.StartDate,
			EndDate:   e.EndDate,
			AllDay:    e.AllDay,
			Calendar:  e.Calendar,
			Location:  e.Location,
			Notes:     e.Notes,
			URL:       e.URL,
			TimeZone:  e.TimeZone,
		}
	}

	return inputs, nil
}
