# cal

A fast, native macOS Calendar CLI built on [go-eventkit](https://github.com/BRO3886/go-eventkit).

Full CRUD for calendar events, natural language dates, recurrence support, import/export, and multiple output formats — all via EventKit (3000x faster than AppleScript).

## Install

```bash
go install github.com/BRO3886/cal/cmd/cal@latest
```

Or build from source:

```bash
git clone https://github.com/BRO3886/cal.git
cd cal
make build
# Binary at ./bin/cal
```

> **Requires macOS.** Uses cgo + EventKit. On first run, macOS will prompt for Calendar access.

## Quick Start

```bash
# Today's agenda
cal today

# Next 7 days
cal upcoming

# Next 30 days, excluding noisy calendars
cal upcoming -d 30 --exclude-calendar Birthdays --exclude-calendar "US Holidays"

# List events in a date range
cal list -f "next monday" -t "next friday"

# Search events
cal search "standup" -c Work

# Show event details (interactive picker)
cal show

# Show by row number from last listing
cal show 2

# Create an event
cal add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Create interactively
cal add -i

# Delete an event (interactive picker)
cal delete
```

## Commands

| Command                          | Description                                       |
| -------------------------------- | ------------------------------------------------- |
| `cal calendars`                  | List all calendars                                |
| `cal calendars create [title]`   | Create a new calendar                             |
| `cal calendars update [name]`    | Update a calendar (rename, recolor)               |
| `cal calendars delete [name]`    | Delete a calendar and all its events              |
| `cal list`                       | List events in a date range                       |
| `cal today`                      | Today's events                                    |
| `cal upcoming`                   | Events in next N days                             |
| `cal show [# or id]`            | Show event details (interactive picker if no arg) |
| `cal add [title]`               | Create an event (`-i` for interactive)            |
| `cal update [# or id]`          | Update an event (`-i` for interactive)            |
| `cal delete [# or id]`          | Delete an event (interactive picker if no arg)    |
| `cal search [query]`            | Search events                                     |
| `cal export`                     | Export events (JSON/CSV/ICS)                      |
| `cal import [file]`             | Import events (JSON/CSV)                          |
| `cal version`                    | Show version info                                 |
| `cal completion`                 | Generate shell completions                        |

## Global Flags

| Flag         | Short | Default | Description                                     |
| ------------ | ----- | ------- | ----------------------------------------------- |
| `--output`   | `-o`  | `table` | Output format: `table`, `json`, `plain`         |
| `--no-color` |       | `false` | Disable color output (also respects `NO_COLOR`) |

## Natural Language Dates

All date flags accept natural language:

| Input                            | Resolves to                          |
| -------------------------------- | ------------------------------------ |
| `today`, `tomorrow`, `yesterday` | Relative dates                       |
| `next monday`, `next friday`     | Next occurrence of weekday           |
| `next week`, `next month`        | Next week Monday / 1st of next month |
| `in 3 hours`, `in 30 minutes`    | Relative time                        |
| `in 5 days`, `in 2 weeks`        | Relative days                        |
| `3pm`, `15:00`, `3:30pm`         | Time today                           |
| `friday 2pm`                     | Next Friday at 2pm                   |
| `mar 15`, `march 15`             | Month + day this year                |
| `2026-03-15 14:00`               | ISO 8601 datetime                    |
| `eod`                            | Today 5:00 PM                        |
| `eow`                            | Friday 5:00 PM                       |
| `this week`                      | Sunday 11:59 PM                      |
| `2 hours ago`, `5 days ago`      | Past relative                        |

## Managing Calendars

```bash
# List all calendars (see sources, types, colors)
cal calendars

# Create a new calendar
cal calendars create "Projects" --source iCloud --color "#FF6961"

# Create interactively (pick source from dropdown)
cal calendars create -i

# Rename a calendar
cal calendars update "Projects" --title "Archived"

# Change calendar color
cal calendars update "Projects" --color "#42D692"

# Update interactively
cal calendars update -i

# Delete a calendar (with confirmation)
cal calendars delete "Projects"

# Delete without confirmation
cal calendars delete "Projects" -f
```

## Event Listing

```bash
# Filter by calendar
cal list -c Work -f today -t "in 7 days"

# Exclude calendars
cal upcoming --exclude-calendar Birthdays --exclude-calendar "Holidays in India"

# Search with date range
cal search "meeting" -f "1 month ago" -t "in 1 month"

# Sort by title
cal list -f today -t "next friday" --sort title

# Limit results
cal upcoming -d 30 -n 10

# JSON output for scripting
cal today -o json | jq '.[].title'

# Plain output for grep
cal today -o plain | grep "standup"
```

## Creating Events

```bash
# Quick event
cal add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# All-day event
cal add "Company Holiday" -s 2026-03-15 --all-day -c Work

# With location and alerts
cal add "Dinner" -s "friday 7pm" -e "friday 9pm" \
  -l "The Restaurant, 123 Main St" --alert 1h --alert 15m

# Recurring event
cal add "Weekly Sync" -s "next monday 10am" -e "next monday 11am" \
  --repeat weekly --repeat-days mon -c Work

# Recurring with end date
cal add "Daily Standup" -s "tomorrow 9am" -e "tomorrow 9:15am" \
  --repeat daily --repeat-until "2026-12-31" -c Work

# With timezone
cal add "NYC Meeting" -s "tomorrow 2pm" -e "tomorrow 3pm" \
  --timezone "America/New_York" -c Work
```

## Updating Events

```bash
# Interactive picker + guided form
cal update -i

# Update by row number from last listing
cal update 2 --title "New Title"

# Reschedule
cal update 3 -s "tomorrow 2pm" -e "tomorrow 3pm"

# Update future occurrences of recurring event
cal update 1 --span future --title "New Series Name"
```

## Deleting Events

```bash
# Interactive picker with confirmation
cal delete

# Delete by row number
cal delete 3

# Skip confirmation
cal delete 3 -f

# Delete future occurrences
cal delete 2 --span future
```

## Export & Import

```bash
# Export to JSON
cal export -f 2026-01-01 -t 2026-12-31 --format json > events.json

# Export to CSV
cal export -c Work --format csv --output-file work-events.csv

# Export to ICS (RFC 5545)
cal export --format ics --output-file calendar.ics

# Import from JSON
cal import events.json

# Import to specific calendar
cal import events.csv -c Personal

# Dry run (preview without creating)
cal import events.json --dry-run
```

## Event Selection

Events can be selected in three ways:

1. **Interactive picker** (no argument): `cal show`, `cal delete`, `cal update` — presents a searchable list
2. **Row number** from the last listing: `cal show 2` picks event #2 from the last `cal ls`/`cal today` output
3. **Full or partial event ID**: `cal show 577B8983-DF44-4665-...` — for scripting and automation

```bash
# List events (shows row numbers)
cal ls
#  #  TIME             TITLE          CALENDAR
#  1  10:00 - 11:00    Standup        Work
#  2  14:00 - 15:00    1:1 with Bob   Work

# Then reference by number
cal show 2
cal update 2 -i
cal delete 1
```

- **JSON output** (`-o json`): Always includes the full event ID for scripting

## Interactive Mode

The `-i` flag on `add` and `update` launches a guided form for step-by-step event creation or editing:

```bash
# Create event interactively (guided form)
cal add -i

# Update event interactively (pick event, then edit fields)
cal update -i

# Pick an event interactively (no argument triggers a searchable picker)
cal show
cal delete
```

Interactive mode uses [charmbracelet/huh](https://github.com/charmbracelet/huh) forms with the Catppuccin theme. The event picker provides a searchable list of upcoming events for quick selection.

## Shell Completions

```bash
# Bash
cal completion bash > /usr/local/etc/bash_completion.d/cal

# Zsh
cal completion zsh > "${fpath[1]}/_cal"

# Fish
cal completion fish > ~/.config/fish/completions/cal.fish
```

## Architecture

Built on [go-eventkit](https://github.com/BRO3886/go-eventkit) — native EventKit bindings via cgo. No AppleScript, no subprocesses. Single binary.

```
cal/
├── cmd/cal/
│   ├── main.go              # Entry point
│   └── commands/             # Cobra commands (one per file)
├── internal/
│   ├── parser/               # Natural language date parsing
│   ├── ui/                   # Output formatting (table/json/plain)
│   └── export/               # JSON/CSV/ICS import/export
├── Makefile
└── go.mod
```

## Documentation

Full documentation is available at [cal.sidv.dev](https://cal.sidv.dev).

## License

[MIT](LICENSE)
