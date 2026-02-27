## flow

Multi-repo workspace manager using git worktrees

### Synopsis

Flow manages multi-repo development workspaces using git worktrees.
A YAML manifest defines which repos/branches belong together, and `flow render` materializes the workspace.

### Options

```
  -h, --help      help for flow
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow delete](flow_delete.md)	 - Delete one or more workspaces and their worktrees
* [flow edit](flow_edit.md)	 - Open flow configuration files in editor
* [flow exec](flow_exec.md)	 - Run a command from the workspace directory
* [flow init](flow_init.md)	 - Create a new empty workspace
* [flow list](flow_list.md)	 - List all workspaces
* [flow open](flow_open.md)	 - Open a shell in the workspace directory
* [flow render](flow_render.md)	 - Create worktrees from workspace state file
* [flow reset](flow_reset.md)	 - Reset a config file to its default value
* [flow status](flow_status.md)	 - Show workspace status
* [flow version](flow_version.md)	 - Print the version

