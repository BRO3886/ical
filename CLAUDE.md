# cal — CLI for macOS Calendar

## Non-Negotiables
- **Conventional Commits**: ALL commits MUST follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`. No exceptions.

## What is this?
Go CLI wrapping macOS Calendar via `go-eventkit`. Native EventKit bindings for 3000x faster reads than AppleScript. Single binary. Provides CRUD for events/calendars, natural language dates, recurrence rules, import/export, and multiple output formats.

**Repository**: `github.com/BRO3886/cal`

## Architecture
```
cal/
├── cmd/cal/
│   ├── main.go                  # Entry point (macOS check, version)
│   └── commands/                # Cobra CLI commands (one file per command)
│       ├── root.go              # Root cmd + global flags (--output, --no-color)
│       ├── calendars.go         # List calendars
│       ├── list.go              # List events (date range, filters)
│       ├── show.go              # Show single event detail
│       ├── add.go               # Create event (flags + interactive)
│       ├── update.go            # Update event
│       ├── delete.go            # Delete event (with confirmation)
│       ├── today.go             # Shortcut: today's events
│       ├── upcoming.go          # Shortcut: next N days
│       ├── search.go            # Search events
│       ├── export.go            # Export events (JSON/CSV/ICS)
│       └── import.go            # Import events (JSON/CSV)
├── internal/
│   ├── ui/                      # Output formatting (table/json/plain)
│   │   └── output.go
│   ├── export/                  # JSON/CSV/ICS import/export
│   │   ├── json.go
│   │   ├── csv.go
│   │   └── ics.go
│   └── parser/                  # Natural language date parsing
│       ├── date.go
│       └── date_test.go
├── journals/                    # Engineering journals
├── docs/
│   └── prd/                     # Product requirements
├── Makefile
├── go.mod
└── README.md
```

## Key Dependencies
- `github.com/BRO3886/go-eventkit` — calendar bindings (the whole point)
- `github.com/spf13/cobra` — CLI framework
- `github.com/olekukonko/tablewriter` v1.x — table output (new API: `NewTable()`, `.Header()`, `.Append()`, `.Render()`)
- `github.com/fatih/color` — terminal colors

## Critical: Architecture Rules
- **All reads/writes go through `go-eventkit/calendar`** — no direct EventKit or AppleScript
- **Single binary** — go-eventkit compiles EventKit via cgo into the binary
- **macOS only** — exit gracefully with error on other platforms
- Events require date ranges for queries — no unbounded fetches
- `eventIdentifier` is the stable ID (not `calendarItemIdentifier`)
- Attendees/organizer are read-only (Apple limitation)
- Subscribed/birthday calendars are read-only
- `--output json|table|plain` on all list/show commands
- `NO_COLOR` env var respected

## Libraries
- `spf13/cobra` — CLI framework
- `olekukonko/tablewriter` v1.x — **new API**: `NewTable()`, `.Header()`, `.Append()`, `.Render()` (NOT the old `SetHeader`/`SetBorder` API)
- `fatih/color` — terminal colors
- `olekukonko/tablewriter/tw` — alignment constants (`tw.AlignLeft`)

## Build & Test
```bash
go build -o bin/cal ./cmd/cal    # Build (compiles EventKit via cgo)
go test ./...                    # Unit tests
make build                       # Via Makefile
make completions                 # bash/zsh/fish
```

## Conventions
- Short IDs: first 8 chars of full event identifier displayed in tables
- Prefix matching: users can pass partial IDs to show/update/delete
- All list/show commands support `-o json|table|plain`
- Date display: human-readable by default, ISO 8601 in JSON
- Confirmation prompt for delete, `--force` to skip
- Natural language dates: "today", "tomorrow", "next friday", "in 3 hours", etc.
- Recurrence display: human-readable ("Every 2 weeks on Mon, Wed")
- Color coding: calendar colors shown, all-day events highlighted

## Journal
Engineering journals live in `journals/` dir. See `.claude/commands/journal.md` for the journaling command.
