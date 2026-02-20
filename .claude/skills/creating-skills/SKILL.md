---
name: creating-skills
description: How to create Claude Code skills for this project. Use when adding new skills, creating SKILL.md files, or setting up skill directories.
disable-model-invocation: true
---

# Creating Skills

Skills are stored in `.claude/skills/<skill-name>/SKILL.md` and extend what Claude can do in this project.

## Directory Structure

Each skill is a directory with `SKILL.md` as the entrypoint:

```
.claude/skills/<skill-name>/
├── SKILL.md           # Main instructions (required)
├── reference.md       # Detailed docs (optional, loaded when needed)
├── examples.md        # Example output (optional)
└── scripts/
    └── helper.sh      # Script Claude can execute (optional)
```

## SKILL.md Format

Every `SKILL.md` has two parts: YAML frontmatter and markdown content.

```yaml
---
name: my-skill
description: What this skill does and when to use it
---

Your instructions here...
```

## Frontmatter Fields

All fields are optional. Only `description` is recommended.

| Field                      | Default     | Description |
|----------------------------|-------------|-------------|
| `name`                     | dir name    | Display name. Lowercase, numbers, hyphens only (max 64 chars). Becomes the `/slash-command`. |
| `description`              | 1st paragraph | What the skill does and when to use it. Claude uses this to decide when to load it. |
| `argument-hint`            | —           | Hint for autocomplete, e.g. `[issue-number]`. |
| `disable-model-invocation` | `false`     | `true` = only user can invoke via `/name`. Prevents Claude from auto-loading. |
| `user-invocable`           | `true`      | `false` = hidden from `/` menu. Only Claude can invoke. Use for background knowledge. |
| `allowed-tools`            | —           | Tools Claude can use without asking permission, e.g. `Read, Grep, Glob`. |
| `model`                    | —           | Model override when skill is active. |
| `context`                  | —           | `fork` = run in isolated subagent context. |
| `agent`                    | —           | Subagent type when `context: fork`. Options: `Explore`, `Plan`, `general-purpose`, or custom. |

## Invocation Control

| Frontmatter                      | User can invoke | Claude can invoke |
|----------------------------------|-----------------|-------------------|
| (default)                        | Yes             | Yes               |
| `disable-model-invocation: true` | Yes             | No                |
| `user-invocable: false`          | No              | Yes               |

## String Substitutions

Use these in skill content for dynamic values:

- `$ARGUMENTS` — all arguments passed when invoking
- `$ARGUMENTS[N]` or `$N` — specific argument by index (0-based)
- `${CLAUDE_SESSION_ID}` — current session ID

## Dynamic Context

The `!` backtick syntax runs shell commands before content is sent to Claude:

```markdown
Current branch: !`git branch --show-current`
Recent commits: !`git log --oneline -5`
```

## Types of Skills

**Reference skills** — background knowledge Claude applies automatically:

```yaml
---
name: api-conventions
description: API design patterns for this codebase
user-invocable: false
---

When writing API endpoints...
```

**Task skills** — actions invoked with `/skill-name`:

```yaml
---
name: deploy
description: Deploy the application
disable-model-invocation: true
---

Deploy $ARGUMENTS to production:
1. Run tests
2. Build
3. Push
```

## Supporting Files

Keep `SKILL.md` under 500 lines. Move detailed reference material to separate files and reference them:

```markdown
## Additional resources

- For complete API details, see [reference.md](reference.md)
- For usage examples, see [examples.md](examples.md)
```

## Checklist for New Skills

1. Create directory: `mkdir -p .claude/skills/<skill-name>`
2. Write `SKILL.md` with frontmatter and instructions
3. Choose invocation model: both (default), user-only (`disable-model-invocation: true`), or claude-only (`user-invocable: false`)
4. Add supporting files if content exceeds 500 lines
5. Test: invoke with `/skill-name` or ask Claude something that matches the description
