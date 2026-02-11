
## Overview

cal is a native macOS CLI that manages Calendar events through [go-eventkit](https://github.com/BRO3886/go-eventkit), a Go library providing direct EventKit bindings via cgo. This gives cal near-native performance — roughly 3000x faster than AppleScript-based alternatives.

The entire application compiles to a single binary with no runtime dependencies beyond macOS itself.

## How It Works

```
User → cal CLI (Cobra) → go-eventkit → EventKit (Objective-C via cgo) → Calendar.app store
```

1. The user invokes a command via the Cobra CLI framework
2. Commands call functions in the `go-eventkit/calendar` package
3. go-eventkit uses cgo to call EventKit's Objective-C APIs directly
4. EventKit reads from and writes to the same store that Calendar.app uses

There is no IPC, no subprocess spawning, and no Apple Events bridge. The Go binary links against EventKit at compile time.

## Project Structure

```
cal/
├── cmd/cal/
│   ├── main.go                  # Entry point (macOS check, version injection)
│   └── commands/                # One file per Cobra command
│       ├── root.go              # Root command + global flags (--output, --no-color)
│       ├── calendars.go         # List calendars
│       ├── list.go              # List events (date range, filters)
│       ├── show.go              # Show single event detail
│       ├── add.go               # Create event (flags + interactive -i)
│       ├── update.go            # Update event (flags + interactive -i)
│       ├── delete.go            # Delete event + pickEvent() helper
│       ├── helpers.go           # Shared helpers
│       ├── today.go             # Shortcut: today's events
│       ├── upcoming.go          # Next N days
│       ├── search.go            # Search events
│       ├── export.go            # Export events (JSON/CSV/ICS)
│       └── import.go            # Import events (JSON/CSV)
├── internal/
│   ├── parser/                  # Natural language date parsing
│   │   ├── date.go
│   │   └── date_test.go
│   ├── ui/                      # Output formatting (table/json/plain)
│   │   └── output.go
│   └── export/                  # Import/export logic
│       ├── json.go
│       ├── csv.go
│       └── ics.go
├── Makefile
└── go.mod
```

## Key Dependencies

| Package                        | Purpose                              |
|--------------------------------|--------------------------------------|
| `github.com/BRO3886/go-eventkit` | Native EventKit bindings (cgo)    |
| `github.com/spf13/cobra`      | CLI framework                        |
| `github.com/olekukonko/tablewriter` | Table output formatting         |
| `github.com/fatih/color`       | Terminal colors                     |
| `github.com/charmbracelet/huh` | Interactive forms and select menus  |

## Design Decisions

### Row Numbers Instead of Short IDs

Calendar event identifiers in EventKit share a common prefix per calendar — the UUID before the `:` separator is the calendar ID, not the event ID. This makes short ID prefixes useless for disambiguation when events belong to the same calendar.

Instead, cal uses sequential row numbers (`#1`, `#2`, ...) displayed in table output. These numbers are cached to `~/.cal-last-list` so subsequent commands like `cal show 2` or `cal delete 1` can reference events from the last listing.

### Three Event Selection Methods

1. **Interactive picker** — No arguments triggers a searchable list powered by `charmbracelet/huh`
2. **Row number** — Numeric argument maps to cached row from last listing
3. **Event ID** — Full or partial `eventIdentifier` for scripting and automation

### End-of-Day Bumping

When `--to` resolves to midnight (00:00:00), cal bumps it to 23:59:59 so that `--to "feb 12"` includes all events on February 12. Without this, midnight would exclude the entire day.

### UTC to Local Conversion

EventKit returns all times in UTC. cal converts them to local time using the event's timezone (or the system timezone) for display. JSON output preserves ISO 8601 timestamps.

## Limitations

These are Apple-imposed constraints, not bugs:

- **Attendees and organizer are read-only** — EventKit does not allow modifying attendee lists
- **Subscribed calendars are read-only** — Cannot create or modify events in subscribed calendars
- **Birthday calendars are read-only** — The Birthdays calendar is auto-generated
- **macOS only** — EventKit is an Apple framework; cal exits gracefully on other platforms
- **Date ranges required** — EventKit requires bounded queries; cal does not support unbounded event fetches

