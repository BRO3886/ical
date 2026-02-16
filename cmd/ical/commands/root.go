package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/ical/internal/skills"
	"github.com/BRO3886/ical/internal/update"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	noColor      bool
)

// updateResult receives the background update check result (if any).
var updateResultCh = make(chan *update.Result, 1)

var rootCmd = &cobra.Command{
	Use:   "ical",
	Short: "A fast, native macOS Calendar CLI",
	Long:  "ical — a fast, native macOS Calendar CLI built on EventKit.\nProvides full CRUD for calendar events, natural language dates,\nrecurrence support, import/export, and multiple output formats.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}

		// Start background update check
		if shouldCheckForUpdate(cmd) {
			go func() {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					updateResultCh <- nil
					return
				}
				updateResultCh <- update.Check(homeDir, versionStr)
			}()
		} else {
			updateResultCh <- nil
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		printUpdateNotice(cmd)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, plain")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

func Execute() error {
	return rootCmd.Execute()
}

// shouldCheckForUpdate returns false for commands/contexts where the check should be skipped.
func shouldCheckForUpdate(cmd *cobra.Command) bool {
	// Skip if env var set
	if os.Getenv("ICAL_NO_UPDATE_CHECK") != "" {
		return false
	}

	// Skip for dev builds
	if versionStr == "" || versionStr == "dev" {
		return false
	}

	// Skip for meta commands
	name := cmd.Name()
	if name == "version" || name == "completion" || name == "skills" {
		return false
	}

	// Skip if --output json (scripting context)
	if outputFormat == "json" {
		return false
	}

	// Skip if stdout is not a TTY (piped output)
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}

	return true
}

// printUpdateNotice prints update and skills staleness notices to stderr.
func printUpdateNotice(_ *cobra.Command) {
	// Collect update result (non-blocking — if goroutine isn't done, skip)
	var result *update.Result
	select {
	case result = <-updateResultCh:
	default:
		// Goroutine still running, don't wait
		result = nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	yellow := color.New(color.FgYellow)

	if result != nil && result.HasUpdate {
		fmt.Fprintln(os.Stderr)
		yellow.Fprintf(os.Stderr, "A new version of ical is available: %s → %s\n", versionStr, result.Latest)
		fmt.Fprintf(os.Stderr, "Update: curl -fsSL https://ical.sidv.dev/install | bash\n")
	}

	// Check skills staleness (local only, no HTTP)
	printSkillsStalenessNotice(homeDir)
}

// printSkillsStalenessNotice checks if installed skills are outdated.
func printSkillsStalenessNotice(homeDir string) {
	if versionStr == "" || versionStr == "dev" {
		return
	}

	targets := skills.InstalledTargets(skills.DefaultTargets(homeDir))
	for _, t := range targets {
		installed := skills.InstalledVersion(t)
		if installed != "" && installed != versionStr {
			yellow := color.New(color.FgYellow)
			fmt.Fprintln(os.Stderr)
			yellow.Fprintf(os.Stderr, "Installed skills are outdated (%s). Run: ical skills install\n", installed)
			return // Only show once
		}
	}
}
