# :ocean: flow

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/milldr)

CLI for managing multi-repo development workspaces using git worktrees.

Working across multiple repos means repetitive setup, scattered branches, and cleanup debt. Flow uses a YAML state file to define which repos and branches belong together, then materializes the workspace with bare clone caching and git worktrees.

![flow demo](demo.gif)

## Motivation

Flow is designed to be called by other AI agents — tools like [OpenClaw](https://github.com/openclaw) that integrate with Slack, Linear, or other services to understand context and then programmatically create state files and initialize workspaces. Prompt OpenClaw with a Slack thread or Linear ticket and let your agent create the workspace using Flow 🌊

Agents should call deterministic tools rather than relying on freeform interpretation with skills. This leads to more consistent results and reproducible environments.

## Quickstart

<details>
<summary><strong>Install</strong></summary>

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

</details>

<details>
<summary><strong>Usage</strong></summary>

### 1. Create a workspace

Flow creates a workspace with an empty state file.

```bash
flow init
```

### 2. Add repos

Open the state file in `$EDITOR` and define which repos and branches belong together.

```bash
flow edit state calm-delta
```

### 3. Render it

Flow fetches each repo into a shared bare clone cache, then creates lightweight worktrees in the workspace directory. Rendering is idempotent — re-running fetches updates and skips worktrees that already exist.

```bash
flow render calm-delta
```

### 4. Start working

Launch your default agent directly into the workspace. Flow reads the `spec.agents` list from your global config and runs the one marked `default: true`.

```bash
flow exec calm-delta
```

See the [spec reference](docs/specs/) for YAML file schemas and the [command reference](docs/commands/) for all commands.

</details>

## What it does

🌳 **Workspaces as code** — Declare git worktrees in a YAML state file for instant, reproducible workspace setup across repos and branches.

🚦 **Status tracking** — Define custom check commands that dynamically resolve the status of each repo in a workspace.

🤖 **AI agent integration** — Generate shared context files and agent instructions across repos so your AI tools have the right skills and knowledge from the start.

## How it works

Flow stores everything under `~/.flow` (override with `$FLOW_HOME`):

```
~/.flow/
├── config.yaml                         # Global config
├── status.yaml                         # Global status spec
├── agents/
│   └── claude/
│       ├── CLAUDE.md                   # Shared agent instructions
│       └── skills/
│           ├── flow-cli/SKILL.md       # Flow CLI skill
│           └── workspace-structure/SKILL.md
├── workspaces/
│   └── calm-delta/                     # Workspace ID
│       ├── state.yaml                  # Workspace manifest (name: vpc-ipv6)
│       ├── status.yaml                 # Optional workspace-specific status spec
│       ├── CLAUDE.md                   # Generated workspace context
│       ├── .claude/
│       │   ├── CLAUDE.md → agents/claude/CLAUDE.md
│       │   └── skills → agents/claude/skills/
│       ├── vpc-service/                # Worktree
│       └── subnet-manager/             # Worktree
└── repos/
    └── github.com/acme/
        ├── vpc-service.git/            # Bare clone
        └── subnet-manager.git/         # Bare clone
```

Bare clones are shared across workspaces. Worktrees are cheap — they share the object store with the bare clone, so multiple workspaces pointing at the same repo don't duplicate data.

See the [spec reference](docs/specs/) for YAML file schemas and the [command reference](docs/commands/) for usage, flags, and GIF demos.

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
