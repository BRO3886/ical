package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BRO3886/cal/internal/ui"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// --- calendars (parent) ---

var calendarsCmd = &cobra.Command{
	Use:     "calendars",
	Aliases: []string{"cals"},
	Short:   "Manage calendars",
	Long:    "List, create, update, and delete calendars. Running without a subcommand lists all calendars.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: list calendars (backwards-compatible)
		return calendarsListCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(calendarsCmd)
}

func handleClientError(err error) error {
	if err.Error() == "calendar: access denied" {
		fmt.Fprintln(os.Stderr, "Calendar access denied. Grant access in System Settings > Privacy & Security > Calendars")
		os.Exit(1)
	}
	return fmt.Errorf("failed to initialize calendar client: %w", err)
}

// --- calendars list ---

var calendarsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all calendars",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		cals, err := client.Calendars()
		if err != nil {
			return fmt.Errorf("failed to list calendars: %w", err)
		}

		ui.PrintCalendars(cals, outputFormat)
		return nil
	},
}

func init() {
	calendarsCmd.AddCommand(calendarsListCmd)
}

// --- calendars create ---

var (
	calCreateTitle       string
	calCreateSource      string
	calCreateColor       string
	calCreateInteractive bool
)

var calendarsCreateCmd = &cobra.Command{
	Use:     "create [title]",
	Aliases: []string{"add", "new"},
	Short:   "Create a new calendar",
	Long: `Creates a new calendar. Title can be passed as argument or via --title flag.

Requires --source to specify the account (e.g., "iCloud", "Gmail").
Run 'cal calendars' to see available sources from existing calendars.

Use -i for interactive mode with guided prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if calCreateInteractive {
			return runCalCreateInteractive()
		}

		title := calCreateTitle
		if title == "" && len(args) > 0 {
			title = strings.Join(args, " ")
		}
		if title == "" {
			return fmt.Errorf("title is required (pass as argument or use --title)")
		}
		if calCreateSource == "" {
			return fmt.Errorf("--source is required (e.g., 'iCloud'). Run 'cal calendars' to see available sources")
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		input := calendar.CreateCalendarInput{
			Title:  title,
			Source: calCreateSource,
			Color:  calCreateColor,
		}

		cal, err := client.CreateCalendar(input)
		if err != nil {
			return fmt.Errorf("failed to create calendar: %w", err)
		}

		ui.PrintCreatedCalendar(cal)
		return nil
	},
}

func init() {
	calendarsCreateCmd.Flags().StringVarP(&calCreateTitle, "title", "T", "", "Calendar title")
	calendarsCreateCmd.Flags().StringVarP(&calCreateSource, "source", "s", "", "Account source (e.g., 'iCloud')")
	calendarsCreateCmd.Flags().StringVar(&calCreateColor, "color", "", "Calendar color (hex, e.g., '#FF6961')")
	calendarsCreateCmd.Flags().BoolVarP(&calCreateInteractive, "interactive", "i", false, "Interactive mode with guided prompts")

	calendarsCmd.AddCommand(calendarsCreateCmd)
}

func runCalCreateInteractive() error {
	client, err := calendar.New()
	if err != nil {
		return handleClientError(err)
	}

	// Discover available sources from existing calendars
	cals, err := client.Calendars()
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}
	sourceOpts := buildSourceOptions(cals)
	if len(sourceOpts) == 0 {
		return fmt.Errorf("no writable calendar sources found")
	}

	var (
		title  string
		source string
		clr    string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Calendar Name").
				Placeholder("Work Projects").
				Value(&title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title is required")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Account").
				Description("Which account should own this calendar?").
				Options(sourceOpts...).
				Value(&source),

			huh.NewInput().
				Title("Color").
				Description("Hex color (e.g., '#FF6961'). Leave empty for default.").
				Placeholder("#FF6961").
				Value(&clr).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return nil
					}
					if !strings.HasPrefix(s, "#") || (len(s) != 7 && len(s) != 4) {
						return fmt.Errorf("invalid hex color (use #RGB or #RRGGBB)")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	input := calendar.CreateCalendarInput{
		Title:  title,
		Source: source,
		Color:  clr,
	}

	cal, err := client.CreateCalendar(input)
	if err != nil {
		return fmt.Errorf("failed to create calendar: %w", err)
	}

	fmt.Println()
	ui.PrintCreatedCalendar(cal)
	return nil
}

// buildSourceOptions extracts unique source names from writable calendars
// and returns huh.Option list for interactive selection.
func buildSourceOptions(calendars []calendar.Calendar) []huh.Option[string] {
	seen := make(map[string]bool)
	var opts []huh.Option[string]
	for _, c := range calendars {
		if c.ReadOnly {
			continue
		}
		if seen[c.Source] {
			continue
		}
		seen[c.Source] = true
		opts = append(opts, huh.NewOption(c.Source, c.Source))
	}
	return opts
}

// --- calendars update ---

var (
	calUpdateTitle       string
	calUpdateColor       string
	calUpdateInteractive bool
)

var calendarsUpdateCmd = &cobra.Command{
	Use:     "update [calendar name]",
	Aliases: []string{"edit", "rename"},
	Short:   "Update a calendar",
	Long: `Updates an existing calendar. Only specified fields are changed.

With no arguments, shows an interactive picker to select the calendar.
Use -i for interactive mode with guided prompts.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		cal, err := pickCalendar(client, args)
		if err != nil {
			return err
		}
		if cal == nil {
			return nil // user cancelled
		}

		if calUpdateInteractive {
			return runCalUpdateInteractive(client, cal)
		}

		input := calendar.UpdateCalendarInput{}
		changed := false

		if cmd.Flags().Changed("title") {
			input.Title = strPtr(calUpdateTitle)
			changed = true
		}
		if cmd.Flags().Changed("color") {
			input.Color = strPtr(calUpdateColor)
			changed = true
		}

		if !changed {
			return fmt.Errorf("nothing to update. Use --title and/or --color, or -i for interactive mode")
		}

		updated, err := client.UpdateCalendar(cal.ID, input)
		if err != nil {
			if errors.Is(err, calendar.ErrImmutable) {
				return fmt.Errorf("calendar %q is immutable (subscribed or birthday calendar)", cal.Title)
			}
			return fmt.Errorf("failed to update calendar: %w", err)
		}

		ui.PrintUpdatedCalendar(updated)
		return nil
	},
}

func init() {
	calendarsUpdateCmd.Flags().StringVarP(&calUpdateTitle, "title", "T", "", "New calendar title")
	calendarsUpdateCmd.Flags().StringVar(&calUpdateColor, "color", "", "New calendar color (hex, e.g., '#FF6961')")
	calendarsUpdateCmd.Flags().BoolVarP(&calUpdateInteractive, "interactive", "i", false, "Interactive mode with guided prompts")

	calendarsCmd.AddCommand(calendarsUpdateCmd)
}

func runCalUpdateInteractive(client *calendar.Client, cal *calendar.Calendar) error {
	title := cal.Title
	clr := cal.Color

	fmt.Printf("Editing calendar: %s (%s)\n\n", cal.Title, cal.Source)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Calendar Name").
				Value(&title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Color").
				Description("Hex color (e.g., '#FF6961'). Leave unchanged to keep current.").
				Value(&clr).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return nil
					}
					if !strings.HasPrefix(s, "#") || (len(s) != 7 && len(s) != 4) {
						return fmt.Errorf("invalid hex color (use #RGB or #RRGGBB)")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	input := calendar.UpdateCalendarInput{}
	changed := false

	if title != cal.Title {
		input.Title = strPtr(title)
		changed = true
	}
	if clr != cal.Color {
		input.Color = strPtr(clr)
		changed = true
	}

	if !changed {
		fmt.Println("No changes made.")
		return nil
	}

	updated, err := client.UpdateCalendar(cal.ID, input)
	if err != nil {
		if errors.Is(err, calendar.ErrImmutable) {
			return fmt.Errorf("calendar %q is immutable (subscribed or birthday calendar)", cal.Title)
		}
		return fmt.Errorf("failed to update calendar: %w", err)
	}

	fmt.Println()
	ui.PrintUpdatedCalendar(updated)
	return nil
}

// --- calendars delete ---

var (
	calDeleteForce bool
)

var calendarsDeleteCmd = &cobra.Command{
	Use:     "delete [calendar name]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a calendar",
	Long: `Permanently deletes a calendar and all its events.

With no arguments, shows an interactive picker to select the calendar.
Asks for confirmation by default. Use --force to skip.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		cal, err := pickCalendar(client, args)
		if err != nil {
			return err
		}
		if cal == nil {
			return nil // user cancelled
		}

		if !calDeleteForce {
			red := color.New(color.FgRed, color.Bold)
			red.Printf("Delete calendar: ")
			fmt.Printf("%s (%s)\n", cal.Title, cal.Source)
			fmt.Println("This will permanently delete the calendar and ALL its events.")
			fmt.Print("Are you sure? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := client.DeleteCalendar(cal.ID); err != nil {
			if errors.Is(err, calendar.ErrImmutable) {
				return fmt.Errorf("calendar %q is immutable (subscribed or birthday calendar)", cal.Title)
			}
			return fmt.Errorf("failed to delete calendar: %w", err)
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("Deleted: ")
		fmt.Printf("%s\n", cal.Title)
		return nil
	},
}

func init() {
	calendarsDeleteCmd.Flags().BoolVarP(&calDeleteForce, "force", "f", false, "Skip confirmation prompt")

	calendarsCmd.AddCommand(calendarsDeleteCmd)
}

// --- shared helpers ---

// pickCalendar finds a calendar by name argument or shows an interactive picker.
func pickCalendar(client *calendar.Client, args []string) (*calendar.Calendar, error) {
	cals, err := client.Calendars()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	if len(args) == 1 {
		name := args[0]
		for _, c := range cals {
			if strings.EqualFold(c.Title, name) {
				return &c, nil
			}
		}
		return nil, fmt.Errorf("no calendar found with name %q", name)
	}

	// Interactive picker — only show writable calendars
	writable := make([]calendar.Calendar, 0)
	for _, c := range cals {
		if !c.ReadOnly {
			writable = append(writable, c)
		}
	}
	if len(writable) == 0 {
		return nil, fmt.Errorf("no writable calendars found")
	}

	options := make([]huh.Option[string], len(writable))
	for i, c := range writable {
		label := fmt.Sprintf("%s (%s) %s", c.Title, c.Source, c.Color)
		options[i] = huh.NewOption(label, c.ID)
	}

	var selectedID string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a calendar").
				Options(options...).
				Value(&selectedID),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("selection error: %w", err)
	}

	for _, c := range writable {
		if c.ID == selectedID {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("calendar not found")
}
