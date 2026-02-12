package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BRO3886/ical/internal/parser"
	"github.com/BRO3886/go-eventkit/calendar"
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
)

var deleteCmd = &cobra.Command{
	Use:     "delete [number or id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete an event",
	Long: `Deletes an event. Asks for confirmation by default.

With no arguments, shows an interactive picker to select the event.
Use --from/--to or --days to control the picker's date range.

With an argument, accepts a row number from the last listing or a full/partial event ID.`,
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

		span := calendar.SpanThisEvent
		if deleteSpan == "future" {
			span = calendar.SpanFutureEvents
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

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.Flags().StringVar(&deleteSpan, "span", "this", "For recurring: this or future")
	deleteCmd.Flags().StringVar(&deleteFrom, "from", "", "Start date for event picker")
	deleteCmd.Flags().StringVar(&deleteTo, "to", "", "End date for event picker")
	deleteCmd.Flags().IntVarP(&deleteDays, "days", "d", 7, "Number of days to show in picker")

	rootCmd.AddCommand(deleteCmd)
}

// pickEvent shows an interactive huh.Select picker for events in a date range.
// Returns nil, nil if user cancelled.
func pickEvent(client *calendar.Client, fromStr, toStr string, days int) (*calendar.Event, error) {
	now := time.Now()

	from := startOfDay(now)
	if fromStr != "" {
		t, err := parser.ParseDate(fromStr)
		if err != nil {
			return nil, fmt.Errorf("invalid --from date: %w", err)
		}
		from = t
	}

	to := from.AddDate(0, 0, days)
	if toStr != "" {
		t, err := parser.ParseDate(toStr)
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
		label := fmt.Sprintf("%s  %s (%s)", parser.FormatTimeRange(start, end, e.AllDay), e.Title, e.Calendar)
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
