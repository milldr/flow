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

Every PR must have exactly one semver label. Use `gh pr create --label <label>`.

| Label | When to use |
|-------|-------------|
| `no-release` | Docs, CI, refactors — no version bump |
| `patch` | Bug fixes |
| `minor` | New features, enhancements |
| `major` | Breaking changes |

## Example

```bash
gh pr create \
  --title "Add workspace clone command" \
  --label minor \
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
