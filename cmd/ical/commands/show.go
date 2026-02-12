package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BRO3886/ical/internal/ui"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var (
	showFrom string
	showTo   string
	showDays int
)

var showCmd = &cobra.Command{
	Use:     "show [number or id]",
	Aliases: []string{"get", "info"},
	Short:   "Show event details",
	Long: `Displays full details for a single event.

With no arguments, shows an interactive picker of upcoming events.
Use --from/--to or --days to control the picker's date range.

With an argument, accepts a row number from the last listing (e.g. 'cal show 2')
or a full/partial event ID.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		var event *calendar.Event
		if len(args) == 1 {
			event, err = findEventByPrefix(client, args[0])
			if err != nil {
				return err
			}
		} else {
			event, err = pickEvent(client, showFrom, showTo, showDays)
			if err != nil {
				return err
			}
			if event == nil {
				return nil // user cancelled
			}
		}

		ui.PrintEventDetail(event, outputFormat)
		return nil
	},
}

func init() {
	showCmd.Flags().StringVarP(&showFrom, "from", "f", "", "Start date for event picker (natural language or ISO 8601)")
	showCmd.Flags().StringVarP(&showTo, "to", "t", "", "End date for event picker")
	showCmd.Flags().IntVarP(&showDays, "days", "d", 7, "Number of days to show in picker")

	rootCmd.AddCommand(showCmd)
}

// findEventByPrefix finds an event by row number from the last listing,
// by exact ID, or by ID prefix matching.
func findEventByPrefix(client *calendar.Client, input string) (*calendar.Event, error) {
	// Check if input is a row number (e.g. "1", "2") from the last listing
	if n, err := strconv.Atoi(input); err == nil && n > 0 {
		if id := ui.LookupRowNumber(n); id != "" {
			event, err := client.Event(id)
			if err == nil {
				return event, nil
			}
			// Event may have been deleted since listing; fall through to search
		}
	}

	// Try exact match
	event, err := client.Event(input)
	if err == nil {
		return event, nil
	}
	if !errors.Is(err, calendar.ErrNotFound) {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}

	// Search recent events for prefix match
	now := time.Now()
	start := now.AddDate(-1, 0, 0) // 1 year back
	end := now.AddDate(1, 0, 0)    // 1 year forward
	events, err := client.Events(start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}

	var matches []calendar.Event
	for _, e := range events {
		if strings.HasPrefix(e.ID, input) {
			matches = append(matches, e)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no event found matching %q", input)
	case 1:
		return &matches[0], nil
	default:
		var sb strings.Builder
		fmt.Fprintf(&sb, "Multiple events match %q. Be more specific:\n", input)
		for _, m := range matches {
			fmt.Fprintf(&sb, "  %s  %s (%s)\n", m.ID, m.Title, m.StartDate.Format("Jan 02 15:04"))
		}
		return nil, fmt.Errorf("%s", sb.String())
	}
}
