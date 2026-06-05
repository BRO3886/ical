package ui

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/calendar"
)

func TestShortID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdefgh-1234-5678", "abcdefgh-1234"},
		{"short", "short"},
		{"1234567890123", "1234567890123"},
		{"", ""},
		{"1234567", "1234567"},
		{"12345678901234", "1234567890123"},
		{"577B8983-DF44-4665-966E-58129A363B3A:20250212", "577B8983-DF44"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ShortID(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatRecurrenceRule(t *testing.T) {
	until := time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		rule eventkit.RecurrenceRule
		want string
	}{
		{
			"daily",
			eventkit.Daily(1),
			"Every day",
		},
		{
			"every 3 days",
			eventkit.Daily(3),
			"Every 3 days",
		},
		{
			"weekly",
			eventkit.Weekly(1),
			"Every week",
		},
		{
			"every 2 weeks on Mon, Wed",
			eventkit.Weekly(2, eventkit.Monday, eventkit.Wednesday),
			"Every 2 weeks on mon, wed",
		},
		{
			"monthly",
			eventkit.Monthly(1),
			"Every month",
		},
		{
			"monthly on 15th",
			eventkit.Monthly(1, 15),
			"Every month on the 15th",
		},
		{
			"every 2 months",
			eventkit.Monthly(2),
			"Every 2 months",
		},
		{
			"yearly",
			eventkit.Yearly(1),
			"Every year",
		},
		{
			"every 2 years",
			eventkit.Yearly(2),
			"Every 2 years",
		},
		{
			"daily until date",
			eventkit.Daily(1).Until(until),
			"Every day until Mar 15, 2027",
		},
		{
			"weekly for 10 occurrences",
			eventkit.Weekly(1).Count(10),
			"Every week for 10 occurrences",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRecurrenceRule(tt.rule)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func TestLocalizeTime(t *testing.T) {
	// UTC 18:00 on 2026-02-11
	utcTime := time.Date(2026, 2, 11, 18, 0, 0, 0, time.UTC)

	// localizeTime always returns time.Local regardless of the tz argument.
	t.Run("always returns local regardless of tz", func(t *testing.T) {
		for _, tz := range []string{"Asia/Kolkata", "America/New_York", "UTC", "", "Invalid/Timezone"} {
			got := localizeTime(utcTime, tz)
			if got.Location() != time.Local {
				t.Errorf("tz=%q: expected local timezone, got %v", tz, got.Location())
			}
			want := utcTime.In(time.Local)
			if got.Hour() != want.Hour() || got.Minute() != want.Minute() {
				t.Errorf("tz=%q: expected %02d:%02d local, got %02d:%02d", tz, want.Hour(), want.Minute(), got.Hour(), got.Minute())
			}
		}
	})
}

func TestLocalizeTimeInZone(t *testing.T) {
	// All tests use fixed reference locations — no time.Local dependency.
	// UTC 18:00 on 2026-02-11 (chosen so wall-clock math is clean across all zones).
	utcTime := time.Date(2026, 2, 11, 18, 0, 0, 0, time.UTC)

	ist := mustLoadLocation("Asia/Kolkata")   // UTC+5:30
	est := mustLoadLocation("America/New_York") // UTC-5 in February
	cst := mustLoadLocation("America/Chicago")  // UTC-6 in February
	gmt := mustLoadLocation("GMT")             // UTC+0

	t.Run("returns nil for empty tz", func(t *testing.T) {
		// Reference location doesn't matter when tz is empty.
		if got := localizeTimeInZone(utcTime, "", ist); got != nil {
			t.Errorf("expected nil for empty tz, got %v", got)
		}
	})

	t.Run("returns nil for invalid tz", func(t *testing.T) {
		if got := localizeTimeInZone(utcTime, "Invalid/Timezone", ist); got != nil {
			t.Errorf("expected nil for invalid tz, got %v", got)
		}
	})

	t.Run("returns nil when event tz matches reference offset (same zone)", func(t *testing.T) {
		// Reference=IST, event tz=Asia/Kolkata → same offset → nil.
		if got := localizeTimeInZone(utcTime, "Asia/Kolkata", ist); got != nil {
			t.Errorf("expected nil (same offset), got %v", got)
		}
		// Reference=EST, event tz=America/New_York → same offset → nil.
		if got := localizeTimeInZone(utcTime, "America/New_York", est); got != nil {
			t.Errorf("expected nil (same offset), got %v", got)
		}
	})

	t.Run("IST reference, event in EST: shows EST wall clock", func(t *testing.T) {
		// UTC 18:00 in EST (UTC-5) = 13:00 EST.
		got := localizeTimeInZone(utcTime, "America/New_York", ist)
		if got == nil {
			t.Fatal("expected non-nil result for EST event vs IST reference")
		}
		if got.Hour() != 13 || got.Minute() != 0 {
			t.Errorf("expected 13:00 EST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("IST reference, event in CST: shows CST wall clock", func(t *testing.T) {
		// UTC 18:00 in CST (UTC-6) = 12:00 CST.
		got := localizeTimeInZone(utcTime, "America/Chicago", ist)
		if got == nil {
			t.Fatal("expected non-nil result for CST event vs IST reference")
		}
		if got.Hour() != 12 || got.Minute() != 0 {
			t.Errorf("expected 12:00 CST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("EST reference, event in IST: shows IST wall clock", func(t *testing.T) {
		// UTC 18:00 in IST (UTC+5:30) = 23:30 IST.
		got := localizeTimeInZone(utcTime, "Asia/Kolkata", est)
		if got == nil {
			t.Fatal("expected non-nil result for IST event vs EST reference")
		}
		if got.Hour() != 23 || got.Minute() != 30 {
			t.Errorf("expected 23:30 IST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("EST reference, event in CST: shows CST wall clock", func(t *testing.T) {
		// UTC 18:00 in CST (UTC-6) = 12:00; EST is UTC-5 = 13:00. Different offsets → non-nil.
		got := localizeTimeInZone(utcTime, "America/Chicago", est)
		if got == nil {
			t.Fatal("expected non-nil result for CST event vs EST reference")
		}
		if got.Hour() != 12 || got.Minute() != 0 {
			t.Errorf("expected 12:00 CST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("GMT reference, event in IST: shows IST wall clock", func(t *testing.T) {
		// UTC 18:00 in IST (UTC+5:30) = 23:30 IST.
		got := localizeTimeInZone(utcTime, "Asia/Kolkata", gmt)
		if got == nil {
			t.Fatal("expected non-nil result for IST event vs GMT reference")
		}
		if got.Hour() != 23 || got.Minute() != 30 {
			t.Errorf("expected 23:30 IST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})

	t.Run("GMT reference, event in GMT: returns nil (same offset)", func(t *testing.T) {
		if got := localizeTimeInZone(utcTime, "GMT", gmt); got != nil {
			t.Errorf("expected nil (same GMT offset), got %v", got)
		}
	})

	t.Run("CST reference, event in EST: shows EST wall clock", func(t *testing.T) {
		// UTC 18:00 in EST (UTC-5) = 13:00. CST is UTC-6 → different → non-nil.
		got := localizeTimeInZone(utcTime, "America/New_York", cst)
		if got == nil {
			t.Fatal("expected non-nil result for EST event vs CST reference")
		}
		if got.Hour() != 13 || got.Minute() != 0 {
			t.Errorf("expected 13:00 EST, got %02d:%02d", got.Hour(), got.Minute())
		}
	})
}

func TestEventDateLabel(t *testing.T) {
	// UTC 15:00 on Sunday 19 Apr 2026. In BST (UTC+1) this is 16:00 Sun;
	// in UTC-negative zones it still falls on Apr 19. The label must reflect
	// whatever location the input carries — the caller localizes upstream.
	utcTime := time.Date(2026, 4, 19, 15, 0, 0, 0, time.UTC)

	bst := mustLoadLocation("Europe/London") // BST in April = UTC+1
	est := mustLoadLocation("America/New_York") // EDT in April = UTC-4
	ist := mustLoadLocation("Asia/Kolkata")   // IST = UTC+5:30

	tests := []struct {
		name string
		in   time.Time
		want string
	}{
		{"UTC 15:00 viewed in BST stays on Sunday", utcTime.In(bst), "Sun 19 Apr"},
		{"UTC 15:00 viewed in EDT stays on Sunday", utcTime.In(est), "Sun 19 Apr"},
		{"UTC 19:30 Sun viewed in IST rolls into Monday",
			time.Date(2026, 4, 19, 19, 30, 0, 0, time.UTC).In(ist),
			"Mon 20 Apr"},
		{"UTC 23:30 Sat viewed in BST rolls into Sunday",
			time.Date(2026, 4, 18, 23, 30, 0, 0, time.UTC).In(bst),
			"Sun 19 Apr"},
		{"UTC 03:30 Sun viewed in EDT stays on Saturday",
			time.Date(2026, 4, 19, 3, 30, 0, 0, time.UTC).In(est),
			"Sat 18 Apr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventDateLabel(tt.in, false)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("showYear includes the year", func(t *testing.T) {
		got := eventDateLabel(utcTime.In(bst), true)
		want := "Sun 19 Apr 2026"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestEventsSpanMultipleYears(t *testing.T) {
	makeEvent := func(year int, month time.Month, day int) calendar.Event {
		return calendar.Event{StartDate: time.Date(year, month, day, 10, 0, 0, 0, time.UTC)}
	}

	tests := []struct {
		name   string
		events []calendar.Event
		want   bool
	}{
		{"empty", nil, false},
		{"single event", []calendar.Event{makeEvent(2026, 5, 1)}, false},
		{"same year", []calendar.Event{makeEvent(2026, 1, 1), makeEvent(2026, 12, 31)}, false},
		{"different years", []calendar.Event{makeEvent(2025, 12, 31), makeEvent(2026, 1, 1)}, true},
		{"three years", []calendar.Event{makeEvent(2024, 6, 1), makeEvent(2025, 6, 1), makeEvent(2026, 6, 1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventsSpanMultipleYears(tt.events)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly10c", 10, "exactly10c"},
		{"this is a long string", 10, "this is..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		// Emoji: display width 2, fits within max — must not be corrupted.
		{"Team lunch 🍕", 40, "Team lunch 🍕"},
		// 38 A's + emoji has display width 40 but byte length 42.
		// Old byte-based code would truncate at byte 37, splitting the emoji
		// and producing invalid UTF-8. Display-width-based code keeps it intact.
		{"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA🍕", 40, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA🍕"},
		// Emoji inside a long title that does need truncation.
		{"🍕 " + "A very long title that exceeds forty characters", 40, "🍕 A very long title that exceeds for..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}
