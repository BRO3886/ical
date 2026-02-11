package commands

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	noColor      bool
)

var rootCmd = &cobra.Command{
	Use:   "cal",
	Short: "A fast, native macOS Calendar CLI",
	Long:  "cal — a fast, native macOS Calendar CLI built on EventKit.\nProvides full CRUD for calendar events, natural language dates,\nrecurrence support, import/export, and multiple output formats.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, plain")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

func Execute() error {
	return rootCmd.Execute()
}
