---
name: ical-cli
description: Manages macOS Calendar events and calendars from the terminal via the ical CLI. Full CRUD for events and calendars with natural-language dates, recurrence, alerts, interactive mode, and JSON/CSV/ICS import/export. Use when the user wants to interact with Apple Calendar from the command line, automate calendar workflows, or build scripts around macOS Calendar.
license: MIT
compatibility: Requires macOS with Calendar.app access and the ical CLI installed (https://ical.sidv.dev)
allowed-tools: Bash(ical *)
argument-hint: "[natural language request]"
---

# ical — macOS Calendar CLI

## Current date context

Resolved fresh every time the skill loads. Prefer these over guessing from context.

- Today: !`date +"%A, %B %-d, %Y"`
- Today (ISO): !`date +"%Y-%m-%d"`
- Local time: !`date +"%H:%M %Z"`
- Tomorrow (ISO): !`date -v+1d +"%Y-%m-%d"`
- Next Monday (ISO, always forward): !`date -v+1d -v+mon +"%Y-%m-%d"`
- Next Friday (ISO, always forward): !`date -v+1d -v+fri +"%Y-%m-%d"`
- End of week (Sunday, ISO): !`date -v+1d -v+sun +"%Y-%m-%d"`

ical also accepts natural-language strings directly (`today`, `tomorrow`, `next friday`, `in 3 hours`, `eow`, `mar 15`). When in doubt, pass the user's own phrasing through — the parser handles it.

## When to use this skill

| User intent | Command |
| --- | --- |
| "What's on my calendar today" | `ical today` |
| "What's coming up this week" | `ical upcoming --days 7` |
| "List events between X and Y" | `ical list --from X --to Y` |
| "Show me event N" | `ical show <row-number>` |
| "Add / schedule / book a meeting" | `ical add "title" --start X --end Y --calendar C` |
| "Move / reschedule an event" | `ical update <row-number> --start X --end Y` |
| "Rename / retitle an event" | `ical update <row-number> --title "new"` |
| "Change an event's notes / location" | `ical update <row-number> --notes "..." --location "..."` |
| "Cancel / delete an event" | `ical delete <row-number> --force` |
| "Find events about X" | `ical search "X" --from today --to "in 30 days"` |
| "Events involving <person>" | `ical list --from today --to "in 14 days" --attendee <name>` |
| "Show only one-off events" | `ical upcoming --days 7 --no-recurring` |
| "List / create / rename / delete calendars" | `ical calendars [create\|update\|delete]` |
| "Export / back up events" | `ical export --format json --output-file backup.json` |

This table covers the common intents. If you need a flag that isn't shown above, **don't guess** — either run `ical <command> --help` (or `-h`) to get the authoritative flag list with defaults, or load [references/commands.md](references/commands.md) for the full reference. `--help` is fast, accurate, and safe to run repeatedly; prefer it over guessing a flag name from convention.

Load [references/commands.md](references/commands.md) when you need every column of a flag (short form, default, type), or when `--help` alone isn't enough context.

Load [references/dates.md](references/dates.md) when a date string fails to parse, or when the user asks what date formats are supported.

## Workflow: identify an event before acting on it

Agents usually can't assume they know the right event ID. The robust pattern:

1. Run a listing (`ical list`, `ical today`, `ical upcoming`) to find the event.
2. Note the row number (`#1`, `#2`...) shown in the output.
3. Act on it by row number: `ical show 2`, `ical update 3 --title "..."`, `ical delete 1 --force`.

Row numbers are cached to `~/.ical-last-list` and stay valid until the next listing command runs. If you need a stable reference across sessions, capture the full event ID with `-o json | jq -r '.[0].id'` and use `--id "<id>"` for exact lookup.

## Gotchas (read before running)

- **`ical delete` prompts for interactive confirmation.** In any non-interactive context, pass `--force`. There is no `--confirm` flag.
- **`ical update` has no `--force` and never confirms.** Run it directly with the flags you want changed.
- **Row numbers reset on every listing.** Running `ical today` invalidates the row numbers from a previous `ical list`.
- **`--id` is exact match only.** No prefix search, no partial match. Pass a full event ID from JSON output.
- **`--id` and positional event args are mutually exclusive.** Pass one or the other.
- **`--repeat-days` only applies to `--repeat weekly`.** With any other frequency the CLI errors out. The recurrence engine silently discards the days otherwise.
- **Timezone abbreviations (EST, CDT, BST, IST...) are rejected** inside date strings. Use `--timezone America/New_York` instead, with IANA names.
- **Event IDs are calendar-scoped.** The UUID before `:` is the calendar ID shared by every event in that calendar. Short prefixes cannot disambiguate events within one calendar — prefer row numbers or `--id "<full>"`.
- **Attendees and organizers are read-only** (Apple EventKit limitation). `ical add` does not accept `--attendee`. The `--attendee` flag on list/search is a filter, not an invite.
- **Subscribed calendars and the Birthdays calendar are read-only.** Event creation against them fails.
- **`--calendar` / `-c` is repeatable.** Pass multiple times to filter by several calendars: `ical list -c Work -c Personal`. Single `-c` is optimized server-side; multiple values filter client-side.
- **Calendar-name matching on `--calendar` and `--exclude-calendar` is case-insensitive and whitespace-trimmed**, so `"  Work "` and `work` both match a calendar named `Work`.
- **EventKit adjusts some hex colors** during save (e.g. `#FF6961` → `#FF8073`). This is CGColor conversion, not a bug.
- **`ical` is macOS-only.** No fallback on Linux or Windows.

## Output formats

All read commands accept `-o`:

- `table` (default) — bordered, colored, human-oriented, with a `Date` column that prints only on day transitions. When events span multiple years, the year is included in the date column
- `json` — ISO 8601 timestamps, full fields, safe for scripts and agents
- `plain` — one event per line, grep-friendly

**Event JSON fields**: `id`, `title`, `start_date`, `end_date`, `all_day`, `calendar`, `calendar_id`, `location`, `notes`, `url`, `status`, `availability`, `organizer`, `attendees`, `recurring`, `recurrence_rules`, `alerts`, `timezone`, `created_at`, `modified_at`.

**Calendar JSON fields**: `id`, `title`, `type`, `color`, `source`, `readOnly`. Note the list key is `title`, not `name`.

### JSON output gotchas

- Dates are **ISO 8601 UTC** (`2026-04-18T15:00:00Z`), not local time. Convert in jq with `fromdate | strftime("%Y-%m-%d %H:%M")` if you need local wall-clock.
- Optional fields (`location`, `url`, `notes`, `organizer`, `attendees`, `recurrence_rules`, `timezone`) are **omitted when empty** — use `.location // ""` in jq rather than assuming the key exists.
- `recurrence_rules` is an **array**. An event with one rule still comes back as `[rule]`. The cheap "is this event repeating" check is the top-level `recurring: true` boolean.
- Inside a recurrence rule, `frequency` is an **integer enum** (`0=daily`, `1=weekly`, `2=monthly`, `3=yearly`), not a string. Compare against the int.
- `alerts[].relativeOffset` is a **negative nanosecond duration** for before-event alerts. 15 minutes before = `-900000000000`. Divide by `-1e9` for seconds, or use `((. / -1000000000) / 60)` in jq for minutes.
- `attendees[].status` is an **integer**, not a string — unlike event-level `status` and `availability` which serialize as strings. Map the int yourself if you need a label.
- `attendees` has no write path — the array is read-only no matter what you do with `ical add` or `ical update`.

## Interactive mode

- `ical add -i` — guided form for title, calendar, dates, location, recurrence, alerts.
- `ical update <n> -i` — pre-filled form with current values.
- `ical show` / `ical update` / `ical delete` with zero args — launches a searchable event picker.

Skip `-i` and zero-arg invocations when running non-interactively — they block on stdin. Agents should always pass explicit flags or a row number.

## Recurrence

```bash
# Daily
ical add "Standup" --start "tomorrow at 9am" --repeat daily

# Every 2 weeks on Mon and Wed (only weekly accepts --repeat-days)
ical add "Team sync" --start "next monday at 10am" \
  --repeat weekly --repeat-interval 2 --repeat-days mon,wed

# Monthly for 6 occurrences
ical add "Review" --start "mar 1 at 2pm" --repeat monthly --repeat-count 6

# Yearly until a date
ical add "Anniversary" --start "jun 15" --repeat yearly --repeat-until "2030-06-15"
```

`ical update <n> --repeat none` removes recurrence. `ical update <n> --span future` changes this and all future occurrences; without it only the single instance is modified.

## Alerts

Repeatable `--alert` flag on `ical add` and `ical update`. Units: `m`, `h`, `d`.

```bash
ical add "Flight" --start "mar 15 at 8am" --alert 1h --alert 1d
```

Calendars contribute their own default alerts to every new event. Two ways to opt out:

```bash
# Explicit --alert already overrides the calendar default — you get
# exactly the alerts you list, no calendar defaults mixed in
ical add "Focus time" --start "tomorrow 2pm" --end "tomorrow 4pm" --alert 15m

# No alerts at all (useful for mirrored busy blocks)
ical add "Busy block" --start "tomorrow 9am" --end "tomorrow 10am" --no-alert
```

Rule of thumb: passing **any** `--alert` gives you exactly those alerts. Passing **no** `--alert` and no `--no-alert` inherits the calendar's default alerts. Passing `--no-alert` gives you zero alerts regardless of the calendar.

## Calendar management

```bash
ical calendars                                              # list (alias: cals)
ical calendars create "Projects" --source iCloud --color "#FF6961"
ical calendars update "Projects" --name "Side projects"
ical calendars delete "Projects" --force
```

`create` requires `--source`. Valid sources come from existing calendars — inspect `ical calendars -o json | jq -r '.[].source' | sort -u` to discover them on the user's machine.

## Common recipes

```bash
# Count today's events
ical today -o json | jq 'length'

# First event of the day
ical today -o json | jq -r '.[0] | "\(.title) at \(.start_date)"'

# Next event involving a teammate
ical list --from today --to "in 14 days" --attendee claire -o json | jq '.[0]'

# Rest of the week, skipping recurring noise
ical list --from today --to "end of week" --no-recurring

# Bulk delete everything matching a search (careful)
ical search "temp" --from today --to "in 7 days" -o json \
  | jq -r '.[].id' \
  | xargs -I {} ical delete --id {} --force

# Weekly agenda export
ical export --from today --to "in 7 days" --format ics --output-file week.ics
```

## Limits

- Attendee invites are not supported (EventKit is read-only for attendees).
- Subscribed and Birthdays calendars are read-only.
- macOS only.
