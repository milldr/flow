# Commands

Each command has a reference page with usage, examples, and a GIF demo.

| Command | Description |
|---------|-------------|
| [init](init.md) | Create a new workspace |
| [list](list.md) | List all workspaces |
| [state](state.md) | Open workspace state file in editor |
| [render](render.md) | Materialize workspace worktrees |
| [exec](exec.md) | Run a command from the workspace directory |
| [delete](delete.md) | Delete a workspace and its worktrees |

## Regenerating GIFs

```bash
make docs
```

GIF recordings are generated with [VHS](https://github.com/charmbracelet/vhs) from `.tape` files in [tapes/](tapes/). Output GIFs go there too.
