---
name: cal-cli
description: Manages macOS Calendar events and calendars from the terminal using the ical CLI. Full CRUD for both events and calendars. Supports natural language dates, recurrence rules, alerts, interactive mode, import/export (JSON/CSV/ICS), and multiple output formats. Use when the user wants to interact with Apple Calendar via command line, automate calendar workflows, or build scripts around macOS Calendar.
metadata:
  author: sidv
  version: "1.0"
compatibility: Requires macOS with Calendar.app. Requires Xcode Command Line Tools for building from source.
---

# ical — CLI for macOS Calendar

A Go CLI that wraps macOS Calendar. Sub-millisecond reads via cgo + EventKit. Single binary, no dependencies at runtime.

## Installation

```bash
go install github.com/BRO3886/ical/cmd/ical@latest
```

Or build from source:

```bash
git clone <repo-url> && cd ical
make build    # produces bin/ical
```

## Quick Start

```bash
# List all calendars (shows sources, colors, types)
ical calendars

# Create a new calendar
ical calendars create "Projects" --source iCloud --color "#FF6961"

# Show today's agenda
ical today

# List events this week
ical list --from today --to "end of week"

# Add an event with natural language dates
ical add "Team standup" --start "tomorrow at 9am" --end "tomorrow at 9:30am" --calendar Work --alert 15m

# Show event details (row number from last listing)
ical show 2

# Delete an event (--force skips confirmation prompt, required in scripts/agents)
ical delete 2 --force

# Search for events
ical search "meeting" --from "30 days ago" --to "next month"

# Export events to ICS
ical export --format ics --from today --to "in 30 days" --output-file events.ics
```

## Command Reference

### Event CRUD

| Command      | Aliases         | Description             |
| ------------ | --------------- | ----------------------- |
| `ical add`    | `create`, `new` | Create an event         |
| `ical show`   | `get`, `info`   | Show full event details |
| `ical update` | `edit`          | Update event properties |
| `ical delete` | `rm`, `remove`  | Delete an event         |

### Event Views

| Command        | Aliases        | Description                    |
| -------------- | -------------- | ------------------------------ |
| `ical list`     | `ls`, `events` | List events in a date range    |
| `ical today`    | —              | Show today's events            |
| `ical upcoming` | `next`, `soon` | Show events in the next N days |

### Search & Export

| Command      | Aliases | Description                             |
| ------------ | ------- | --------------------------------------- |
| `ical search` | `find`  | Search events by title, location, notes |
| `ical export` | —       | Export events to JSON, CSV, or ICS      |
| `ical import` | —       | Import events from JSON or CSV file     |

### Calendar Management

| Command                   | Aliases           | Description                          |
| ------------------------- | ----------------- | ------------------------------------ |
| `ical calendars`           | `icals`            | List all calendars                   |
| `ical calendars create`    | `add`, `new`      | Create a new calendar                |
| `ical calendars update`    | `edit`, `rename`  | Update a calendar (rename, recolor)  |
| `ical calendars delete`    | `rm`, `remove`    | Delete a calendar and all its events |

### Skills & Other

| Command                | Aliases | Description                                        |
| ---------------------- | ------- | -------------------------------------------------- |
| `ical skills install`   | —       | Install ical agent skill for Claude Code / Codex   |
| `ical skills uninstall` | —       | Remove ical agent skill                            |
| `ical skills status`    | —       | Show skill installation status                     |
| `ical version`          | —       | Print version and build info                       |
| `ical completion`       | —       | Generate shell completions (bash/zsh/fish)         |

For full flag details on every command, see [references/commands.md](references/commands.md).

## Key Concepts

### Row Numbers

Event listings display row numbers (`#1`, `#2`, `#3`...) alongside events. These are cached to `~/.ical-last-list` so you can reference them in subsequent commands:

```bash
ical list --from today --to "next week"   # Shows #1, #2, #3...
ical show 2                                # Show details for row #2
ical update 3 --title "New title"          # Update row #3
ical delete 1 --force                      # Delete row #1 (skip confirmation)
ical delete 1                              # Delete row #1 (prompts for confirmation)
```

Row numbers reset each time you run a list/today/upcoming command. With no arguments, `show`, `update`, and `delete` launch an interactive picker instead.

### Event ID flag (for scripts and agents)

When you have a full event ID (from `-o json` output), use `--id` for exact lookup with no prefix matching:

```bash
# Get event ID from JSON output
EVENT_ID=$(ical today -o json | jq -r '.[0].id')

# Use --id for reliable exact lookup
ical show --id "$EVENT_ID"
ical update --id "$EVENT_ID" --title "New title"
ical delete --id "$EVENT_ID" --force
```

> **Important for scripting**:
> - `ical delete` prompts for interactive confirmation by default. Always pass `--force` (or `-f`) when running non-interactively. There is no `--confirm` flag.
> - `ical update` does **not** require confirmation and has **no `--force` flag** — just run it directly with the flags you want to change.
> - `--id` and a positional argument are mutually exclusive — passing both returns an error.

### Natural Language Dates

Date flags (`--from`, `--to`, `--start`, `--end`, `--due`) accept natural language:

```bash
ical list --from today --to "next friday"
ical add "Lunch" --start "tomorrow at noon" --end "tomorrow at 1pm"
ical search "standup" --from "2 weeks ago"
ical upcoming --days 14
```

Supported patterns: `today`, `tomorrow`, `next monday`, `in 3 hours`, `eod`, `eow`, `this week`, `5pm`, `mar 15`, `2 days ago`, and more. See [references/dates.md](references/dates.md) for the full list.

### Interactive Mode

The `add` and `update` commands support `-i` for guided form-based input:

```bash
ical add -i        # Multi-page form: title, calendar, dates, location, recurrence
ical update 2 -i   # Pre-filled form with current event values
```

The `show`, `update`, and `delete` commands accept 0 arguments to launch an interactive event picker:

```bash
ical show          # Pick from upcoming events
ical delete        # Pick an event to delete
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
ical add "Standup" --start "tomorrow at 9am" --repeat daily

# Every 2 weeks on Mon and Wed
ical add "Team sync" --start "next monday at 10am" --repeat weekly --repeat-interval 2 --repeat-days mon,wed

# Monthly for 6 months
ical add "Review" --start "mar 1 at 2pm" --repeat monthly --repeat-count 6

# Yearly until a date
ical add "Anniversary" --start "jun 15" --repeat yearly --repeat-until "2030-06-15"
```

Use `--repeat none` on update to remove recurrence. Use `--span future` to update/delete this and all future occurrences.

### Alerts

Add reminders before an event with the `--alert` flag (repeatable):

```bash
ical add "Meeting" --start "tomorrow at 2pm" --alert 15m          # 15 minutes before
ical add "Flight" --start "mar 15 at 8am" --alert 1h --alert 1d   # 1 hour + 1 day before
```

Supported units: `m` (minutes), `h` (hours), `d` (days).

## Common Workflows

### Daily review

```bash
ical today                                 # See today's agenda
ical upcoming --days 1                     # Same as today
ical list --from today --to "end of week"  # Rest of the week
```

### Weekly planning

```bash
ical upcoming --days 7                           # Full week view
ical add "Planning" --start "monday at 9am" -i  # Add events interactively
```

### Scripting with JSON output

```bash
# Count today's events
ical today -o json | jq 'length'

# Get titles of upcoming events
ical upcoming -o json | jq -r '.[].title'

# Find events on a specific calendar
ical list --from today --to "in 30 days" --calendar Work -o json | jq '.[].title'

# List calendar names (field is "title", not "name")
ical calendars -o json | jq -r '.[].title'

# Get calendar IDs and names
ical calendars -o json | jq -r '.[] | "\(.id) \(.title)"'
```

**Calendar JSON fields**: `id`, `title`, `type`, `color`, `source`, `readOnly`
**Event JSON fields**: `id`, `title`, `start_date`, `end_date`, `calendar`, `calendar_id`, `location`, `notes`, `url`, `all_day`, `recurrence`, `alerts`

### Backup and restore

```bash
# Export all events from the past year
ical export --from "12 months ago" --to "in 12 months" --format json --output-file backup.json

# Export as ICS for other calendar apps
ical export --from today --to "in 6 months" --format ics --output-file events.ics

# Import from backup
ical import backup.json --calendar "Restored"
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
