package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/ical/internal/parser"
	"github.com/BRO3886/ical/internal/ui"
	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/charmbracelet/huh"
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
	addInteractive    bool
)

var addCmd = &cobra.Command{
	Use:     "add [title]",
	Aliases: []string{"create", "new"},
	Short:   "Create a new event",
	Long:    "Creates a new calendar event. Title can be passed as argument or via --title flag.\nUse -i for interactive mode with guided prompts.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addInteractive {
			return runAddInteractive()
		}

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
	addCmd.Flags().BoolVarP(&addInteractive, "interactive", "i", false, "Interactive mode with guided prompts")

	rootCmd.AddCommand(addCmd)
}

func runAddInteractive() error {
	client, err := calendar.New()
	if err != nil {
		return handleClientError(err)
	}

	// Fetch calendars for selection
	cals, err := client.Calendars()
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}
	calOpts := buildCalendarOptions(cals, "")
	if len(calOpts) == 0 {
		return fmt.Errorf("no writable calendars found")
	}

	var (
		title    string
		calName  string
		startStr string
		endStr   string
		allDay   bool
		location string
		notes    string
		urlStr   string
		alertStr string
		tz       string
	)

	// Page 1: Essential details
	essentials := huh.NewGroup(
		huh.NewInput().
			Title("Title").
			Placeholder("Meeting with team").
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
			Description("e.g., 'tomorrow 2pm', '2026-03-15 14:00'").
			Placeholder("tomorrow 2pm").
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
			Description("Leave empty for start + 1 hour").
			Placeholder("tomorrow 3pm").
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

	// Page 2: Optional details
	details := huh.NewGroup(
		huh.NewInput().
			Title("Location").
			Placeholder("optional").
			Value(&location),

		huh.NewInput().
			Title("Notes").
			Placeholder("optional").
			Value(&notes),

		huh.NewInput().
			Title("URL").
			Placeholder("optional").
			Value(&urlStr),

		huh.NewInput().
			Title("Alerts").
			Description("Comma-separated, e.g., '15m, 1h'. Leave empty for none.").
			Placeholder("15m").
			Value(&alertStr),

		huh.NewInput().
			Title("Timezone").
			Description("IANA timezone, leave empty for system default").
			Placeholder("America/New_York").
			Value(&tz).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return nil
				}
				_, err := time.LoadLocation(s)
				if err != nil {
					return fmt.Errorf("invalid timezone: %v", err)
				}
				return nil
			}),
	)

	// Page 3: Recurrence
	var (
		repeat       string
		repeatDays   string
	)

	recurrence := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Repeat").
			Options(
				huh.NewOption("None", ""),
				huh.NewOption("Daily", "daily"),
				huh.NewOption("Weekly", "weekly"),
				huh.NewOption("Monthly", "monthly"),
				huh.NewOption("Yearly", "yearly"),
			).
			Value(&repeat),

		huh.NewInput().
			Title("Repeat days (for weekly)").
			Description("e.g., mon,wed,fri — leave empty to skip").
			Placeholder("mon,wed,fri").
			Value(&repeatDays),
	)

	form := huh.NewForm(essentials, details, recurrence).
		WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	// Build event input
	startTime, _ := parser.ParseDate(startStr) // validated above

	endTime := startTime.Add(time.Hour)
	if allDay {
		endTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day()+1,
			0, 0, 0, 0, startTime.Location())
	} else if strings.TrimSpace(endStr) != "" {
		endTime, _ = parser.ParseDate(endStr) // validated above
	}

	if tz != "" {
		loc, _ := time.LoadLocation(tz) // validated above
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(),
			startTime.Hour(), startTime.Minute(), startTime.Second(), 0, loc)
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(),
			endTime.Hour(), endTime.Minute(), endTime.Second(), 0, loc)
	}

	input := calendar.CreateEventInput{
		Title:     title,
		StartDate: startTime,
		EndDate:   endTime,
		AllDay:    allDay,
		Location:  location,
		Notes:     notes,
		URL:       urlStr,
		Calendar:  calName,
		TimeZone:  tz,
	}

	// Parse alerts
	if strings.TrimSpace(alertStr) != "" {
		for _, a := range strings.Split(alertStr, ",") {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			d, err := parser.ParseAlertDuration(a)
			if err != nil {
				return err
			}
			input.Alerts = append(input.Alerts, calendar.Alert{RelativeOffset: -d})
		}
	}

	// Parse recurrence
	if repeat != "" {
		addRepeat = repeat
		addRepeatInterval = 1
		addRepeatDays = repeatDays
		addRepeatUntil = ""
		addRepeatCount = 0
		rule, err := buildRecurrenceRule()
		if err != nil {
			return err
		}
		input.RecurrenceRules = []eventkit.RecurrenceRule{rule}
	}

	event, err := client.CreateEvent(input)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	fmt.Println()
	ui.PrintCreatedEvent(event)
	return nil
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

// buildCalendarOptions builds huh.Option list from writable calendars.
func buildCalendarOptions(calendars []calendar.Calendar, current string) []huh.Option[string] {
	var opts []huh.Option[string]
	for _, c := range calendars {
		if c.ReadOnly {
			continue
		}
		label := fmt.Sprintf("%s (%s)", c.Title, c.Source)
		opt := huh.NewOption(label, c.Title)
		if c.Title == current {
			opt = opt.Selected(true)
		}
		opts = append(opts, opt)
	}
	return opts
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
