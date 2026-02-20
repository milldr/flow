---
name: flow-cli
description: Flow CLI commands and state file format. Use when working with flow workspaces, editing state files, or running flow commands.
user-invocable: false
---

# Flow CLI

## Common Commands

| Command | Description |
|---------|-------------|
| `flow list` | List all workspaces |
| `flow state <workspace>` | Show workspace state (YAML manifest) |
| `flow render <workspace>` | Materialize workspace (clone repos, create worktrees) |
| `flow exec <workspace> -- <cmd>` | Run a command inside the workspace directory |
| `flow init <workspace>` | Create a new workspace interactively |
| `flow delete <workspace>` | Delete a workspace and its worktrees |

## State File Format

The workspace manifest (`state.yaml`) follows this schema:

```yaml
apiVersion: flow/v1
kind: State
metadata:
  name: my-project          # human-friendly name (optional)
  description: Description  # what this workspace is for (optional)
  created: 2024-01-01T00:00:00Z
spec:
  repos:
    - url: github.com/org/repo-a
      branch: feat/my-feature
      path: ./repo-a          # local directory name (optional, derived from URL)
    - url: github.com/org/repo-b
      branch: main
```

## Editing State

To modify a workspace (add/remove repos, change branches), edit the `state.yaml` file directly, then run `flow render <workspace>` to apply changes.

## Workspace Resolution

Workspaces can be referenced by their ID (directory name like `calm-delta`) or by their metadata name. If a name matches multiple workspaces, flow will report the ambiguity.
