---
name: creating-prs
description: How to create pull requests for this project. Use when opening a PR on GitHub.
---

# Creating PRs

## PR Body Format

```markdown
## What
One-line summary of the change.

## Why
Brief motivation — what problem this solves or what it enables.

## Ref
- Related PRD, issue, or link (if any)
```

## Labels

Every PR needs a label for release-drafter. Use `gh pr create --label <label>` (repeatable) or `gh pr edit --add-label <label>`.

**Semver labels** (determines version bump):

| Label | Bump |
|-------|------|
| `major` | Major |
| `minor`, `feature`, `enhancement` | Minor |
| `patch`, `fix`, `bugfix`, `bug` | Patch |

Default bump (no semver label): **minor**.

**Category labels** (determines changelog section):

| Label | Changelog section |
|-------|-------------------|
| `feature`, `enhancement` | New Features |
| `fix`, `bugfix`, `bug` | Bug Fixes |
| `docs`, `documentation` | Documentation |

## Example

```bash
gh pr create \
  --title "Add workspace clone command" \
  --label feature \
  --body "$(cat <<'EOF'
## What
Adds `flow clone` to clone a workspace to a new ID.

## Why
Users want to duplicate workspace setups without re-editing state files.

## Ref
- [PRD-006](docs/prd/006-clone.md)
EOF
)"
```
