# flow delete

Delete a workspace and its worktrees. Prompts for confirmation unless `--force` is passed.

![flow delete](delete.gif)

## Usage

```
flow delete <workspace>
```

## Examples

```bash
flow delete vpc-ipv6               # by name (prompts for confirmation)
flow delete calm-delta             # by ID
flow delete vpc-ipv6 --force       # skip confirmation
```

## Options

| Flag | Long      | Description              |
| ---- | --------- | ------------------------ |
| `-f` | `--force` | Skip confirmation prompt |
