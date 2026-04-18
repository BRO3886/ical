package commands

import "testing"

func TestNormalizeCalendarName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Work", "work"},
		{"already lowercase", "work", "work"},
		{"leading space", "  Work", "work"},
		{"trailing space", "Work   ", "work"},
		{"both sides padded", "  Work  ", "work"},
		{"tab padded", "\tWork\t", "work"},
		{"internal spaces preserved", "Holidays in India", "holidays in india"},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCalendarName(tt.in)
			if got != tt.want {
				t.Errorf("normalizeCalendarName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
