package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/BRO3886/go-eventkit/dateparser"
	"github.com/BRO3886/ical/internal/ui"
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
	addNoAlert        bool
	addRepeat         string
	addRepeatInterval int
	addRepeatUntil    string
	addRepeatCount    int
	addRepeatDays     string
	addTimezone       string
	addInteractive    bool
	addInvite         []string
	addTravel         string
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

		startTime, err := dateparser.ParseDate(addStart)
		if err != nil {
			return fmt.Errorf("invalid --start date: %w", err)
		}

		endTime := startTime.Add(time.Hour)
		if addEnd != "" {
			endTime, err = dateparser.ParseDate(addEnd)
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
			d, err := dateparser.ParseAlertDuration(a)
			if err != nil {
				return err
			}
			input.Alerts = append(input.Alerts, calendar.Alert{RelativeOffset: -d})
		}

		// Any explicit --alert suppresses the calendar default; --no-alert
		// suppresses it explicitly so a user can create a zero-alert event
		// on a calendar that has a default configured.
		if addNoAlert || len(input.Alerts) > 0 {
			input.SuppressDefaultAlarms = true
		}

		// Parse recurrence
		if addRepeat != "" {
			rule, err := buildRecurrenceRule(addTimezone)
			if err != nil {
				return err
			}
			input.RecurrenceRules = []eventkit.RecurrenceRule{rule}
		}

		// Parse invitees and travel time.
		attendees, err := parseAttendees(addInvite)
		if err != nil {
			return err
		}
		input.Attendees = attendees
		if addTravel != "" {
			d, err := dateparser.ParseAlertDuration(addTravel)
			if err != nil {
				return fmt.Errorf("invalid --travel duration: %w", err)
			}
			input.TravelTime = d
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		if len(input.Attendees) > 0 && !client.AttendeeWritesSupported() {
			return fmt.Errorf("inviting attendees is not supported on this macOS version")
		}

		event, err := client.CreateEvent(input)
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		ui.PrintCreatedEvent(event)
		if len(input.Attendees) > 0 {
			fmt.Printf("Invited %d attendee(s); invitations are sent by the calendar account.\n", len(input.Attendees))
		}
		return nil
	},
}

// parseAttendees converts --invite values into attendee inputs. Each value is
// either a bare email ("a@x.com") or a named form ("Alice <a@x.com>"). It
// returns an error for entries with no usable email address.
func parseAttendees(invites []string) ([]calendar.AttendeeInput, error) {
	if len(invites) == 0 {
		return nil, nil
	}
	out := make([]calendar.AttendeeInput, 0, len(invites))
	for _, raw := range invites {
		name, email := splitNameEmail(raw)
		if !looksLikeEmail(email) {
			return nil, fmt.Errorf("invalid --invite value %q (expected an email or \"Name <email>\"; pass --invite once per person)", raw)
		}
		out = append(out, calendar.AttendeeInput{Name: name, Email: email})
	}
	return out, nil
}

// looksLikeEmail does a deliberately permissive check: exactly one "@" with
// non-empty local and domain parts, a dot in the domain, and no internal
// whitespace or commas. The comma/space guard catches the common mistake of
// packing several addresses into one --invite value, which would otherwise be
// saved as a single malformed attendee.
func looksLikeEmail(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t,;") {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at != strings.LastIndexByte(s, '@') {
		return false
	}
	domain := s[at+1:]
	return len(domain) >= 3 && strings.Contains(domain, ".") &&
		!strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".")
}

// splitNameEmail parses "Name <email>" or a bare email. The returned name is
// empty for the bare form.
func splitNameEmail(raw string) (name, email string) {
	s := strings.TrimSpace(raw)
	if i := strings.LastIndex(s, "<"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j >= 0 {
			email = strings.TrimSpace(s[i+1 : i+j])
			name = strings.TrimSpace(s[:i])
			return name, email
		}
	}
	return "", s
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
	addCmd.Flags().BoolVar(&addNoAlert, "no-alert", false, "Create with zero alerts (overrides calendar default alerts)")
	addCmd.Flags().StringVarP(&addRepeat, "repeat", "r", "", "Recurrence: daily, weekly, monthly, yearly")
	addCmd.Flags().IntVar(&addRepeatInterval, "repeat-interval", 1, "Recurrence interval")
	addCmd.Flags().StringVar(&addRepeatUntil, "repeat-until", "", "Recurrence end date")
	addCmd.Flags().IntVar(&addRepeatCount, "repeat-count", 0, "Recurrence occurrence count")
	addCmd.Flags().StringVar(&addRepeatDays, "repeat-days", "", "Days for weekly recurrence (e.g., mon,wed,fri)")
	addCmd.Flags().StringVar(&addTimezone, "timezone", "", "IANA timezone (e.g., America/New_York)")
	addCmd.Flags().StringArrayVar(&addInvite, "invite", nil, "Invite an attendee by email or \"Name <email>\" — repeatable (sends an invitation)")
	addCmd.Flags().StringVar(&addTravel, "travel", "", "Travel time before the event (e.g., 30m, 1h)")
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
				_, err := dateparser.ParseDate(s)
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
				_, err := dateparser.ParseDate(s)
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
		repeat     string
		repeatDays string
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
	startTime, _ := dateparser.ParseDate(startStr) // validated above

	endTime := startTime.Add(time.Hour)
	if allDay {
		endTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day()+1,
			0, 0, 0, 0, startTime.Location())
	} else if strings.TrimSpace(endStr) != "" {
		endTime, _ = dateparser.ParseDate(endStr) // validated above
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
			d, err := dateparser.ParseAlertDuration(a)
			if err != nil {
				return err
			}
			input.Alerts = append(input.Alerts, calendar.Alert{RelativeOffset: -d})
		}
	}

	// Matches the non-interactive path: any explicit alert suppresses the
	// calendar default, and the --no-alert CLI flag (if passed together
	// with -i) forces zero alerts.
	if addNoAlert || len(input.Alerts) > 0 {
		input.SuppressDefaultAlarms = true
	}

	// Parse recurrence
	if repeat != "" {
		addRepeat = repeat
		addRepeatInterval = 1
		addRepeatDays = repeatDays
		addRepeatUntil = ""
		addRepeatCount = 0
		rule, err := buildRecurrenceRule(tz)
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

// buildRecurrenceRule builds the recurrence rule from the addRepeat* globals.
// timezone is the event's IANA zone (may be empty) and is used only to compute
// the end-of-day --repeat-until bound in the right zone; callers must pass it
// explicitly since this is shared by add and update, which track the zone in
// different globals.
func buildRecurrenceRule(timezone string) (eventkit.RecurrenceRule, error) {
	interval := addRepeatInterval
	if interval < 1 {
		interval = 1
	}

	freq := strings.ToLower(addRepeat)
	if addRepeatDays != "" && freq != "weekly" {
		return eventkit.RecurrenceRule{}, fmt.Errorf("--repeat-days is only valid with --repeat weekly, got --repeat %s", freq)
	}

	var rule eventkit.RecurrenceRule

	switch freq {
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
		t, err := dateparser.ParseDate(addRepeatUntil)
		if err != nil {
			return rule, fmt.Errorf("invalid --repeat-until: %w", err)
		}
		var loc *time.Location
		if timezone != "" {
			loc, err = time.LoadLocation(timezone)
			if err != nil {
				return rule, fmt.Errorf("invalid timezone %q: %w", timezone, err)
			}
		}
		rule = rule.Until(repeatUntilBound(t, loc))
	}

	if addRepeatCount > 0 {
		rule = rule.Count(addRepeatCount)
	}

	if err := rule.Validate(); err != nil {
		return rule, fmt.Errorf("invalid recurrence rule: %w", err)
	}

	return rule, nil
}

// repeatUntilBound returns the inclusive UNTIL bound for a parsed --repeat-until
// value. A date-only value parses to midnight; an RRULE UNTIL at 00:00 excludes
// any occurrence later that same day, so the series ends a day early (issue #39).
// Snapping a midnight bound to end-of-day keeps an occurrence that falls anywhere
// on the named calendar day. When the event carries an explicit timezone,
// end-of-day is computed in that zone so the absolute instant lines up with the
// occurrence's wall-clock day.
func repeatUntilBound(t time.Time, loc *time.Location) time.Time {
	if loc != nil {
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
	}
	return endOfDayIfMidnight(t)
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
