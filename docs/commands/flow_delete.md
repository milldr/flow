## flow delete

Delete one or more workspaces and their worktrees

![flow delete](tapes/delete.gif)


```
flow delete <workspace> [workspace...] [flags]
```

### Examples

```
  flow delete calm-delta
  flow delete calm-delta warm-brook --force
  flow delete ws1 ws2 ws3
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for delete
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow](flow.md)	 - Multi-repo workspace manager using git worktrees

