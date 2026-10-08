package ui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Notices contains the information selected for display after a command.
type Notices struct {
	Version            string
	LatestVersion      string
	UpdateCommand      string
	StaleSkillsVersion string
}

// PrintNotices renders update and skill notices to w.
func PrintNotices(w io.Writer, notices Notices) {
	yellow := color.New(color.FgYellow)
	if notices.LatestVersion != "" {
		fmt.Fprintln(w)
		yellow.Fprintf(w, "A new version of ical is available: %s → %s\n", notices.Version, notices.LatestVersion)
		fmt.Fprintf(w, "Update: %s\n", notices.UpdateCommand)
	}
	if notices.StaleSkillsVersion != "" {
		fmt.Fprintln(w)
		yellow.Fprintf(w, "Installed skills are outdated (%s). Run: ical skills install\n", notices.StaleSkillsVersion)
	}
}
