# Complete Command Reference

## ical calendars

Manage calendars. Running without a subcommand lists all calendars.

```bash
ical calendars
ical cals
ical calendars -o json
```

Aliases: `cals`

### ical calendars create

Create a new calendar. Title can be passed as a positional argument or via `--title`.

```bash
ical calendars create "Projects" --source iCloud --color "#FF6961"
ical calendars create -i
```

| Flag              | Short | Description                                | Default |
| ----------------- | ----- | ------------------------------------------ | ------- |
| `--title`         | `-T`  | Calendar title                             | —       |
| `--source`        | `-s`  | Account source (required, e.g., "iCloud")  | —       |
| `--color`         | —     | Calendar color (hex, e.g., "#FF6961")      | —       |
| `--interactive`   | `-i`  | Interactive mode with guided prompts       | false   |

Aliases: `add`, `new`

### ical calendars update

Update an existing calendar (rename, recolor). With no arguments, shows an interactive picker.

```bash
ical calendars update "Projects" --title "Archived" --color "#8295AF"
ical calendars update "Projects" -i
ical calendars update -i
```

| Flag              | Short | Description                              | Default |
| ----------------- | ----- | ---------------------------------------- | ------- |
| `--title`         | `-T`  | New calendar title                       | —       |
| `--color`         | —     | New calendar color (hex)                 | —       |
| `--interactive`   | `-i`  | Interactive mode with guided prompts     | false   |

Aliases: `edit`, `rename`

### ical calendars delete

Permanently delete a calendar and all its events. With no arguments, shows an interactive picker.

```bash
ical calendars delete "Projects"
ical calendars delete "Projects" --force
ical calendars delete
```

| Flag      | Short | Description              | Default |
| --------- | ----- | ------------------------ | ------- |
| `--force` | `-f`  | Skip confirmation prompt | false   |

Aliases: `rm`, `remove`

---

## ical list

List events within a date range. Defaults to today if no range specified.

```bash
ical list
ical list --from today --to "next friday"
ical list --calendar Work --sort title
ical ls --from "mar 1" --to "mar 31" --all-day
ical list --from today --to "in 30 days" --exclude-calendar Birthdays -o json
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

## ical today

Show today's events. Shortcut for `ical list --from today --to tomorrow`.

```bash
ical today
ical today --calendar Work
ical today -o json
ical today --exclude-calendar Birthdays
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

## ical upcoming

Show events in the next N days. Shortcut for `ical list --from today --to "in N days"`.

```bash
ical upcoming
ical upcoming --days 14
ical upcoming --calendar Work -o json
ical next --days 3
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

## ical show

Display full details for a single event. With no arguments, shows an interactive picker.

```bash
ical show 2                                      # Row number from last listing
ical show                                        # Interactive event picker
ical show ABC12345 -o json                       # Full or partial event ID
ical show --id "577B8983-DF44:ABC123" -o json   # Exact event ID (agents: use this)
```

| Flag     | Short | Description                                  | Default |
| -------- | ----- | -------------------------------------------- | ------- |
| `--id`   | —     | Full event ID (exact match, no prefix search) | —       |
| `--from` | `-f`  | Start date for event picker                  | Today   |
| `--to`   | `-t`  | End date for event picker                    | —       |
| `--days` | `-d`  | Number of days to show in picker             | 7       |

Event selection methods:

1. **No argument** — interactive huh picker of upcoming events
2. **Number** (e.g., `2`) — row number from the last `list`/`today`/`upcoming` output
3. **String** — full or partial event ID (positional arg)
4. **`--id`** — full event ID, exact match only (preferred for scripts/agents)

> `--id` and a positional argument are mutually exclusive — passing both is an error.

Aliases: `get`, `info`

---

## ical add

Create a new calendar event. Title can be passed as a positional argument or via `--title`.

```bash
ical add "Team standup" --start "tomorrow at 9am" --end "tomorrow at 9:30am" --calendar Work
ical add --title "Lunch" --start "today at noon" --end "today at 1pm" --location "Cafe"
ical add "Flight" --start "mar 15 at 8am" --end "mar 15 at 11am" --alert 1h --alert 1d
ical add "Weekly sync" --start "next monday at 10am" --repeat weekly --repeat-days mon
ical add -i   # Interactive mode
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

## ical update

Update an existing event. Only specified fields are changed.

```bash
ical update 2 --title "New title"
ical update 3 --start "tomorrow at 10am" --end "tomorrow at 11am"
ical update 1 --location ""              # Clear location
ical update 1 --alert none               # Clear alerts
ical update 1 --repeat none              # Remove recurrence
ical update 2 --span future --start "next monday at 9am"  # Update future occurrences
ical update -i                           # Interactive mode with picker
ical update --id "577B8983-DF44:ABC123" --title "New title"  # Exact ID (agents: use this)
```

| Flag                | Short | Description                                  | Default |
| ------------------- | ----- | -------------------------------------------- | ------- |
| `--id`              | —     | Full event ID (exact match, no prefix search) | —       |
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

Event selection: same as `show` (no args = picker, number = row, string = event ID, `--id` = exact).

> `--id` and a positional argument are mutually exclusive — passing both is an error.
> `update` has **no** `--force` flag and requires no confirmation — changes apply immediately.

Aliases: `edit`

---

## ical delete

Delete an event. Asks for confirmation by default.

```bash
ical delete 1 --force                              # Row number, skip confirmation
ical rm 2 --force                                  # Alias
ical delete                                        # Interactive picker
ical delete 3 --span future                        # Delete this and future occurrences
ical delete --id "577B8983-DF44:ABC123" --force   # Exact event ID (agents: use this)
```

| Flag      | Short | Description                                  | Default |
| --------- | ----- | -------------------------------------------- | ------- |
| `--id`    | —     | Full event ID (exact match, no prefix search) | —       |
| `--force` | `-f`  | Skip confirmation prompt                     | false   |
| `--span`  | —     | For recurring: "this" or "future"            | this    |
| `--from`  | —     | Start date for event picker                  | Today   |
| `--to`    | —     | End date for event picker                    | —       |
| `--days`  | `-d`  | Number of days to show in picker             | 7       |

Event selection: same as `show` (no args = picker, number = row, string = event ID, `--id` = exact).

> `--id` and a positional argument are mutually exclusive — passing both is an error.
> `update` does **not** have a `--force` flag — it applies changes immediately without confirmation.

Aliases: `rm`, `remove`

---

## ical search

Search events by title, location, and notes within a date range.

```bash
ical search "meeting"
ical search "standup" --calendar Work
ical search "dentist" --from "jan 1" --to "dec 31" --limit 5
ical find "lunch" -o json
```

| Flag         | Short | Description                 | Default       |
| ------------ | ----- | --------------------------- | ------------- |
| `--from`     | `-f`  | Start of search range       | 30 days ago   |
| `--to`       | `-t`  | End of search range         | 30 days ahead |
| `--calendar` | `-c`  | Filter by calendar name     | All calendars |
| `--limit`    | `-n`  | Max results (0 = unlimited) | 0             |

Aliases: `find`

---

## ical export

Export events to JSON, CSV, or ICS format.

```bash
ical export > events.json
ical export --format ics --output-file events.ics
ical export --calendar Work --from today --to "in 6 months" --format csv
```

| Flag            | Short | Description                     | Default       |
| --------------- | ----- | ------------------------------- | ------------- |
| `--from`        | `-f`  | Start date                      | 30 days ago   |
| `--to`          | `-t`  | End date                        | 30 days ahead |
| `--calendar`    | `-c`  | Filter by calendar name         | All calendars |
| `--format`      | —     | Format: json, csv, ics          | json          |
| `--output-file` | —     | Write to file instead of stdout | stdout        |

---

## ical import

Import events from a JSON or CSV file. Format is auto-detected from file extension.

```bash
ical import events.json
ical import events.csv --calendar "Imported"
ical import backup.json --dry-run
ical import data.json --force
```

| Flag         | Short | Description                             | Default           |
| ------------ | ----- | --------------------------------------- | ----------------- |
| `--calendar` | `-c`  | Override target calendar for all events | Original calendar |
| `--dry-run`  | —     | Preview without creating events         | false             |
| `--force`    | `-f`  | Skip confirmation prompt                | false             |

---

## ical skills

Manage AI agent skills. The ical binary embeds its own agent skill files and can install them directly into the skills directory of supported AI coding agents.

### ical skills install

Install the ical agent skill for AI coding agents. Without `--agent`, shows an interactive multi-select picker.

```bash
ical skills install                # Interactive — pick agents
ical skills install --agent claude # Direct — Claude Code only
ical skills install --agent codex  # Direct — Codex CLI only
ical skills install --agent all    # Both agents
```

| Flag      | Short | Description                          | Default     |
| --------- | ----- | ------------------------------------ | ----------- |
| `--agent` | —     | Target agent: claude, codex, or all  | Interactive |

Supported targets:
- `claude` → `~/.claude/skills/ical-cli/` (Claude Code, Copilot, Cursor, OpenCode, Augment)
- `codex` → `~/.agents/skills/ical-cli/` (Codex CLI, Copilot, Windsurf, OpenCode, Augment)

### ical skills uninstall

Remove the ical agent skill. Same `--agent` flag and interactive picker as install.

```bash
ical skills uninstall
ical skills uninstall --agent claude
```

| Flag      | Short | Description                          | Default     |
| --------- | ----- | ------------------------------------ | ----------- |
| `--agent` | —     | Target agent: claude, codex, or all  | Interactive |

### ical skills status

Show where skills are installed and whether they match the current binary version.

```bash
ical skills status
```

---

## ical version

Print version and build information. Also shows a notice if a newer version is available or if installed skills are outdated.

```bash
ical version
```

Output format: `ical <version> (commit <hash>, built <date>)`

---

## ical completion

Generate shell completion scripts.

```bash
ical completion bash > /usr/local/etc/bash_completion.d/ical
ical completion zsh > "${fpath[1]}/_ical"
ical completion fish > ~/.config/fish/completions/ical.fish
```

---

## Global Flags

These flags are available on all commands:

| Flag         | Short | Description                       | Default |
| ------------ | ----- | --------------------------------- | ------- |
| `--output`   | `-o`  | Output format: table, json, plain | table   |
| `--no-color` | —     | Disable color output              | false   |

The `NO_COLOR` environment variable is also respected.

Set `ICAL_NO_UPDATE_CHECK=1` to disable the background update check.
