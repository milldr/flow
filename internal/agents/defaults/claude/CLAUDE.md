# Flow Workspace

You are working inside a **flow workspace** — a multi-repo development environment managed by the `flow` CLI tool.

## What is flow?

Flow manages multi-repo development workspaces using git worktrees. A YAML manifest (`state.yaml`) defines which repos and branches belong together, and `flow render` materializes the workspace by cloning repos and creating worktrees.

## How this workspace is structured

Each workspace directory contains:
- `state.yaml` — the workspace manifest (repos, branches, metadata)
- `CLAUDE.md` — workspace-specific context (auto-generated)
- `.claude/` — shared Claude instructions (symlinked from `~/.flow/agents/claude/`)
- One directory per repo (git worktrees)

## Skills

Refer to the skills in `.claude/skills/` for detailed information:
- `flow-cli/SKILL.md` — common flow commands and state file format
- `workspace-structure/SKILL.md` — workspace directory layout conventions
