package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/cal/internal/parser"
	"github.com/BRO3886/cal/internal/ui"
	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var (
	addTitle          string
	addStart          string
	addEnd            string
	addAllDay         bool
	addCalendar       string
	addLocation       string
	addNotes          string
	addURL            string
	addAlerts         []string
	addRepeat         string
	addRepeatInterval int
	addRepeatUntil    string
	addRepeatCount    int
	addRepeatDays     string
	addTimezone       string
)

var addCmd = &cobra.Command{
	Use:     "add [title]",
	Aliases: []string{"create", "new"},
	Short:   "Create a new event",
	Long:    "Creates a new calendar event. Title can be passed as argument or via --title flag.",
	RunE: func(cmd *cobra.Command, args []string) error {
		title := addTitle
		if title == "" && len(args) > 0 {
			title = strings.Join(args, " ")
		}
		if title == "" {
			return fmt.Errorf("title is required (pass as argument or use --title)")
		}

		if addStart == "" {
			return fmt.Errorf("--start is required")
		}

		startTime, err := parser.ParseDate(addStart)
		if err != nil {
			return fmt.Errorf("invalid --start date: %w", err)
		}

		endTime := startTime.Add(time.Hour)
		if addEnd != "" {
			endTime, err = parser.ParseDate(addEnd)
			if err != nil {
				return fmt.Errorf("invalid --end date: %w", err)
			}
		}

		if addTimezone != "" {
			loc, err := time.LoadLocation(addTimezone)
			if err != nil {
				return fmt.Errorf("invalid timezone %q: %w", addTimezone, err)
			}
			startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(),
				startTime.Hour(), startTime.Minute(), startTime.Second(), 0, loc)
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(),
				endTime.Hour(), endTime.Minute(), endTime.Second(), 0, loc)
		}

		input := calendar.CreateEventInput{
			Title:     title,
			StartDate: startTime,
			EndDate:   endTime,
			AllDay:    addAllDay,
			Location:  addLocation,
			Notes:     addNotes,
			URL:       addURL,
			Calendar:  addCalendar,
			TimeZone:  addTimezone,
		}

		// Parse alerts
		for _, a := range addAlerts {
			d, err := parser.ParseAlertDuration(a)
			if err != nil {
				return err
			}
			input.Alerts = append(input.Alerts, calendar.Alert{RelativeOffset: -d})
		}

		// Parse recurrence
		if addRepeat != "" {
			rule, err := buildRecurrenceRule()
			if err != nil {
				return err
			}
			input.RecurrenceRules = []eventkit.RecurrenceRule{rule}
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		event, err := client.CreateEvent(input)
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		ui.PrintCreatedEvent(event)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addTitle, "title", "T", "", "Event title")
	addCmd.Flags().StringVarP(&addStart, "start", "s", "", "Start date/time (required)")
	addCmd.Flags().StringVarP(&addEnd, "end", "e", "", "End date/time (default: start + 1h)")
	addCmd.Flags().BoolVarP(&addAllDay, "all-day", "a", false, "Create as all-day event")
	addCmd.Flags().StringVarP(&addCalendar, "calendar", "c", "", "Calendar name")
	addCmd.Flags().StringVarP(&addLocation, "location", "l", "", "Location string")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Notes/description")
	addCmd.Flags().StringVarP(&addURL, "url", "u", "", "URL to attach")
	addCmd.Flags().StringArrayVar(&addAlerts, "alert", nil, "Alert before event (e.g., 15m, 1h, 1d) — repeatable")
	addCmd.Flags().StringVarP(&addRepeat, "repeat", "r", "", "Recurrence: daily, weekly, monthly, yearly")
	addCmd.Flags().IntVar(&addRepeatInterval, "repeat-interval", 1, "Recurrence interval")
	addCmd.Flags().StringVar(&addRepeatUntil, "repeat-until", "", "Recurrence end date")
	addCmd.Flags().IntVar(&addRepeatCount, "repeat-count", 0, "Recurrence occurrence count")
	addCmd.Flags().StringVar(&addRepeatDays, "repeat-days", "", "Days for weekly recurrence (e.g., mon,wed,fri)")
	addCmd.Flags().StringVar(&addTimezone, "timezone", "", "IANA timezone (e.g., America/New_York)")

	rootCmd.AddCommand(addCmd)
}

func buildRecurrenceRule() (eventkit.RecurrenceRule, error) {
	interval := addRepeatInterval
	if interval < 1 {
		interval = 1
	}

	var rule eventkit.RecurrenceRule

	switch strings.ToLower(addRepeat) {
	case "daily":
		rule = eventkit.Daily(interval)
	case "weekly":
		days, err := parseRepeatDays(addRepeatDays)
		if err != nil {
			return rule, err
		}
		rule = eventkit.Weekly(interval, days...)
	case "monthly":
		rule = eventkit.Monthly(interval)
	case "yearly":
		rule = eventkit.Yearly(interval)
	default:
		return rule, fmt.Errorf("invalid repeat value %q (use daily, weekly, monthly, yearly)", addRepeat)
	}

	if addRepeatUntil != "" {
		t, err := parser.ParseDate(addRepeatUntil)
		if err != nil {
			return rule, fmt.Errorf("invalid --repeat-until: %w", err)
		}
		rule = rule.Until(t)
	}

	if addRepeatCount > 0 {
		rule = rule.Count(addRepeatCount)
	}

	return rule, nil
}

var weekdayMap = map[string]eventkit.Weekday{
	"sun": eventkit.Sunday, "sunday": eventkit.Sunday,
	"mon": eventkit.Monday, "monday": eventkit.Monday,
	"tue": eventkit.Tuesday, "tuesday": eventkit.Tuesday,
	"wed": eventkit.Wednesday, "wednesday": eventkit.Wednesday,
	"thu": eventkit.Thursday, "thursday": eventkit.Thursday,
	"fri": eventkit.Friday, "friday": eventkit.Friday,
	"sat": eventkit.Saturday, "saturday": eventkit.Saturday,
}

func parseRepeatDays(s string) ([]eventkit.Weekday, error) {
	if s == "" {
		return nil, nil
	}

	parts := strings.Split(s, ",")
	days := make([]eventkit.Weekday, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		d, ok := weekdayMap[p]
		if !ok {
			return nil, fmt.Errorf("unknown day %q (use mon,tue,wed,thu,fri,sat,sun)", p)
		}
		days = append(days, d)
	}
	return days, nil
}
