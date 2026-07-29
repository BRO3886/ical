# ical — Agent Skills Distribution & Update Check

> **Status: PLANNED**

## Problem

The `ical` CLI ships with agent skills (`skills/cal-cli/`) that teach AI coding agents how to use it — but users must manually symlink or copy these files into their agent's skills directory. There's also no way to know when a new version is available.

## Goals

1. **`ical skills install`** — embed skill files in the binary via `go:embed`, write them to the correct agent skills directory
2. **Background update check** — on any command, periodically check GitHub for a newer release and print a notice to stderr
3. **Post-install prompt** — after `scripts/install.sh` runs, prompt the user to install skills

## Non-Goals

- Self-update command (`ical update`) — too many install methods (go install, curl script, manual build) with different binary locations and permissions. Let the install script be the update mechanism.
- Publishing to a skills marketplace (future)
- MCP server bundling (separate concern)
- Windows/Linux support (macOS only)

---

## Design

### 1. Embedded Skills (`go:embed`)

Embed the entire `skills/cal-cli/` directory into the binary at build time:

```go
//go:embed skills/cal-cli
var embeddedSkills embed.FS
```

This means `SKILL.md`, `references/commands.md`, and `references/dates.md` are all baked into the binary. No external files needed at runtime.

**Build consideration**: The `go:embed` directive requires the `skills/` directory to be accessible relative to the Go source file that declares it. Options:
- (a) Place the embed directive in `cmd/ical/main.go` and move/symlink `skills/` under `cmd/ical/`
- (b) Create a dedicated `internal/skills/` package with an `embed.go` that references a copy of the skills dir
- (c) Use a top-level `embed.go` in the module root and import it from commands

Recommended: **(c)** — a top-level `skills.go` file in the module root (`/skills.go`) that embeds `skills/cal-cli` and exports it. Commands import from there. Keeps the source of truth in one place.

### 2. `ical skills` Command

New parent command with subcommands:

#### `ical skills install [--agent <agent>]`

Writes embedded skill files to the agent's skills directory.

**Supported agents and their paths:**

| Agent | Directory | Who reads it |
|-------|-----------|--------------|
| `claude` | `~/.claude/skills/ical-cli/` | Claude Code, Copilot, Cursor, OpenCode, Augment |
| `codex` | `~/.agents/skills/ical-cli/` | Codex CLI, Copilot, Windsurf, OpenCode, Augment |

**Interactive flow (default — no `--agent` flag):**

Uses `charmbracelet/huh` for a multi-select form so users can pick one or more targets:

```
$ ical skills install

  Install ical skill for which AI agents?

  > [x] Claude Code   (~/.claude/skills/)
    [x] Codex CLI     (~/.agents/skills/)

  ✓ Installed ical-cli skill to ~/.claude/skills/ical-cli/
  ✓ Installed ical-cli skill to ~/.agents/skills/ical-cli/

  Files: SKILL.md, references/commands.md, references/dates.md
  The skill will be available in your next session.
```

**Behavior:**
1. If `--agent` flag provided (`claude`, `codex`, or `all`), skip interactive prompt and install directly
2. If no flag, detect which agents are present (check if `~/.claude/` or `~/.agents/` dirs exist)
3. Pre-check detected agents in the `huh.MultiSelect` form (undetected agents still shown but unchecked)
4. If only one agent detected AND running non-interactively (piped stdin), default to that agent without prompting
5. Create directory, write all embedded files
6. Print confirmation with path per agent

**huh form details:**
- Use `huh.NewMultiSelect[string]` with options for each agent
- Theme: `ThemeCatppuccin()` (consistent with existing interactive forms)
- Validation: at least one agent must be selected
- The form shows the install path next to each agent name for transparency

**Non-interactive / CI mode:** When stdin is not a TTY (e.g., piped from install script), fall back to `--agent` flag or auto-detect. Never block on a prompt in a pipe.

**Overwrite policy:** Always overwrite. Skills are versioned with the binary — the binary is the source of truth. No merge, no diff.

#### `ical skills uninstall [--agent <agent>]`

Removes the skill directory from the agent's skills location.

**Interactive flow (default — no `--agent` flag):**

```
$ ical skills uninstall

  Uninstall ical skill from which AI agents?

  > [x] Claude Code   (~/.claude/skills/ical-cli/ — installed, v0.5.0)
    [ ] Codex CLI     (~/.agents/skills/ical-cli/ — not installed)

  ✓ Removed ical-cli skill from ~/.claude/skills/ical-cli/
```

Only agents with an existing installation are pre-checked. Uses the same `huh.MultiSelect` pattern as install.

#### `ical skills status`

Shows where skills are installed and whether they're up to date.

```
$ ical skills status
ical-cli skill (v0.5.0):
  ~/.claude/skills/ical-cli/  ✓ installed (v0.5.0)
  ~/.agents/skills/ical-cli/  ✗ not installed
```

**Version tracking:** Write a `.ical-version` file inside the installed skill directory containing the binary version. `status` compares this against the running binary's version.

### 3. Background Update Check

On any command, periodically check if a newer version exists and print a one-line notice to stderr. Non-blocking, non-intrusive.

**State file:** `~/.cache/ical/update-check` — a simple file containing:
```
checked_at=2026-02-16T10:00:00Z
latest=v0.5.0
```

**Check logic (runs in `PersistentPreRun` of root command):**

1. Read `~/.cache/ical/update-check`. If `checked_at` is less than 24 hours ago, use cached `latest` value — no HTTP request.
2. If stale or missing, spawn a goroutine that:
   - Hits `https://api.github.com/repos/BRO3886/ical/releases/latest` (no auth needed)
   - Parses the `tag_name` field
   - Writes the result + timestamp to `~/.cache/ical/update-check`
3. In `PersistentPostRun`, if the goroutine completed and `latest > current`:
   ```
   A new version of ical is available: v0.4.0 → v0.5.0
   Update: curl -fsSL https://ical.sidv.dev/install | bash
   ```
   Printed to **stderr** so it doesn't interfere with piped output (e.g., `ical ls -o json | jq`).

**Skip conditions — do NOT check when:**
- `ICAL_NO_UPDATE_CHECK=1` env var is set (for CI/scripts)
- `--output json` is used (scripting context)
- The command is `version` or `completion` (meta commands)
- Stdout is not a TTY (piped output)

**Goroutine timeout:** 2 seconds max. If GitHub is slow or unreachable, silently give up. Never delay the user's command.

#### Skills staleness notice

When skills are installed but their `.ical-version` doesn't match the running binary, piggyback on the same stderr notice:

```
A new version of ical is available: v0.4.0 → v0.5.0
Update: curl -fsSL https://ical.sidv.dev/install | bash

Installed skills are outdated (v0.4.0). Run: ical skills install
```

The skills staleness check is local-only (no HTTP) — just compare `.ical-version` file content against the binary's version. This runs even when the update check is cached/skipped.

### 4. Post-Install Prompt

After `scripts/install.sh` places the binary, delegate to `ical skills install` which handles the interactive huh form natively:

```bash
# After binary installation succeeds:
echo ""
echo "ical can install an AI agent skill that teaches Claude Code / Codex how to use it."
printf "Install agent skill now? [Y/n] "
read -r answer
if [ "$answer" != "n" ] && [ "$answer" != "N" ]; then
    ical skills install
fi
```

The shell script only handles the yes/no gate. Once the user says yes, `ical skills install` takes over with the full huh multi-select form — the user picks which agents to install for, sees the paths, and gets confirmation. This avoids duplicating agent detection logic in bash.

---

## Command Tree (Updated)

```
ical
├── calendars [list|create|update|delete]
├── list (ls)
├── today
├── upcoming
├── show
├── add
├── update
├── delete
├── search
├── export
├── import
├── skills [install|uninstall|status]   ← NEW
├── version
└── completion
```

---

## Implementation Plan

### Phase 1: Embedded Skills + `ical skills install`
1. Create `skills.go` at module root with `go:embed skills/cal-cli`
2. Add `cmd/ical/commands/skills.go` with `install`, `uninstall`, `status` subcommands
3. Interactive agent selection with `huh.NewMultiSelect`
4. Write embedded files to `~/.claude/skills/ical-cli/` and/or `~/.agents/skills/ical-cli/`
5. Include `.ical-version` file for version tracking
6. Add `--agent` flag for non-interactive use

### Phase 2: Background Update Check
1. Add update check logic in root command's `PersistentPreRun` (goroutine) and `PersistentPostRun` (print notice)
2. State file `~/.cache/ical/update-check` with 24-hour TTL
3. GitHub API call with 2-second timeout
4. Skills staleness check (local, no HTTP)
5. Stderr output, skip conditions (env var, json output, piped, meta commands)

### Phase 3: Post-Install Integration
1. Update `scripts/install.sh` to prompt for skill install after binary placement
2. Update `website/static/install` to match

---

## Edge Cases

- **Skills dir doesn't exist**: Create it (and parent dirs) with `os.MkdirAll`
- **Existing symlinked skill**: Remove symlink before writing files (user may have been using dev symlink)
- **GitHub unreachable**: Update check goroutine times out after 2s, silently gives up
- **Rate limiting on GitHub API**: Unauthenticated limit is 60/hr — with 24h cache, this is never hit
- **State file unwritable**: Silently skip (e.g., weird permissions on home dir)
- **Concurrent commands**: Two `ical` invocations racing to write `~/.cache/ical/update-check` — harmless, last writer wins
- **Version comparison**: Use semver parsing, not string comparison (`v0.10.0 > v0.9.0`)
- **Non-interactive stdin**: Detect with `os.Stdin.Stat()` checking for `ModeCharDevice`, fall back to `--agent` flag

## Open Questions

1. Should `ical skills install` also support installing to `.claude/skills/` in the current project directory (project-scoped) as an option?
2. Should the update check notice include a "don't show again" hint (`export ICAL_NO_UPDATE_CHECK=1`)?
