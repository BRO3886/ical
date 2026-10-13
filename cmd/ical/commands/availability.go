package commands

import (
	"fmt"
	"strings"

	"github.com/BRO3886/go-eventkit/calendar"
)

// parseAvailability maps an --availability value to the event's free/busy
// ("Show As") setting.
func parseAvailability(s string) (calendar.Availability, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "busy":
		return calendar.AvailabilityBusy, nil
	case "free":
		return calendar.AvailabilityFree, nil
	case "tentative":
		return calendar.AvailabilityTentative, nil
	case "unavailable":
		return calendar.AvailabilityUnavailable, nil
	default:
		return 0, fmt.Errorf("invalid --availability %q (want busy, free, tentative or unavailable)", s)
	}
}
