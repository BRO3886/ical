package ical

import "embed"

// EmbeddedSkills contains the agent skills files (SKILL.md + references/)
// baked into the binary at build time.
//
//go:embed skills/ical-cli
var EmbeddedSkills embed.FS
