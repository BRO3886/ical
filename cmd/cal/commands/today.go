package commands

import (
	"time"

	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's events",
	Long:  "Shortcut for 'cal list --from today --to tomorrow'. Shows the day's agenda.",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		from := startOfDay(now)
		to := startOfDay(now.AddDate(0, 0, 1))
		return listEvents(from, to)
	},
}

func init() {
	todayCmd.Flags().StringVarP(&listCalendar, "calendar", "c", "", "Filter by calendar name")
	todayCmd.Flags().StringVar(&listCalendarID, "calendar-id", "", "Filter by calendar ID")
	todayCmd.Flags().StringVarP(&listSearch, "search", "s", "", "Search title, location, notes")
	todayCmd.Flags().BoolVar(&listAllDay, "all-day", false, "Show only all-day events")
	todayCmd.Flags().StringVar(&listSort, "sort", "start", "Sort by: start, end, title, calendar")
	todayCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "Max events to display")

	rootCmd.AddCommand(todayCmd)
}
