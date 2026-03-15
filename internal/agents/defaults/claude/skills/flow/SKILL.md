---
name: flow
description: Flow workspace management — commands, state format, rendering, and PRs. Load this skill at the start of every session.
user-invocable: false
---

# Flow

## Task Initialization

When launched into a fresh workspace with no repos:

1. **Edit `state.yaml`** with the task details:
   - `metadata.name` — short kebab-case name (e.g., `vpc-ipv6`)
   - `metadata.description` — one-line summary
   - `spec.repos` — repos needed (see State Format below)

2. **Run `flow render <workspace>`** — clones repos and creates branches. No flags needed.

3. **Start working** in the repo directories.

### State Format

```yaml
apiVersion: flow/v1
kind: State
metadata:
  name: my-project
  description: What this workspace is for
  created: "2026-01-01T00:00:00Z"
spec:
  repos:
    - url: github.com/org/repo     # git remote URL
      branch: feat/my-feature       # branch to work on
      base: staging                  # optional — create branch from this (default: repo's default branch)
      path: ./repo                   # optional — local directory name (derived from URL)
```

`branch` is the branch you work on. `base` is where it's created from. Omit `base` to branch from the repo's default (e.g., `main`).

### Example with custom base

```yaml
spec:
  repos:
    - url: github.com/acme/apps
      branch: qa-1
      base: staging
```

## Commands

| Command | Description |
|---------|-------------|
| `flow render <ws>` | Create fresh branches from base |
| `flow render <ws> --reset=false` | Use existing remote branches (errors if missing) |
| `flow list` | List all workspaces |
| `flow edit state <ws>` | Open state file in editor |
| `flow open <ws>` | Open shell in workspace |
| `flow exec <ws> -- <cmd>` | Run command in workspace |
| `flow delete <ws>` | Delete workspace and worktrees |

## Render Behavior

- **Default**: creates a fresh branch from `origin/{base}`, resetting if it already exists.
- **`--reset=false`**: uses an existing remote branch as-is. Errors if branch doesn't exist.
- **Additive**: re-render only processes new repos — existing worktrees are untouched.

## Pushing and Creating PRs

Push before creating a PR — flow branches are local until pushed.

```bash
git push -u origin HEAD
```

In worktrees, `gh` can't auto-detect the repo. Always pass `--repo` and `--head`:

```bash
gh pr create --repo <owner>/<repo> --head <branch> --title "..." --body "..."
```

## Workspace Structure

```
~/.flow/
├── config.yaml              # global config
├── agents/claude/            # shared Claude instructions
├── repos/                    # bare clone cache (shared across workspaces)
└── workspaces/<id>/
    ├── state.yaml            # workspace manifest (source of truth)
    ├── CLAUDE.md             # auto-generated context
    ├── .claude/ → shared     # symlinks to agents/claude/
    ├── repo-a/               # git worktree
    └── repo-b/               # git worktree
```

Workspaces are referenced by ID (e.g., `calm-delta`) or by `metadata.name`.
