package commands

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/BRO3886/ical/internal/update"
)

func TestPrintUpdateNoticeSkipsSkillsStalenessForJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	skillDir := filepath.Join(tmpDir, ".claude", "skills", "ical-cli")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("test skill\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, ".ical-version"), []byte("v0.10.1\n"), 0o644); err != nil {
		t.Fatalf("write version: %v", err)
	}

	oldOutputFormat := outputFormat
	oldVersionStr := versionStr
	oldUpdateResultCh := updateResultCh
	oldStderr := os.Stderr
	defer func() {
		outputFormat = oldOutputFormat
		versionStr = oldVersionStr
		updateResultCh = oldUpdateResultCh
		os.Stderr = oldStderr
	}()

	outputFormat = "json"
	versionStr = "v0.10.2"
	updateResultCh = make(chan *update.Result, 1)
	updateResultCh <- nil

	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stderr = writeEnd

	printUpdateNotice(rootCmd)

	if err := writeEnd.Close(); err != nil {
		t.Fatalf("close stderr pipe: %v", err)
	}
	got, err := io.ReadAll(readEnd)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if string(got) != "" {
		t.Fatalf("stderr = %q, want empty for json output", got)
	}
}
