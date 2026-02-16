---
title: "Commands"
description: "Complete reference for every ical command, flag, and option."
weight: 2
---

## Overview

ical provides commands for managing macOS Calendar events and calendars. Every command that displays data supports `--output` (`-o`) with `table`, `json`, or `plain` formats.

| Command                          | Description                                       |
|----------------------------------|---------------------------------------------------|
| `ical calendars`                  | List all calendars                                |
| `ical calendars create [title]`   | Create a new calendar                             |
| `ical calendars update [name]`    | Update a calendar (rename, recolor)               |
| `ical calendars delete [name]`    | Delete a calendar and all its events              |
| `ical list`                       | List events in a date range                       |
| `ical today`                      | Today's events                                    |
| `ical upcoming`                   | Events in next N days                             |
| `ical show [# or id]`            | Show event details                                |
| `ical add [title]`               | Create an event                                   |
| `ical update [# or id]`          | Update an event                                   |
| `ical delete [# or id]`          | Delete an event                                   |
| `ical search [query]`            | Search events                                     |
| `ical export`                     | Export events (JSON/CSV/ICS)                      |
| `ical import [file]`             | Import events (JSON/CSV)                          |
| `ical skills install`             | Install AI agent skill (Claude Code / Codex)      |
| `ical skills uninstall`           | Remove AI agent skill                             |
| `ical skills status`              | Show skill installation status                    |
| `ical version`                    | Show version info                                 |
| `ical completion`                 | Generate shell completions                        |

## Global Flags

These flags are available on all commands:

| Flag         | Short | Default | Description                                     |
|--------------|-------|---------|--------------------------------------------------|
| `--output`   | `-o`  | `table` | Output format: `table`, `json`, `plain`          |
| `--no-color` |       | `false` | Disable color output (also respects `NO_COLOR`)  |

---

## ical calendars

Manage calendars. Running without a subcommand lists all calendars.

```bash
ical calendars
ical cals
ical calendars -o json
```

Displays the calendar name, source (iCloud, Google, etc.), type, color, and read-only status. Useful for finding the exact calendar name to pass to `-c` and discovering available sources for creating new calendars.

---

### ical calendars create

Create a new calendar. Requires `--source` to specify the account.

```bash
# Create with flags
ical calendars create "Projects" --source iCloud --color "#FF6961"

# Interactive mode
ical calendars create -i
```

#### Flags

| Flag              | Short | Description                                      |
|-------------------|-------|--------------------------------------------------|
| `--title`         | `-T`  | Calendar title                                   |
| `--source`        | `-s`  | Account source — required (e.g., "iCloud")       |
| `--color`         |       | Calendar color (hex, e.g., "#FF6961")            |
| `--interactive`   | `-i`  | Interactive mode with guided prompts             |

Run `ical calendars` to see available sources from existing calendars.

---

### ical calendars update

Update an existing calendar (rename or recolor).

```bash
# Update by name
ical calendars update "Projects" --title "Archived" --color "#8295AF"

# Interactive mode (guided form)
ical calendars update "Projects" -i

# Interactive picker (no argument)
ical calendars update -i
```

#### Flags

| Flag              | Short | Description                                  |
|-------------------|-------|----------------------------------------------|
| `--title`         | `-T`  | New calendar title                           |
| `--color`         |       | New calendar color (hex, e.g., "#42D692")    |
| `--interactive`   | `-i`  | Interactive mode with guided prompts         |

Subscribed and birthday calendars cannot be updated (immutable).

---

### ical calendars delete

Permanently delete a calendar and all its events.

```bash
# Delete by name (with confirmation)
ical calendars delete "Projects"

# Skip confirmation
ical calendars delete "Projects" --force

# Interactive picker (no argument)
ical calendars delete
```

#### Flags

| Flag      | Short | Description                |
|-----------|-------|----------------------------|
| `--force` | `-f`  | Skip confirmation prompt   |

Subscribed and birthday calendars cannot be deleted (immutable).

---

## ical list

List events within a date range.

```bash
ical list -f "next monday" -t "next friday"
ical list -f today -t "in 7 days" -c Work
```

### Flags

| Flag                  | Short | Description                        |
|-----------------------|-------|------------------------------------|
| `--from`              | `-f`  | Start date (natural language)      |
| `--to`                | `-t`  | End date (natural language)        |
| `--calendar`          | `-c`  | Filter by calendar name            |
| `--exclude-calendar`  |       | Exclude calendar (repeatable)      |
| `--limit`             | `-n`  | Maximum number of results          |
| `--sort`              |       | Sort by: `title`, `time`, `calendar` |

Events are displayed with row numbers (`#1`, `#2`, ...) that can be used with `show`, `update`, and `delete`. The row mapping is cached to `~/.ical-last-list` so subsequent commands can reference events by number.

```bash
# List, then act on event #2
ical list -f today -t "next friday"
ical show 2
```

---

## ical today

Show today's events. A convenience shortcut for `ical list -f today -t today`.

```bash
ical today
ical today -c Work
ical today -o json
```

### Flags

| Flag                  | Short | Description                   |
|-----------------------|-------|-------------------------------|
| `--calendar`          | `-c`  | Filter by calendar name       |
| `--exclude-calendar`  |       | Exclude calendar (repeatable) |

---

## ical upcoming

Show events for the next N days (default: 7).

```bash
ical upcoming
ical upcoming -d 30
ical upcoming -d 14 -c Work --exclude-calendar Birthdays
```

### Flags

| Flag                  | Short | Default | Description                        |
|-----------------------|-------|---------|------------------------------------|
| `--days`              | `-d`  | `7`     | Number of days to look ahead       |
| `--calendar`          | `-c`  |         | Filter by calendar name            |
| `--exclude-calendar`  |       |         | Exclude calendar (repeatable)      |
| `--limit`             | `-n`  |         | Maximum number of results          |
| `--sort`              |       |         | Sort by: `title`, `time`, `calendar` |

---

## ical show

Display detailed information about a single event.

```bash
# Interactive picker (no argument)
ical show

# By row number from last listing
ical show 2

# By event ID
ical show 577B8983-DF44-4665-B0F9-ABCD1234
```

### Event Selection

Events can be selected three ways:

1. **Interactive picker** — Run with no argument to get a searchable list
2. **Row number** — Use `#N` from the last `list`, `today`, or `upcoming` output
3. **Event ID** — Pass a full or partial `eventIdentifier` (for scripting)

The show command displays title, calendar, start/end times, location, notes, alerts, recurrence rules, URL, and attendees.

---

## ical add

Create a new calendar event.

```bash
# With flags
ical add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Interactive guided form
ical add -i
```

### Flags

| Flag              | Short | Description                                |
|-------------------|-------|--------------------------------------------|
| `--start`         | `-s`  | Start time (natural language)              |
| `--end`           | `-e`  | End time (natural language)                |
| `--calendar`      | `-c`  | Calendar to add to                         |
| `--location`      | `-l`  | Event location                             |
| `--all-day`       |       | Create an all-day event                    |
| `--alert`         |       | Add alert before event (repeatable)        |
| `--timezone`      |       | Timezone (e.g., `America/New_York`)        |
| `--repeat`        |       | Recurrence: `daily`, `weekly`, `monthly`, `yearly` |
| `--repeat-days`   |       | Days for weekly recurrence (repeatable)    |
| `--repeat-until`  |       | Recurrence end date                        |
| `--interactive`   | `-i`  | Use guided interactive form                |

### Examples

```bash
# All-day event
ical add "Company Holiday" -s 2026-03-15 --all-day -c Work

# With location and multiple alerts
ical add "Dinner" -s "friday 7pm" -e "friday 9pm" \
  -l "The Restaurant, 123 Main St" --alert 1h --alert 15m

# Weekly recurring event
ical add "Weekly Sync" -s "next monday 10am" -e "next monday 11am" \
  --repeat weekly --repeat-days mon -c Work

# Recurring with end date
ical add "Daily Standup" -s "tomorrow 9am" -e "tomorrow 9:15am" \
  --repeat daily --repeat-until "2026-12-31" -c Work

# With timezone
ical add "NYC Meeting" -s "tomorrow 2pm" -e "tomorrow 3pm" \
  --timezone "America/New_York" -c Work
```

### Interactive Mode

The `-i` flag launches a guided form where you fill in each field step by step. The form uses the Catppuccin theme and supports calendar selection from a dropdown.

---

## ical update

Update an existing event.

```bash
# Interactive picker + guided form
ical update -i

# Update by row number
ical update 2 --title "New Title"

# Reschedule
ical update 3 -s "tomorrow 2pm" -e "tomorrow 3pm"

# Update future occurrences of recurring event
ical update 1 --span future --title "New Series Name"
```

### Flags

| Flag              | Short | Description                                |
|-------------------|-------|--------------------------------------------|
| `--title`         |       | New title                                  |
| `--start`         | `-s`  | New start time (natural language)          |
| `--end`           | `-e`  | New end time (natural language)            |
| `--calendar`      | `-c`  | Move to different calendar                 |
| `--location`      | `-l`  | New location                               |
| `--all-day`       |       | Toggle all-day status                      |
| `--alert`         |       | Replace alerts (repeatable)                |
| `--timezone`      |       | New timezone                               |
| `--repeat`        |       | New recurrence pattern                     |
| `--repeat-days`   |       | New recurrence days                        |
| `--repeat-until`  |       | New recurrence end date                    |
| `--span`          |       | Apply to: `this` or `future` occurrences   |
| `--interactive`   | `-i`  | Use guided interactive form                |

---

## ical delete

Delete an event with a confirmation prompt.

```bash
# Interactive picker with confirmation
ical delete

# Delete by row number
ical delete 3

# Skip confirmation
ical delete 3 -f

# Delete future occurrences of recurring event
ical delete 2 --span future
```

### Flags

| Flag       | Short | Description                              |
|------------|-------|------------------------------------------|
| `--force`  | `-f`  | Skip confirmation prompt                 |
| `--span`   |       | Apply to: `this` or `future` occurrences |

---

## ical search

Search events by title and description.

```bash
ical search "standup"
ical search "meeting" -c Work -f "1 month ago" -t "in 1 month"
```

### Flags

| Flag                  | Short | Description                        |
|-----------------------|-------|------------------------------------|
| `--from`              | `-f`  | Start date (natural language)      |
| `--to`                | `-t`  | End date (natural language)        |
| `--calendar`          | `-c`  | Filter by calendar name            |
| `--exclude-calendar`  |       | Exclude calendar (repeatable)      |
| `--limit`             | `-n`  | Maximum number of results          |
| `--sort`              |       | Sort by: `title`, `time`, `calendar` |

---

## ical export

Export events to a file or stdout.

```bash
# Export to JSON
ical export -f 2026-01-01 -t 2026-12-31 --format json > events.json

# Export to CSV
ical export -c Work --format csv --output-file work-events.csv

# Export to ICS (RFC 5545)
ical export --format ics --output-file calendar.ics
```

### Flags

| Flag             | Short | Default | Description                    |
|------------------|-------|---------|--------------------------------|
| `--format`       |       | `json`  | Export format: `json`, `csv`, `ics` |
| `--from`         | `-f`  |         | Start date filter              |
| `--to`           | `-t`  |         | End date filter                |
| `--calendar`     | `-c`  |         | Filter by calendar             |
| `--output-file`  |       |         | Save to file (default: stdout) |

### Formats

- **JSON**: Full event data including IDs, timestamps, recurrence rules
- **CSV**: Tabular format suitable for spreadsheets
- **ICS**: RFC 5545 iCalendar format, compatible with any calendar app

---

## ical import

Import events from a JSON or CSV file.

```bash
ical import events.json
ical import events.csv -c Personal
ical import events.json --dry-run
```

### Flags

| Flag          | Short | Description                            |
|---------------|-------|----------------------------------------|
| `--calendar`  | `-c`  | Target calendar for imported events    |
| `--dry-run`   |       | Preview import without creating events |

The format is auto-detected from the file extension (`.json` or `.csv`).

---

## ical skills

Manage the embedded AI agent skill. ical ships with an [agent skill](https://agentskills.io) baked into the binary that teaches AI coding agents how to use it.

### ical skills install

Install the skill to the selected agent's skill directory. Without `--agent`, shows an interactive picker.

```bash
# Interactive — pick which agents
ical skills install

# Direct
ical skills install --agent claude   # → ~/.claude/skills/ical-cli/
ical skills install --agent codex    # → ~/.agents/skills/ical-cli/
ical skills install --agent all      # Both
```

#### Flags

| Flag      | Description                          |
|-----------|--------------------------------------|
| `--agent` | Target agent: `claude`, `codex`, or `all` |

Supported targets:
- **claude** → `~/.claude/skills/ical-cli/` — works with Claude Code, GitHub Copilot, Cursor, OpenCode, Augment
- **codex** → `~/.agents/skills/ical-cli/` — works with Codex CLI, GitHub Copilot, Windsurf, OpenCode, Augment

---

### ical skills uninstall

Remove the skill from the selected agent's skill directory.

```bash
ical skills uninstall
ical skills uninstall --agent claude
```

#### Flags

| Flag      | Description                          |
|-----------|--------------------------------------|
| `--agent` | Target agent: `claude`, `codex`, or `all` |

---

### ical skills status

Show where skills are installed and whether they match the current binary version.

```bash
ical skills status
```

If installed skills are outdated (from a previous version), ical will show a warning and prompt you to run `ical skills install` to update them.

---

## ical version

Display the installed version of ical.

```bash
ical version
```

---

## ical completion

Generate shell completion scripts.

```bash
# Bash
ical completion bash > /usr/local/etc/bash_completion.d/ical

# Zsh
ical completion zsh > "${fpath[1]}/_ical"

# Fish
ical completion fish > ~/.config/fish/completions/ical.fish
```

After generating, restart your shell or source the completion file to activate.
