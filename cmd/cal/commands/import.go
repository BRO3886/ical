package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BRO3886/cal/internal/export"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	importCalendar string
	importDryRun   bool
	importForce    bool
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import events from file",
	Long:  "Imports events from JSON or CSV files.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		ext := strings.ToLower(filepath.Ext(filename))
		var inputs []calendar.CreateEventInput

		switch ext {
		case ".json":
			inputs, err = export.ParseJSON(f)
		case ".csv":
			inputs, err = export.ParseCSV(f)
		default:
			return fmt.Errorf("unsupported file format %q (use .json or .csv)", ext)
		}
		if err != nil {
			return fmt.Errorf("failed to parse file: %w", err)
		}

		if importCalendar != "" {
			for i := range inputs {
				inputs[i].Calendar = importCalendar
			}
		}

		if importDryRun {
			fmt.Printf("Dry run: would create %d events\n", len(inputs))
			for _, input := range inputs {
				fmt.Printf("  - %s (%s - %s) [%s]\n",
					input.Title,
					input.StartDate.Format("Jan 02 15:04"),
					input.EndDate.Format("Jan 02 15:04"),
					input.Calendar,
				)
			}
			return nil
		}

		if !importForce {
			fmt.Printf("Import %d events? [y/N] ", len(inputs))
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		created := 0
		errors := 0
		for _, input := range inputs {
			_, err := client.CreateEvent(input)
			if err != nil {
				yellow := color.New(color.FgYellow)
				yellow.Fprintf(os.Stderr, "Warning: failed to create %q: %v\n", input.Title, err)
				errors++
				continue
			}
			created++
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Printf("Created %d events", created)
		if errors > 0 {
			fmt.Printf(", %d errors", errors)
		}
		fmt.Println()

		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&importCalendar, "calendar", "c", "", "Override target calendar for all events")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview without creating")
	importCmd.Flags().BoolVarP(&importForce, "force", "f", false, "Skip confirmation prompt")

	rootCmd.AddCommand(importCmd)
}
