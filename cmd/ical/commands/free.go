package commands

import (
	"fmt"
	"sort"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/go-eventkit/dateparser"
	"github.com/spf13/cobra"
)

var (
	freeFrom string
	freeTo   string
)

var freeCmd = &cobra.Command{
	Use:   "free <email> [email...]",
	Short: "Look up free/busy availability for people",
	Long: `Shows free/busy availability for one or more email addresses over a time range.

Requires a calendar account whose server supports availability lookups
(Exchange or Google Workspace). iCloud does not support free/busy, so an
iCloud-only setup will report that no account supports the lookup.

Defaults to the next 24 hours; use --from/--to to change the window.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		if freeFrom != "" {
			t, err := dateparser.ParseDate(freeFrom)
			if err != nil {
				return fmt.Errorf("invalid --from: %w", err)
			}
			start = t
		}
		end := start.Add(24 * time.Hour)
		if freeTo != "" {
			t, err := dateparser.ParseDate(freeTo)
			if err != nil {
				return fmt.Errorf("invalid --to: %w", err)
			}
			end = endOfDayIfMidnight(t)
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}
		if !client.AvailabilitySupported() {
			return fmt.Errorf("availability lookup is not supported on this macOS version")
		}

		result, err := client.RequestAvailability(args, start, end)
		if err != nil {
			return fmt.Errorf("availability lookup failed: %w", err)
		}

		printAvailability(result, args)
		return nil
	},
}

func init() {
	freeCmd.Flags().StringVarP(&freeFrom, "from", "f", "", "Start of the window (default: now)")
	freeCmd.Flags().StringVarP(&freeTo, "to", "t", "", "End of the window (default: +24h)")
	rootCmd.AddCommand(freeCmd)
}

// printAvailability renders the per-address spans in the order the addresses
// were requested, so output is deterministic regardless of map iteration.
func printAvailability(result map[string][]calendar.AvailabilitySpan, addresses []string) {
	for _, addr := range addresses {
		spans := result[addr]
		busy := filterBusySpans(spans)
		if len(busy) == 0 {
			fmt.Printf("%s: free for the whole window\n", addr)
			continue
		}
		fmt.Printf("%s:\n", addr)
		for _, s := range busy {
			fmt.Printf("  %s – %s  %s\n",
				s.Start.In(time.Local).Format("Mon 02 Jan 15:04"),
				s.End.In(time.Local).Format("15:04"),
				s.Type)
		}
	}
}

// filterBusySpans drops free spans and sorts the rest by start time. Only
// busy/tentative/unavailable periods are interesting when scheduling.
func filterBusySpans(spans []calendar.AvailabilitySpan) []calendar.AvailabilitySpan {
	out := make([]calendar.AvailabilitySpan, 0, len(spans))
	for _, s := range spans {
		if s.Type != calendar.AvailabilityTypeFree {
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}
