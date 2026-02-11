
## Requirements

- **macOS** (any recent version)
- **Go 1.24+** (for `go install`)
- Calendar access permission (macOS will prompt on first run)

cal uses cgo to compile native EventKit bindings directly into the binary. It does not work on Linux or Windows.

## Installation

### Via go install

```bash
go install github.com/BRO3886/cal/cmd/cal@latest
```

### From source

```bash
git clone https://github.com/BRO3886/cal.git
cd cal
make build
# Binary at ./bin/cal
```

## First Run

On the first invocation, macOS will display a permission dialog asking for Calendar access. Grant it — cal needs this to read and write events.

```bash
cal today
```

This shows all events for today in a table format with row numbers, times, titles, and calendar names.

## Basic Usage

```bash
# Today's agenda
cal today

# Next 7 days
cal upcoming

# List events in a date range
cal list -f "next monday" -t "next friday"

# Search events by title
cal search "standup" -c Work

# Show event details (interactive picker)
cal show

# Create an event
cal add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Create interactively with a guided form
cal add -i

# Delete an event (interactive picker with confirmation)
cal delete
```

## Output Formats

All list and show commands support three output formats via the `--output` (or `-o`) flag:

| Format  | Description                            |
|---------|----------------------------------------|
| `table` | Human-readable table with colors       |
| `json`  | Structured JSON (ISO 8601 dates, full event IDs) |
| `plain` | Simple line-based output for grepping  |

```bash
# JSON output for scripting
cal today -o json | jq '.[].title'

# Plain output for grep
cal today -o plain | grep "standup"
```

## Color Support

cal respects the `NO_COLOR` environment variable. You can also pass `--no-color` to disable colored output.

## Shell Completions

Generate completions for your shell:

```bash
# Bash
cal completion bash > /usr/local/etc/bash_completion.d/cal

# Zsh
cal completion zsh > "${fpath[1]}/_cal"

# Fish
cal completion fish > ~/.config/fish/completions/cal.fish
```

