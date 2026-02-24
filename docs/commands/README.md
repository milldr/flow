# Commands

Auto-generated CLI reference with GIF demos. See [flow.md](flow.md) for the command index.

| Command | Description |
|---------|-------------|
| [flow init](flow_init.md) | Create a new workspace |
| [flow list](flow_list.md) | List all workspaces |
| [flow edit](flow_edit.md) | Edit flow configuration files |
| [flow render](flow_render.md) | Create worktrees from workspace state file |
| [flow exec](flow_exec.md) | Run a command from the workspace directory |
| [flow status](flow_status.md) | Show workspace status |
| [flow reset](flow_reset.md) | Reset a config file to its default value |
| [flow delete](flow_delete.md) | Delete a workspace and its worktrees |
| [flow version](flow_version.md) | Print the version |

## Regenerating

```bash
make docs
```

- Reference docs are auto-generated from cobra command definitions (`make gendocs`)
- GIF recordings are generated with [VHS](https://github.com/charmbracelet/vhs) from `.tape` files in [tapes/](tapes/)
