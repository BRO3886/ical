---
title: "Getting Started"
description: "Install ical and start managing your macOS Calendar from the terminal."
weight: 1
---

## Requirements

- **macOS** (any recent version)
- **Go 1.24+** (for `go install`)
- Calendar access permission (macOS will prompt on first run)

iical uses cgo to compile native EventKit bindings directly into the binary. It does not work on Linux or Windows.

## Installation

### Via go install

```bash
go install github.com/BRO3886/ical/cmd/ical@latest
```

### From source

```bash
git clone https://github.com/BRO3886/ical.git
cd ical
make build
# Binary at ./bin/ical
```

## First Run

On the first invocation, macOS will display a permission dialog asking for Calendar access. Grant it — ical needs this to read and write events.

```bash
iical today
```

This shows all events for today in a table format with row numbers, times, titles, and calendar names.

## Basic Usage

```bash
# Today's agenda
iical today

# Next 7 days
iical upcoming

# List events in a date range
iical list -f "next monday" -t "next friday"

# Search events by title
iical search "standup" -c Work

# Show event details (interactive picker)
iical show

# Create an event
iical add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Create interactively with a guided form
iical add -i

# Delete an event (interactive picker with confirmation)
iical delete
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
iical today -o json | jq '.[].title'

# Plain output for grep
iical today -o plain | grep "standup"
```

## Interactive Mode

iical supports two kinds of interactive workflows, both powered by [charmbracelet/huh](https://github.com/charmbracelet/huh) with the Catppuccin theme.

### Guided Forms

The `-i` flag on `add` and `update` launches a step-by-step form where you fill in each field — title, calendar, start/end time, location, alerts, and recurrence — with validation and dropdowns.

```bash
# Create an event interactively
iical add -i

# Update an event interactively (pick event, then edit fields)
iical update -i
```

### Event Picker

Running `show`, `update`, or `delete` with no argument opens a searchable picker that lists your upcoming events. Type to filter, then press Enter to select.

```bash
# Pick an event to view
iical show

# Pick an event to delete (with confirmation)
iical delete

# Pick an event to update
iical update
```

You can also combine the picker with the guided form:

```bash
# Pick an event, then edit it in a form
iical update -i
```

## Color Support

iical respects the `NO_COLOR` environment variable. You can also pass `--no-color` to disable colored output.

## Shell Completions

Generate completions for your shell:

```bash
# Bash
iical completion bash > /usr/local/etc/bash_completion.d/ical

# Zsh
iical completion zsh > "${fpath[1]}/_ical"

# Fish
iical completion fish > ~/.config/fish/completions/ical.fish
```
