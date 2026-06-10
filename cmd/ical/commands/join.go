package commands

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/ical/internal/ui"
	"github.com/spf13/cobra"
)

var (
	joinPrint bool
	joinDays  int
)

var joinCmd = &cobra.Command{
	Use:   "join [number or id]",
	Short: "Open the meeting link of the current or next event",
	Long: `Opens the video-conference link (Zoom, Meet, Teams, FaceTime, ...) of an event.

With no arguments, picks the event you most likely want to join: a meeting
happening right now, or else the next upcoming event that has a conference
link (searched over the next --days days).

With an argument, accepts a row number from the last listing or a
full/partial event ID, same as 'ical show'.`,
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
			if event.ConferenceURL == "" {
				return fmt.Errorf("event %q has no conference link", event.Title)
			}
		} else {
			now := time.Now()
			// EventKit's range predicate uses overlap semantics, so any
			// ongoing event is returned as long as the range starts before it
			// ends. Start at local midnight anyway so all-day events for
			// "today" are reliably in range across timezones.
			y, m, d := now.Date()
			dayStart := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
			events, err := client.Events(dayStart, now.AddDate(0, 0, joinDays))
			if err != nil {
				return fmt.Errorf("failed to fetch events: %w", err)
			}
			event = nextJoinableEvent(events, now)
			if event == nil {
				return fmt.Errorf("no event with a conference link in the next %d day(s)", joinDays)
			}
		}

		if joinPrint {
			fmt.Println(event.ConferenceURL)
			return nil
		}
		if outputFormat == "json" {
			ui.PrintEventDetail(event, "json")
			return nil
		}

		start := event.StartDate.In(time.Local)
		fmt.Fprintf(os.Stderr, "Joining %q (%s)\n", event.Title, start.Format("Mon 15:04"))
		fmt.Println(event.ConferenceURL)
		if err := exec.Command("open", event.ConferenceURL).Run(); err != nil {
			return fmt.Errorf("failed to open conference link: %w", err)
		}
		return nil
	},
}

func init() {
	joinCmd.Flags().BoolVarP(&joinPrint, "print", "p", false, "Print the conference link instead of opening it")
	joinCmd.Flags().IntVarP(&joinDays, "days", "d", 7, "How many days ahead to look for the next meeting")

	rootCmd.AddCommand(joinCmd)
}

// nextJoinableEvent picks the event whose conference link the user most
// likely wants: an ongoing event (started, not yet ended) with a link —
// preferring the most recently started one — or else the upcoming linked
// event that starts soonest. Returns nil if no event has a conference link.
func nextJoinableEvent(events []calendar.Event, now time.Time) *calendar.Event {
	var ongoing, upcoming []calendar.Event
	for _, e := range events {
		if e.ConferenceURL == "" {
			continue
		}
		switch {
		case !e.StartDate.After(now) && e.EndDate.After(now):
			ongoing = append(ongoing, e)
		case e.StartDate.After(now):
			upcoming = append(upcoming, e)
		}
	}
	if len(ongoing) > 0 {
		sort.SliceStable(ongoing, func(i, j int) bool {
			return ongoing[i].StartDate.After(ongoing[j].StartDate)
		})
		return &ongoing[0]
	}
	if len(upcoming) > 0 {
		sort.SliceStable(upcoming, func(i, j int) bool {
			return upcoming[i].StartDate.Before(upcoming[j].StartDate)
		})
		return &upcoming[0]
	}
	return nil
}
