Create a new release for the `ical` project.

## Step 1: Analyze Changes

Run these commands to understand what changed since the last release:

```bash
git tag --sort=-v:refname | head -1   # Latest tag
git log <latest-tag>..HEAD --oneline  # Commits since last release
git diff <latest-tag>..HEAD --stat    # Files changed
```

Read the commit messages carefully. Classify the release:

### Semver Rules
- **MAJOR** (vX.0.0): Breaking changes — removed commands, renamed flags, changed default behavior, dropped compatibility
- **MINOR** (v0.X.0): New features, new commands, new flags, new packages. No breaking changes.
- **PATCH** (v0.0.X): Bug fixes, docs-only changes, performance improvements, dependency bumps. No new features.

Look for commits prefixed with:
- `feat(...)` → MINOR bump (or MAJOR if `feat!` or contains `BREAKING CHANGE`)
- `fix(...)` → PATCH bump
- `docs(...)`, `chore(...)`, `refactor(...)`, `perf(...)` → PATCH bump
- `!` suffix or `BREAKING CHANGE` in body → MAJOR bump

If the current version is pre-1.0 (v0.x.y), breaking changes bump MINOR not MAJOR.

**Present the proposed version to the user and ask for confirmation before proceeding.**

## Step 2: Run Tests

```bash
go test ./...
```

All tests must pass. Do not proceed if any test fails.

## Step 3: Tag and Push

**Tag BEFORE building** — `make release` uses `git describe --tags` to embed the version in the binary. If you build before tagging, the binary will report the wrong version (e.g. `v0.5.0-1-gabcdef` instead of `v0.5.1`).

```bash
git tag v<VERSION>
git push origin v<VERSION>
```

## Step 4: Build Release Binaries

```bash
make release
```

This produces `bin/ical-darwin-arm64.tar.gz` and `bin/ical-darwin-amd64.tar.gz`.

Verify the version is correct before uploading:

```bash
cd /tmp && tar -xzf /path/to/bin/ical-darwin-arm64.tar.gz && ./ical version && rm ./ical
```

The output must show exactly `ical v<VERSION>` with no commit suffix.

## Step 5: Create GitHub Release

Use `gh release create` with the format below. Generate the notes by analyzing the commits from Step 1.

### Release Notes Format

```
gh release create v<VERSION> \
  bin/ical-darwin-arm64.tar.gz \
  bin/ical-darwin-amd64.tar.gz \
  --title "v<VERSION>" \
  --notes "$(cat <<'EOF'
<RELEASE_NOTES>
EOF
)"
```

### Release Notes Template

```markdown
## Breaking Changes

<!-- ONLY include this section if there are breaking changes. Delete entirely otherwise. -->

- **<what broke>** — <migration instructions>

## What's New

<!-- Group related changes under descriptive subheadings. Use ### for major features, bullet points for smaller changes. -->

### <Feature Name>

<1-3 sentence description of what it does and why it matters.>

```bash
<example usage>
```

### Other Changes

- <bullet for smaller changes: bug fixes, refactors, dep bumps>
- <bullet>

## Install / Update

```bash
curl -fsSL https://ical.sidv.dev/install | bash
```

Or via Go:

```bash
go install github.com/BRO3886/ical/cmd/ical@v<VERSION>
```

**Full Changelog**: https://github.com/BRO3886/ical/compare/v<PREV>...v<VERSION>
```

### Rules for Release Notes

1. **Lead with breaking changes** if any — users need to see these first
2. **Group by feature, not by file** — users care about capabilities, not internal structure
3. **Include code examples** for new commands or flags — show, don't just tell
4. **Keep it scannable** — subheadings, bullets, code blocks. No prose walls.
5. **Always end with install instructions** and full changelog link
6. **No internal implementation details** — users don't care about package names or refactors unless they affect the CLI surface
7. **Mention new env vars and flags** — these are user-facing API
