package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BRO3886/ical/internal/update"
	"github.com/fatih/color"
)

// interactiveConditions is the baseline: a released binary, table output, and a
// human at a terminal. Each case below turns exactly one condition off.
func interactiveConditions() noticeConditions {
	return noticeConditions{
		commandPath:  "ical list",
		outputFormat: "table",
		version:      "v0.12.1",
		suppressed:   false,
		interactive:  true,
	}
}

func TestShouldShowNotices(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*noticeConditions)
		want   bool
	}{
		{
			name:   "interactive table output shows notices",
			mutate: func(c *noticeConditions) {},
			want:   true,
		},
		{
			name:   "plain output still shows notices",
			mutate: func(c *noticeConditions) { c.outputFormat = "plain" },
			want:   true,
		},
		{
			name:   "json output is a scripting context",
			mutate: func(c *noticeConditions) { c.outputFormat = "json" },
			want:   false,
		},
		{
			name:   "redirected stderr has no reader",
			mutate: func(c *noticeConditions) { c.interactive = false },
			want:   false,
		},
		{
			name:   "ICAL_NO_UPDATE_CHECK suppresses every notice",
			mutate: func(c *noticeConditions) { c.suppressed = true },
			want:   false,
		},
		{
			name:   "dev builds have no release to compare against",
			mutate: func(c *noticeConditions) { c.version = "dev" },
			want:   false,
		},
		{
			name:   "unset version behaves like a dev build",
			mutate: func(c *noticeConditions) { c.version = "" },
			want:   false,
		},
		{
			name:   "version command reports versions itself",
			mutate: func(c *noticeConditions) { c.commandPath = "ical version" },
			want:   false,
		},
		{
			name:   "completion output is machine consumed",
			mutate: func(c *noticeConditions) { c.commandPath = "ical completion zsh" },
			want:   false,
		},
		{
			name:   "skills parent command",
			mutate: func(c *noticeConditions) { c.commandPath = "ical skills" },
			want:   false,
		},
		{
			// Regression: matching on the leaf name alone reports "status" here,
			// letting the notice through on the very command that explains it.
			name:   "nested skills status is still a meta command",
			mutate: func(c *noticeConditions) { c.commandPath = "ical skills status" },
			want:   false,
		},
		{
			name:   "nested skills install is still a meta command",
			mutate: func(c *noticeConditions) { c.commandPath = "ical skills install" },
			want:   false,
		},
		{
			// json must be decided on its own rather than shadowed by a
			// redirected stderr — the two conditions are independent.
			name: "json output suppresses even at a terminal",
			mutate: func(c *noticeConditions) {
				c.outputFormat = "json"
				c.interactive = true
			},
			want: false,
		},
		{
			name:   "an ordinary subcommand is not a meta command",
			mutate: func(c *noticeConditions) { c.commandPath = "ical calendars list" },
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := interactiveConditions()
			tt.mutate(&c)

			if got := shouldShowNotices(c); got != tt.want {
				t.Errorf("shouldShowNotices(%+v) = %v, want %v", c, got, tt.want)
			}
		})
	}
}

func TestIsMetaCommand(t *testing.T) {
	tests := []struct {
		commandPath string
		want        bool
	}{
		{"", false},
		{"ical", false},
		{"ical list", false},
		{"ical calendars list", false},
		{"ical version", true},
		{"ical skills", true},
		{"ical skills status", true},
		{"ical skills uninstall", true},
		{"ical completion fish", true},
		// Exact word match only — a command merely starting with a meta name
		// must not be swept up.
		{"ical versions", false},
		{"ical skillsets list", false},
	}

	for _, tt := range tests {
		t.Run(tt.commandPath, func(t *testing.T) {
			if got := isMetaCommand(tt.commandPath); got != tt.want {
				t.Errorf("isMetaCommand(%q) = %v, want %v", tt.commandPath, got, tt.want)
			}
		})
	}
}

// TestCurrentNoticeConditionsWalksRealCommandTree pins the wiring, not just the
// rule: cobra hands the executed leaf to PersistentPostRun, so a check on
// cmd.Name() alone would read "status" here and miss the meta command entirely.
func TestCurrentNoticeConditionsWalksRealCommandTree(t *testing.T) {
	tests := []struct {
		args     []string
		wantPath string
		wantMeta bool
	}{
		{[]string{"skills", "status"}, "ical skills status", true},
		{[]string{"skills", "install"}, "ical skills install", true},
		{[]string{"version"}, "ical version", true},
		{[]string{"calendars", "list"}, "ical calendars list", false},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd, _, err := rootCmd.Find(tt.args)
			if err != nil {
				t.Fatalf("find %v: %v", tt.args, err)
			}

			got := currentNoticeConditions(cmd)
			if got.commandPath != tt.wantPath {
				t.Errorf("commandPath = %q, want %q", got.commandPath, tt.wantPath)
			}
			if isMetaCommand(got.commandPath) != tt.wantMeta {
				t.Errorf("isMetaCommand(%q) = %v, want %v", got.commandPath, !tt.wantMeta, tt.wantMeta)
			}
		})
	}
}

func TestCurrentNoticeConditionsReadsSuppressionEnv(t *testing.T) {
	t.Setenv("ICAL_NO_UPDATE_CHECK", "")
	if currentNoticeConditions(rootCmd).suppressed {
		t.Error("suppressed = true with an empty ICAL_NO_UPDATE_CHECK, want false")
	}

	t.Setenv("ICAL_NO_UPDATE_CHECK", "1")
	if !currentNoticeConditions(rootCmd).suppressed {
		t.Error("suppressed = false with ICAL_NO_UPDATE_CHECK=1, want true")
	}
}

func TestIsTerminal(t *testing.T) {
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer readEnd.Close()
	defer writeEnd.Close()

	if isTerminal(writeEnd) {
		t.Error("isTerminal(pipe) = true, want false")
	}

	regular, err := os.CreateTemp(t.TempDir(), "notice")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer regular.Close()

	if isTerminal(regular) {
		t.Error("isTerminal(regular file) = true, want false")
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer devNull.Close()

	// os.DevNull is a character device, so it reads as a terminal. Notices sent
	// there are discarded rather than polluting anything.
	if !isTerminal(devNull) {
		t.Errorf("isTerminal(%s) = false, want true", os.DevNull)
	}

	closed, err := os.CreateTemp(t.TempDir(), "closed")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	closed.Close()

	if isTerminal(closed) {
		t.Error("isTerminal(closed file) = true, want false — a failed Stat must not show notices")
	}
}

func TestCollectUpdateResult(t *testing.T) {
	t.Run("empty channel does not block", func(t *testing.T) {
		if got := collectUpdateResult(make(chan *update.Result, 1)); got != nil {
			t.Errorf("collectUpdateResult(empty) = %+v, want nil", got)
		}
	})

	t.Run("returns a buffered result", func(t *testing.T) {
		ch := make(chan *update.Result, 1)
		want := &update.Result{HasUpdate: true, Latest: "v0.13.0"}
		ch <- want

		if got := collectUpdateResult(ch); got != want {
			t.Errorf("collectUpdateResult = %+v, want %+v", got, want)
		}
	})

	t.Run("a nil result is passed through", func(t *testing.T) {
		ch := make(chan *update.Result, 1)
		ch <- nil

		if got := collectUpdateResult(ch); got != nil {
			t.Errorf("collectUpdateResult = %+v, want nil", got)
		}
	})
}

// installSkillFor writes SKILL.md for one agent under homeDir. An empty version
// omits the .ical-version file, which is how a pre-versioning install looks.
func installSkillFor(t *testing.T, homeDir, agentDir, version string) {
	t.Helper()

	skillDir := filepath.Join(homeDir, agentDir, "skills", "ical-cli")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("test skill\n"), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	if version == "" {
		return
	}
	if err := os.WriteFile(filepath.Join(skillDir, ".ical-version"), []byte(version+"\n"), 0o644); err != nil {
		t.Fatalf("write version: %v", err)
	}
}

// installSkill writes a skill whose recorded version is installedVersion. A
// version that differs from the binary version is what makes it stale.
func installSkill(t *testing.T, installedVersion string) string {
	t.Helper()

	homeDir := t.TempDir()
	installSkillFor(t, homeDir, ".claude", installedVersion)

	return homeDir
}

func TestPrintSkillsStalenessNotice(t *testing.T) {
	color.NoColor = true

	tests := []struct {
		name             string
		installedVersion string
		binaryVersion    string
		wantNotice       bool
	}{
		{
			name:             "outdated skills produce a notice",
			installedVersion: "v0.10.1",
			binaryVersion:    "v0.12.1",
			wantNotice:       true,
		},
		{
			name:             "matching versions stay silent",
			installedVersion: "v0.12.1",
			binaryVersion:    "v0.12.1",
			wantNotice:       false,
		},
		{
			name:             "dev builds stay silent",
			installedVersion: "v0.10.1",
			binaryVersion:    "dev",
			wantNotice:       false,
		},
		{
			name:             "unset version stays silent",
			installedVersion: "v0.10.1",
			binaryVersion:    "",
			wantNotice:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := installSkill(t, tt.installedVersion)

			var buf bytes.Buffer
			printSkillsStalenessNotice(&buf, tt.binaryVersion, homeDir)

			got := strings.Contains(buf.String(), "Installed skills are outdated")
			if got != tt.wantNotice {
				t.Errorf("notice printed = %v, want %v (output = %q)", got, tt.wantNotice, buf.String())
			}
			if tt.wantNotice && !strings.Contains(buf.String(), tt.installedVersion) {
				t.Errorf("notice %q does not name installed version %q", buf.String(), tt.installedVersion)
			}
		})
	}
}

func TestPrintSkillsStalenessNoticeEdgeCases(t *testing.T) {
	color.NoColor = true

	t.Run("no skills installed stays silent", func(t *testing.T) {
		var buf bytes.Buffer
		printSkillsStalenessNotice(&buf, "v0.12.1", t.TempDir())

		if buf.String() != "" {
			t.Errorf("output = %q, want empty when no skill is installed", buf.String())
		}
	})

	t.Run("missing home directory stays silent", func(t *testing.T) {
		var buf bytes.Buffer
		printSkillsStalenessNotice(&buf, "v0.12.1", filepath.Join(t.TempDir(), "does-not-exist"))

		if buf.String() != "" {
			t.Errorf("output = %q, want empty for an absent home directory", buf.String())
		}
	})

	t.Run("install without a version file stays silent", func(t *testing.T) {
		homeDir := t.TempDir()
		installSkillFor(t, homeDir, ".claude", "")

		var buf bytes.Buffer
		printSkillsStalenessNotice(&buf, "v0.12.1", homeDir)

		if buf.String() != "" {
			t.Errorf("output = %q, want empty when the version file is absent", buf.String())
		}
	})

	t.Run("version file whitespace is ignored", func(t *testing.T) {
		homeDir := t.TempDir()
		installSkillFor(t, homeDir, ".claude", "  v0.12.1  ")

		var buf bytes.Buffer
		printSkillsStalenessNotice(&buf, "v0.12.1", homeDir)

		if buf.String() != "" {
			t.Errorf("output = %q, want empty — padded version matches the binary", buf.String())
		}
	})

	t.Run("several stale agents produce one notice", func(t *testing.T) {
		homeDir := t.TempDir()
		installSkillFor(t, homeDir, ".claude", "v0.10.1")
		installSkillFor(t, homeDir, ".codex", "v0.10.1")

		var buf bytes.Buffer
		printSkillsStalenessNotice(&buf, "v0.12.1", homeDir)

		if got := strings.Count(buf.String(), "Installed skills are outdated"); got != 1 {
			t.Errorf("notice count = %d, want 1 (output = %q)", got, buf.String())
		}
	})

	t.Run("one stale agent among current ones still warns", func(t *testing.T) {
		homeDir := t.TempDir()
		installSkillFor(t, homeDir, ".claude", "v0.12.1")
		installSkillFor(t, homeDir, ".codex", "v0.10.1")

		var buf bytes.Buffer
		printSkillsStalenessNotice(&buf, "v0.12.1", homeDir)

		if !strings.Contains(buf.String(), "v0.10.1") {
			t.Errorf("output = %q, want a notice naming the stale v0.10.1 install", buf.String())
		}
	})
}

func TestPrintNoticesUpdateAvailable(t *testing.T) {
	color.NoColor = true

	// Skills are current, so only the update notice can fire.
	homeDir := installSkill(t, "v0.12.1")

	var buf bytes.Buffer
	printNotices(&buf, "v0.12.1", &update.Result{HasUpdate: true, Latest: "v0.13.0"}, homeDir)

	out := buf.String()
	if !strings.Contains(out, "A new version of ical is available") {
		t.Errorf("output %q is missing the update notice", out)
	}
	if !strings.Contains(out, "v0.13.0") {
		t.Errorf("output %q does not name the latest version", out)
	}
	if strings.Contains(out, "Installed skills are outdated") {
		t.Errorf("output %q has a staleness notice for up-to-date skills", out)
	}
}

func TestPrintNoticesSilentWhenNothingToReport(t *testing.T) {
	color.NoColor = true

	homeDir := installSkill(t, "v0.12.1")

	tests := []struct {
		name   string
		result *update.Result
	}{
		{"no result from the background check", nil},
		{"check ran and found nothing", &update.Result{HasUpdate: false}},
		{"check reports a latest version but no update", &update.Result{HasUpdate: false, Latest: "v0.12.1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printNotices(&buf, "v0.12.1", tt.result, homeDir)

			if buf.String() != "" {
				t.Errorf("output = %q, want empty when nothing is stale and no update exists", buf.String())
			}
		})
	}
}

// Both notices can fire in the same run; neither may hide the other.
func TestPrintNoticesUpdateAndStalenessTogether(t *testing.T) {
	color.NoColor = true

	homeDir := installSkill(t, "v0.10.1")

	var buf bytes.Buffer
	printNotices(&buf, "v0.12.1", &update.Result{HasUpdate: true, Latest: "v0.13.0"}, homeDir)

	out := buf.String()
	if !strings.Contains(out, "A new version of ical is available") {
		t.Errorf("output %q is missing the update notice", out)
	}
	if !strings.Contains(out, "Installed skills are outdated") {
		t.Errorf("output %q is missing the staleness notice", out)
	}
}

// A dev build reports neither notice, whatever the background check returned.
func TestPrintNoticesDevBuildSuppressesStaleness(t *testing.T) {
	color.NoColor = true

	homeDir := installSkill(t, "v0.10.1")

	var buf bytes.Buffer
	printNotices(&buf, "dev", nil, homeDir)

	if buf.String() != "" {
		t.Errorf("output = %q, want empty for a dev build", buf.String())
	}
}
