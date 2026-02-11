package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/cal/internal/parser"
	"github.com/BRO3886/cal/internal/ui"
	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/charmbracelet/huh"
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
	updateInteractive    bool
)

var updateCmd = &cobra.Command{
	Use:     "update [number or id]",
	Aliases: []string{"edit"},
	Short:   "Update an event",
	Long: `Updates an existing event. Only specified fields are changed.

With no arguments, shows an interactive picker to select the event.
With an argument, accepts a row number from the last listing or a full/partial event ID.
Use -i for interactive mode with guided prompts.`,
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
		} else {
			event, err = pickEvent(client, "", "", 7)
			if err != nil {
				return err
			}
			if event == nil {
				return nil
			}
		}

		if updateInteractive {
			return runUpdateInteractive(client, event)
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
	updateCmd.Flags().BoolVarP(&updateInteractive, "interactive", "i", false, "Interactive mode with guided prompts")

	rootCmd.AddCommand(updateCmd)
}

func runUpdateInteractive(client *calendar.Client, event *calendar.Event) error {
	start := localizeEventTime(event.StartDate, event.TimeZone)
	end := localizeEventTime(event.EndDate, event.TimeZone)

	// Pre-fill with current values
	title := event.Title
	calName := event.Calendar
	startStr := start.Format("2006-01-02 15:04")
	endStr := end.Format("2006-01-02 15:04")
	allDay := event.AllDay
	location := event.Location
	notes := event.Notes
	urlStr := event.URL
	alertStr := buildAlertSummary(event.Alerts)
	if alertStr == "none" {
		alertStr = ""
	}

	// Fetch calendars for selection
	cals, err := client.Calendars()
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}
	calOpts := buildCalendarOptions(cals, event.Calendar)
	if len(calOpts) == 0 {
		return fmt.Errorf("no writable calendars found")
	}

	fmt.Printf("Editing: %s (ID: %s)\n\n", event.Title, ui.ShortID(event.ID))

	// Page 1: Core fields
	core := huh.NewGroup(
		huh.NewInput().
			Title("Title").
			Value(&title).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("title is required")
				}
				return nil
			}),

		huh.NewSelect[string]().
			Title("Calendar").
			Options(calOpts...).
			Value(&calName),

		huh.NewInput().
			Title("Start").
			Description("Natural language or '2006-01-02 15:04' format").
			Value(&startStr).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("start date is required")
				}
				_, err := parser.ParseDate(s)
				if err != nil {
					return fmt.Errorf("invalid date: %v", err)
				}
				return nil
			}),

		huh.NewInput().
			Title("End").
			Value(&endStr).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return nil
				}
				_, err := parser.ParseDate(s)
				if err != nil {
					return fmt.Errorf("invalid date: %v", err)
				}
				return nil
			}),

		huh.NewConfirm().
			Title("All day?").
			Value(&allDay),
	)

	// Page 2: Details
	details := huh.NewGroup(
		huh.NewInput().
			Title("Location").
			Description("Type - to clear").
			Value(&location),

		huh.NewInput().
			Title("Notes").
			Description("Type - to clear").
			Value(&notes),

		huh.NewInput().
			Title("URL").
			Description("Type - to clear").
			Value(&urlStr),

		huh.NewInput().
			Title("Alerts").
			Description("Comma-separated, e.g., '15m, 1h'. Type 'none' to clear.").
			Value(&alertStr),
	)

	// Page 3: Recurring event span
	spanVal := "this"
	var groups []*huh.Group
	groups = append(groups, core, details)

	if event.Recurring {
		spanGroup := huh.NewGroup(
			huh.NewSelect[string]().
				Title("Apply changes to").
				Description("This is a recurring event").
				Options(
					huh.NewOption("This event only", "this"),
					huh.NewOption("This and future events", "future"),
				).
				Value(&spanVal),
		)
		groups = append(groups, spanGroup)
	}

	form := huh.NewForm(groups...).
		WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	// Build update input — only set fields that changed
	input := calendar.UpdateEventInput{}

	if title != event.Title {
		input.Title = strPtr(title)
	}

	if calName != event.Calendar {
		input.Calendar = strPtr(calName)
	}

	newStart, _ := parser.ParseDate(startStr) // validated above
	if !newStart.Equal(event.StartDate) {
		input.StartDate = &newStart
	}

	if strings.TrimSpace(endStr) != "" {
		newEnd, _ := parser.ParseDate(endStr) // validated above
		if !newEnd.Equal(event.EndDate) {
			input.EndDate = &newEnd
		}
	}

	if allDay != event.AllDay {
		input.AllDay = &allDay
	}

	// Handle clearable fields (- to clear)
	if location == "-" {
		empty := ""
		input.Location = &empty
	} else if location != event.Location {
		input.Location = strPtr(location)
	}

	if notes == "-" {
		empty := ""
		input.Notes = &empty
	} else if notes != event.Notes {
		input.Notes = strPtr(notes)
	}

	if urlStr == "-" {
		empty := ""
		input.URL = &empty
	} else if urlStr != event.URL {
		input.URL = strPtr(urlStr)
	}

	// Parse alerts if changed
	alertStr = strings.TrimSpace(alertStr)
	currentAlertStr := buildAlertSummary(event.Alerts)
	if currentAlertStr == "none" {
		currentAlertStr = ""
	}
	if alertStr != currentAlertStr {
		if alertStr == "" || strings.ToLower(alertStr) == "none" {
			empty := []calendar.Alert{}
			input.Alerts = &empty
		} else {
			alerts := make([]calendar.Alert, 0)
			for _, a := range strings.Split(alertStr, ",") {
				a = strings.TrimSpace(a)
				if a == "" {
					continue
				}
				d, err := parser.ParseAlertDuration(a)
				if err != nil {
					return err
				}
				alerts = append(alerts, calendar.Alert{RelativeOffset: -d})
			}
			input.Alerts = &alerts
		}
	}

	span := calendar.SpanThisEvent
	if spanVal == "future" {
		span = calendar.SpanFutureEvents
	}

	updated, err := client.UpdateEvent(event.ID, input, span)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	fmt.Println()
	ui.PrintUpdatedEvent(updated)
	return nil
}

// buildAlertSummary returns a human-readable summary of current alerts.
func buildAlertSummary(alerts []calendar.Alert) string {
	if len(alerts) == 0 {
		return "none"
	}
	parts := make([]string, len(alerts))
	for i, a := range alerts {
		d := a.RelativeOffset
		if d < 0 {
			d = -d
		}
		parts[i] = formatAlertDurationShort(d) + " before"
	}
	return strings.Join(parts, ", ")
}

// formatAlertDurationShort returns a compact alert duration string like "15m", "1h", "1d".
func formatAlertDurationShort(d time.Duration) string {
	if d >= 24*time.Hour {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if d >= time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh", hours)
	}
	mins := int(d.Minutes())
	return fmt.Sprintf("%dm", mins)
}

// localizeEventTime converts event time to local using the event's timezone.
func localizeEventTime(t time.Time, tz string) time.Time {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return t.In(loc)
		}
	}
	return t.In(time.Local)
}

func strPtr(s string) *string {
	return &s
}
