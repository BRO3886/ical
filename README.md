# ical

A fast, native macOS Calendar CLI built on [go-eventkit](https://github.com/BRO3886/go-eventkit).

Full CRUD for calendar events, natural language dates, recurrence support, import/export, and multiple output formats — all via EventKit (3000x faster than AppleScript).

## Install

**Quick install (recommended):**

```bash
curl -fsSL https://ical.sidv.dev/install | bash
```

**Via Go:**

```bash
go install github.com/BRO3886/ical/cmd/ical@latest
```

> Requires Go 1.21+ and Xcode Command Line Tools (`xcode-select --install`).

**Manual download:**

Apple Silicon:
```bash
curl -sSL https://github.com/BRO3886/ical/releases/latest/download/ical-darwin-arm64.tar.gz | tar xzf -
sudo mv ical /usr/local/bin/
```

Intel:
```bash
curl -sSL https://github.com/BRO3886/ical/releases/latest/download/ical-darwin-amd64.tar.gz | tar xzf -
sudo mv ical /usr/local/bin/
```

**Build from source:**

```bash
git clone https://github.com/BRO3886/ical.git
cd ical
make build
# Binary at ./bin/ical
```

> **Requires macOS.** Uses cgo + EventKit. On first run, macOS will prompt for Calendar access.

## Quick Start

```bash
# Today's agenda
ical today

# Next 7 days
ical upcoming

# Next 30 days, excluding noisy calendars
ical upcoming -d 30 --exclude-calendar Birthdays --exclude-calendar "US Holidays"

# List events in a date range
ical list -f "next monday" -t "next friday"

# Search events
ical search "standup" -c Work

# Show event details (interactive picker)
ical show

# Show by row number from last listing
ical show 2

# Create an event
ical add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Create interactively
ical add -i

# Delete an event (interactive picker)
ical delete
```

## Commands

| Command                          | Description                                       |
| -------------------------------- | ------------------------------------------------- |
| `ical calendars`                  | List all calendars                                |
| `ical calendars create [title]`   | Create a new calendar                             |
| `ical calendars update [name]`    | Update a calendar (rename, recolor)               |
| `ical calendars delete [name]`    | Delete a calendar and all its events              |
| `ical list`                       | List events in a date range                       |
| `ical today`                      | Today's events                                    |
| `ical upcoming`                   | Events in next N days                             |
| `ical show [# or id]`            | Show event details (interactive picker if no arg) |
| `ical add [title]`               | Create an event (`-i` for interactive)            |
| `ical update [# or id]`          | Update an event (`-i` for interactive)            |
| `ical delete [# or id]`          | Delete an event (interactive picker if no arg)    |
| `ical search [query]`            | Search events                                     |
| `ical export`                     | Export events (JSON/CSV/ICS)                      |
| `ical import [file]`             | Import events (JSON/CSV)                          |
| `ical skills install`             | Install AI agent skill (Claude Code / Codex)      |
| `ical skills uninstall`           | Remove AI agent skill                             |
| `ical skills status`              | Show skill installation status                    |
| `ical version`                    | Show version info                                 |
| `ical completion`                 | Generate shell completions                        |

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
ical calendars

# Create a new calendar
ical calendars create "Projects" --source iCloud --color "#FF6961"

# Create interactively (pick source from dropdown)
ical calendars create -i

# Rename a calendar
ical calendars update "Projects" --title "Archived"

# Change calendar color
ical calendars update "Projects" --color "#42D692"

# Update interactively
ical calendars update -i

# Delete a calendar (with confirmation)
ical calendars delete "Projects"

# Delete without confirmation
ical calendars delete "Projects" -f
```

## Event Listing

```bash
# Filter by calendar
ical list -c Work -f today -t "in 7 days"

# Exclude calendars
ical upcoming --exclude-calendar Birthdays --exclude-calendar "Holidays in India"

# Search with date range
ical search "meeting" -f "1 month ago" -t "in 1 month"

# Sort by title
ical list -f today -t "next friday" --sort title

# Limit results
ical upcoming -d 30 -n 10

# JSON output for scripting
ical today -o json | jq '.[].title'

# Plain output for grep
ical today -o plain | grep "standup"
```

## Creating Events

```bash
# Quick event
ical add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# All-day event
ical add "Company Holiday" -s 2026-03-15 --all-day -c Work

# With location and alerts
ical add "Dinner" -s "friday 7pm" -e "friday 9pm" \
  -l "The Restaurant, 123 Main St" --alert 1h --alert 15m

# Recurring event
ical add "Weekly Sync" -s "next monday 10am" -e "next monday 11am" \
  --repeat weekly --repeat-days mon -c Work

# Recurring with end date
ical add "Daily Standup" -s "tomorrow 9am" -e "tomorrow 9:15am" \
  --repeat daily --repeat-until "2026-12-31" -c Work

# With timezone
ical add "NYC Meeting" -s "tomorrow 2pm" -e "tomorrow 3pm" \
  --timezone "America/New_York" -c Work
```

## Updating Events

```bash
# Interactive picker + guided form
ical update -i

# Update by row number from last listing
ical update 2 --title "New Title"

# Reschedule
ical update 3 -s "tomorrow 2pm" -e "tomorrow 3pm"

# Update future occurrences of recurring event
ical update 1 --span future --title "New Series Name"
```

## Deleting Events

```bash
# Interactive picker with confirmation
ical delete

# Delete by row number
ical delete 3

# Skip confirmation
ical delete 3 -f

# Delete future occurrences
ical delete 2 --span future
```

## Export & Import

```bash
# Export to JSON
ical export -f 2026-01-01 -t 2026-12-31 --format json > events.json

# Export to CSV
ical export -c Work --format csv --output-file work-events.csv

# Export to ICS (RFC 5545)
ical export --format ics --output-file calendar.ics

# Import from JSON
ical import events.json

# Import to specific calendar
ical import events.csv -c Personal

# Dry run (preview without creating)
ical import events.json --dry-run
```

## Event Selection

Events can be selected in three ways:

1. **Interactive picker** (no argument): `ical show`, `ical delete`, `ical update` — presents a searchable list
2. **Row number** from the last listing: `ical show 2` picks event #2 from the last `ical ls`/`ical today` output
3. **Full or partial event ID**: `ical show 577B8983-DF44-4665-...` — for scripting and automation

```bash
# List events (shows row numbers)
ical ls
#  #  TIME             TITLE          CALENDAR
#  1  10:00 - 11:00    Standup        Work
#  2  14:00 - 15:00    1:1 with Bob   Work

# Then reference by number
ical show 2
ical update 2 -i
ical delete 1
```

- **JSON output** (`-o json`): Always includes the full event ID for scripting

## Interactive Mode

The `-i` flag on `add` and `update` launches a guided form for step-by-step event creation or editing:

```bash
# Create event interactively (guided form)
ical add -i

# Update event interactively (pick event, then edit fields)
ical update -i

# Pick an event interactively (no argument triggers a searchable picker)
ical show
ical delete
```

Interactive mode uses [charmbracelet/huh](https://github.com/charmbracelet/huh) forms with the Catppuccin theme. The event picker provides a searchable list of upcoming events for quick selection.

## AI Agent Skills

ical ships with an embedded [agent skill](https://agentskills.io) that teaches AI coding agents (Claude Code, Codex CLI, Copilot, Cursor, etc.) how to use it effectively.

```bash
# Install the skill (interactive — pick which agents)
ical skills install

# Install for a specific agent
ical skills install --agent claude   # → ~/.claude/skills/ical-cli/
ical skills install --agent codex    # → ~/.agents/skills/ical-cli/
ical skills install --agent all      # Both

# Check installation status
ical skills status

# Remove the skill
ical skills uninstall
```

The skill is automatically kept in sync with the binary version. After updating ical, run `ical skills install` to update the skill files. ical will show a notice if installed skills are outdated.

## Shell Completions

```bash
# Bash
ical completion bash > /usr/local/etc/bash_completion.d/ical

# Zsh
ical completion zsh > "${fpath[1]}/_ical"

# Fish
ical completion fish > ~/.config/fish/completions/ical.fish
```

## Architecture

Built on [go-eventkit](https://github.com/BRO3886/go-eventkit) — native EventKit bindings via cgo. No AppleScript, no subprocesses. Single binary.

```
ical/
├── cmd/ical/
│   ├── main.go              # Entry point
│   └── commands/             # Cobra commands (one per file)
├── internal/
│   ├── parser/               # Natural language date parsing
│   ├── ui/                   # Output formatting (table/json/plain)
│   ├── export/               # JSON/CSV/ICS import/export
│   ├── skills/               # Agent skill install/uninstall logic
│   └── update/               # Background update check
├── skills/cal-cli/           # Embedded agent skill (baked into binary)
├── Makefile
└── go.mod
```

## Documentation

Full documentation is available at [ical.sidv.dev](https://ical.sidv.dev).

## License

[MIT](LICENSE)
