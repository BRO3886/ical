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
	updateTitle          string
	updateStart          string
	updateEnd            string
	updateAllDay         string
	updateCalendar       string
	updateLocation       string
	updateNotes          string
	updateURL            string
	updateAlerts         []string
	updateTimezone       string
	updateSpan           string
	updateRepeat         string
	updateRepeatInterval int
	updateRepeatUntil    string
	updateRepeatCount    int
	updateRepeatDays     string
)

var updateCmd = &cobra.Command{
	Use:     "update [id]",
	Aliases: []string{"edit"},
	Short:   "Update an event",
	Long:    "Updates an existing event. Only specified fields are changed.",
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

		input := calendar.UpdateEventInput{}

		if cmd.Flags().Changed("title") {
			input.Title = strPtr(updateTitle)
		}
		if cmd.Flags().Changed("start") {
			t, err := parser.ParseDate(updateStart)
			if err != nil {
				return fmt.Errorf("invalid --start: %w", err)
			}
			input.StartDate = &t
		}
		if cmd.Flags().Changed("end") {
			t, err := parser.ParseDate(updateEnd)
			if err != nil {
				return fmt.Errorf("invalid --end: %w", err)
			}
			input.EndDate = &t
		}
		if cmd.Flags().Changed("all-day") {
			b := updateAllDay == "true"
			input.AllDay = &b
		}
		if cmd.Flags().Changed("calendar") {
			input.Calendar = strPtr(updateCalendar)
		}
		if cmd.Flags().Changed("location") {
			input.Location = strPtr(updateLocation)
		}
		if cmd.Flags().Changed("notes") {
			input.Notes = strPtr(updateNotes)
		}
		if cmd.Flags().Changed("url") {
			input.URL = strPtr(updateURL)
		}
		if cmd.Flags().Changed("timezone") {
			if updateTimezone != "" {
				if _, err := time.LoadLocation(updateTimezone); err != nil {
					return fmt.Errorf("invalid timezone %q: %w", updateTimezone, err)
				}
			}
			input.TimeZone = strPtr(updateTimezone)
		}
		if cmd.Flags().Changed("alert") {
			if len(updateAlerts) == 1 && updateAlerts[0] == "none" {
				empty := []calendar.Alert{}
				input.Alerts = &empty
			} else {
				alerts := make([]calendar.Alert, 0, len(updateAlerts))
				for _, a := range updateAlerts {
					d, err := parser.ParseAlertDuration(a)
					if err != nil {
						return err
					}
					alerts = append(alerts, calendar.Alert{RelativeOffset: -d})
				}
				input.Alerts = &alerts
			}
		}
		if cmd.Flags().Changed("repeat") {
			if strings.ToLower(updateRepeat) == "none" {
				empty := []eventkit.RecurrenceRule{}
				input.RecurrenceRules = &empty
			} else {
				// Temporarily set the add flags for buildRecurrenceRule
				addRepeat = updateRepeat
				addRepeatInterval = updateRepeatInterval
				addRepeatUntil = updateRepeatUntil
				addRepeatCount = updateRepeatCount
				addRepeatDays = updateRepeatDays
				rule, err := buildRecurrenceRule()
				if err != nil {
					return err
				}
				rules := []eventkit.RecurrenceRule{rule}
				input.RecurrenceRules = &rules
			}
		}

		span := calendar.SpanThisEvent
		if updateSpan == "future" {
			span = calendar.SpanFutureEvents
		}

		updated, err := client.UpdateEvent(event.ID, input, span)
		if err != nil {
			return fmt.Errorf("failed to update event: %w", err)
		}

		ui.PrintUpdatedEvent(updated)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateTitle, "title", "T", "", "New title")
	updateCmd.Flags().StringVarP(&updateStart, "start", "s", "", "New start date/time")
	updateCmd.Flags().StringVarP(&updateEnd, "end", "e", "", "New end date/time")
	updateCmd.Flags().StringVarP(&updateAllDay, "all-day", "a", "", "Set all-day (true/false)")
	updateCmd.Flags().StringVarP(&updateCalendar, "calendar", "c", "", "Move to calendar (by name)")
	updateCmd.Flags().StringVarP(&updateLocation, "location", "l", "", "New location (empty to clear)")
	updateCmd.Flags().StringVarP(&updateNotes, "notes", "n", "", "New notes (empty to clear)")
	updateCmd.Flags().StringVarP(&updateURL, "url", "u", "", "New URL (empty to clear)")
	updateCmd.Flags().StringArrayVar(&updateAlerts, "alert", nil, "Replace alerts (repeatable, 'none' to clear)")
	updateCmd.Flags().StringVar(&updateTimezone, "timezone", "", "New timezone")
	updateCmd.Flags().StringVar(&updateSpan, "span", "this", "For recurring events: this or future")
	updateCmd.Flags().StringVarP(&updateRepeat, "repeat", "r", "", "Set/change recurrence (none to remove)")
	updateCmd.Flags().IntVar(&updateRepeatInterval, "repeat-interval", 1, "Change recurrence interval")
	updateCmd.Flags().StringVar(&updateRepeatUntil, "repeat-until", "", "Change recurrence end date")
	updateCmd.Flags().IntVar(&updateRepeatCount, "repeat-count", 0, "Change recurrence count")
	updateCmd.Flags().StringVar(&updateRepeatDays, "repeat-days", "", "Change recurrence days")

	rootCmd.AddCommand(updateCmd)
}

func strPtr(s string) *string {
	return &s
}
