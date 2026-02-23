# PRD-007: `flow edit` Command

**Depends on:** [002-mvp.md](./002-mvp.md), [006-status.md](./006-status.md)

## Summary

Introduce `flow edit` as a parent command for opening flow configuration files in `$EDITOR`. Migrate the existing `flow state` command to `flow edit state` and add `flow edit config` for the global config file.

## Motivation

Flow currently has `flow state <workspace>` to open a workspace's state file in an editor. PRD-006 adds `flow edit status` for status specs. Rather than scattering editor commands across unrelated top-level verbs, consolidate them under a single `flow edit` parent command. This gives a consistent, discoverable pattern: `flow edit <thing>`.

## Commands

| Command | Description |
|---------|-------------|
| `flow edit state <workspace>` | Open workspace state file in editor |
| `flow edit config` | Open global config file in editor |
| `flow edit status` | Defined in PRD-006 |
| `flow edit status <workspace>` | Defined in PRD-006 |

---

### `flow edit state <workspace>`

Open the workspace's `state.yaml` in `$EDITOR`. The `<workspace>` argument accepts either an ID or name, resolved via `resolveWorkspace`. Identical behavior to the current `flow state` command.

**Example:**
```bash
flow edit state calm-delta    # Opens ~/.flow/workspaces/<id>/state.yaml in $EDITOR
```

---

### `flow edit config`

Open the global config file (`~/.flow/config.yaml`) in `$EDITOR`.

**Example:**
```bash
flow edit config    # Opens ~/.flow/config.yaml in $EDITOR
```

---

## Migration

**Breaking change:** `flow state` is removed. Use `flow edit state` instead. This is acceptable since flow is pre-release.

## Acceptance Criteria

- [ ] `flow edit state <workspace>` opens state file in editor
- [ ] `flow edit config` opens global config in editor
- [ ] `flow state` is removed (no alias)
- [ ] `flow edit` with no subcommand prints help listing available subcommands
- [ ] `make build`, `make lint`, `make test` all pass
