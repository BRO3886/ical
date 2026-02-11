
## Overview

cal provides 13 commands for managing macOS Calendar events. Every command that displays data supports `--output` (`-o`) with `table`, `json`, or `plain` formats.

| Command                | Description                                       |
|------------------------|---------------------------------------------------|
| `cal calendars`        | List all calendars                                |
| `cal list`             | List events in a date range                       |
| `cal today`            | Today's events                                    |
| `cal upcoming`         | Events in next N days                             |
| `cal show [# or id]`   | Show event details                               |
| `cal add [title]`      | Create an event                                  |
| `cal update [# or id]` | Update an event                                  |
| `cal delete [# or id]` | Delete an event                                  |
| `cal search [query]`   | Search events                                    |
| `cal export`           | Export events (JSON/CSV/ICS)                     |
| `cal import [file]`    | Import events (JSON/CSV)                         |
| `cal version`          | Show version info                                |
| `cal completion`       | Generate shell completions                       |

## Global Flags

These flags are available on all commands:

| Flag         | Short | Default | Description                                     |
|--------------|-------|---------|--------------------------------------------------|
| `--output`   | `-o`  | `table` | Output format: `table`, `json`, `plain`          |
| `--no-color` |       | `false` | Disable color output (also respects `NO_COLOR`)  |

---

## cal calendars

List all available calendars on this Mac.

```bash
cal calendars
```

Displays the calendar name, source (iCloud, Google, etc.), and type. Useful for finding the exact calendar name to pass to `-c`.

---

## cal list

List events within a date range.

```bash
cal list -f "next monday" -t "next friday"
cal list -f today -t "in 7 days" -c Work
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

Events are displayed with row numbers (`#1`, `#2`, ...) that can be used with `show`, `update`, and `delete`. The row mapping is cached to `~/.cal-last-list` so subsequent commands can reference events by number.

```bash
# List, then act on event #2
cal list -f today -t "next friday"
cal show 2
```

---

## cal today

Show today's events. A convenience shortcut for `cal list -f today -t today`.

```bash
cal today
cal today -c Work
cal today -o json
```

### Flags

| Flag                  | Short | Description                   |
|-----------------------|-------|-------------------------------|
| `--calendar`          | `-c`  | Filter by calendar name       |
| `--exclude-calendar`  |       | Exclude calendar (repeatable) |

---

## cal upcoming

Show events for the next N days (default: 7).

```bash
cal upcoming
cal upcoming -d 30
cal upcoming -d 14 -c Work --exclude-calendar Birthdays
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

## cal show

Display detailed information about a single event.

```bash
# Interactive picker (no argument)
cal show

# By row number from last listing
cal show 2

# By event ID
cal show 577B8983-DF44-4665-B0F9-ABCD1234
```

### Event Selection

Events can be selected three ways:

1. **Interactive picker** — Run with no argument to get a searchable list
2. **Row number** — Use `#N` from the last `list`, `today`, or `upcoming` output
3. **Event ID** — Pass a full or partial `eventIdentifier` (for scripting)

The show command displays title, calendar, start/end times, location, notes, alerts, recurrence rules, URL, and attendees.

---

## cal add

Create a new calendar event.

```bash
# With flags
cal add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Interactive guided form
cal add -i
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
cal add "Company Holiday" -s 2026-03-15 --all-day -c Work

# With location and multiple alerts
cal add "Dinner" -s "friday 7pm" -e "friday 9pm" \
  -l "The Restaurant, 123 Main St" --alert 1h --alert 15m

# Weekly recurring event
cal add "Weekly Sync" -s "next monday 10am" -e "next monday 11am" \
  --repeat weekly --repeat-days mon -c Work

# Recurring with end date
cal add "Daily Standup" -s "tomorrow 9am" -e "tomorrow 9:15am" \
  --repeat daily --repeat-until "2026-12-31" -c Work

# With timezone
cal add "NYC Meeting" -s "tomorrow 2pm" -e "tomorrow 3pm" \
  --timezone "America/New_York" -c Work
```

### Interactive Mode

The `-i` flag launches a guided form where you fill in each field step by step. The form uses the Catppuccin theme and supports calendar selection from a dropdown.

---

## cal update

Update an existing event.

```bash
# Interactive picker + guided form
cal update -i

# Update by row number
cal update 2 --title "New Title"

# Reschedule
cal update 3 -s "tomorrow 2pm" -e "tomorrow 3pm"

# Update future occurrences of recurring event
cal update 1 --span future --title "New Series Name"
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

## cal delete

Delete an event with a confirmation prompt.

```bash
# Interactive picker with confirmation
cal delete

# Delete by row number
cal delete 3

# Skip confirmation
cal delete 3 -f

# Delete future occurrences of recurring event
cal delete 2 --span future
```

### Flags

| Flag       | Short | Description                              |
|------------|-------|------------------------------------------|
| `--force`  | `-f`  | Skip confirmation prompt                 |
| `--span`   |       | Apply to: `this` or `future` occurrences |

---

## cal search

Search events by title and description.

```bash
cal search "standup"
cal search "meeting" -c Work -f "1 month ago" -t "in 1 month"
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

## cal export

Export events to a file or stdout.

```bash
# Export to JSON
cal export -f 2026-01-01 -t 2026-12-31 --format json > events.json

# Export to CSV
cal export -c Work --format csv --output-file work-events.csv

# Export to ICS (RFC 5545)
cal export --format ics --output-file calendar.ics
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

## cal import

Import events from a JSON or CSV file.

```bash
cal import events.json
cal import events.csv -c Personal
cal import events.json --dry-run
```

### Flags

| Flag          | Short | Description                            |
|---------------|-------|----------------------------------------|
| `--calendar`  | `-c`  | Target calendar for imported events    |
| `--dry-run`   |       | Preview import without creating events |

The format is auto-detected from the file extension (`.json` or `.csv`).

---

## cal version

Display the installed version of cal.

```bash
cal version
```

---

## cal completion

Generate shell completion scripts.

```bash
# Bash
cal completion bash > /usr/local/etc/bash_completion.d/cal

# Zsh
cal completion zsh > "${fpath[1]}/_cal"

# Fish
cal completion fish > ~/.config/fish/completions/cal.fish
```

After generating, restart your shell or source the completion file to activate.

