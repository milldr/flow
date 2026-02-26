---
name: flow-cli
description: Flow CLI commands and state file format. Use when working with flow workspaces, editing state files, or running flow commands.
user-invocable: false
---

# Flow CLI

## Task Initialization

When you are launched into a fresh workspace with no repos, your first job is to set up the workspace for the task. The user may paste a message from Slack, dictate a prompt, or describe the work in plain language. Follow these steps:

1. **Understand the task**: Read the user's prompt. Identify which repos are needed and what the work involves.

2. **Edit `state.yaml`**: Update the workspace manifest to configure the task:
   - `metadata.name` — a short, kebab-case name for the task (e.g., `vpc-ipv6`, `fix-auth-bug`)
   - `metadata.description` — a one-line summary of the task
   - `spec.repos` — add the repos needed for this task:
     - `url` — the git remote URL
     - `branch` — if the user specifies a branch, use it. Otherwise, create a descriptive feature branch name (e.g., `feat/add-ipv6-support`). Flow will automatically create the branch from the repo's default branch if it doesn't already exist in the remote.

3. **Render the workspace**: Run `flow render <workspace>` (using either the workspace ID or name). This clones repos and checks out the specified branches.

4. **Start working**: The repos are now available as directories in the workspace. Begin the task.

### Example

Given a prompt like _"Add IPv6 support to the VPC service — repo is github.com/acme/vpc-service"_:

```yaml
apiVersion: flow/v1
kind: State
metadata:
  name: vpc-ipv6
  description: Add IPv6 support to the VPC service
  created: "2026-02-24T00:00:00Z"
spec:
  repos:
    - url: github.com/acme/vpc-service
      branch: feat/ipv6-support
```

Then run `flow render vpc-ipv6` to materialize the workspace and start coding.

## Common Commands

| Command | Description |
|---------|-------------|
| `flow list` | List all workspaces |
| `flow edit state <workspace>` | Open workspace state file in editor |
| `flow render <workspace>` | Clone repos and check out branches (creates worktrees) |
| `flow open <workspace>` | Open a shell in the workspace directory |
| `flow exec <workspace> -- <cmd>` | Run a command inside the workspace directory |
| `flow init <workspace>` | Create a new workspace interactively |
| `flow delete <workspace>` | Delete a workspace and its worktrees |
| `flow reset status` | Reset global status spec to default |
| `flow reset config` | Reset global config to default |
| `flow reset state <workspace>` | Reset workspace state to default (preserves name/description) |
| `flow reset skills` | Reset shared agent skills to their defaults |

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
      branch: feat/my-feature   # the branch you want to work on
      path: ./repo-a            # local directory name (optional, derived from URL)
    - url: github.com/org/repo-b
      branch: main
```

**Important**: `branch` is the branch you want to work on, not a base branch. If the branch already exists in the remote, `flow render` checks it out. If it doesn't exist, flow automatically creates it off the repo's default branch (e.g., `main`). You do not need to clone repos yourself — `flow render` handles cloning, branch creation, and checkout.

## Pushing and Creating PRs

After committing your changes, you must push the branch to the remote **before** creating a PR. Flow renders branches locally — they don't exist on GitHub until you push.

```bash
# From inside the repo directory:
git push -u origin HEAD
```

Then create the PR:

```bash
gh pr create --title "..." --body "..."
```

**Common mistake**: running `gh pr create` without pushing first. This produces:

```
aborted: you must first push the current branch to a remote, or use the --head flag
```

Always push first. The `-u` flag sets up tracking so subsequent pushes are just `git push`.

## Editing State

To modify a workspace (add/remove repos, change branches), edit the `state.yaml` file directly, then run `flow render <workspace>` to apply changes. Flow will clone any new repos and check out the specified branches.

## Workspace Resolution

Workspaces can be referenced by their ID (directory name like `calm-delta`) or by their metadata name. If a name matches multiple workspaces, flow will report the ambiguity.
