package ical

import (
	"io/fs"
	"testing"
)

func TestEmbeddedSkillsContainsSkillMD(t *testing.T) {
	data, err := fs.ReadFile(EmbeddedSkills, "skills/cal-cli/SKILL.md")
	if err != nil {
		t.Fatalf("failed to read embedded SKILL.md: %v", err)
	}
	if len(data) == 0 {
		t.Error("embedded SKILL.md is empty")
	}
	// Verify it contains the expected frontmatter
	content := string(data)
	if content[:3] != "---" {
		t.Errorf("SKILL.md does not start with YAML frontmatter, starts with: %q", content[:20])
	}
}

func TestEmbeddedSkillsContainsReferences(t *testing.T) {
	files := []string{
		"skills/cal-cli/references/commands.md",
		"skills/cal-cli/references/dates.md",
	}

	for _, f := range files {
		data, err := fs.ReadFile(EmbeddedSkills, f)
		if err != nil {
			t.Errorf("failed to read embedded %s: %v", f, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("embedded %s is empty", f)
		}
	}
}

func TestEmbeddedSkillsWalkDir(t *testing.T) {
	var files []string
	err := fs.WalkDir(EmbeddedSkills, "skills/cal-cli", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}

	// Should have at least SKILL.md + 2 reference files
	if len(files) < 3 {
		t.Errorf("expected at least 3 embedded files, got %d: %v", len(files), files)
	}
}
