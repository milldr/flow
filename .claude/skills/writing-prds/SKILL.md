---
name: writing-prds
description: How to write PRDs for this project. Use when creating a new product requirement document, planning a feature, or adding to docs/prd/.
---

# Writing PRDs

PRDs live in `docs/prd/` and follow a consistent format. Each new feature gets a PRD before implementation.

PRDs capture a decision at a point in time. As the project evolves, implementations may diverge from the original PRD — that's expected. Don't update old PRDs to match current code. Instead, write a new PRD when the direction changes.

## File Naming

PRDs are numbered sequentially: `001-vision.md`, `002-mvp.md`, `003-prompt.md`, etc. Use the next available number.

## Required Sections

Follow the structure of existing PRDs (read one from `docs/prd/` for reference). At minimum:

```markdown
# PRD-NNN: Title

**Depends on:** [NNN-name.md](./NNN-name.md)

## Summary
One paragraph explaining what this change does.

## Motivation
Why this change is needed. What problem it solves.

## Changes
### Subsection per change area
Details of what changes, with examples and diagrams where helpful.

## Files
| File | Action |
|------|--------|
| `path/to/file.go` | Created — description |
| `path/to/other.go` | Modified — description |

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] `make build`, `make lint`, `make test` all pass
```

## After Creating

Update `docs/prd/README.md` to add a row to the index table:

```markdown
| [NNN](NNN-title.md) | Title |
```

## Conventions

- **Depends on** links to the prerequisite PRD
- **Files table** lists every file that will be created or modified
- **Acceptance criteria** uses checkboxes — checked off as implementation progresses
- Always include `make build`, `make lint`, `make test` as a criterion
- Keep the summary concise — motivation section has the detail
