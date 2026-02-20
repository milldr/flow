---
name: workspace-structure
description: Flow workspace directory layout and key concepts. Use when navigating a flow workspace, understanding the directory structure, or working with git worktrees.
user-invocable: false
---

# Workspace Structure

## Flow Home Directory

Flow stores all data under `$FLOW_HOME` (defaults to `~/.flow/`):

```
~/.flow/
├── config.yaml           # global flow configuration
├── agents/               # AI agent configurations
│   └── claude/           # shared Claude Code instructions
├── workspaces/           # all workspace directories
│   └── <id>/             # one directory per workspace
└── repos/                # bare clone cache (shared across workspaces)
```

## Workspace Directory

Each workspace is identified by a random ID (e.g., `calm-delta`) and contains:

```
~/.flow/workspaces/<id>/
├── state.yaml            # workspace manifest (source of truth)
├── CLAUDE.md             # auto-generated workspace context
├── .claude/              # Claude Code configuration
│   ├── CLAUDE.md         # → symlink to shared instructions
│   └── skills/           # → symlink to shared skills
├── repo-a/               # git worktree
└── repo-b/               # git worktree
```

## Key Concepts

- **Render handles everything**: `flow render` clones repos, creates branches, and checks out worktrees. You never need to clone or checkout repos manually.
- **Branch = desired branch**: the `branch` field in `state.yaml` is the branch you want to work on. If it exists, flow checks it out. If it doesn't exist, flow creates it off the repo's default branch (e.g., `main`).
- **Worktrees, not clones**: repos are bare-cloned once to `~/.flow/repos/` and shared across workspaces via git worktrees. This saves disk space and enables multiple workspaces to use the same repo at different branches.
- **State-driven**: the `state.yaml` file is the source of truth. Edit it and re-render to change the workspace.
- **Idempotent render**: running `flow render` multiple times is safe — it skips repos that are already checked out and fetches updates for existing bare clones.
