# :ocean: flow

> [!WARNING]
> This project is still in draft and under active development. APIs and commands may change.

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
✅ Created workspace: calm-delta
   Location: ~/.flow/workspaces/calm-delta/

Next steps:
  flow state calm-delta     # Add repos to state file
  flow render calm-delta    # Create worktrees
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
📦 Rendering workspace: calm-delta

  [1/2] github.com/acme/vpc-service
        └── Worktree: ./vpc-service (feature/ipv6) ✓
  [2/2] github.com/acme/subnet-manager
        └── Worktree: ./subnet-manager (feature/ipv6) ✓

✅ Workspace ready

  flow exec calm-delta -- cursor .   # Open in editor
```

Flow fetches each repo into a bare clone cache (`~/.flow/repos/`), then creates lightweight worktrees in the workspace directory. Running `render` again is idempotent — it fetches updates and skips worktrees that already exist.

## Commands

### `flow init [name]`

Create a new empty workspace. If no name is given, one is generated automatically (e.g. `calm-delta`, `bold-creek`).

```bash
flow init              # generated name
flow init vpc-ipv6     # explicit name
```

### `flow list`

List all workspaces.

```bash
flow list
```

```
NAME        DESCRIPTION                       REPOS   CREATED
vpc-ipv6    IPv6 support across VPC services   3      2h ago
auth-v2     Auth service rewrite               2      5d ago
```

**Aliases:** `flow ls`

### `flow render <name>`

Materialize the workspace. Clones repos (or fetches if cached), creates worktrees for each repo/branch pair.

```bash
flow render vpc-ipv6
```

Idempotent — safe to run repeatedly. Existing worktrees are skipped, cached repos are fetched.

### `flow state <name>`

Open the workspace state file in your editor (`$EDITOR` or `vim`).

```bash
flow state vpc-ipv6
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

### `flow exec <name> -- <command>`

Run a command from the workspace directory.

```bash
flow exec vpc-ipv6 -- cursor .
flow exec vpc-ipv6 -- git status
flow exec vpc-ipv6 -- ls -la
```

### `flow delete <name>`

Delete a workspace and its worktrees. Prompts for confirmation unless `--force` is passed.

```bash
flow delete vpc-ipv6
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
│   └── vpc-ipv6/
│       ├── state.yaml          # Workspace manifest
│       ├── vpc-service/        # Worktree
│       └── subnet-manager/     # Worktree
└── cache/
    ├── acme-vpc-service.git/       # Bare clone
    └── acme-subnet-manager.git/    # Bare clone
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

## License

MIT
