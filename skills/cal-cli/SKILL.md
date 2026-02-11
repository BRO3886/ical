---
name: cal-cli
description: Manages macOS Calendar events and calendars from the terminal using the cal CLI. Full CRUD for both events and calendars. Supports natural language dates, recurrence rules, alerts, interactive mode, import/export (JSON/CSV/ICS), and multiple output formats. Use when the user wants to interact with Apple Calendar via command line, automate calendar workflows, or build scripts around macOS Calendar.
metadata:
  author: sidv
  version: "1.0"
compatibility: Requires macOS with Calendar.app. Requires Xcode Command Line Tools for building from source.
---

# cal — CLI for macOS Calendar

A Go CLI that wraps macOS Calendar. Sub-millisecond reads via cgo + EventKit. Single binary, no dependencies at runtime.

## Installation

```bash
go install github.com/BRO3886/cal/cmd/cal@latest
```

Or build from source:

```bash
git clone <repo-url> && cd cal
make build    # produces bin/cal
```

## Quick Start

```bash
# List all calendars (shows sources, colors, types)
cal calendars

# Create a new calendar
cal calendars create "Projects" --source iCloud --color "#FF6961"

# Show today's agenda
cal today

# List events this week
cal list --from today --to "end of week"

# Add an event with natural language dates
cal add "Team standup" --start "tomorrow at 9am" --end "tomorrow at 9:30am" --calendar Work --alert 15m

# Show event details (row number from last listing)
cal show 2

# Search for events
cal search "meeting" --from "30 days ago" --to "next month"

# Export events to ICS
cal export --format ics --from today --to "in 30 days" --output-file events.ics
```

## Command Reference

### Event CRUD

| Command      | Aliases         | Description             |
| ------------ | --------------- | ----------------------- |
| `cal add`    | `create`, `new` | Create an event         |
| `cal show`   | `get`, `info`   | Show full event details |
| `cal update` | `edit`          | Update event properties |
| `cal delete` | `rm`, `remove`  | Delete an event         |

### Event Views

| Command        | Aliases        | Description                    |
| -------------- | -------------- | ------------------------------ |
| `cal list`     | `ls`, `events` | List events in a date range    |
| `cal today`    | —              | Show today's events            |
| `cal upcoming` | `next`, `soon` | Show events in the next N days |

### Search & Export

| Command      | Aliases | Description                             |
| ------------ | ------- | --------------------------------------- |
| `cal search` | `find`  | Search events by title, location, notes |
| `cal export` | —       | Export events to JSON, CSV, or ICS      |
| `cal import` | —       | Import events from JSON or CSV file     |

### Calendar Management

| Command                   | Aliases           | Description                          |
| ------------------------- | ----------------- | ------------------------------------ |
| `cal calendars`           | `cals`            | List all calendars                   |
| `cal calendars create`    | `add`, `new`      | Create a new calendar                |
| `cal calendars update`    | `edit`, `rename`  | Update a calendar (rename, recolor)  |
| `cal calendars delete`    | `rm`, `remove`    | Delete a calendar and all its events |

### Other

| Command          | Aliases | Description                                |
| ---------------- | ------- | ------------------------------------------ |
| `cal version`    | —       | Print version and build info               |
| `cal completion` | —       | Generate shell completions (bash/zsh/fish) |

For full flag details on every command, see [references/commands.md](references/commands.md).

## Key Concepts

### Row Numbers

Event listings display row numbers (`#1`, `#2`, `#3`...) alongside events. These are cached to `~/.cal-last-list` so you can reference them in subsequent commands:

```bash
cal list --from today --to "next week"   # Shows #1, #2, #3...
cal show 2                                # Show details for row #2
cal update 3 --title "New title"          # Update row #3
cal delete 1                              # Delete row #1
```

Row numbers reset each time you run a list/today/upcoming command. With no arguments, `show`, `update`, and `delete` launch an interactive picker instead.

### Natural Language Dates

Date flags (`--from`, `--to`, `--start`, `--end`, `--due`) accept natural language:

```bash
cal list --from today --to "next friday"
cal add "Lunch" --start "tomorrow at noon" --end "tomorrow at 1pm"
cal search "standup" --from "2 weeks ago"
cal upcoming --days 14
```

Supported patterns: `today`, `tomorrow`, `next monday`, `in 3 hours`, `eod`, `eow`, `this week`, `5pm`, `mar 15`, `2 days ago`, and more. See [references/dates.md](references/dates.md) for the full list.

### Interactive Mode

The `add` and `update` commands support `-i` for guided form-based input:

```bash
cal add -i        # Multi-page form: title, calendar, dates, location, recurrence
cal update 2 -i   # Pre-filled form with current event values
```

The `show`, `update`, and `delete` commands accept 0 arguments to launch an interactive event picker:

```bash
cal show          # Pick from upcoming events
cal delete        # Pick an event to delete
```

### Output Formats

All read commands support `-o` / `--output`:

- **table** (default) — formatted table with borders and color
- **json** — machine-readable JSON (ISO 8601 dates)
- **plain** — simple text, one item per line

The `NO_COLOR` environment variable and `--no-color` flag are respected.

### Recurrence

Events can repeat with flexible rules:

```bash
# Daily standup
cal add "Standup" --start "tomorrow at 9am" --repeat daily

# Every 2 weeks on Mon and Wed
cal add "Team sync" --start "next monday at 10am" --repeat weekly --repeat-interval 2 --repeat-days mon,wed

# Monthly for 6 months
cal add "Review" --start "mar 1 at 2pm" --repeat monthly --repeat-count 6

# Yearly until a date
cal add "Anniversary" --start "jun 15" --repeat yearly --repeat-until "2030-06-15"
```

Use `--repeat none` on update to remove recurrence. Use `--span future` to update/delete this and all future occurrences.

### Alerts

Add reminders before an event with the `--alert` flag (repeatable):

```bash
cal add "Meeting" --start "tomorrow at 2pm" --alert 15m          # 15 minutes before
cal add "Flight" --start "mar 15 at 8am" --alert 1h --alert 1d   # 1 hour + 1 day before
```

Supported units: `m` (minutes), `h` (hours), `d` (days).

## Common Workflows

### Daily review

```bash
cal today                                 # See today's agenda
cal upcoming --days 1                     # Same as today
cal list --from today --to "end of week"  # Rest of the week
```

### Weekly planning

```bash
cal upcoming --days 7                           # Full week view
cal add "Planning" --start "monday at 9am" -i  # Add events interactively
```

### Scripting with JSON output

```bash
# Count today's events
cal today -o json | jq 'length'

# Get titles of upcoming events
cal upcoming -o json | jq -r '.[].title'

# Find events on a specific calendar
cal list --from today --to "in 30 days" --calendar Work -o json | jq '.[].title'
```

### Backup and restore

```bash
# Export all events from the past year
cal export --from "12 months ago" --to "in 12 months" --format json --output-file backup.json

# Export as ICS for other calendar apps
cal export --from today --to "in 6 months" --format ics --output-file events.ics

# Import from backup
cal import backup.json --calendar "Restored"
```

## Public Go API

For programmatic access to macOS Calendar, use [`go-eventkit`](https://github.com/BRO3886/go-eventkit) directly:

```go
import "github.com/BRO3886/go-eventkit/calendar"

client, _ := calendar.New()
events, _ := client.Events(from, to, calendar.WithCalendarName("Work"))
event, _ := client.CreateEvent(calendar.CreateEventInput{
    Title:        "Team Meeting",
    StartDate:    start,
    EndDate:      end,
    CalendarName: "Work",
})
```

See [go-eventkit docs](https://github.com/BRO3886/go-eventkit) for the full API surface.

## Limitations

- **macOS only** — requires EventKit framework via cgo
- **No attendee management** — attendees and organizer are read-only (Apple limitation)
- **Subscribed/birthday calendars are read-only** — cannot create events on these
- **Event IDs are calendar-scoped** — the UUID prefix before `:` is the calendar ID, not event-specific. Use row numbers or the interactive picker instead of raw IDs
