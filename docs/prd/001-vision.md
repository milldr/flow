# PRD-001: Flow Vision

## Problem

Modern development often spans multiple repositories:
- A library and the applications that consume it
- Infrastructure code and application code
- Backend services and frontend clients
- Terraform modules and the stacks that use them

Managing this is painful:

1. **Setup friction** — Clone multiple repos, checkout the right branches, remember which branches belong together
2. **Context loss** — Switch tasks, forget which repos/branches were part of that work
3. **Cleanup debt** — Old branches and clones accumulate, unclear what's safe to delete
4. **No reproducibility** — Can't easily recreate a workspace or share setup with teammates

Git worktrees help but are manual. Existing tools (monorepo tooling, IDE workspaces) either assume single-repo or add too much complexity.

## Solution

**Flow** is a CLI tool that manages multi-repo development workspaces using git worktrees.

A workspace is defined by a **state file** — a YAML manifest listing which repos and branches belong together. Run `flow state render` and your environment is ready.

```yaml
# state.yaml
apiVersion: flow/v1
kind: State
metadata:
  name: vpc-ipv6
  description: "Add IPv6 support to VPC module"

spec:
  repos:
    - url: github.com/cloudposse/atmos
      branch: feat/ipv6-support
      path: ./atmos
    - url: github.com/cloudposse/docs
      branch: feat/ipv6-support
      path: ./docs
```

## Philosophy

### State-driven
The state file is the source of truth. The workspace is a rendered output of the state. Change the state, re-render, environment updates.

### Config-as-code
Workspace definitions are YAML files. They can be versioned, shared, templated. No hidden state in GUI or databases.

### Multi-repo native
Built from the ground up for projects spanning multiple repositories. Not an afterthought bolted onto a single-repo tool.

### Worktree-based
Uses git worktrees for lightweight, fast checkouts. Multiple workspaces can share the same bare clone cache.

### Minimal by default
Core functionality only. Integrations (GitHub, Linear, Slack) are optional add-ons, not required dependencies.

## Core Concepts

### Workspace
An isolated development environment containing one or more repository worktrees. Identified by a name, described by a state file.

### State File
A YAML manifest (`state.yaml`) that defines workspace configuration. Human-readable and editable.

### Bare Clone Cache
Flow maintains bare clones of repositories in `~/.flow/repos/`. Worktrees are created from these, making subsequent workspace creation fast.

### Render
The act of creating/updating worktrees to match the state file. Idempotent — safe to run repeatedly.

## Target User

Developers who:
- Work across multiple repositories regularly
- Want reproducible, documented workspace setups
- Prefer CLI tools over GUI
- Value config-as-code approaches

## Success Criteria

- Create a multi-repo workspace in < 30 seconds (after initial clone)
- State file is human-readable and editable without docs
- Single binary, no runtime dependencies
- Works on macOS and Linux
