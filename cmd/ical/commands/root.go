package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

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
		if shouldShowNotices(currentNoticeConditions(cmd)) {
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
		if !shouldShowNotices(currentNoticeConditions(cmd)) {
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return
		}

		printNotices(os.Stderr, versionStr, collectUpdateResult(updateResultCh), homeDir)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, plain")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

func Execute() error {
	return rootCmd.Execute()
}

// metaCommands never emit post-run notices: their whole output is either
// machine-consumed (completion) or is itself a version/skills report.
var metaCommands = map[string]bool{"version": true, "completion": true, "skills": true}

// noticeConditions is the full input to shouldShowNotices. Keeping it a plain
// value lets the rule be driven from a table instead of package globals.
type noticeConditions struct {
	commandPath  string
	outputFormat string
	version      string
	suppressed   bool
	interactive  bool
}

// currentNoticeConditions samples the process state that shouldShowNotices needs.
func currentNoticeConditions(cmd *cobra.Command) noticeConditions {
	return noticeConditions{
		commandPath:  cmd.CommandPath(),
		outputFormat: outputFormat,
		version:      versionStr,
		suppressed:   os.Getenv("ICAL_NO_UPDATE_CHECK") != "",
		// Notices are written to stderr, so stderr — not stdout — decides
		// whether a human is there to read them.
		interactive: isTerminal(os.Stderr),
	}
}

// shouldShowNotices reports whether post-run notices belong in this invocation.
// It gates both the background update check and the printing of any notice, so
// the two can never disagree about what counts as a scripted context.
func shouldShowNotices(c noticeConditions) bool {
	if c.suppressed {
		return false
	}

	if c.version == "" || c.version == "dev" {
		return false
	}

	if isMetaCommand(c.commandPath) {
		return false
	}

	if c.outputFormat == "json" {
		return false
	}

	return c.interactive
}

// isMetaCommand matches on the whole command path, so nested subcommands such
// as "ical skills status" are covered by the "skills" entry. Matching on the
// leaf name alone would silently miss every subcommand of a meta command.
func isMetaCommand(commandPath string) bool {
	for _, name := range strings.Fields(commandPath) {
		if metaCommands[name] {
			return true
		}
	}
	return false
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// collectUpdateResult reads the background check result without blocking. A nil
// result means the goroutine is still in flight and this run says nothing.
func collectUpdateResult(ch <-chan *update.Result) *update.Result {
	select {
	case result := <-ch:
		return result
	default:
		return nil
	}
}

// printNotices writes the update and skills staleness notices.
func printNotices(w io.Writer, version string, result *update.Result, homeDir string) {
	if result != nil && result.HasUpdate {
		yellow := color.New(color.FgYellow)
		fmt.Fprintln(w)
		yellow.Fprintf(w, "A new version of ical is available: %s → %s\n", version, result.Latest)
		fmt.Fprintf(w, "Update: curl -fsSL https://ical.sidv.dev/install | bash\n")
	}

	// Check skills staleness (local only, no HTTP)
	printSkillsStalenessNotice(w, version, homeDir)
}

// printSkillsStalenessNotice checks if installed skills are outdated.
func printSkillsStalenessNotice(w io.Writer, version, homeDir string) {
	if version == "" || version == "dev" {
		return
	}

	targets := skills.InstalledTargets(skills.DefaultTargets(homeDir))
	for _, t := range targets {
		installed := skills.InstalledVersion(t)
		if installed != "" && installed != version {
			yellow := color.New(color.FgYellow)
			fmt.Fprintln(w)
			yellow.Fprintf(w, "Installed skills are outdated (%s). Run: ical skills install\n", installed)
			return // Only show once
		}
	}
}
