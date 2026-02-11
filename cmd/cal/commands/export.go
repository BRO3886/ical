package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/BRO3886/cal/internal/export"
	"github.com/BRO3886/cal/internal/parser"
	"github.com/BRO3886/go-eventkit/calendar"
	"github.com/spf13/cobra"
)

var (
	exportFrom       string
	exportTo         string
	exportCalendar   string
	exportFormatFlag string
	exportOutputFile string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export events",
	Long:  "Exports events to JSON, CSV, or ICS format.",
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		from := now.AddDate(0, 0, -30)
		if exportFrom != "" {
			t, err := parser.ParseDate(exportFrom)
			if err != nil {
				return fmt.Errorf("invalid --from date: %w", err)
			}
			from = t
		}

		to := now.AddDate(0, 0, 30)
		if exportTo != "" {
			t, err := parser.ParseDate(exportTo)
			if err != nil {
				return fmt.Errorf("invalid --to date: %w", err)
			}
			to = t
		}

		client, err := calendar.New()
		if err != nil {
			return handleClientError(err)
		}

		var opts []calendar.ListOption
		if exportCalendar != "" {
			opts = append(opts, calendar.WithCalendar(exportCalendar))
		}

		events, err := client.Events(from, to, opts...)
		if err != nil {
			return fmt.Errorf("failed to fetch events: %w", err)
		}

		w := os.Stdout
		if exportOutputFile != "" {
			f, err := os.Create(exportOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			w = f
		}

		switch exportFormatFlag {
		case "csv":
			return export.CSV(events, w)
		case "ics":
			return export.ICS(events, w)
		default:
			return export.JSON(events, w)
		}
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFrom, "from", "f", "", "Start date (default: 30 days ago)")
	exportCmd.Flags().StringVarP(&exportTo, "to", "t", "", "End date (default: 30 days ahead)")
	exportCmd.Flags().StringVarP(&exportCalendar, "calendar", "c", "", "Filter by calendar")
	exportCmd.Flags().StringVar(&exportFormatFlag, "format", "json", "Format: json, csv, ics")
	exportCmd.Flags().StringVar(&exportOutputFile, "output-file", "", "Write to file instead of stdout")

	rootCmd.AddCommand(exportCmd)
}
