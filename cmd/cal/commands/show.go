package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/cal/internal/ui"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show [id]",
	Aliases: []string{"get", "info"},
	Short:   "Show event details",
	Long:    "Displays full details for a single event. Supports ID prefix matching.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		event, err := findEventByPrefix(client, args[0])
		if err != nil {
			return err
		}

		ui.PrintEventDetail(event, outputFormat)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}

// findEventByPrefix finds an event by ID prefix matching.
// First tries exact match, then searches recent events for prefix matches.
func findEventByPrefix(client *calendar.Client, prefix string) (*calendar.Event, error) {
	// Try exact match first
	event, err := client.Event(prefix)
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
		if strings.HasPrefix(e.ID, prefix) {
			matches = append(matches, e)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no event found with ID %q", prefix)
	case 1:
		return &matches[0], nil
	default:
		var sb strings.Builder
		fmt.Fprintf(&sb, "Multiple events match %q. Did you mean:\n", prefix)
		for _, m := range matches {
			fmt.Fprintf(&sb, "  %s  %s (%s)\n", ui.ShortID(m.ID), m.Title, m.StartDate.Format("Jan 02 15:04"))
		}
		return nil, fmt.Errorf("%s", sb.String())
	}
}
