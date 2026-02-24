## flow reset state

Reset a workspace state file to its default

![flow reset_state](tapes/reset_state.gif)


```
flow reset state <workspace> [flags]
```

### Examples

```
  flow reset state vpc-ipv6
  flow reset state vpc-ipv6 --force
```

### Options

```
  -f, --force   Skip confirmation prompt
  -h, --help    help for state
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow reset](flow_reset.md)	 - Reset a config file to its default value

