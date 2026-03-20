package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	ical "github.com/BRO3886/ical"
	"github.com/BRO3886/ical/internal/skills"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var skillsAgentFlag string

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage AI agent skills for ical",
	Long: `Install, uninstall, and check the status of the ical agent skill.

The ical skill teaches AI coding agents (Claude Code, Codex CLI, etc.)
how to use ical effectively. It includes command references, date parsing
docs, and usage examples.`,
}

func init() {
	rootCmd.AddCommand(skillsCmd)
}

// --- skills install ---

var skillsInstallDryRun bool

var skillsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install ical skill for AI agents",
	Long: `Installs the ical agent skill to the selected AI agent's skill directory.

Supported agents:
  claude    → ~/.claude/skills/ical-cli/    (Claude Code, Copilot, Cursor, OpenCode, Augment)
  codex     → ~/.codex/skills/ical-cli/     (Codex CLI)
  openclaw  → ~/.openclaw/skills/ical-cli/  (OpenClaw)
  others    → ~/.agents/skills/ical-cli/    (Copilot, Windsurf, OpenCode, Augment)

Without --agent, shows an interactive picker to select which agents to install for.

The skill files are documentation (SKILL.md + reference pages) that teach AI
agents how to invoke ical. They contain the same information published at
https://ical.sidv.dev/docs — use --dry-run to preview what will be written.`,
	RunE: runSkillsInstall,
}

func init() {
	skillsInstallCmd.Flags().StringVar(&skillsAgentFlag, "agent", "", "Agent target: claude, codex, openclaw, others, or all")
	skillsInstallCmd.Flags().BoolVar(&skillsInstallDryRun, "dry-run", false, "Preview what would be installed without writing anything")
	skillsCmd.AddCommand(skillsInstallCmd)
}

func runSkillsInstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	allTargets := skills.DefaultTargets(homeDir)
	targets, err := resolveTargets(allTargets, skillsAgentFlag, homeDir, "install")
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil // user cancelled
	}

	// Dry-run: show what would be written and exit
	if skillsInstallDryRun {
		return printDryRun(ical.EmbeddedSkills, targets, homeDir)
	}

	// Show disclaimer and ask for confirmation (interactive only).
	// Non-interactive invocations (piped stdin, --agent in CI) skip this
	// intentionally so scripts and agents can install without blocking.
	if isInteractive() {
		if !confirmInstall(targets, homeDir) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	return installToTargets(ical.EmbeddedSkills, targets, homeDir)
}

// confirmInstall shows a disclaimer about what skills install does and asks
// for explicit confirmation. Returns true if the user agrees.
func confirmInstall(targets []skills.AgentTarget, homeDir string) bool {
	cyan := color.New(color.FgCyan)
	cyan.Println("\nThe skill files are documentation that teaches AI agents how to use ical.")
	cyan.Println("They contain the same content published at https://ical.sidv.dev/docs")
	fmt.Println("\nThe following paths will be created:")
	for _, t := range targets {
		fmt.Printf("  • %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
	}
	fmt.Println("\nUse --dry-run to preview the exact files without writing anything.")

	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Proceed with installation?").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return false
	}
	return confirm
}

// printDryRun lists the files that would be written without actually writing them.
func printDryRun(embeddedFS fs.FS, targets []skills.AgentTarget, homeDir string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Println("Dry run — no files will be written.")
	fmt.Println()

	// Collect embedded file paths
	var files []string
	err := fs.WalkDir(embeddedFS, "skills/ical-cli", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel("skills/ical-cli", path)
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk embedded skills: %w", err)
	}

	for _, t := range targets {
		displayDir := skills.DisplayPath(skills.SkillDir(t), homeDir)
		fmt.Printf("%s (%s):\n", t.Name, displayDir)
		for _, f := range files {
			fmt.Printf("  %s\n", f)
		}
		fmt.Printf("  %s\n\n", skills.VersionFileName)
	}

	fmt.Println("These are the same docs published at https://ical.sidv.dev/docs")
	fmt.Println("Run without --dry-run to install.")
	return nil
}

func installToTargets(embeddedFS fs.FS, targets []skills.AgentTarget, homeDir string) error {
	green := color.New(color.FgGreen, color.Bold)

	for _, t := range targets {
		written, err := skills.Install(embeddedFS, t, versionStr)
		if err != nil {
			return fmt.Errorf("failed to install for %s: %w", t.Name, err)
		}

		green.Print("✓ ")
		fmt.Printf("Installed ical-cli skill to %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
		fmt.Printf("  Files: %s\n", strings.Join(written, ", "))
	}

	fmt.Println("\nThe skill will be available in your next session.")
	return nil
}

// --- skills uninstall ---

var skillsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall ical skill from AI agents",
	Long:  `Removes the ical agent skill from the selected AI agent's skill directory.`,
	RunE:  runSkillsUninstall,
}

func init() {
	skillsUninstallCmd.Flags().StringVar(&skillsAgentFlag, "agent", "", "Agent target: claude, codex, openclaw, others, or all")
	skillsCmd.AddCommand(skillsUninstallCmd)
}

func runSkillsUninstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	allTargets := skills.DefaultTargets(homeDir)
	targets, err := resolveTargets(allTargets, skillsAgentFlag, homeDir, "uninstall")
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil // user cancelled
	}

	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)

	for _, t := range targets {
		removed, err := skills.Uninstall(t)
		if err != nil {
			return fmt.Errorf("failed to uninstall from %s: %w", t.Name, err)
		}
		if removed {
			green.Print("✓ ")
			fmt.Printf("Removed ical-cli skill from %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
		} else {
			yellow.Print("- ")
			fmt.Printf("Not installed at %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
		}
	}

	return nil
}

// --- skills status ---

var skillsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show skill installation status",
	RunE:  runSkillsStatus,
}

func init() {
	skillsCmd.AddCommand(skillsStatusCmd)
}

func runSkillsStatus(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	allTargets := skills.DefaultTargets(homeDir)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	fmt.Printf("ical-cli skill (binary %s):\n", versionStr)
	for _, t := range allTargets {
		displayDir := skills.DisplayPath(skills.SkillDir(t), homeDir)
		if !skills.IsInstalled(t) {
			red.Printf("  ✗ ")
			fmt.Printf("%-12s %s (not installed)\n", t.Name, displayDir)
			continue
		}
		installed := skills.InstalledVersion(t)
		if installed == "" {
			yellow.Printf("  ? ")
			fmt.Printf("%-12s %s (installed, unknown version)\n", t.Name, displayDir)
		} else if installed != versionStr {
			yellow.Printf("  ⚠ ")
			fmt.Printf("%-12s %s (installed %s, outdated)\n", t.Name, displayDir, installed)
		} else {
			green.Printf("  ✓ ")
			fmt.Printf("%-12s %s (installed %s)\n", t.Name, displayDir, installed)
		}
	}

	return nil
}

// --- shared helpers ---

// resolveTargets determines which agent targets to operate on.
// If --agent is specified, uses that directly.
// Otherwise, shows an interactive huh.MultiSelect picker.
func resolveTargets(allTargets []skills.AgentTarget, agentFlag, homeDir, action string) ([]skills.AgentTarget, error) {
	// If --agent flag provided, resolve directly
	if agentFlag != "" {
		return resolveAgentFlag(allTargets, agentFlag)
	}

	// Check if stdin is a TTY — if not, auto-detect
	if !isInteractive() {
		detected := skills.DetectAgents(allTargets)
		if len(detected) == 0 {
			// Default to claude
			return allTargets[:1], nil
		}
		return detected, nil
	}

	// Interactive multi-select
	return runAgentPicker(allTargets, homeDir, action)
}

func resolveAgentFlag(allTargets []skills.AgentTarget, flag string) ([]skills.AgentTarget, error) {
	flag = strings.ToLower(strings.TrimSpace(flag))
	if flag == "all" {
		return allTargets, nil
	}
	for _, t := range allTargets {
		if t.Key == flag {
			return []skills.AgentTarget{t}, nil
		}
	}
	return nil, fmt.Errorf("unknown agent %q (valid: claude, codex, openclaw, others, all)", flag)
}

func runAgentPicker(allTargets []skills.AgentTarget, homeDir, action string) ([]skills.AgentTarget, error) {
	detected := skills.DetectAgents(allTargets)
	detectedKeys := make(map[string]bool)
	for _, d := range detected {
		detectedKeys[d.Key] = true
	}

	// Build options
	options := make([]huh.Option[string], len(allTargets))
	for i, t := range allTargets {
		label := fmt.Sprintf("%-12s (%s)", t.Name, skills.DisplayPath(t.BaseDir, homeDir))
		if action == "uninstall" && skills.IsInstalled(t) {
			v := skills.InstalledVersion(t)
			if v != "" {
				label += fmt.Sprintf(" — installed, %s", v)
			} else {
				label += " — installed"
			}
		}
		options[i] = huh.NewOption(label, t.Key)
	}

	// Pre-select based on context
	var preselected []string
	if action == "uninstall" {
		// Pre-check installed agents
		for _, t := range allTargets {
			if skills.IsInstalled(t) {
				preselected = append(preselected, t.Key)
			}
		}
	} else {
		// Pre-check detected agents
		for _, t := range allTargets {
			if detectedKeys[t.Key] {
				preselected = append(preselected, t.Key)
			}
		}
		// If none detected, pre-check all
		if len(preselected) == 0 {
			for _, t := range allTargets {
				preselected = append(preselected, t.Key)
			}
		}
	}

	var selected []string
	title := fmt.Sprintf("%s ical skill for which AI agents?",
		strings.ToUpper(action[:1])+action[1:])

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(options...).
				Value(&selected).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return fmt.Errorf("select at least one agent")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCatppuccin())

	// Set preselected values
	selected = preselected

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("selection error: %w", err)
	}

	// Map selected keys back to targets
	selectedMap := make(map[string]bool)
	for _, s := range selected {
		selectedMap[s] = true
	}

	var targets []skills.AgentTarget
	for _, t := range allTargets {
		if selectedMap[t.Key] {
			targets = append(targets, t)
		}
	}

	return targets, nil
}

// isInteractive returns true if stdin is a terminal.
func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
