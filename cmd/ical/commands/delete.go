package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/go-eventkit/dateparser"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	deleteForce bool
	deleteSpan  string
	deleteFrom  string
	deleteTo    string
	deleteDays  int
	deleteID    string
)

var deleteCmd = &cobra.Command{
	Use:     "delete [number or id...]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete one or more events",
	Long: `Deletes one or more events. Asks for confirmation by default.

With no arguments, shows an interactive picker to select the event.
Use --from/--to or --days to control the picker's date range.

With one argument, accepts a row number from the last listing or a full/partial event ID.
With multiple arguments, performs a batch delete using row numbers or event IDs.
Use --id for exact event ID lookup (no prefix matching, single event only).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		idFlagSet := cmd.Flags().Changed("id")
		if idFlagSet && len(args) > 0 {
			return fmt.Errorf("cannot use both --id and a positional argument")
		}

		span := calendar.SpanThisEvent
		if deleteSpan == "future" {
			span = calendar.SpanFutureEvents
		}

		// Batch delete: multiple positional args
		if len(args) > 1 {
			return runBatchDelete(client, args, span)
		}

		// Single event delete
		var event *calendar.Event
		if idFlagSet {
			event, err = client.Event(deleteID)
			if err != nil {
				return fmt.Errorf("event not found: %w", err)
			}
		} else if len(args) == 1 {
			event, err = findEventByPrefix(client, args[0])
			if err != nil {
				return err
			}
		} else {
			event, err = pickEvent(client, deleteFrom, deleteTo, deleteDays)
			if err != nil {
				return err
			}
			if event == nil {
				return nil // user cancelled
			}
		}

		if !deleteForce {
			red := color.New(color.FgRed, color.Bold)
			red.Printf("Delete event: ")
			fmt.Printf("%s (%s)\n", event.Title, event.StartDate.Format("Jan 02 15:04"))

			fmt.Print("Are you sure? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteEvent(event.ID, span); err != nil {
			return fmt.Errorf("failed to delete event: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("Deleted: ")
		fmt.Printf("%s\n", event.Title)
		return nil
	},
}

// runBatchDelete resolves multiple args to events and deletes them in a single batch call.
func runBatchDelete(client *calendar.Client, args []string, span calendar.Span) error {
	// Resolve all args to events first
	events := make([]*calendar.Event, 0, len(args))
	for _, arg := range args {
		event, err := findEventByPrefix(client, arg)
		if err != nil {
			return fmt.Errorf("resolving %q: %w", arg, err)
		}
		events = append(events, event)
	}

	// Confirmation
	if !deleteForce {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("Delete %d events:\n", len(events))
		for _, e := range events {
			fmt.Printf("  - %s (%s)\n", e.Title, e.StartDate.Format("Jan 02 15:04"))
		}

		fmt.Print("Are you sure? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Collect IDs and delete in batch
	ids := make([]string, len(events))
	nameByID := make(map[string]string, len(events))
	for i, e := range events {
		ids[i] = e.ID
		nameByID[e.ID] = e.Title
	}

	errs := client.DeleteEvents(ids, span)

	green := color.New(color.FgGreen, color.Bold)
	redC := color.New(color.FgRed, color.Bold)
	var failed int
	for _, id := range ids {
		if err, ok := errs[id]; ok && err != nil {
			redC.Print("Failed: ")
			fmt.Printf("%s — %v\n", nameByID[id], err)
			failed++
		} else {
			green.Print("Deleted: ")
			fmt.Printf("%s\n", nameByID[id])
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d events failed to delete", failed, len(ids))
	}
	return nil
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.Flags().StringVar(&deleteSpan, "span", "this", "For recurring: this or future")
	deleteCmd.Flags().StringVar(&deleteFrom, "from", "", "Start date for event picker")
	deleteCmd.Flags().StringVar(&deleteTo, "to", "", "End date for event picker")
	deleteCmd.Flags().IntVarP(&deleteDays, "days", "d", 7, "Number of days to show in picker")
	deleteCmd.Flags().StringVar(&deleteID, "id", "", "Full event ID (exact match, no prefix search)")

	rootCmd.AddCommand(deleteCmd)
}

// pickEvent shows an interactive huh.Select picker for events in a date range.
// Returns nil, nil if user cancelled.
func pickEvent(client *calendar.Client, fromStr, toStr string, days int) (*calendar.Event, error) {
	now := time.Now()

	from := startOfDay(now)
	if fromStr != "" {
		t, err := dateparser.ParseDate(fromStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --from date: %w", err)
		}
		from = t
	}

	to := from.AddDate(0, 0, days)
	if toStr != "" {
		t, err := dateparser.ParseDate(toStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --to date: %w", err)
		}
		to = endOfDayIfMidnight(t)
	}

	events, err := client.Events(from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	if len(events) == 0 {
		fmt.Println("No events found in the selected range.")
		return nil, nil
	}

	options := make([]huh.Option[string], len(events))
	for i, e := range events {
		start := localizeEventTime(e.StartDate, e.TimeZone)
		end := localizeEventTime(e.EndDate, e.TimeZone)
		label := fmt.Sprintf("%s  %s (%s)", dateparser.FormatTimeRange(start, end, e.AllDay), e.Title, e.Calendar)
		options[i] = huh.NewOption(label, e.ID)
	}

	var selectedID string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an event").
				Options(options...).
				Value(&selectedID),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("selection error: %w", err)
	}

	event, err := client.Event(selectedID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}

	return event, nil
}
