## flow archive

Archive a workspace (remove worktrees, keep state)

### Synopsis

Archive a workspace by removing its worktrees to free up branches,
while preserving the state file. Archived workspaces are hidden from
flow status by default (use --all to see them).

Use --closed to archive all workspaces with "closed" status at once.

```
flow archive [workspace] [flags]
```

### Examples

```
  flow archive my-workspace    # Archive a single workspace
  flow archive --closed        # Archive all closed workspaces
```

### Options

```
      --closed   Archive all workspaces with closed status
  -f, --force    Skip confirmation prompt
  -h, --help     help for archive
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow](flow.md)	 - Multi-repo workspace manager using git worktrees

