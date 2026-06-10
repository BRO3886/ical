package commands

import (
	"fmt"
	"strings"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var rsvpCmd = &cobra.Command{
	Use:   "rsvp <accepted|declined|tentative> [number or id]",
	Short: "Respond to an event invitation",
	Long: `Sets your RSVP status on an event invitation.

The first argument is your response: accepted, declined, or tentative
(aliases: yes/no/maybe). The second selects the event — a row number from the
last listing or a full/partial event ID. With no event argument, an
interactive picker is shown.

On a server-backed calendar (iCloud, Exchange, Google) this sends your reply
to the organizer.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := parseRSVPStatus(args[0])
		if err != nil {
			return err
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}
		if !client.RSVPSupported() {
			return fmt.Errorf("RSVP is not supported on this macOS version")
		}

		var event *calendar.Event
		if len(args) == 2 {
			event, err = findEventByPrefix(client, args[1])
		} else {
			// Invitations are commonly weeks out, so the picker scans a much
			// wider window than the self-event commands (show/delete) do.
			event, err = pickEvent(client, "", "", 90)
		}
		if err != nil {
			return err
		}
		if event == nil {
			return nil // cancelled
		}

		if err := client.RespondToInvitation(event.ID, status); err != nil {
			return fmt.Errorf("failed to RSVP: %w", err)
		}
		fmt.Printf("RSVP'd %s to %q\n", status, event.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rsvpCmd)
}

// parseRSVPStatus maps a user-facing response word to a ParticipantStatus.
func parseRSVPStatus(s string) (calendar.ParticipantStatus, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "accepted", "accept", "yes", "y":
		return calendar.ParticipantStatusAccepted, nil
	case "declined", "decline", "no", "n":
		return calendar.ParticipantStatusDeclined, nil
	case "tentative", "maybe", "m":
		return calendar.ParticipantStatusTentative, nil
	default:
		return 0, fmt.Errorf("invalid RSVP response %q (use accepted, declined, or tentative)", s)
	}
}
