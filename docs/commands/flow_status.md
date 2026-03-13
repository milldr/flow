## flow status

Show workspace status

![flow status](tapes/status.gif)


### Synopsis

Show the resolved status of workspaces.

Without arguments, shows all workspaces with their statuses.
Archived workspaces are hidden by default; use --all to include them.
With a workspace argument, shows a detailed per-repo status breakdown.

```
flow status [workspace] [flags]
```

### Examples

```
  flow status                  # Show all workspace statuses
  flow status --all             # Include archived workspaces
  flow status vpc-ipv6          # Show per-repo breakdown
```

### Options

```
  -a, --all    Include archived workspaces
  -h, --help   help for status
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow](flow.md)	 - Multi-repo workspace manager using git worktrees

