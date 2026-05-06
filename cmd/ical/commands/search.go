package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/go-eventkit/dateparser"
	"github.com/BRO3886/ical/internal/ui"
	"github.com/spf13/cobra"
)

var (
	searchFrom        string
	searchTo          string
	searchCalendars   []string
	searchLimit       int
	searchAttendee    string
	searchNoRecurring bool
)

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Aliases: []string{"find"},
	Short:   "Search events",
	Long:    "Searches events within a date range by title, location, and notes.",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		now := time.Now()
		from := now.AddDate(0, 0, -30) // 30 days ago
		if searchFrom != "" {
			t, err := dateparser.ParseDate(searchFrom)
			if err != nil {
				return fmt.Errorf("invalid --from date: %w", err)
			}
			from = t
		}

		to := now.AddDate(0, 0, 30) // 30 days ahead
		if searchTo != "" {
			t, err := dateparser.ParseDate(searchTo)
			if err != nil {
				return fmt.Errorf("invalid --to date: %w", err)
			}
			to = endOfDayIfMidnight(t)
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		opts := []calendar.ListOption{calendar.WithSearch(query)}
		normalized := normalizeCalendarNames(searchCalendars)
		if len(normalized) == 1 {
			opts = append(opts, calendar.WithCalendar(normalized[0]))
		} else if len(normalized) > 1 {
			opts = append(opts, calendar.WithCalendars(normalized))
		}

		events, err := client.Events(from, to, opts...)
		if err != nil {
			return fmt.Errorf("failed to search events: %w", err)
		}

		if searchAttendee != "" {
			filtered := make([]calendar.Event, 0, len(events))
			for _, e := range events {
				if attendeeMatches(e, searchAttendee) {
					filtered = append(filtered, e)
				}
			}
			events = filtered
		}

		if searchNoRecurring {
			events = filterRecurring(events)
		}

		if searchLimit > 0 && len(events) > searchLimit {
			events = events[:searchLimit]
		}

		ui.PrintEvents(events, outputFormat)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchFrom, "from", "f", "", "Start of search range (default: 30 days ago)")
	searchCmd.Flags().StringVarP(&searchTo, "to", "t", "", "End of search range (default: 30 days ahead)")
	searchCmd.Flags().StringArrayVarP(&searchCalendars, "calendar", "c", nil, "Filter by calendar name (repeatable)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 0, "Max results")
	searchCmd.Flags().StringVarP(&searchAttendee, "attendee", "a", "", "Filter by attendee or organizer name/email")
	searchCmd.Flags().BoolVar(&searchNoRecurring, "no-recurring", false, "Hide recurring events")

	rootCmd.AddCommand(searchCmd)
}
