# Complete Command Reference

## cal calendars

Manage calendars. Running without a subcommand lists all calendars.

```bash
cal calendars
cal cals
cal calendars -o json
```

Aliases: `cals`

### cal calendars create

Create a new calendar. Title can be passed as a positional argument or via `--title`.

```bash
cal calendars create "Projects" --source iCloud --color "#FF6961"
cal calendars create -i
```

| Flag              | Short | Description                                | Default |
| ----------------- | ----- | ------------------------------------------ | ------- |
| `--title`         | `-T`  | Calendar title                             | —       |
| `--source`        | `-s`  | Account source (required, e.g., "iCloud")  | —       |
| `--color`         | —     | Calendar color (hex, e.g., "#FF6961")      | —       |
| `--interactive`   | `-i`  | Interactive mode with guided prompts       | false   |

Aliases: `add`, `new`

### cal calendars update

Update an existing calendar (rename, recolor). With no arguments, shows an interactive picker.

```bash
cal calendars update "Projects" --title "Archived" --color "#8295AF"
cal calendars update "Projects" -i
cal calendars update -i
```

| Flag              | Short | Description                              | Default |
| ----------------- | ----- | ---------------------------------------- | ------- |
| `--title`         | `-T`  | New calendar title                       | —       |
| `--color`         | —     | New calendar color (hex)                 | —       |
| `--interactive`   | `-i`  | Interactive mode with guided prompts     | false   |

Aliases: `edit`, `rename`

### cal calendars delete

Permanently delete a calendar and all its events. With no arguments, shows an interactive picker.

```bash
cal calendars delete "Projects"
cal calendars delete "Projects" --force
cal calendars delete
```

| Flag      | Short | Description              | Default |
| --------- | ----- | ------------------------ | ------- |
| `--force` | `-f`  | Skip confirmation prompt | false   |

Aliases: `rm`, `remove`

---

## cal list

List events within a date range. Defaults to today if no range specified.

```bash
cal list
cal list --from today --to "next friday"
cal list --calendar Work --sort title
cal ls --from "mar 1" --to "mar 31" --all-day
cal list --from today --to "in 30 days" --exclude-calendar Birthdays -o json
```

| Flag                 | Short | Description                               | Default         |
| -------------------- | ----- | ----------------------------------------- | --------------- |
| `--from`             | `-f`  | Start date (natural language or ISO 8601) | Today           |
| `--to`               | `-t`  | End date (natural language or ISO 8601)   | From + 24 hours |
| `--calendar`         | `-c`  | Filter by calendar name                   | All calendars   |
| `--calendar-id`      | —     | Filter by calendar ID                     | —               |
| `--search`           | `-s`  | Search title, location, notes             | —               |
| `--all-day`          | —     | Show only all-day events                  | false           |
| `--sort`             | —     | Sort by: start, end, title, calendar      | start           |
| `--limit`            | `-n`  | Max events to display (0 = unlimited)     | 0               |
| `--exclude-calendar` | —     | Exclude calendars by name (repeatable)    | —               |

Aliases: `ls`, `events`

---

## cal today

Show today's events. Shortcut for `cal list --from today --to tomorrow`.

```bash
cal today
cal today --calendar Work
cal today -o json
cal today --exclude-calendar Birthdays
```

| Flag                 | Short | Description                            | Default       |
| -------------------- | ----- | -------------------------------------- | ------------- |
| `--calendar`         | `-c`  | Filter by calendar name                | All calendars |
| `--calendar-id`      | —     | Filter by calendar ID                  | —             |
| `--search`           | `-s`  | Search title, location, notes          | —             |
| `--all-day`          | —     | Show only all-day events               | false         |
| `--sort`             | —     | Sort by: start, end, title, calendar   | start         |
| `--limit`            | `-n`  | Max events to display (0 = unlimited)  | 0             |
| `--exclude-calendar` | —     | Exclude calendars by name (repeatable) | —             |

---

## cal upcoming

Show events in the next N days. Shortcut for `cal list --from today --to "in N days"`.

```bash
cal upcoming
cal upcoming --days 14
cal upcoming --calendar Work -o json
cal next --days 3
```

| Flag                 | Short | Description                            | Default       |
| -------------------- | ----- | -------------------------------------- | ------------- |
| `--days`             | `-d`  | Number of days to look ahead           | 7             |
| `--calendar`         | `-c`  | Filter by calendar name                | All calendars |
| `--calendar-id`      | —     | Filter by calendar ID                  | —             |
| `--search`           | `-s`  | Search title, location, notes          | —             |
| `--all-day`          | —     | Show only all-day events               | false         |
| `--sort`             | —     | Sort by: start, end, title, calendar   | start         |
| `--limit`            | `-n`  | Max events to display (0 = unlimited)  | 0             |
| `--exclude-calendar` | —     | Exclude calendars by name (repeatable) | —             |

Aliases: `next`, `soon`

---

## cal show

Display full details for a single event. With no arguments, shows an interactive picker.

```bash
cal show 2                    # Row number from last listing
cal show                      # Interactive event picker
cal show ABC12345 -o json     # Full or partial event ID
```

| Flag     | Short | Description                      | Default |
| -------- | ----- | -------------------------------- | ------- |
| `--from` | `-f`  | Start date for event picker      | Today   |
| `--to`   | `-t`  | End date for event picker        | —       |
| `--days` | `-d`  | Number of days to show in picker | 7       |

Event selection methods:

1. **No argument** — interactive huh picker of upcoming events
2. **Number** (e.g., `2`) — row number from the last `list`/`today`/`upcoming` output
3. **String** — full or partial event ID

Aliases: `get`, `info`

---

## cal add

Create a new calendar event. Title can be passed as a positional argument or via `--title`.

```bash
cal add "Team standup" --start "tomorrow at 9am" --end "tomorrow at 9:30am" --calendar Work
cal add --title "Lunch" --start "today at noon" --end "today at 1pm" --location "Cafe"
cal add "Flight" --start "mar 15 at 8am" --end "mar 15 at 11am" --alert 1h --alert 1d
cal add "Weekly sync" --start "next monday at 10am" --repeat weekly --repeat-days mon
cal add -i   # Interactive mode
```

| Flag                | Short | Description                                         | Default        |
| ------------------- | ----- | --------------------------------------------------- | -------------- |
| `--title`           | `-T`  | Event title                                         | —              |
| `--start`           | `-s`  | Start date/time (required)                          | —              |
| `--end`             | `-e`  | End date/time                                       | Start + 1 hour |
| `--all-day`         | `-a`  | Create as all-day event                             | false          |
| `--calendar`        | `-c`  | Calendar name                                       | System default |
| `--location`        | `-l`  | Location string                                     | —              |
| `--notes`           | `-n`  | Notes/description                                   | —              |
| `--url`             | `-u`  | URL to attach                                       | —              |
| `--alert`           | —     | Alert before event (e.g., 15m, 1h, 1d) — repeatable | —              |
| `--repeat`          | `-r`  | Recurrence: daily, weekly, monthly, yearly          | —              |
| `--repeat-interval` | —     | Recurrence interval                                 | 1              |
| `--repeat-until`    | —     | Recurrence end date                                 | —              |
| `--repeat-count`    | —     | Number of occurrences                               | 0              |
| `--repeat-days`     | —     | Days for weekly recurrence (e.g., mon,wed,fri)      | —              |
| `--timezone`        | —     | IANA timezone (e.g., America/New_York)              | —              |
| `--interactive`     | `-i`  | Interactive mode with guided prompts                | false          |

Aliases: `create`, `new`

---

## cal update

Update an existing event. Only specified fields are changed.

```bash
cal update 2 --title "New title"
cal update 3 --start "tomorrow at 10am" --end "tomorrow at 11am"
cal update 1 --location ""              # Clear location
cal update 1 --alert none               # Clear alerts
cal update 1 --repeat none              # Remove recurrence
cal update 2 --span future --start "next monday at 9am"  # Update future occurrences
cal update -i                           # Interactive mode with picker
```

| Flag                | Short | Description                                  | Default |
| ------------------- | ----- | -------------------------------------------- | ------- |
| `--title`           | `-T`  | New title                                    | —       |
| `--start`           | `-s`  | New start date/time                          | —       |
| `--end`             | `-e`  | New end date/time                            | —       |
| `--all-day`         | `-a`  | Set all-day: "true" or "false"               | —       |
| `--calendar`        | `-c`  | Move to calendar (by name)                   | —       |
| `--location`        | `-l`  | New location (empty string to clear)         | —       |
| `--notes`           | `-n`  | New notes (empty string to clear)            | —       |
| `--url`             | `-u`  | New URL (empty string to clear)              | —       |
| `--alert`           | —     | Replace alerts (repeatable, `none` to clear) | —       |
| `--timezone`        | —     | New IANA timezone                            | —       |
| `--span`            | —     | For recurring: "this" or "future"            | this    |
| `--repeat`          | `-r`  | Set/change recurrence ("none" to remove)     | —       |
| `--repeat-interval` | —     | Change recurrence interval                   | 1       |
| `--repeat-until`    | —     | Change recurrence end date                   | —       |
| `--repeat-count`    | —     | Change recurrence count                      | 0       |
| `--repeat-days`     | —     | Change recurrence days                       | —       |
| `--interactive`     | `-i`  | Interactive mode with guided prompts         | false   |

Event selection: same as `show` (no args = picker, number = row, string = event ID).

Aliases: `edit`

---

## cal delete

Delete an event. Asks for confirmation by default.

```bash
cal delete 1
cal rm 2 --force
cal delete                    # Interactive picker
cal delete 3 --span future    # Delete this and future occurrences
```

| Flag      | Short | Description                       | Default |
| --------- | ----- | --------------------------------- | ------- |
| `--force` | `-f`  | Skip confirmation prompt          | false   |
| `--span`  | —     | For recurring: "this" or "future" | this    |
| `--from`  | —     | Start date for event picker       | Today   |
| `--to`    | —     | End date for event picker         | —       |
| `--days`  | `-d`  | Number of days to show in picker  | 7       |

Event selection: same as `show` (no args = picker, number = row, string = event ID).

Aliases: `rm`, `remove`

---

## cal search

Search events by title, location, and notes within a date range.

```bash
cal search "meeting"
cal search "standup" --calendar Work
cal search "dentist" --from "jan 1" --to "dec 31" --limit 5
cal find "lunch" -o json
```

| Flag         | Short | Description                 | Default       |
| ------------ | ----- | --------------------------- | ------------- |
| `--from`     | `-f`  | Start of search range       | 30 days ago   |
| `--to`       | `-t`  | End of search range         | 30 days ahead |
| `--calendar` | `-c`  | Filter by calendar name     | All calendars |
| `--limit`    | `-n`  | Max results (0 = unlimited) | 0             |

Aliases: `find`

---

## cal export

Export events to JSON, CSV, or ICS format.

```bash
cal export > events.json
cal export --format ics --output-file events.ics
cal export --calendar Work --from today --to "in 6 months" --format csv
```

| Flag            | Short | Description                     | Default       |
| --------------- | ----- | ------------------------------- | ------------- |
| `--from`        | `-f`  | Start date                      | 30 days ago   |
| `--to`          | `-t`  | End date                        | 30 days ahead |
| `--calendar`    | `-c`  | Filter by calendar name         | All calendars |
| `--format`      | —     | Format: json, csv, ics          | json          |
| `--output-file` | —     | Write to file instead of stdout | stdout        |

---

## cal import

Import events from a JSON or CSV file. Format is auto-detected from file extension.

```bash
cal import events.json
cal import events.csv --calendar "Imported"
cal import backup.json --dry-run
cal import data.json --force
```

| Flag         | Short | Description                             | Default           |
| ------------ | ----- | --------------------------------------- | ----------------- |
| `--calendar` | `-c`  | Override target calendar for all events | Original calendar |
| `--dry-run`  | —     | Preview without creating events         | false             |
| `--force`    | `-f`  | Skip confirmation prompt                | false             |

---

## cal version

Print version and build information.

```bash
cal version
```

Output format: `cal <version> (commit <hash>, built <date>)`

---

## cal completion

Generate shell completion scripts.

```bash
cal completion bash > /usr/local/etc/bash_completion.d/cal
cal completion zsh > "${fpath[1]}/_cal"
cal completion fish > ~/.config/fish/completions/cal.fish
```

---

## Global Flags

These flags are available on all commands:

| Flag         | Short | Description                       | Default |
| ------------ | ----- | --------------------------------- | ------- |
| `--output`   | `-o`  | Output format: table, json, plain | table   |
| `--no-color` | —     | Disable color output              | false   |

The `NO_COLOR` environment variable is also respected.
