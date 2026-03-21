# Natural Language Date Reference

Date flags (`--from`, `--to`, `--start`, `--end`) on ical commands accept natural language dates. Parsing is provided by the shared [`go-eventkit/dateparser`](https://github.com/BRO3886/go-eventkit) package.

## Keywords

| Input                 | Resolves to                   |
| --------------------- | ----------------------------- |
| `today`               | Today at midnight (00:00)     |
| `tomorrow`            | Tomorrow at midnight          |
| `yesterday`           | Yesterday at midnight         |
| `now`                 | Current date and time         |
| `eod` / `end of day`  | Today at 5:00 PM              |
| `eow` / `end of week` | Next Friday at 5:00 PM        |
| `this week`           | Next Sunday at 23:59          |
| `next week`           | Next Monday at midnight       |
| `next month`          | 1st of next month at midnight |

## Relative Time

Pattern: `in <number> <unit>`

| Input                        | Example result     |
| ---------------------------- | ------------------ |
| `in 5 minutes` / `in 5 mins` | 5 minutes from now |
| `in 3 hours` / `in 3 hrs`    | 3 hours from now   |
| `in 2 days`                  | 2 days from now    |
| `in 1 week`                  | 7 days from now    |
| `in 2 months`                | 2 months from now  |

## Past References

Pattern: `<number> <unit> ago`

| Input           | Example result       |
| --------------- | -------------------- |
| `5 minutes ago` | 5 minutes before now |
| `2 hours ago`   | 2 hours before now   |
| `3 days ago`    | 3 days before now    |
| `1 week ago`    | 7 days before now    |
| `6 months ago`  | 6 months before now  |

## Weekday Patterns

| Input                | Resolves to             |
| -------------------- | ----------------------- |
| `next monday`        | Next Monday at midnight |
| `next friday at 2pm` | Next Friday at 2:00 PM  |
| `next fri 3:30pm`    | Next Friday at 3:30 PM  |
| `monday`             | Next Monday at midnight |
| `friday 2pm`         | Next Friday at 2:00 PM  |

Supports full names and 3-letter abbreviations: `sun`, `mon`, `tue`, `wed`, `thu`, `fri`, `sat`.

## Keyword + Time

| Input             | Resolves to         |
| ----------------- | ------------------- |
| `today at 5pm`    | Today at 5:00 PM    |
| `today 3:30pm`    | Today at 3:30 PM    |
| `tomorrow at 9am` | Tomorrow at 9:00 AM |
| `tomorrow 14:00`  | Tomorrow at 2:00 PM |

## Month + Day

Both month-first and day-first ordering are supported:

| Input              | Resolves to             |
| ------------------ | ----------------------- |
| `mar 15`           | March 15 at midnight    |
| `march 15 2pm`     | March 15 at 2:00 PM    |
| `dec 31 11:59pm`   | December 31 at 11:59 PM |
| `21 mar`           | March 21 at midnight    |
| `21 mar 2pm`       | March 21 at 2:00 PM    |
| `21 march at 2pm`  | March 21 at 2:00 PM    |
| `21 march 2026`    | March 21, 2026          |

Supports full and abbreviated month names (jan/january through dec/december). Day-first formats also accept an optional year.

## Time Only

| Input    | Resolves to      |
| -------- | ---------------- |
| `5pm`    | Today at 5:00 PM |
| `9am`    | Today at 9:00 AM |
| `3:30pm` | Today at 3:30 PM |
| `17:00`  | Today at 5:00 PM |

## Standard Date Formats

| Format            | Example                                       |
| ----------------- | --------------------------------------------- |
| RFC 3339          | `2026-02-15T14:30:00Z`                        |
| ISO date          | `2026-02-15`                                  |
| ISO datetime      | `2026-02-15 14:30` / `2026-02-15T14:30`       |
| ISO with seconds  | `2026-02-15 14:30:00` / `2026-02-15T14:30:00` |
| 12-hour           | `2026-02-15 2:30PM`                           |
| US format         | `02/15/2026`                                  |
| US with time      | `02/15/2026 14:30` / `02/15/2026 2:30PM`      |
| Month name        | `Feb 15, 2026` / `February 15, 2026`          |
| Month name + time | `Feb 15, 2026 2:30PM` / `Feb 15, 2026 14:30`  |
| Day-first         | `15 Feb 2026` / `15 February 2026`            |

## Timezone Handling

**Timezone abbreviations (CDT, CST, EST, etc.) cannot be embedded in date strings** — the parser will reject them. Use the `--timezone` flag on `add` or `update` with an IANA timezone name:

```bash
# WRONG — will fail
ical add "Meeting" --start "2026-06-17 at 2pm CDT"

# CORRECT — use IANA timezone
ical add "Meeting" --start "2026-06-17 14:00" --timezone "America/Chicago"
```

Accepted values: any IANA name (e.g., `America/Chicago`, `America/New_York`, `Europe/Madrid`). Times are stored correctly and displayed in the machine's local timezone.

## Alert Durations

The `--alert` flag on `add` and `update` accepts duration strings:

| Input | Meaning           |
| ----- | ----------------- |
| `15m` | 15 minutes before |
| `1h`  | 1 hour before     |
| `2d`  | 2 days before     |
