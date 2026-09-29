package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/dateparser"
)

// allDayEnd turns an all-day --end date into the stored end: the last second
// of that day. The date is inclusive — "--start Fri --end Sun" covers Sunday —
// for both add and update. Any time of day on end is ignored.
func allDayEnd(start, end time.Time) (time.Time, error) {
	last := time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
	firstDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	if last.Before(firstDay) {
		return time.Time{}, fmt.Errorf("--end %s is before --start %s", end.Format("2006-01-02"), start.Format("2006-01-02"))
	}
	return last, nil
}

var compoundDurationPart = regexp.MustCompile(`(\d+)([dhm])`)

// parseDuration parses an alert or travel duration. It accepts everything
// dateparser.ParseAlertDuration does ("15m", "1h", "2 days") plus compound
// forms such as "1h10m", "2h 15m" and "1d2h".
func parseDuration(s string) (time.Duration, error) {
	if d, err := dateparser.ParseAlertDuration(s); err == nil {
		return d, nil
	}
	compact := strings.ToLower(strings.Join(strings.Fields(s), ""))
	parts := compoundDurationPart.FindAllStringSubmatch(compact, -1)
	matched := 0
	var total time.Duration
	for _, p := range parts {
		matched += len(p[0])
		n, _ := strconv.Atoi(p[1])
		switch p[2] {
		case "d":
			total += time.Duration(n) * 24 * time.Hour
		case "h":
			total += time.Duration(n) * time.Hour
		case "m":
			total += time.Duration(n) * time.Minute
		}
	}
	if len(parts) == 0 || matched != len(compact) {
		return 0, fmt.Errorf("invalid duration %q (use e.g. 15m, 1h, 1h10m, 1d)", s)
	}
	return total, nil
}

// searchEmptyNotice explains an empty search result, including the range
// searched, so a default ±30-day window is never mistaken for "no such event".
func searchEmptyNotice(query string, from, to time.Time, explicitRange bool) string {
	msg := fmt.Sprintf("No events matching %q between %s and %s.", query, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if !explicitRange {
		msg += " Search defaults to 30 days either side of today; pass --from/--to to widen it."
	}
	return msg
}
