# PRD-005: Global Config & Claude Workspace Instructions

**Depends on:** [002-mvp.md](./002-mvp.md)

## Summary

Add a global config file (`$FLOW_HOME/config.yaml`) and auto-generate Claude Code instructions for every rendered workspace. When running `flow exec <ws> -- claude`, Claude Code will understand the workspace structure, what repos are present, and how flow works.

## Motivation

Flow workspaces currently have no AI context. Running `flow exec <ws> -- claude` drops Claude Code into a multi-repo workspace with no understanding of:

- What flow is or how it manages workspaces
- Which repos are present and on what branches
- How the directory structure is organized
- Common flow commands for managing state

Additionally, flow has no global config file — all configuration is derived from paths. Adding `$FLOW_HOME/config.yaml` provides a foundation for future settings.

## Changes

### Global config

Auto-create `$FLOW_HOME/config.yaml` on first run with default values:

```yaml
apiVersion: flow/v1
kind: Config
```

The config is loaded on every CLI invocation. If missing, it's created with defaults. If present, it's loaded as-is.

### Shared agent files

Create `$FLOW_HOME/agents/claude/` with shared context files:

- `CLAUDE.md` — what flow is, how workspaces work, pointers to skills
- `skills/flow-cli/SKILL.md` — common flow commands and state file format
- `skills/workspace-structure/SKILL.md` — workspace directory layout

These files are written **only if missing** — user edits are preserved.

### Per-workspace Claude setup

On `flow render`, generate workspace-specific Claude files:

- `<workspace>/CLAUDE.md` — workspace-specific context (name, description, repo table). **Always overwritten** to reflect current state.
- `<workspace>/.claude/CLAUDE.md` → symlink to shared `CLAUDE.md`
- `<workspace>/.claude/skills/` → symlink to shared `skills/`

## Directory Structure

```
~/.flow/
├── config.yaml
├── agents/
│   └── claude/
│       ├── CLAUDE.md
│       └── skills/
│           ├── flow-cli/
│           │   └── SKILL.md
│           └── workspace-structure/
│               └── SKILL.md
├── workspaces/
│   └── <id>/
│       ├── state.yaml
│       ├── CLAUDE.md                  # auto-generated, workspace-specific
│       ├── .claude/
│       │   ├── CLAUDE.md  →  shared   # symlink
│       │   └── skills/    →  shared   # symlink
│       ├── repo-a/
│       └── repo-b/
└── repos/
```

## Files

| File | Action |
|------|--------|
| `docs/prd/005-config-and-claude.md` | Created — this PRD |
| `docs/prd/README.md` | Modified — add PRD-005 row |
| `internal/config/flowconfig.go` | Created — `FlowConfig` type, load/save/defaults |
| `internal/config/flowconfig_test.go` | Created — tests |
| `internal/config/config.go` | Modified — `AgentsDir`, `ConfigFile`, `FlowConfig` fields |
| `internal/config/config_test.go` | Modified — tests for new fields |
| `internal/agents/content.go` | Created — `go:embed` directives for default content |
| `internal/agents/defaults/claude/CLAUDE.md` | Created — shared context |
| `internal/agents/defaults/claude/skills/flow-cli.md` | Created — flow CLI skill |
| `internal/agents/defaults/claude/skills/workspace-structure.md` | Created — workspace structure skill |
| `internal/agents/claude.go` | Created — `EnsureSharedAgent()`, `SetupWorkspaceClaude()` |
| `internal/agents/claude_test.go` | Created — tests |
| `internal/cmd/root.go` | Modified — call `EnsureSharedAgent()` |
| `internal/workspace/workspace.go` | Modified — call `SetupWorkspaceClaude()` in `Render()` |
| `internal/workspace/workspace_test.go` | Modified — update `testService`, add assertions |

## Design Decisions

- **Shared files never overwritten** — user may customize them
- **Workspace CLAUDE.md always overwritten** — fully generated from state; users customize via shared CLAUDE.md instead
- **`agents/` dir is provider-agnostic** — future `agents/cursor/`, `agents/codex/` etc.
- **Claude setup runs on render, not init** — at render time we know the repos
- **`EnsureSharedAgent` called from `root.go`** not `EnsureDirs()` — avoids import cycle
- **Symlinks use absolute paths** — reliable regardless of working directory

## Acceptance Criteria

- [ ] `$FLOW_HOME/config.yaml` auto-created on first run
- [ ] `$FLOW_HOME/agents/claude/` populated with default files
- [ ] Shared files preserved across runs (not overwritten)
- [ ] `flow render` generates workspace-specific `CLAUDE.md`
- [ ] `.claude/CLAUDE.md` symlinks to shared context
- [ ] `.claude/skills/` symlinks to shared skills
- [ ] Re-render is idempotent (no errors, symlinks unchanged)
- [ ] `make build`, `make lint`, `make test` all pass
