# flow exec

Run a command from the workspace directory.

![flow exec](exec.gif)

## Usage

```
flow exec <workspace> -- <command>
```

## Examples

```bash
flow exec vpc-ipv6 -- cursor .       # open in Cursor
flow exec vpc-ipv6 -- claude          # start Claude Code
flow exec calm-delta -- git status    # by ID
flow exec vpc-ipv6 -- ls -la
```
