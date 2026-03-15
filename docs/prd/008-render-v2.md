# PRD-008: Render v2

**Depends on:** [002-mvp.md](./002-mvp.md), [005-config-and-claude.md](./005-config-and-claude.md)

## Summary

Overhaul `flow render` to make reset-from-base the default behavior, improve idempotency for already-rendered repos, support the `base` field in state, and fix the interactive prompt failure when running under non-TTY contexts (e.g. Claude Code).

## Motivation

The current render command has several UX gaps exposed by real-world usage:

1. **Default behavior requires a TTY.** When no `--reset` flag is passed, render prompts interactively. This fails inside Claude Code (`open /dev/tty: device not configured`), which is the primary way users interact with flow. The generated Claude skills tell the agent to run `flow render`, but don't mention the `--reset` flag — so the agent hits the TTY error every time.

2. **Re-render touches existing repos unnecessarily.** Running `flow render` a second time (e.g. after adding a new repo to state) resets or fast-forwards all existing worktrees. Users expect additive behavior: add a repo to state, re-render, and only the new repo is materialized.

3. **Reset semantics are unclear.** `--reset` creates a branch from the repo's default branch (e.g. `main`), but users often want to branch from a specific base like `staging`. The `base` field exists in the state schema but the skills and docs don't mention it, so agents and users don't know to use it.

4. **`--reset=false` silently succeeds when the branch doesn't exist.** If the user specifies a branch that doesn't exist in the remote and passes `--reset=false`, the behavior is confusing. It should fail clearly.

### Use case that doesn't work today

> "Set up a flow workspace for atmos-pro on the `qa-1` branch, based on `staging`"

The agent writes the correct state file with `branch: qa-1` and `base: staging`, then runs `flow render`. Without `--reset`, render hits a TTY prompt and crashes. The agent guesses flags (`--non-interactive`, `--yes`, `-y`) before the user tells it to use `--reset`. The generated skills should have made this seamless.

## Changes

### 1. Remove the interactive prompt — make `--reset` default true with no prompt

**Current behavior (three modes):**

| Flag | Behavior |
|------|----------|
| `--reset` | Reset branches from base |
| `--reset=false` | Use existing remote branches |
| *(no flag)* | Prompt user interactively |

**New behavior (two modes):**

| Flag | Behavior |
|------|----------|
| `--reset` *(default, or no flag)* | Reset branches from base |
| `--reset=false` | Use existing remote branches |

The prompt mode is removed entirely. `flow render <workspace>` without flags behaves identically to `flow render <workspace> --reset`.

### 2. Skip already-rendered repos on re-render

When `flow render` is run on a workspace that has already been rendered, repos whose worktrees already exist should be **skipped entirely** — no fetch, no reset, no fast-forward. Only repos that are new (no worktree exists) should be processed.

This makes render additive: users can add repos to `state.yaml` and re-render without affecting existing checkouts.

**Status messages:**

| Scenario | Message |
|----------|---------|
| New repo, reset | `repo (branch, reset from base) ✓` |
| New repo, new branch | `repo (branch, new branch from base) ✓` |
| New repo, existing remote branch | `repo (branch) ✓` |
| Already rendered | `repo (branch) exists, skipped` |

To force a full re-render of all repos (including already-rendered ones), users should delete the workspace and re-create it, or manually remove the worktree directory before re-rendering.

### 3. Reset creates a clean branch from the upstream base

When `--reset` is active (the default), render should:

1. Fetch the bare repo to get latest remote state
2. Resolve the base branch: use `repo.base` from state if set, otherwise the repo's default branch
3. Delete the local branch if it already exists
4. Create a new worktree with a fresh branch starting from `origin/{base}`

This ensures the new branch always starts from the **latest remote state** of the base branch — not a stale local ref.

### 4. `--reset=false` requires the branch to exist

When `--reset=false` is passed:

- If the branch exists in the remote: check it out as a worktree, done
- If the branch does **not** exist in the remote: fail with a clear error message

**Error message:**
```
✗ Branch "qa-1" does not exist in remote for repo "apps"
  Hint: remove --reset=false to create a new branch, or push the branch to the remote first
```

The `base` field is irrelevant when `--reset=false` — we're using an existing branch, not creating one.

### 5. Update generated Claude skills

Update the flow-cli skill (`agents/defaults/claude/skills/flow-cli.md`) to:

- Document the `base` field in the state schema
- Show `flow render` examples that reflect the new defaults (no need for `--reset` flag in the common case)
- Remove references to interactive prompting
- Make it clear that render is additive for already-rendered repos

## Files

| File | Action |
|------|--------|
| `docs/prd/008-render-v2.md` | Created — this PRD |
| `docs/prd/README.md` | Modified — add PRD-008 row |
| `internal/cmd/render.go` | Modified — remove prompt mode, simplify flag handling |
| `internal/workspace/workspace.go` | Modified — skip already-rendered repos, error on missing branch with `--reset=false` |
| `internal/workspace/workspace_test.go` | Modified — update tests for new behavior |
| `internal/agents/defaults/claude/skills/flow-cli.md` | Modified — document `base` field, update render examples |

## Design Decisions

- **No prompt mode** — flow is primarily used via AI agents that don't have TTY access. The prompt added friction in the most common path. Two explicit modes (`--reset` and `--reset=false`) are sufficient and unambiguous.
- **Skip, don't update, existing worktrees** — the previous "idempotent update" behavior (fetch + reset to remote) was surprising. Users re-render to add repos, not to update existing ones. If they want to update a worktree, they can `git pull` inside it like any normal repo.
- **Fail on missing branch with `--reset=false`** — silent fallback to creating a branch defeats the purpose of `--reset=false`, which explicitly means "use what exists." A clear error with a hint is more helpful.
- **`base` field already exists in schema** — no schema changes needed, just documentation and skill updates.

## Acceptance Criteria

- [ ] `flow render <workspace>` without flags resets branches from base (no prompt)
- [ ] `flow render <workspace> --reset=false` uses existing remote branches
- [ ] `flow render <workspace> --reset=false` errors if branch doesn't exist in remote
- [ ] Re-render skips repos whose worktrees already exist
- [ ] Re-render processes only newly added repos
- [ ] Reset creates branch from `origin/{base}` where `base` is `repo.base` or default branch
- [ ] Claude skills document the `base` field and new render defaults
- [ ] No TTY required for any render path
- [ ] `make build`, `make lint`, `make test` all pass
