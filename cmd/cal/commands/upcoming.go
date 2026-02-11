package commands

import (
	"time"

	"github.com/spf13/cobra"
)

var upcomingDays int

var upcomingCmd = &cobra.Command{
	Use:     "upcoming",
	Aliases: []string{"next", "soon"},
	Short:   "Show events in the next N days",
	Long:    "Shortcut for 'cal list' with --from today --to 'in N days'.",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		from := startOfDay(now)
		to := startOfDay(now.AddDate(0, 0, upcomingDays))
		return listEvents(from, to)
	},
}

func init() {
	upcomingCmd.Flags().IntVarP(&upcomingDays, "days", "d", 7, "Number of days to look ahead")
	upcomingCmd.Flags().StringVarP(&listCalendar, "calendar", "c", "", "Filter by calendar name")
	upcomingCmd.Flags().StringVar(&listCalendarID, "calendar-id", "", "Filter by calendar ID")
	upcomingCmd.Flags().StringVarP(&listSearch, "search", "s", "", "Search title, location, notes")
	upcomingCmd.Flags().BoolVar(&listAllDay, "all-day", false, "Show only all-day events")
	upcomingCmd.Flags().StringVar(&listSort, "sort", "start", "Sort by: start, end, title, calendar")
	upcomingCmd.Flags().IntVarP(&listLimit, "limit", "n", 0, "Max events to display")

	rootCmd.AddCommand(upcomingCmd)
}
