package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/cal/internal/ui"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var calendarsCmd = &cobra.Command{
	Use:     "calendars",
	Aliases: []string{"cals"},
	Short:   "List all calendars",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		cals, err := client.Calendars()
		if err != nil {
			return fmt.Errorf("failed to list calendars: %w", err)
		}

		ui.PrintCalendars(cals, outputFormat)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(calendarsCmd)
}

func handleClientError(err error) error {
	if err.Error() == "calendar: access denied" {
		fmt.Fprintln(os.Stderr, "Calendar access denied. Grant access in System Settings > Privacy & Security > Calendars")
		os.Exit(1)
	}
	return fmt.Errorf("failed to initialize calendar client: %w", err)
}
