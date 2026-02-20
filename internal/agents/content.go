// Package agents manages AI agent configuration files for workspaces.
package agents

import _ "embed"

//go:embed defaults/claude/CLAUDE.md
var defaultClaudeMD []byte

//go:embed defaults/claude/skills/flow-cli/SKILL.md
var defaultFlowCLI []byte

//go:embed defaults/claude/skills/workspace-structure/SKILL.md
var defaultWorkspaceStructure []byte
