package commands

import (
	"fmt"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "List pending event invitations",
	Long:  "Lists event invitations awaiting your response (the Calendar.app notification inbox).",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		invitations, err := client.PendingInvitations()
		if err != nil {
			return fmt.Errorf("failed to fetch invitations: %w", err)
		}

		if len(invitations) == 0 {
			fmt.Println("No pending invitations.")
			return nil
		}

		for i, inv := range invitations {
			when := inv.Start.In(time.Local).Format("Mon 02 Jan 15:04")
			if inv.AllDay {
				when = inv.Start.In(time.Local).Format("Mon 02 Jan") + " (all day)"
			}
			fmt.Printf("%d. %s\n   %s", i+1, inv.Title, when)
			if inv.Organizer != "" {
				fmt.Printf(" · from %s", inv.Organizer)
			}
			if inv.Location != "" {
				fmt.Printf(" · %s", inv.Location)
			}
			fmt.Println()
		}
		fmt.Println("\nRespond with: ical rsvp accepted|declined|tentative <id>")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inboxCmd)
}
