## flow edit status

Open status spec file in editor

![flow edit_status](tapes/edit_status.gif)


### Synopsis

Open a status spec file in your editor.

Without arguments, opens the global status spec (~/.flow/status.yaml).
With a workspace argument, opens the workspace-specific status spec.

```
flow edit status [workspace] [flags]
```

### Examples

```
  flow edit status             # Edit global status spec
  flow edit status vpc-ipv6     # Edit workspace status spec
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose debug output
```

### SEE ALSO

* [flow edit](flow_edit.md)	 - Open flow configuration files in editor

