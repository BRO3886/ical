---
title: "Getting Started"
description: "Install ical and start managing your macOS Calendar from the terminal."
weight: 1
---

## Requirements

- **macOS** (any recent version)
- **Go 1.24+** (for `go install`)
- Calendar access permission (macOS will prompt on first run)

ical uses cgo to compile native EventKit bindings directly into the binary. It does not work on Linux or Windows.

## Installation

### Quick install (recommended)

```bash
curl -fsSL https://ical.sidv.dev/install | bash
```

Downloads the latest release binary and installs to `/usr/local/bin`.

### Via go install

```bash
go install github.com/BRO3886/ical/cmd/ical@latest
```

> Requires Go 1.21+ and Xcode Command Line Tools (`xcode-select --install`).

### Manual download

Apple Silicon:
```bash
curl -LO https://github.com/BRO3886/ical/releases/latest/download/ical-darwin-arm64.tar.gz
tar xzf ical-darwin-arm64.tar.gz
sudo mv ical /usr/local/bin/
```

Intel:
```bash
curl -LO https://github.com/BRO3886/ical/releases/latest/download/ical-darwin-amd64.tar.gz
tar xzf ical-darwin-amd64.tar.gz
sudo mv ical /usr/local/bin/
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
ical today
```

This shows all events for today in a table format with row numbers, times, titles, and calendar names.

## Basic Usage

```bash
# Today's agenda
ical today

# Next 7 days
ical upcoming

# List events in a date range
ical list -f "next monday" -t "next friday"

# Search events by title
ical search "standup" -c Work

# Show event details (interactive picker)
ical show

# Create an event
ical add "Team Standup" -s "tomorrow 9am" -e "tomorrow 9:30am" -c Work

# Create interactively with a guided form
ical add -i

# Delete an event (interactive picker with confirmation)
ical delete
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
ical today -o json | jq '.[].title'

# Plain output for grep
ical today -o plain | grep "standup"
```

## Interactive Mode

ical supports two kinds of interactive workflows, both powered by [charmbracelet/huh](https://github.com/charmbracelet/huh) with the Catppuccin theme.

### Guided Forms

The `-i` flag on `add` and `update` launches a step-by-step form where you fill in each field — title, calendar, start/end time, location, alerts, and recurrence — with validation and dropdowns.

```bash
# Create an event interactively
ical add -i

# Update an event interactively (pick event, then edit fields)
ical update -i
```

### Event Picker

Running `show`, `update`, or `delete` with no argument opens a searchable picker that lists your upcoming events. Type to filter, then press Enter to select.

```bash
# Pick an event to view
ical show

# Pick an event to delete (with confirmation)
ical delete

# Pick an event to update
ical update
```

You can also combine the picker with the guided form:

```bash
# Pick an event, then edit it in a form
ical update -i
```

## Color Support

ical respects the `NO_COLOR` environment variable. You can also pass `--no-color` to disable colored output.

## Shell Completions

Generate completions for your shell:

```bash
# Bash
ical completion bash > /usr/local/etc/bash_completion.d/ical

# Zsh
ical completion zsh > "${fpath[1]}/_ical"

# Fish
ical completion fish > ~/.config/fish/completions/ical.fish
```
