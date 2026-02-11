package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

// CSV exports events as CSV.
func CSV(events []calendar.Event, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"ID", "Title", "Start", "End", "AllDay", "Calendar", "Location", "Notes", "URL", "Status", "Recurring", "Timezone"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, e := range events {
		record := []string{
			e.ID,
			e.Title,
			e.StartDate.Format(time.RFC3339),
			e.EndDate.Format(time.RFC3339),
			strconv.FormatBool(e.AllDay),
			e.Calendar,
			e.Location,
			e.Notes,
			e.URL,
			e.Status.String(),
			strconv.FormatBool(e.Recurring),
			e.TimeZone,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// ParseCSV reads a CSV file and returns CreateEventInput slice.
func ParseCSV(r io.Reader) ([]calendar.CreateEventInput, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file has no data rows")
	}

	// Find column indices from header
	header := records[0]
	cols := make(map[string]int)
	for i, h := range header {
		cols[h] = i
	}

	var inputs []calendar.CreateEventInput
	for i, record := range records[1:] {
		title := getCol(record, cols, "Title")
		if title == "" {
			continue
		}

		startStr := getCol(record, cols, "Start")
		endStr := getCol(record, cols, "End")
		if startStr == "" || endStr == "" {
			return nil, fmt.Errorf("row %d: missing Start or End", i+2)
		}

		startTime, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid Start %q: %w", i+2, startStr, err)
		}
		endTime, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid End %q: %w", i+2, endStr, err)
		}

		allDay, _ := strconv.ParseBool(getCol(record, cols, "AllDay"))

		inputs = append(inputs, calendar.CreateEventInput{
			Title:     title,
			StartDate: startTime,
			EndDate:   endTime,
			AllDay:    allDay,
			Calendar:  getCol(record, cols, "Calendar"),
			Location:  getCol(record, cols, "Location"),
			Notes:     getCol(record, cols, "Notes"),
			URL:       getCol(record, cols, "URL"),
			TimeZone:  getCol(record, cols, "Timezone"),
		})
	}

	return inputs, nil
}

func getCol(record []string, cols map[string]int, name string) string {
	idx, ok := cols[name]
	if !ok || idx >= len(record) {
		return ""
	}
	return record[idx]
}
