package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/calendar"
)

// rowCacheEnvKeys identify the session a command runs in, most specific
// first. Row numbers cached by one session must not resolve in another:
// with parallel terminals or agents, `ical delete 3` could otherwise act on
// a row from someone else's listing.
var rowCacheEnvKeys = []string{
	"ICAL_SESSION",           // explicit override
	"CLAUDE_CODE_SESSION_ID", // one Claude Code session (may span panes)
	"TMUX_PANE",
	"WEZTERM_PANE",
	"KITTY_WINDOW_ID",
	"ITERM_SESSION_ID",
	"TERM_SESSION_ID", // Terminal.app tab
}

// rowCacheMaxAge is how long an unused session cache file is kept.
const rowCacheMaxAge = 7 * 24 * time.Hour

// rowCacheSessionKey returns "NAME=value" for the first session variable set,
// or "" when none is.
func rowCacheSessionKey(getenv func(string) string) string {
	for _, k := range rowCacheEnvKeys {
		if v := getenv(k); v != "" {
			return k + "=" + v
		}
	}
	return ""
}

// rowCachePath returns the row cache file for a session. Without a session
// key it falls back to the legacy shared ~/.ical-last-list.
func rowCachePath(home, sessionKey string) string {
	if sessionKey == "" {
		return filepath.Join(home, ".ical-last-list")
	}
	sum := sha256.Sum256([]byte(sessionKey))
	return filepath.Join(home, ".cache", "ical", "rows", hex.EncodeToString(sum[:8]))
}

func lastListPath() string {
	home, _ := os.UserHomeDir()
	return rowCachePath(home, rowCacheSessionKey(os.Getenv))
}

// RowRef is one cached row: the event ID and, for a recurring event, the
// original start of the occurrence that row showed. Every occurrence of a
// series shares its ID, so the ID alone always means the first occurrence.
type RowRef struct {
	ID         string
	Occurrence *time.Time
}

// formatRow renders a row as "ID" or "ID<TAB>occurrence (RFC 3339)".
func formatRow(e calendar.Event) string {
	if (e.Recurring || e.IsDetached) && e.OccurrenceDate != nil {
		return e.ID + "\t" + e.OccurrenceDate.UTC().Format(time.RFC3339Nano)
	}
	return e.ID
}

// parseRow reads a line written by formatRow, or a bare ID from older versions.
func parseRow(line string) RowRef {
	id, occ, found := strings.Cut(line, "\t")
	ref := RowRef{ID: id}
	if found {
		if t, err := time.Parse(time.RFC3339Nano, occ); err == nil {
			ref.Occurrence = &t
		}
	}
	return ref
}

// SaveLastList writes event rows to this session's cache so row numbers can
// be used later.
func SaveLastList(events []calendar.Event) {
	ids := make([]string, len(events))
	for i, e := range events {
		ids[i] = formatRow(e)
	}
	path := lastListPath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(strings.Join(ids, "\n")+"\n"), 0o644)
	pruneRowCaches(filepath.Dir(path))
}

// pruneRowCaches removes session cache files that have not been written for
// rowCacheMaxAge, so one file per past session does not pile up.
func pruneRowCaches(dir string) {
	if filepath.Base(dir) != "rows" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-rowCacheMaxAge)
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// LookupRow returns the row for a 1-based row number from this session's
// last listing, and false if there is no such row.
func LookupRow(n int) (RowRef, bool) {
	data, err := os.ReadFile(lastListPath())
	if err != nil {
		return RowRef{}, false
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if n < 1 || n > len(lines) || lines[n-1] == "" {
		return RowRef{}, false
	}
	return parseRow(lines[n-1]), true
}
