package commands

import (
	"strings"
	"testing"
	"time"
)

// All-day --end names the last day of the event (inclusive) for both add and
// update, matching `list --to`.
func TestAllDayEnd(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	day := func(d int) time.Time { return time.Date(2026, 12, d, 0, 0, 0, 0, ny) }
	lastSecond := func(d int) time.Time { return time.Date(2026, 12, d, 23, 59, 59, 0, ny) }

	tests := []struct {
		name       string
		start, end time.Time
		want       time.Time
		wantErr    bool
	}{
		{"Fri to Sun covers Sun", day(11), day(13), lastSecond(13), false},
		{"single day", day(11), day(11), lastSecond(11), false},
		{"time on the end date is ignored", day(11), day(13).Add(15 * time.Hour), lastSecond(13), false},
		{"end before start", day(11), day(10), time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := allDayEnd(tt.start, tt.end)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"15m", 15 * time.Minute, false},
		{"70m", 70 * time.Minute, false},
		{"1h", time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"1h10m", 70 * time.Minute, false},
		{"1H30M", 90 * time.Minute, false},
		{"2h 15m", 135 * time.Minute, false},
		{"1d2h", 26 * time.Hour, false},
		{"0m", 0, false},
		{"", 0, true},
		{"soon", 0, true},
		{"-5m", 0, true},
		{"1.5h", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchEmptyNotice(t *testing.T) {
	from := time.Date(2026, 8, 30, 0, 0, 0, 0, time.Local)
	to := time.Date(2026, 10, 29, 23, 59, 59, 0, time.Local)

	got := searchEmptyNotice("Show Weekend", from, to, false)
	for _, want := range []string{`"Show Weekend"`, "2026-08-30", "2026-10-29", "--from/--to"} {
		if !strings.Contains(got, want) {
			t.Errorf("notice %q lacks %q", got, want)
		}
	}
	if got := searchEmptyNotice("x", from, to, true); strings.Contains(got, "--from/--to") {
		t.Errorf("explicit range should not suggest widening: %q", got)
	}
}
