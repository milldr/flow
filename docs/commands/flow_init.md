## flow init

Create a new empty workspace

![flow init](tapes/init.gif)


### Synopsis

Create a new empty workspace with a generated ID.
Optionally provide a human-friendly name as the first argument.
By default, launches the configured agent in the new workspace.

```
flow init [name] [flags]
```

### Examples

```
  flow init
  flow init my-project
  flow init my-project --no-exec
```

### Options

```
  -h, --help      help for init
      --no-exec   skip launching the agent after creation
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow](flow.md)	 - Multi-repo workspace manager using git worktrees

