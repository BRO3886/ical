package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/ical/internal/ui"
	"github.com/spf13/cobra"
)

var (
	showFrom string
	showTo   string
	showDays int
	showID   string
)

var showCmd = &cobra.Command{
	Use:     "show [number or id]",
	Aliases: []string{"get", "info"},
	Short:   "Show event details",
	Long: `Displays full details for a single event.

With no arguments, shows an interactive picker of upcoming events.
Use --from/--to or --days to control the picker's date range.

With an argument, accepts a row number from the last listing (e.g. 'ical show 2')
or a full/partial event ID. Use --id for exact event ID lookup (no prefix matching).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		idFlagSet := cmd.Flags().Changed("id")
		if idFlagSet && len(args) > 0 {
			return fmt.Errorf("cannot use both --id and a positional argument")
		}

		var event *calendar.Event
		if idFlagSet {
			event, err = findEventByID(client, showID)
			if err != nil {
				return fmt.Errorf("event not found: %w", err)
			}
		} else if len(args) == 1 {
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
	showCmd.Flags().StringVar(&showID, "id", "", "Full event ID (exact match, no prefix search)")

	rootCmd.AddCommand(showCmd)
}

type eventGetter interface {
	Event(id string) (*calendar.Event, error)
}

type eventLookup interface {
	eventGetter
	Events(start, end time.Time, opts ...calendar.ListOption) ([]calendar.Event, error)
}

func findEventByID(client eventGetter, id string) (*calendar.Event, error) {
	if id == "" {
		return nil, fmt.Errorf("event ID is empty: %w", calendar.ErrNotFound)
	}
	event, err := client.Event(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}
	// A stale identifier must never substitute another event before a write.
	if event == nil || event.ID != id {
		return nil, fmt.Errorf("event ID did not match: %w", calendar.ErrNotFound)
	}
	return event, nil
}

// findEventByPrefix finds an event by cached row number, exact ID, or unique prefix.
func findEventByPrefix(client eventLookup, input string) (*calendar.Event, error) {
	if input == "" {
		return nil, fmt.Errorf("event ID is empty: %w", calendar.ErrNotFound)
	}
	if n, err := strconv.Atoi(input); err == nil && n > 0 {
		if id := ui.LookupRowNumber(n); id != "" {
			event, err := findEventByID(client, id)
			if err != nil {
				return nil, fmt.Errorf("event from row %d is no longer available: %w", n, err)
			}
			return event, nil
		}
	}

	event, err := findEventByID(client, input)
	if err == nil {
		return event, nil
	}
	if !errors.Is(err, calendar.ErrNotFound) {
		return nil, err
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
		return nil, fmt.Errorf("no event found matching %q: %w", input, calendar.ErrNotFound)
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
