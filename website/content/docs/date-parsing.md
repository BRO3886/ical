---
title: "Date Parsing"
description: "Natural language date input reference for all ical commands."
weight: 3
---

## Overview

Every flag that accepts a date (`--from`, `--to`, `--start`, `--end`, `--repeat-until`) supports natural language input. The built-in parser handles 20+ patterns without any external dependencies.

All relative expressions are evaluated at the moment the command runs.

## Supported Patterns

### Relative Days

| Input        | Resolves to              |
|--------------|--------------------------|
| `today`      | Start of today           |
| `tomorrow`   | Start of tomorrow        |
| `yesterday`  | Start of yesterday       |

### Weekdays

| Input           | Resolves to                        |
|-----------------|-------------------------------------|
| `next monday`   | Next occurrence of that weekday    |
| `next friday`   | Next occurrence of that weekday    |
| `friday`        | Next Friday (same as `next friday`) |

### Relative Periods

| Input         | Resolves to                      |
|---------------|----------------------------------|
| `next week`   | Monday of next week              |
| `next month`  | 1st of next month                |
| `this week`   | End of this week (Sunday 11:59 PM) |

### Relative Time

| Input              | Resolves to                |
|--------------------|----------------------------|
| `in 3 hours`       | 3 hours from now           |
| `in 30 minutes`    | 30 minutes from now        |
| `in 5 days`        | 5 days from now            |
| `in 2 weeks`       | 14 days from now           |
| `in 1 month`       | 1 month from now           |

### Past Relative

| Input             | Resolves to               |
|-------------------|---------------------------|
| `2 hours ago`     | 2 hours before now        |
| `5 days ago`      | 5 days before now         |
| `1 month ago`     | 1 month before now        |

### Time of Day

| Input     | Resolves to          |
|-----------|----------------------|
| `3pm`     | Today at 3:00 PM     |
| `15:00`   | Today at 3:00 PM     |
| `3:30pm`  | Today at 3:30 PM     |
| `9am`     | Today at 9:00 AM     |

### Weekday + Time

| Input           | Resolves to                |
|-----------------|----------------------------|
| `friday 2pm`    | Next Friday at 2:00 PM     |
| `monday 9am`    | Next Monday at 9:00 AM     |

### Month + Day

| Input        | Resolves to                  |
|--------------|------------------------------|
| `mar 15`     | March 15 of this year        |
| `march 15`   | March 15 of this year       |
| `dec 25`     | December 25 of this year    |

### ISO 8601

| Input                | Resolves to             |
|----------------------|-------------------------|
| `2026-03-15`         | March 15, 2026          |
| `2026-03-15 14:00`   | March 15, 2026 at 2 PM |

### Shorthand Keywords

| Input  | Resolves to              |
|--------|--------------------------|
| `eod`  | Today at 5:00 PM         |
| `eow`  | Friday at 5:00 PM        |

## End-of-Day Behavior

When a date is used with the `--to` flag and resolves to midnight (00:00:00), ical automatically bumps it to 23:59:59 of that day. This ensures that `--to "feb 12"` includes all events on February 12, not just those before midnight.

This applies to `list`, `search`, `export`, and the interactive event picker.

## Usage Examples

```bash
# Events from today through end of next Friday
ical list -f today -t "next friday"

# Events in the past month
ical list -f "1 month ago" -t today

# Create event starting in 2 hours
ical add "Quick call" -s "in 2 hours" -e "in 3 hours"

# Search events around a specific date
ical search "review" -f "mar 1" -t "mar 31"

# Export this week's events
ical export -f today -t "this week" --format json
```
