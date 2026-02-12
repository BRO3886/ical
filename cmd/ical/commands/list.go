package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/BRO3886/ical/internal/parser"
	"github.com/BRO3886/ical/internal/ui"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var (
	listFrom            string
	listTo              string
	listCalendar        string
	listCalendarID      string
	listSearch          string
	listAllDay          bool
	listSort            string
	listLimit           int
	listExcludeCalendar []string
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "events"},
	Short:   "List events in a date range",
	Long:    "List events within a date range. Defaults to today if no range specified.",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()

		from := startOfDay(now)
		if listFrom != "" {
			t, err := parser.ParseDate(listFrom)
			if err != nil {
				return fmt.Errorf("invalid --from date: %w", err)
			}
			from = t
		}

		to := from.Add(24 * time.Hour)
		if listTo != "" {
			t, err := parser.ParseDate(listTo)
			if err != nil {
				return fmt.Errorf("invalid --to date: %w", err)
			}
			to = endOfDayIfMidnight(t)
		}

		return listEvents(from, to)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listFrom, "from", "f", "", "Start date (natural language or ISO 8601)")
	listCmd.Flags().StringVarP(&listTo, "to", "t", "", "End date (natural language or ISO 8601)")
	listCmd.Flags().StringVarP(&listCalendar, "calendar", "c", "", "Filter by calendar name")
	listCmd.Flags().StringVar(&listCalendarID, "calendar-id", "", "Filter by calendar ID")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "Search title, location, notes")
	listCmd.Flags().BoolVar(&listAllDay, "all-day", false, "Show only all-day events")
	listCmd.Flags().StringVar(&listSort, "sort", "start", "Sort by: start, end, title, calendar")
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "Max events to display")
	listCmd.Flags().StringArrayVar(&listExcludeCalendar, "exclude-calendar", nil, "Exclude calendars by name (repeatable)")

	rootCmd.AddCommand(listCmd)
}

func listEvents(from, to time.Time) error {
	client, err := calendar.New()
	if err != nil {
		return handleClientError(err)
	}

	opts := buildListOptions()
	events, err := client.Events(from, to, opts...)
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	if listAllDay {
		filtered := make([]calendar.Event, 0)
		for _, e := range events {
			if e.AllDay {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}

	if len(listExcludeCalendar) > 0 {
		excluded := make(map[string]bool, len(listExcludeCalendar))
		for _, c := range listExcludeCalendar {
			excluded[strings.ToLower(c)] = true
		}
		filtered := make([]calendar.Event, 0, len(events))
		for _, e := range events {
			if !excluded[strings.ToLower(e.Calendar)] {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}

	sortEvents(events, listSort)

	if listLimit > 0 && len(events) > listLimit {
		events = events[:listLimit]
	}

	ui.PrintEvents(events, outputFormat)
	return nil
}

func buildListOptions() []calendar.ListOption {
	var opts []calendar.ListOption
	if listCalendar != "" {
		opts = append(opts, calendar.WithCalendar(listCalendar))
	}
	if listCalendarID != "" {
		opts = append(opts, calendar.WithCalendarID(listCalendarID))
	}
	if listSearch != "" {
		opts = append(opts, calendar.WithSearch(listSearch))
	}
	return opts
}

func sortEvents(events []calendar.Event, sortBy string) {
	switch sortBy {
	case "end":
		sort.Slice(events, func(i, j int) bool {
			return events[i].EndDate.Before(events[j].EndDate)
		})
	case "title":
		sort.Slice(events, func(i, j int) bool {
			return events[i].Title < events[j].Title
		})
	case "calendar":
		sort.Slice(events, func(i, j int) bool {
			return events[i].Calendar < events[j].Calendar
		})
	default: // "start"
		sort.Slice(events, func(i, j int) bool {
			return events[i].StartDate.Before(events[j].StartDate)
		})
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// endOfDayIfMidnight bumps a midnight time to 23:59:59 so that --to "feb 12"
// means "through the end of Feb 12" rather than "up to the start of Feb 12".
// If the time has an explicit hour/minute (not midnight), it's left as-is.
func endOfDayIfMidnight(t time.Time) time.Time {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
	}
	return t
}
