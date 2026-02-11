# cal — macOS Calendar CLI

## Overview

`cal` is a fast, native macOS Calendar CLI built on [go-eventkit](https://github.com/BRO3886/go-eventkit). It provides full CRUD for calendar events, natural language date parsing, recurrence support, import/export, and multiple output formats — all via EventKit (3000x faster than AppleScript).

**Module**: `github.com/BRO3886/cal`
**Primary dependency**: `github.com/BRO3886/go-eventkit/calendar`

## Design Principles

1. **Fast by default** — EventKit reads in <50ms, no subprocess overhead
2. **Dual mode** — every write command supports both flag-based (scriptable) and interactive (`-i`) modes
3. **Consistent output** — `--output table|json|plain` on every list/show command
4. **Familiar UX** — follows patterns from `rem` (output formats, ID prefix matching, confirmation prompts) and `gtasks` (dual-mode, filtering, sorting)
5. **No surprises** — confirmation prompts on destructive ops, `--force` to skip, `--dry-run` on imports

## Command Reference

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `table` | Output format: `table`, `json`, `plain` |
| `--no-color` | | `false` | Disable color output (also respects `NO_COLOR` env) |

---

### `cal calendars` — List all calendars

Lists all calendars across all accounts (iCloud, Google, Exchange, subscriptions, birthdays).

**Aliases**: `cals`

**Output columns** (table): Name, Source, Type, Color, ReadOnly
**JSON fields**: `id`, `title`, `source`, `type`, `color`, `read_only`

**Examples**:
```bash
cal calendars                     # Table of all calendars
cal calendars -o json             # JSON array
cal calendars -o json | jq '.[].title'
```

---

### `cal list` — List events in a date range

Queries events within a date range. Defaults to today if no range specified.

**Aliases**: `ls`, `events`

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--from` | `-f` | today 00:00 | Start date (NL or ISO 8601) |
| `--to` | `-t` | from + 24h | End date (NL or ISO 8601) |
| `--calendar` | `-c` | all | Filter by calendar name |
| `--calendar-id` | | | Filter by calendar ID |
| `--search` | `-s` | | Search title, location, notes |
| `--all-day` | | `false` | Show only all-day events |
| `--sort` | | `start` | Sort by: `start`, `end`, `title`, `calendar` |
| `--limit` | `-n` | 0 (all) | Max events to display |

**Output columns** (table): Time, Title, Calendar, Location, Duration
- All-day events: show "All Day" instead of time range
- Multi-day events: show date range
- Recurring events: show recurrence icon/indicator

**Examples**:
```bash
cal list                          # Today's events
cal list -f tomorrow              # Tomorrow's events
cal list -f "next monday" -t "next friday"  # Work week
cal list -f 2026-03-01 -t 2026-03-31       # March 2026
cal list -c Work                  # Only "Work" calendar
cal list -s standup               # Search for "standup"
cal list -f today -t "in 7 days" --sort title
cal list -o json | jq '.[] | {title, start: .start_date}'
```

---

### `cal today` — Today's events

Shortcut for `cal list --from today --to tomorrow`. Shows the day's agenda.

**Flags**: Same as `cal list` (except `--from`/`--to`)

**Examples**:
```bash
cal today                         # Today's full agenda
cal today -c Personal             # Today's personal events
cal today -o plain                # Plain text for scripting
```

---

### `cal upcoming` — Events in next N days

Shortcut for `cal list` with `--from today --to "in N days"`.

**Aliases**: `next`, `soon`

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--days` | `-d` | 7 | Number of days to look ahead |
| *(plus all `list` flags except `--from`/`--to`)* |

**Examples**:
```bash
cal upcoming                      # Next 7 days
cal upcoming -d 30                # Next 30 days
cal upcoming -c Work -d 14        # Work events, 2 weeks
```

---

### `cal show [id]` — Show event details

Displays full details for a single event. Supports ID prefix matching.

**Aliases**: `get`, `info`

**Output fields**:
- Title, Calendar, Status
- Start → End (with timezone, duration)
- All Day (if applicable)
- Location + Structured Location (lat/long if present)
- URL
- Notes (full, not truncated)
- Recurrence rules (human-readable)
- Alerts
- Attendees (name, email, RSVP status)
- Organizer
- Created/Modified timestamps

**Examples**:
```bash
cal show a1b2c3d4                 # By short ID prefix
cal show a1b2c3d4-e5f6-...       # By full ID
cal show a1b2c3d4 -o json        # JSON output
```

---

### `cal add [title]` — Create an event

Creates a new calendar event. Title can be passed as argument or via `--title` flag.

**Aliases**: `create`, `new`

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-T` | (arg) | Event title |
| `--start` | `-s` | | Start date/time (NL or ISO 8601) **required** |
| `--end` | `-e` | start + 1h | End date/time (NL or ISO 8601) |
| `--all-day` | `-a` | `false` | Create as all-day event |
| `--calendar` | `-c` | default | Calendar name |
| `--location` | `-l` | | Location string |
| `--notes` | `-n` | | Notes/description |
| `--url` | `-u` | | URL to attach |
| `--alert` | | | Alert before event (e.g., "15m", "1h", "1d") — repeatable |
| `--repeat` | `-r` | | Recurrence: `daily`, `weekly`, `monthly`, `yearly` |
| `--repeat-interval` | | 1 | Recurrence interval (e.g., 2 for "every 2 weeks") |
| `--repeat-until` | | | Recurrence end date |
| `--repeat-count` | | | Recurrence occurrence count |
| `--repeat-days` | | | Days for weekly recurrence (e.g., "mon,wed,fri") |
| `--timezone` | | system | IANA timezone (e.g., "America/New_York") |
| `--interactive` | `-i` | `false` | Interactive mode with prompts |

**Recurrence flag details**:
- `--repeat daily` — every day
- `--repeat weekly --repeat-days mon,wed,fri` — MWF
- `--repeat monthly --repeat-interval 2` — every 2 months
- `--repeat yearly` — annually
- `--repeat-until "2026-12-31"` — stop on date
- `--repeat-count 10` — stop after 10 occurrences

**Interactive mode** (`-i`): Prompts for title, start, end, calendar (select from list), location, notes, recurrence, alerts.

**Output**: Prints created event details (ID, title, time, calendar).

**Examples**:
```bash
# Quick event
cal add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# All-day event
cal add "Company Holiday" -s 2026-03-15 --all-day -c Work

# Recurring event
cal add "Weekly Sync" -s "next monday 10am" -e "next monday 11am" \
  --repeat weekly --repeat-days mon -c Work

# With location and alerts
cal add "Dinner" -s "friday 7pm" -e "friday 9pm" \
  -l "The Restaurant, 123 Main St" --alert 1h --alert 15m

# Interactive
cal add -i
```

---

### `cal update [id]` — Update an event

Updates an existing event. Only specified fields are changed.

**Aliases**: `edit`

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-T` | | New title |
| `--start` | `-s` | | New start date/time |
| `--end` | `-e` | | New end date/time |
| `--all-day` | `-a` | | Set all-day (true/false) |
| `--calendar` | `-c` | | Move to calendar (by name) |
| `--location` | `-l` | | New location (`""` to clear) |
| `--notes` | `-n` | | New notes (`""` to clear) |
| `--url` | `-u` | | New URL (`""` to clear) |
| `--alert` | | | Replace alerts (repeatable, `none` to clear) |
| `--timezone` | | | New timezone |
| `--span` | | `this` | For recurring events: `this` or `future` |
| `--repeat` | `-r` | | Set/change recurrence (`none` to remove) |
| `--repeat-interval` | | | Change recurrence interval |
| `--repeat-until` | | | Change recurrence end date |
| `--repeat-count` | | | Change recurrence count |
| `--repeat-days` | | | Change recurrence days |
| `--interactive` | `-i` | `false` | Interactive mode (shows current values) |

**Interactive mode** (`-i`): Shows current values in `[brackets]`, user presses Enter to keep or types new value.

**Examples**:
```bash
cal update a1b2c3d4 --title "Updated Title"
cal update a1b2c3d4 -s "tomorrow 2pm" -e "tomorrow 3pm"
cal update a1b2c3d4 --location ""     # Clear location
cal update a1b2c3d4 --span future --title "New Series Name"
cal update a1b2c3d4 -i                # Interactive
```

---

### `cal delete [id]` — Delete an event

Deletes an event. Asks for confirmation by default.

**Aliases**: `rm`, `remove`

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | `-f` | `false` | Skip confirmation prompt |
| `--span` | | `this` | For recurring: `this` or `future` |

**Examples**:
```bash
cal delete a1b2c3d4               # With confirmation
cal delete a1b2c3d4 -f            # Force delete
cal delete a1b2c3d4 --span future # Delete this + all future occurrences
```

---

### `cal search [query]` — Search events

Searches events within a date range by title, location, and notes.

**Aliases**: `find`

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--from` | `-f` | 30 days ago | Start of search range |
| `--to` | `-t` | 30 days ahead | End of search range |
| `--calendar` | `-c` | all | Filter by calendar |
| `--limit` | `-n` | 0 (all) | Max results |

**Examples**:
```bash
cal search standup                # Search +-30 days
cal search "team meeting" -c Work
cal search dentist -f "6 months ago" -t "in 6 months"
```

---

### `cal export` — Export events

Exports events to JSON, CSV, or ICS format.

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--from` | `-f` | 30 days ago | Start date |
| `--to` | `-t` | 30 days ahead | End date |
| `--calendar` | `-c` | all | Filter by calendar |
| `--format` | | `json` | Format: `json`, `csv`, `ics` |
| `--output-file` | | stdout | Write to file instead of stdout |

**JSON fields**: `id`, `title`, `start_date`, `end_date`, `all_day`, `calendar`, `calendar_id`, `location`, `structured_location` (if present), `notes`, `url`, `status`, `availability`, `organizer`, `attendees`, `recurring`, `recurrence_rules`, `alerts`, `timezone`, `created_at`, `modified_at`

**CSV columns**: ID, Title, Start, End, AllDay, Calendar, Location, Notes, URL, Status, Recurring, Timezone

**ICS**: Standard iCalendar format (RFC 5545) — interoperable with Google Calendar, Outlook, etc. Includes VEVENT components with RRULE for recurrence.

**Examples**:
```bash
cal export                                    # JSON to stdout
cal export --format csv --output-file events.csv
cal export -c Work -f 2026-01-01 -t 2026-12-31 --format ics --output-file work-2026.ics
cal export -o json | jq '.[] | select(.recurring)'
```

---

### `cal import [file]` — Import events

Imports events from JSON or CSV files.

**Flags**:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--calendar` | `-c` | | Override target calendar for all events |
| `--dry-run` | | `false` | Preview without creating |
| `--force` | `-f` | `false` | Skip confirmation prompt |

**Behavior**:
- Detects format by file extension (`.json`, `.csv`)
- IDs are NOT reimported (EventKit assigns new IDs)
- Per-event errors are warnings (import continues)
- Shows summary: `Created X events, Y errors`

**Examples**:
```bash
cal import events.json                        # Import from JSON
cal import events.csv -c Personal             # Override calendar
cal import events.json --dry-run              # Preview
```

---

### `cal version` — Show version

Displays version and build info.

```bash
cal version
# cal v0.1.0 (built 2026-02-11)
```

---

### `cal completion` — Shell completions

Generates shell completion scripts.

```bash
cal completion bash > /usr/local/etc/bash_completion.d/cal
cal completion zsh > "${fpath[1]}/_cal"
cal completion fish > ~/.config/fish/completions/cal.fish
```

---

## Natural Language Date Parsing

Supported patterns (same as `rem`, extended for time-of-day):

| Pattern | Example | Resolves to |
|---------|---------|-------------|
| `today` | `today` | Today 00:00 |
| `tomorrow` | `tomorrow` | Tomorrow 00:00 |
| `yesterday` | `yesterday` | Yesterday 00:00 |
| `next [day]` | `next friday` | Next Friday 00:00 |
| `next week` | `next week` | Next Monday 00:00 |
| `next month` | `next month` | 1st of next month |
| `in X hours` | `in 3 hours` | Now + 3h |
| `in X minutes` | `in 30 minutes` | Now + 30m |
| `in X days` | `in 5 days` | Today + 5d |
| `[time]` | `3pm`, `15:00` | Today at time |
| `[day] [time]` | `friday 2pm` | Next Friday at 2pm |
| `[date] [time]` | `2026-03-15 14:00` | Specific datetime |
| `eod` | `eod` | Today 17:00 |
| `eow` | `eow` | Friday 17:00 |
| `X ago` | `2 hours ago` | Now - 2h |
| `[month] [day]` | `mar 15`, `march 15` | Mar 15 current year |

Time-of-day is critical for calendar events (unlike reminders). When a time is specified with a date, it combines them. When only a date is given (no time), it defaults to 00:00 for start and 23:59 for end (or marks as all-day if `--all-day` is set).

---

## Output Formatting

### Table (default)
- Uses `olekukonko/tablewriter` v1.x (new API)
- Color: calendar color dot, priority-based status colors
- Truncation: title 40 chars, location 25 chars in table
- All-day events grouped at top
- Duration shown as human-readable ("1h 30m", "All Day", "3 days")
- `NO_COLOR` env var and `--no-color` flag respected

### JSON
- Minified, pipeable to `jq`
- Full event detail (no truncation)
- Dates in ISO 8601 / RFC 3339 format
- Null for unset optional fields

### Plain
- One event per line: `[HH:MM-HH:MM] Title (Calendar) @ Location`
- All-day: `[All Day] Title (Calendar)`
- Useful for scripting, piping to `grep`, notifications

---

## ID Handling

- **Display**: First 8 chars of `eventIdentifier` in tables
- **Input**: Prefix matching — any unique prefix resolves to the full event
- **Ambiguity**: If prefix matches multiple events, error with "did you mean?" listing matches
- **JSON output**: Always full ID

---

## Recurrence Display

Human-readable recurrence descriptions in show/detail views:

| Rule | Display |
|------|---------|
| Daily, interval 1 | "Every day" |
| Daily, interval 3 | "Every 3 days" |
| Weekly, Mon/Wed/Fri | "Every week on Mon, Wed, Fri" |
| Weekly, interval 2, Tue | "Every 2 weeks on Tue" |
| Monthly, day 15 | "Every month on the 15th" |
| Monthly, last Friday | "Every month on the last Fri" |
| Yearly, interval 1 | "Every year" |
| + until date | "... until Mar 15, 2027" |
| + count | "... for 10 occurrences" |

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Not macOS | Exit with "cal requires macOS" |
| TCC denied | Exit with "Calendar access denied. Grant access in System Settings > Privacy & Security > Calendars" |
| Event not found | "No event found with ID 'abc123'" |
| Ambiguous ID prefix | "Multiple events match 'a1b2'. Did you mean: ..." (list matches) |
| Read-only calendar | "Cannot modify events in read-only calendar 'Holidays'" |
| Invalid date | "Could not parse date: 'not-a-date'. Try formats like 'tomorrow 2pm' or '2026-03-15 14:00'" |
| Import failure (per-event) | Warning, continue importing remaining events |

---

## Implementation Phases

### Phase 1: Core CLI (MVP)
Priority: Ship a usable CLI fast.

1. **Project scaffolding**: `go.mod`, `main.go`, `root.go`, Makefile
2. **`cal calendars`**: List calendars (table/json/plain)
3. **`cal list`** + **`cal today`** + **`cal upcoming`**: Event listing with filters
4. **`cal show`**: Event detail view
5. **`cal add`**: Create events (flag-based first, then interactive)
6. **`cal update`**: Update events
7. **`cal delete`**: Delete events with confirmation
8. **`cal search`**: Search events
9. **Output formatting** (`internal/ui/`): table, JSON, plain
10. **Date parsing** (`internal/parser/`): NL date parser
11. **`cal version`** + **`cal completion`**

### Phase 2: Import/Export
1. **`cal export`**: JSON + CSV export
2. **`cal import`**: JSON + CSV import with dry-run
3. **ICS export**: RFC 5545 iCalendar format

### Phase 3: Polish
1. **Interactive mode** (`-i`) for add/update
2. **Recurrence flags** for add/update
3. **Shell completions** (dynamic calendar name completion)
4. **README.md** with full documentation
5. **Makefile** with cross-compile targets

---

## go-eventkit Calendar API Reference

Quick reference for the underlying library methods used by each command.

### Client Initialization
```go
client, err := calendar.New()  // Requests TCC access
```

### Commands → API Mapping

| Command | API Call | Notes |
|---------|----------|-------|
| `cal calendars` | `client.Calendars()` | Returns `[]Calendar` |
| `cal list` | `client.Events(start, end, opts...)` | Requires date range |
| `cal today` | `client.Events(todayStart, tomorrowStart)` | |
| `cal upcoming` | `client.Events(now, now.Add(N*24h))` | |
| `cal show [id]` | `client.Event(id)` | Returns `*Event` |
| `cal add` | `client.CreateEvent(input)` | Returns `*Event` with ID |
| `cal update [id]` | `client.UpdateEvent(id, input, span)` | Span: this/future |
| `cal delete [id]` | `client.DeleteEvent(id, span)` | Span: this/future |
| `cal search [q]` | `client.Events(start, end, WithSearch(q))` | |
| `cal export` | `client.Events(start, end, opts...)` | + marshal |
| `cal import` | `client.CreateEvent(input)` per event | Loop with error collection |

### Functional Options for `Events()`
```go
calendar.WithCalendar("Work")      // Filter by calendar name
calendar.WithCalendarID("abc-123") // Filter by calendar ID
calendar.WithSearch("standup")     // Search title, location, notes
```

### Input Types
```go
// CreateEventInput — all fields for new events
calendar.CreateEventInput{
    Title:              "Meeting",           // required
    StartDate:          start,               // required
    EndDate:            end,                 // required
    AllDay:             false,
    Location:           "Room 3",
    Notes:              "Bring laptop",
    URL:                "https://zoom.us/j/123",
    Calendar:           "Work",              // name, empty = default
    Alerts:             []calendar.Alert{{RelativeOffset: -15 * time.Minute}},
    TimeZone:           "America/New_York",
    RecurrenceRules:    []eventkit.RecurrenceRule{eventkit.Weekly(1, eventkit.Monday)},
    StructuredLocation: &eventkit.StructuredLocation{Title: "Apple Park", Latitude: 37.3349, Longitude: -122.0090},
}

// UpdateEventInput — pointer fields, nil = don't change
calendar.UpdateEventInput{
    Title:    strPtr("New Title"),
    Location: strPtr(""),          // empty string clears
}

// Span for recurring events
calendar.SpanThisEvent     // only this occurrence
calendar.SpanFutureEvents  // this + all future
```

### Key Types
```go
// Event fields available for display
event.ID                    // stable identifier
event.Title, event.Notes, event.URL, event.Location
event.StartDate, event.EndDate, event.AllDay
event.Calendar, event.CalendarID
event.Status                // None/Confirmed/Tentative/Canceled
event.Availability          // Busy/Free/Tentative/Unavailable
event.Organizer             // read-only string
event.Attendees             // []Attendee (read-only)
event.Recurring             // bool
event.RecurrenceRules       // []eventkit.RecurrenceRule
event.IsDetached            // modified occurrence
event.StructuredLocation    // *eventkit.StructuredLocation
event.Alerts                // []Alert
event.TimeZone              // IANA timezone
event.CreatedAt, event.ModifiedAt

// Calendar fields
cal.ID, cal.Title, cal.Type, cal.Color, cal.Source, cal.ReadOnly

// RecurrenceRule (from eventkit package)
eventkit.Daily(1)                                    // Every day
eventkit.Weekly(2, eventkit.Monday, eventkit.Friday)  // Every 2 weeks on Mon/Fri
eventkit.Monthly(1, 15)                              // Monthly on 15th
eventkit.Yearly(1)                                   // Yearly
rule.Until(time.Date(2027, 3, 15, 0, 0, 0, 0, time.Local))  // End date
rule.Count(10)                                       // End after N occurrences
```

---

## Comparison with rem

| Aspect | cal | rem |
|--------|-----|-----|
| **Data source** | go-eventkit (EventKit for all ops) | Own bridge (EventKit reads, AppleScript writes) |
| **Write speed** | <10ms (EventKit native) | 200-500ms (AppleScript) |
| **Recurrence** | Full CRUD (read + write) | Not supported |
| **Locations** | Structured (lat/long) + string | Not applicable |
| **Attendees** | Read-only display | Not applicable |
| **Export formats** | JSON, CSV, ICS | JSON, CSV |
| **Date ranges** | Required for queries | Optional (fetch all) |
| **Interactive** | `-i` flag on add/update | `-i` on add only |
| **ID matching** | Prefix matching | Prefix matching |
| **Output formats** | table/json/plain | table/json/plain |
