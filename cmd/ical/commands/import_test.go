package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportNoAlertDryRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "event.ics")
	if err := os.WriteFile(path, []byte("BEGIN:VCALENDAR\nBEGIN:VEVENT\nSUMMARY:Import test\nDTSTART:20260819T153500Z\nDTEND:20260819T164000Z\nEND:VEVENT\nEND:VCALENDAR\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { importDryRun = false; importNoAlert = false; rootCmd.SetArgs(nil) })
	rootCmd.SetArgs([]string{"import", path, "--dry-run", "--no-alert"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
