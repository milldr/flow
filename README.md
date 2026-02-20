# :ocean: flow

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/milldr)

CLI for managing multi-repo development workspaces using git worktrees.

Working across multiple repos means repetitive setup, scattered branches, and cleanup debt. Flow uses a YAML state file to define which repos and branches belong together, then materializes the workspace with bare clone caching and git worktrees.

![flow demo](demo.gif)

## Quickstart

### 1. Install

```bash
brew install milldr/tap/flow
```

Or with Go:

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

## State file

Each workspace is defined by a `state.yaml` file. Run `flow state <workspace>` to open it in your editor.

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

| Field | Description |
|-------|-------------|
| `metadata.name` | Optional human-friendly name |
| `metadata.description` | Optional description |
| `spec.repos[].url` | Git remote URL |
| `spec.repos[].branch` | Branch to check out |
| `spec.repos[].path` | Directory name in the workspace (defaults to repo name) |

See the full [command reference](docs/commands/) for usage, flags, examples, and GIF demos.

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
