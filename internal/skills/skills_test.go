package skills

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestDefaultTargets(t *testing.T) {
	targets := DefaultTargets("/home/user")
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	if targets[0].Key != "claude" {
		t.Errorf("expected first target key 'claude', got %q", targets[0].Key)
	}
	if targets[0].BaseDir != "/home/user/.claude/skills" {
		t.Errorf("expected claude base dir '/home/user/.claude/skills', got %q", targets[0].BaseDir)
	}

	if targets[1].Key != "codex" {
		t.Errorf("expected second target key 'codex', got %q", targets[1].Key)
	}
	if targets[1].BaseDir != "/home/user/.agents/skills" {
		t.Errorf("expected codex base dir '/home/user/.agents/skills', got %q", targets[1].BaseDir)
	}
}

func TestSkillDir(t *testing.T) {
	target := AgentTarget{BaseDir: "/home/user/.claude/skills"}
	got := SkillDir(target)
	want := "/home/user/.claude/skills/ical-cli"
	if got != want {
		t.Errorf("SkillDir() = %q, want %q", got, want)
	}
}

func TestDisplayPath(t *testing.T) {
	tests := []struct {
		path    string
		homeDir string
		want    string
	}{
		{"/home/user/.claude/skills/ical-cli", "/home/user", "~/.claude/skills/ical-cli"},
		{"/other/path/skills", "/home/user", "/other/path/skills"},
		{"/home/user", "/home/user", "~"},
	}

	for _, tt := range tests {
		got := DisplayPath(tt.path, tt.homeDir)
		if got != tt.want {
			t.Errorf("DisplayPath(%q, %q) = %q, want %q", tt.path, tt.homeDir, got, tt.want)
		}
	}
}

// newTestFS creates a mock embedded filesystem matching the skills/cal-cli structure.
func newTestFS() *fstest.MapFS {
	return &fstest.MapFS{
		"skills/cal-cli/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: cal-cli\n---\n# Test skill\n"),
		},
		"skills/cal-cli/references/commands.md": &fstest.MapFile{
			Data: []byte("# Commands\n"),
		},
		"skills/cal-cli/references/dates.md": &fstest.MapFile{
			Data: []byte("# Dates\n"),
		},
	}
}

func TestInstall(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	fs := newTestFS()
	written, err := Install(fs, target, "v1.0.0")
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Check files were written
	if len(written) != 3 {
		t.Errorf("expected 3 files written, got %d: %v", len(written), written)
	}

	// Verify SKILL.md content
	data, err := os.ReadFile(filepath.Join(SkillDir(target), "SKILL.md"))
	if err != nil {
		t.Fatalf("failed to read installed SKILL.md: %v", err)
	}
	if string(data) != "---\nname: cal-cli\n---\n# Test skill\n" {
		t.Errorf("SKILL.md content mismatch: %q", string(data))
	}

	// Verify references directory
	data, err = os.ReadFile(filepath.Join(SkillDir(target), "references", "commands.md"))
	if err != nil {
		t.Fatalf("failed to read installed commands.md: %v", err)
	}
	if string(data) != "# Commands\n" {
		t.Errorf("commands.md content mismatch: %q", string(data))
	}

	// Verify version file
	data, err = os.ReadFile(filepath.Join(SkillDir(target), ".ical-version"))
	if err != nil {
		t.Fatalf("failed to read version file: %v", err)
	}
	if string(data) != "v1.0.0\n" {
		t.Errorf("version file content = %q, want %q", string(data), "v1.0.0\n")
	}
}

func TestInstallOverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	// First install
	fs := newTestFS()
	_, err := Install(fs, target, "v1.0.0")
	if err != nil {
		t.Fatalf("first Install() error: %v", err)
	}

	// Second install with different version
	_, err = Install(fs, target, "v2.0.0")
	if err != nil {
		t.Fatalf("second Install() error: %v", err)
	}

	// Version should be updated
	v := InstalledVersion(target)
	if v != "v2.0.0" {
		t.Errorf("InstalledVersion() = %q, want %q", v, "v2.0.0")
	}
}

func TestInstallOverwritesSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	// Create a symlink at the skill dir path
	skillDir := SkillDir(target)
	if err := os.MkdirAll(filepath.Dir(skillDir), 0o755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	linkTarget := filepath.Join(tmpDir, "some-other-dir")
	if err := os.MkdirAll(linkTarget, 0o755); err != nil {
		t.Fatalf("failed to create link target: %v", err)
	}
	if err := os.Symlink(linkTarget, skillDir); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Install should replace the symlink
	fs := newTestFS()
	_, err := Install(fs, target, "v1.0.0")
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Should now be a regular directory, not a symlink
	info, err := os.Lstat(skillDir)
	if err != nil {
		t.Fatalf("failed to stat skill dir: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("skill dir is still a symlink after install")
	}

	if !IsInstalled(target) {
		t.Error("IsInstalled() = false after install")
	}
}

func TestUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	// Install first
	fs := newTestFS()
	_, err := Install(fs, target, "v1.0.0")
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	// Uninstall
	removed, err := Uninstall(target)
	if err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}
	if !removed {
		t.Error("Uninstall() returned false, expected true")
	}

	// Verify gone
	if IsInstalled(target) {
		t.Error("IsInstalled() = true after uninstall")
	}
}

func TestUninstallNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	removed, err := Uninstall(target)
	if err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}
	if removed {
		t.Error("Uninstall() returned true for non-existent skill")
	}
}

func TestIsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	// Not installed
	if IsInstalled(target) {
		t.Error("IsInstalled() = true before install")
	}

	// Install
	fs := newTestFS()
	_, _ = Install(fs, target, "v1.0.0")

	// Installed
	if !IsInstalled(target) {
		t.Error("IsInstalled() = false after install")
	}
}

func TestInstalledVersion(t *testing.T) {
	tmpDir := t.TempDir()
	target := AgentTarget{
		Name:    "Test Agent",
		Key:     "test",
		BaseDir: filepath.Join(tmpDir, "skills"),
	}

	// Not installed
	if v := InstalledVersion(target); v != "" {
		t.Errorf("InstalledVersion() = %q, want empty", v)
	}

	// Install
	fs := newTestFS()
	_, _ = Install(fs, target, "v1.2.3")

	if v := InstalledVersion(target); v != "v1.2.3" {
		t.Errorf("InstalledVersion() = %q, want %q", v, "v1.2.3")
	}
}

func TestDetectAgents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create ~/.claude/ directory (simulate Claude Code installed)
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Don't create ~/.agents/ (simulate Codex not installed)

	targets := DefaultTargets(tmpDir)
	detected := DetectAgents(targets)

	if len(detected) != 1 {
		t.Fatalf("expected 1 detected agent, got %d", len(detected))
	}
	if detected[0].Key != "claude" {
		t.Errorf("expected detected agent 'claude', got %q", detected[0].Key)
	}
}

func TestDetectAgentsNonePresent(t *testing.T) {
	tmpDir := t.TempDir()
	targets := DefaultTargets(tmpDir)
	detected := DetectAgents(targets)

	if len(detected) != 0 {
		t.Errorf("expected 0 detected agents, got %d", len(detected))
	}
}

func TestDetectAgentsBothPresent(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".claude", "skills"), 0o755)
	os.MkdirAll(filepath.Join(tmpDir, ".agents", "skills"), 0o755)

	targets := DefaultTargets(tmpDir)
	detected := DetectAgents(targets)

	if len(detected) != 2 {
		t.Errorf("expected 2 detected agents, got %d", len(detected))
	}
}

func TestInstalledTargets(t *testing.T) {
	tmpDir := t.TempDir()
	targets := DefaultTargets(tmpDir)

	// Install only for claude
	fs := newTestFS()
	_, _ = Install(fs, targets[0], "v1.0.0")

	installed := InstalledTargets(targets)
	if len(installed) != 1 {
		t.Fatalf("expected 1 installed target, got %d", len(installed))
	}
	if installed[0].Key != "claude" {
		t.Errorf("expected installed target 'claude', got %q", installed[0].Key)
	}
}
