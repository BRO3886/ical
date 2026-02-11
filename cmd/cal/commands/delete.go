package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	deleteForce bool
	deleteSpan  string
)

var deleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete an event",
	Long:    "Deletes an event. Asks for confirmation by default.",
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

	rootCmd.AddCommand(deleteCmd)
}
