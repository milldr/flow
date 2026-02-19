# :ocean: flow

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/milldr)

> [!IMPORTANT]
> This project is pre-release. APIs and commands may change.

CLI for managing multi-repo development workspaces using git worktrees.

Working across multiple repos means repetitive setup, scattered branches, and cleanup debt. Flow uses a YAML state file to define which repos and branches belong together, then materializes the workspace with bare clone caching and git worktrees.

![flow demo](demo.gif)

## Quickstart

### 1. Install

```bash
go install github.com/milldr/flow/cmd/flow@latest
```

Or build from source:

```bash
git clone https://github.com/milldr/flow.git
cd flow
make install
```

### 2. Create a workspace

```bash
flow init
```

```
✓ Created workspace calm-delta

  Next: flow state calm-delta
```

### 3. Add repos

```bash
flow state calm-delta     # Open state.yaml in $EDITOR
```

### 4. Render it

```bash
flow render calm-delta
```

```
✓ Workspace ready

  flow exec calm-delta -- <command>
  flow exec calm-delta -- cursor .
  flow exec calm-delta -- claude
```

Flow fetches each repo into a bare clone cache (`~/.flow/repos/`), then creates lightweight worktrees in the workspace directory. Running `render` again is idempotent — it fetches updates and skips worktrees that already exist.

## Commands

Every workspace has a generated **ID** (e.g. `calm-delta`) and an optional **name** (e.g. `vpc-ipv6`). All commands that take a `<workspace>` argument accept either the ID or the name. If a name matches multiple workspaces, an interactive selector is shown.

### `flow init [name]`

Create a new empty workspace. A unique ID is always generated. If a name is provided, it's associated with the workspace.

```bash
flow init              # ID only (e.g. calm-delta)
flow init vpc-ipv6     # ID + name
```

### `flow list`

List all workspaces.

```bash
flow list
```

```
ID             NAME       DESCRIPTION                        REPOS  CREATED
calm-delta     vpc-ipv6   IPv6 support across VPC services   3      2h ago
bold-creek     auth-v2    Auth service rewrite               2      5d ago
```

**Aliases:** `flow ls`

### `flow render <workspace>`

Materialize the workspace. Clones repos (or fetches if cached), creates worktrees for each repo/branch pair.

```bash
flow render vpc-ipv6       # by name
flow render calm-delta     # by ID
```

Idempotent — safe to run repeatedly. Existing worktrees are skipped, cached repos are fetched.

### `flow state <workspace>`

Open the workspace state file in your editor (`$EDITOR` or `vim`).

```bash
flow state vpc-ipv6        # by name
flow state calm-delta      # by ID
```

The state file is YAML:

```yaml
apiVersion: flow/v1
kind: State
metadata:
  name: vpc-ipv6
  description: IPv6 support across VPC services
  created: "2026-02-18T12:00:00Z"
spec:
  repos:
    - url: git@github.com:acme/vpc-service.git
      branch: feature/ipv6
      path: vpc-service
    - url: git@github.com:acme/subnet-manager.git
      branch: feature/ipv6
      path: subnet-manager
```

### `flow exec <workspace> -- <command>`

Run a command from the workspace directory.

```bash
flow exec vpc-ipv6 -- cursor .       # by name
flow exec calm-delta -- git status    # by ID
flow exec vpc-ipv6 -- ls -la
```

### `flow delete <workspace>`

Delete a workspace and its worktrees. Prompts for confirmation unless `--force` is passed.

```bash
flow delete vpc-ipv6       # by name
flow delete calm-delta     # by ID
```

**Options:**

| Flag | Long      | Description              |
| ---- | --------- | ------------------------ |
| `-f` | `--force` | Skip confirmation prompt |

### Global options

| Flag | Long        | Description                |
| ---- | ----------- | -------------------------- |
| `-v` | `--verbose` | Enable verbose debug output |

## How it works

Flow stores everything under `~/.flow` (override with `$FLOW_HOME`):

```
~/.flow/
├── workspaces/
│   └── calm-delta/                     # Workspace ID
│       ├── state.yaml                  # Workspace manifest (name: vpc-ipv6)
│       ├── vpc-service/                # Worktree
│       └── subnet-manager/             # Worktree
└── cache/
    ├── acme-vpc-service.git/           # Bare clone
    └── acme-subnet-manager.git/        # Bare clone
```

Bare clones are shared across workspaces. Worktrees are cheap — they share the object store with the bare clone, so multiple workspaces pointing at the same repo don't duplicate data.

## Requirements

- Go 1.25+
- Git 2.20+ (worktree support)

## Development

```bash
git clone https://github.com/milldr/flow.git
cd flow

# Build
make build

# Run tests
make test

# Lint
make lint

# Build release snapshot
make snapshot
```

## Support

I build and maintain projects like flow in my free time as personal hobbies. They're completely free and always will be. If you find this useful and want to show some support, feel free to buy me a coffee:

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/milldr)

## License

MIT
