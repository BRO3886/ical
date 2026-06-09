package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit"
)

// TestBuildRecurrenceRule verifies that buildRecurrenceRule constructs valid rules
// and that Validate() catches invalid configurations.
func TestBuildRecurrenceRule(t *testing.T) {
	tests := []struct {
		name           string
		repeat         string
		interval       int
		days           string
		repeatUntil    string
		repeatCount    int
		wantErr        bool
		wantFrequency  eventkit.RecurrenceFrequency
		wantInterval   int
	}{
		{
			name:          "daily",
			repeat:        "daily",
			interval:      1,
			wantFrequency: eventkit.FrequencyDaily,
			wantInterval:  1,
		},
		{
			name:          "weekly with days",
			repeat:        "weekly",
			interval:      2,
			days:          "mon,wed,fri",
			wantFrequency: eventkit.FrequencyWeekly,
			wantInterval:  2,
		},
		{
			name:          "monthly",
			repeat:        "monthly",
			interval:      1,
			wantFrequency: eventkit.FrequencyMonthly,
			wantInterval:  1,
		},
		{
			name:          "yearly",
			repeat:        "yearly",
			interval:      1,
			wantFrequency: eventkit.FrequencyYearly,
			wantInterval:  1,
		},
		{
			name:        "with count",
			repeat:      "daily",
			interval:    1,
			repeatCount: 10,
			wantFrequency: eventkit.FrequencyDaily,
			wantInterval:  1,
		},
		{
			name:        "with until",
			repeat:      "weekly",
			interval:    1,
			repeatUntil: "2026-12-31",
			wantFrequency: eventkit.FrequencyWeekly,
			wantInterval:  1,
		},
		{
			name:    "invalid repeat type",
			repeat:  "biweekly",
			interval: 1,
			wantErr: true,
		},
		{
			name:    "invalid repeat days",
			repeat:  "weekly",
			interval: 1,
			days:    "foo,bar",
			wantErr: true,
		},
		{
			name:        "invalid repeat until",
			repeat:      "daily",
			interval:    1,
			repeatUntil: "not-a-date",
			wantErr:     true,
		},
		{
			name:     "repeat-days with daily errors",
			repeat:   "daily",
			interval: 1,
			days:     "mon",
			wantErr:  true,
		},
		{
			name:     "repeat-days with monthly errors",
			repeat:   "monthly",
			interval: 1,
			days:     "mon,wed",
			wantErr:  true,
		},
		{
			name:     "repeat-days with yearly errors",
			repeat:   "yearly",
			interval: 1,
			days:     "fri",
			wantErr:  true,
		},
		{
			name:     "zero interval defaults to 1",
			repeat:   "daily",
			interval: 0,
			wantFrequency: eventkit.FrequencyDaily,
			wantInterval:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set package-level vars (used by buildRecurrenceRule)
			addRepeat = tt.repeat
			addRepeatInterval = tt.interval
			addRepeatDays = tt.days
			addRepeatUntil = tt.repeatUntil
			addRepeatCount = tt.repeatCount

			rule, err := buildRecurrenceRule("")
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rule.Frequency != tt.wantFrequency {
				t.Errorf("frequency: got %v, want %v", rule.Frequency, tt.wantFrequency)
			}
			if rule.Interval != tt.wantInterval {
				t.Errorf("interval: got %d, want %d", rule.Interval, tt.wantInterval)
			}
		})
	}
}

// TestRepeatUntilBound verifies that a date-only --repeat-until is snapped to
// end-of-day so an occurrence later that same day is kept (issue #39), while an
// explicit time is preserved.
func TestRepeatUntilBound(t *testing.T) {
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatalf("load Asia/Kolkata: %v", err)
	}

	tests := []struct {
		name string
		in   time.Time
		loc  *time.Location
		want time.Time
	}{
		{
			name: "date-only midnight local bumped to end of day",
			in:   time.Date(2026, 8, 14, 0, 0, 0, 0, time.Local),
			loc:  nil,
			want: time.Date(2026, 8, 14, 23, 59, 59, 0, time.Local),
		},
		{
			name: "explicit time left untouched",
			in:   time.Date(2026, 8, 14, 21, 30, 0, 0, time.Local),
			loc:  nil,
			want: time.Date(2026, 8, 14, 21, 30, 0, 0, time.Local),
		},
		{
			name: "date-only reinterpreted in event timezone, end of day",
			in:   time.Date(2026, 8, 14, 0, 0, 0, 0, time.Local),
			loc:  ist,
			want: time.Date(2026, 8, 14, 23, 59, 59, 0, ist),
		},
		{
			name: "explicit time reinterpreted in event timezone, not bumped",
			in:   time.Date(2026, 8, 14, 21, 30, 0, 0, time.Local),
			loc:  ist,
			want: time.Date(2026, 8, 14, 21, 30, 0, 0, ist),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repeatUntilBound(tt.in, tt.loc)
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBuildRecurrenceRuleUntilTimezone verifies that the timezone passed to
// buildRecurrenceRule reaches the date-only --repeat-until bound, so the
// end-of-day instant is computed in the event's zone. This is the path the
// update command relies on (it tracks the zone in a different global than add).
func TestBuildRecurrenceRuleUntilTimezone(t *testing.T) {
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Fatalf("load Asia/Kolkata: %v", err)
	}

	addRepeat = "weekly"
	addRepeatInterval = 1
	addRepeatDays = ""
	addRepeatUntil = "2026-08-14"
	addRepeatCount = 0

	rule, err := buildRecurrenceRule("Asia/Kolkata")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.End == nil || rule.End.EndDate == nil {
		t.Fatal("expected an UNTIL bound, got none")
	}

	want := time.Date(2026, 8, 14, 23, 59, 59, 0, ist)
	if !rule.End.EndDate.Equal(want) {
		t.Errorf("UNTIL bound: got %v, want %v", rule.End.EndDate, want)
	}
}

// TestRecurrenceRuleValidate verifies that go-eventkit's Validate() catches
// invalid recurrence configurations before they hit EventKit.
func TestRecurrenceRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    eventkit.RecurrenceRule
		wantErr bool
	}{
		{
			"valid daily",
			eventkit.Daily(1),
			false,
		},
		{
			"valid weekly with days",
			eventkit.Weekly(1, eventkit.Monday, eventkit.Wednesday),
			false,
		},
		{
			"valid yearly",
			eventkit.Yearly(1),
			false,
		},
		{
			"valid daily with count",
			eventkit.Daily(1).Count(10),
			false,
		},
		{
			"zero interval",
			eventkit.RecurrenceRule{
				Frequency: eventkit.FrequencyDaily,
				Interval:  0,
			},
			true,
		},
		{
			"days of week on daily (invalid)",
			eventkit.RecurrenceRule{
				Frequency: eventkit.FrequencyDaily,
				Interval:  1,
				DaysOfTheWeek: []eventkit.RecurrenceDayOfWeek{
					{DayOfTheWeek: eventkit.Monday},
				},
			},
			true,
		},
		{
			"days of month on weekly (invalid)",
			eventkit.RecurrenceRule{
				Frequency:      eventkit.FrequencyWeekly,
				Interval:       1,
				DaysOfTheMonth: []int{15},
			},
			true,
		},
		{
			"months of year on daily (invalid)",
			eventkit.RecurrenceRule{
				Frequency:       eventkit.FrequencyDaily,
				Interval:        1,
				MonthsOfTheYear: []int{3, 6},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestParseRepeatDays verifies weekday string parsing.
func TestParseRepeatDays(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{"empty", "", 0, false},
		{"single", "mon", 1, false},
		{"multiple", "mon,wed,fri", 3, false},
		{"with spaces", "mon, wed, fri", 3, false},
		{"full names", "monday,wednesday", 2, false},
		{"invalid", "foo", 0, true},
		{"mixed invalid", "mon,foo", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			days, err := parseRepeatDays(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(days) != tt.wantLen {
				t.Errorf("got %d days, want %d", len(days), tt.wantLen)
			}
		})
	}
}
