## flow exec

Run a command from the workspace directory

```
flow exec <workspace> -- <command> [flags]
```

### Examples

```
  flow exec calm-delta -- cursor .
  flow exec calm-delta -- git status
  flow exec calm-delta -- ls -la
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow](flow.md)	 - Multi-repo workspace manager using git worktrees

