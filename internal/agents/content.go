// Package agents manages AI agent configuration files for workspaces.
package agents

import _ "embed"

//go:embed defaults/claude/CLAUDE.md
var defaultClaudeMD []byte

//go:embed defaults/claude/skills/flow/SKILL.md
var defaultFlowSkill []byte
