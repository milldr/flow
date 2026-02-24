# PRD-006: Workspace Status

**Depends on:** [002-mvp.md](./002-mvp.md)

## Summary

Add a `flow status` command that dynamically resolves the current status of each workspace by running user-defined check commands. Statuses (e.g., "in-progress", "in-review", "closed") are not hardcoded — they are defined in a spec file along with the shell commands that determine when each status applies.

## Motivation

Flow manages the lifecycle of multi-repo workspaces, but currently has no concept of where a workspace is in the development workflow. Users manually track whether their work is in progress, under review, merged, etc. This is exactly the kind of state that can be derived automatically from external systems (GitHub PRs, CI status, deployment state).

By letting users define their own statuses and the shell commands that resolve them, flow becomes workflow-aware without being workflow-opinionated. A user working with GitHub PRs defines checks against `gh`. A user with a different workflow defines different checks. Flow just runs the commands and reports the results.

## Concepts

### Status Spec

A YAML file that defines the available statuses and how to check for them.

**Precedence:** The global spec at `~/.flow/status.yaml` serves as the default for all workspaces. A workspace can override this by placing a `status.yaml` in its workspace directory (`~/.flow/workspaces/<id>/status.yaml`). When a workspace-level spec exists, it **fully replaces** the global spec for that workspace — there is no merging.

```yaml
apiVersion: flow/v1
kind: Status
statuses:
  - name: string       # Required. Status identifier (used in output).
    description: string # Optional. Human-readable description.
    check: string      # Required (unless default). Shell command, exit 0 = match.
    default: bool      # Optional. Fallback if no checks match. Only one allowed.
```

**Rules:**
- At least one status must have `default: true`
- A default status does not need a `check` field
- Status names must be unique
- Statuses are evaluated top-to-bottom (first match wins per repo)

**Example:**
```yaml
apiVersion: flow/v1
kind: Status
statuses:
  - name: closed
    description: PR merged
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state merged --json number | jq -e 'length > 0' > /dev/null 2>&1

  - name: in-review
    description: PR open
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length > 0' > /dev/null 2>&1

  - name: in-progress
    description: Active development
    default: true
```

### Resolution Model

Statuses are evaluated **per-repo** using the check commands. Each check is a shell command executed via `sh -c` with environment variables providing context:

| Variable | Value |
|----------|-------|
| `FLOW_REPO_URL` | Repo URL from state (e.g., `github.com/org/repo`) |
| `FLOW_REPO_BRANCH` | Branch name |
| `FLOW_REPO_PATH` | Worktree path |
| `FLOW_WORKSPACE_ID` | Workspace ID |
| `FLOW_WORKSPACE_NAME` | Workspace name |

**Per-repo resolution:** Statuses are checked in spec order (top to bottom). The first check that exits 0 determines the repo's status. If no check matches, the `default: true` status is used.

**Workspace-level aggregation:** The workspace status is the **least advanced** status across all its repos. "Least advanced" means latest in spec order. If repo A is "closed" (index 0) and repo B is "in-review" (index 1), the workspace status is "in-review" — the workspace isn't done until all repos are done.

### Why First-Match-Wins

The spec order matters. List the most advanced/specific statuses first, with the default fallback last. This is analogous to pattern matching or firewall rules — intuitive for users who work with config-as-code tools.

## Commands

### `flow status`

Show all workspaces with their resolved statuses.

**Behavior:**
1. List all workspaces
2. For each workspace, load its status spec (workspace-level if present, otherwise global)
3. For each workspace, resolve status (run checks per-repo, aggregate)
4. Display table

**Output:**
```
  Resolving workspace statuses...

NAME           STATUS        REPOS   CREATED
vpc-ipv6       in-review     2       2h ago
webhook-fix    in-progress   1       3d ago
auth-refactor  closed        3       5d ago
```

### `flow status <workspace>`

Show detailed per-repo status breakdown for a single workspace. The `<workspace>` argument accepts either an ID or name, resolved via `resolveWorkspace`.

**Behavior:**
1. Resolve workspace by ID or name
2. Load status spec (workspace-level if present, otherwise global)
3. Resolve status for each repo
4. Display workspace status and per-repo breakdown

**Output:**
```
  Resolving status for vpc-ipv6...

Status: in-review

  REPO                             BRANCH              STATUS
  github.com/cloudposse/atmos      feat/ipv6-support   closed
  github.com/cloudposse/docs       feat/ipv6-support   in-review
```

### `flow status --init`

Generate a starter status spec file at `~/.flow/status.yaml` with a commented example. Opens in `$EDITOR` after creation.

**Output:**
```
✓ Created status spec at ~/.flow/status.yaml
```

### `flow edit status`

Open the global status spec (`~/.flow/status.yaml`) in `$EDITOR`. If the file doesn't exist, create a starter spec first (same as `--init`).

**Example:**
```bash
flow edit status    # Opens ~/.flow/status.yaml in $EDITOR
```

### `flow edit status <workspace>`

Open a workspace-specific status spec in `$EDITOR`. The `<workspace>` argument accepts either an ID or name, resolved via `resolveWorkspace`. If the file doesn't exist, create it by copying the global spec as a starting point.

**Example:**
```bash
flow edit status vpc-ipv6    # Opens ~/.flow/workspaces/<id>/status.yaml in $EDITOR
```

## Execution Details

**Concurrency:** Checks run concurrently across repos within a workspace and across workspaces. Each check has a timeout (default: 10 seconds) to prevent hanging on network calls.

**Error handling:** A check that exits non-zero means "this status doesn't match" — there's no distinction between "check failed" and "doesn't apply." If all non-default checks fail for a repo, the default status is used. This is simple and predictable.

**Spec lookup:** For each workspace, check for `~/.flow/workspaces/<id>/status.yaml` first. If not found, fall back to `~/.flow/status.yaml`. If neither exists, print a message suggesting `flow status --init` and exit.

**Spinner:** Since checks may involve network calls, show a spinner while resolving. Use the existing `ui.RunWithSpinner` pattern.

## Directory Structure

```
~/.flow/
├── config.yaml
├── status.yaml              # NEW — global default status spec
├── agents/
├── repos/
└── workspaces/
    └── <id>/
        ├── state.yaml
        ├── status.yaml      # NEW — optional workspace-level override
        └── ...
```

## Example Workflows

### Basic PR workflow
```yaml
# ~/.flow/status.yaml
apiVersion: flow/v1
kind: Status
statuses:
  - name: closed
    description: PR merged
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state merged --json number | jq -e 'length > 0' > /dev/null 2>&1

  - name: in-review
    description: PR open
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length > 0' > /dev/null 2>&1

  - name: in-progress
    description: Active development
    default: true
```

```bash
flow status
# NAME           STATUS        REPOS   CREATED
# vpc-ipv6       in-review     2       2h ago
# webhook-fix    in-progress   1       3d ago

flow status vpc-ipv6
# Status: in-review
#
#   REPO                             BRANCH              STATUS
#   github.com/cloudposse/atmos      feat/ipv6-support   closed
#   github.com/cloudposse/docs       feat/ipv6-support   in-review
```

### CI/CD pipeline workflow
```yaml
apiVersion: flow/v1
kind: Status
statuses:
  - name: deployed
    description: Changes live in production
    check: gh run list --repo "https://$FLOW_REPO_URL" --branch "$FLOW_REPO_BRANCH" --workflow deploy --status completed --json conclusion -q '.[0].conclusion == "success"' | grep -q true

  - name: approved
    description: PR approved, awaiting merge
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --json reviewDecision -q '.[0].reviewDecision == "APPROVED"' | grep -q true

  - name: in-review
    description: PR open
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length > 0' > /dev/null 2>&1

  - name: draft
    description: Draft PR
    check: gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --json isDraft -q '.[0].isDraft' | grep -q true

  - name: in-progress
    description: Active development
    default: true
```

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Spec location | Global default + per-workspace override | Global spec covers the common case; workspace-level `status.yaml` fully replaces it when present |
| Check mechanism | Shell commands via `sh -c` | Maximum flexibility; works with any tool (`gh`, `curl`, custom scripts) |
| Context passing | Environment variables | Simpler than Go templates; natural for shell commands |
| Resolution order | First-match-wins, top-to-bottom | Familiar pattern (routing tables, pattern matching); most specific first |
| Aggregation | Least-advanced repo wins | Intuitive: workspace isn't done until all repos are done |
| Error model | Non-zero exit = doesn't match | Simple; no need to distinguish error types for MVP |
| Check timeout | 10 seconds per check | Prevents hanging on network issues |

## Open Questions

These are design choices to confirm before implementation:

1. **Should `flow list` also show status?** Option A: Keep `flow status` separate (fast `list` vs slower `status`). Option B: Add a `--status` flag to `flow list`. Recommendation: Keep separate for MVP — status checks involve network calls and add latency.

2. **Check timeout configurability?** Could add a `timeout` field per status or globally in the spec. Recommendation: Hardcode 10s for MVP, make configurable later.

## Acceptance Criteria

- [ ] `flow status --init` creates a starter `~/.flow/status.yaml` and opens in editor
- [ ] `flow status` resolves and displays statuses for all workspaces
- [ ] `flow status <workspace>` shows per-repo status breakdown
- [ ] Statuses are evaluated top-to-bottom, first match wins per repo
- [ ] Workspace status is the least-advanced repo status
- [ ] Check commands receive correct environment variables
- [ ] Checks run concurrently with a timeout
- [ ] Default status is used when no checks match
- [ ] Workspace-level `status.yaml` fully overrides the global spec for that workspace
- [ ] Missing spec file (no workspace-level or global) produces helpful error with `--init` suggestion
- [ ] `flow edit status` opens global status spec in editor (creates starter if missing)
- [ ] `flow edit status <workspace>` opens workspace-level status spec in editor (copies global as starting point if missing)
- [ ] Spec validation: requires unique names, exactly one default, valid structure
- [ ] `make build`, `make lint`, `make test` all pass
