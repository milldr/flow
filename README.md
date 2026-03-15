# :ocean: flow

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/milldr)

CLI for managing multi-repo development workspaces using git worktrees.

Working across multiple repos means repetitive setup, scattered branches, and cleanup debt. ${\color{teal}\textbf{flow}}$ uses a YAML state file to define which repos and branches belong together, then materializes the workspace with bare clone caching and git worktrees.

## Features

🌳 **Workspaces as code** — Declare repos and branches in a YAML state file. Repos are cloned once into a shared cache and checked out as lightweight git worktrees, so multiple workspaces pointing at the same repo don't duplicate data.

🚦 **Status tracking** — Define custom check commands that dynamically resolve the status of each repo in a workspace.

🤖 **AI agent integration** — Generate shared context files and agent instructions across repos so your AI tools have the right skills and knowledge from the start.

## Why ${\color{teal}\textbf{flow}}$?

AI agents work best when they have deterministic tools instead of freeform instructions. Asking an agent to "set up a multi-repo workspace" produces inconsistent results — but giving it a CLI that manages YAML state files, bare clone caches, and git worktrees produces the same result every time.

${\color{teal}\textbf{flow}}$ is that deterministic layer. It gives agents (and humans) a small set of reliable commands for workspace lifecycle: create, define repos in YAML, render worktrees, check status. Agents call these tools through embedded skills rather than interpreting setup instructions on their own.

Beyond workspace creation, ${\color{teal}\textbf{flow}}$ centralizes agent skills across all your workspaces and lets you check the status of many workstreams in parallel. It is not opinionated about which agent or editor you use — configure Claude, Cursor, or anything else in a single config file. Everything is customizable: status checks, agent skills, workspace structure.

### Describe the problem, ${\color{teal}\textbf{flow}}$ does the rest

`flow init` creates a workspace pre-configured with agent instructions and skills, then launches you directly into your preferred agent — Claude, Cursor, or anything else you configure. The agent reads the embedded flow skill to understand how workspaces work, plus any custom skills you've added (like resolving repos by friendly names). Describe what you're working on and the agent handles the rest: editing the state file, rendering worktrees, and getting everything ready to go.

![flow init](demo.gif)

### Track all your workstreams in one place

`flow status` gives you a live view across all active workspaces. Statuses are fully customizable — they're just one-line shell commands in a config file, not hardcoded logic. Have your AI generate complex status checks (query GitHub PRs, check CI, inspect git state) and hide them behind simple labels. Nothing in ${\color{teal}\textbf{flow}}$ is opinionated; anything that could be is configurable.

![flow status](demo-status.gif)

## Getting started

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

### Create a workspace and start working

${\color{teal}\textbf{flow}}$ creates a workspace and launches your configured agent. Describe what you're working on and the agent handles the rest.

```bash
flow init
```

The agent reads its embedded skills to edit `state.yaml`, run `flow render`, and begin working in the repos.

See the [spec reference](docs/specs/) for YAML file schemas and the [command reference](docs/commands/) for all commands.

## How it works

${\color{teal}\textbf{flow}}$ stores everything under `~/.flow` (override with `$FLOW_HOME`):

```
~/.flow/
├── config.yaml                         # Global config
├── status.yaml                         # Global status spec
├── agents/
│   └── claude/
│       ├── CLAUDE.md                   # Shared agent instructions
│       └── skills/
│           ├── flow/SKILL.md           # Built-in: workspace management
│           └── find-repo/SKILL.md      # Your own custom skill
├── workspaces/
│   └── calm-delta/                     # Workspace ID
│       ├── state.yaml                  # Workspace manifest (name: vpc-ipv6)
│       ├── status.yaml                 # Optional workspace-specific status spec
│       ├── CLAUDE.md                   # Generated workspace context
│       ├── .claude/
│       │   ├── CLAUDE.md → agents/claude/CLAUDE.md
│       │   └── skills/                # Consolidated from shared + repo skills
│       ├── vpc-service/                # Worktree
│       └── subnet-manager/             # Worktree
└── repos/
    └── github.com/acme/
        ├── vpc-service.git/            # Bare clone
        └── subnet-manager.git/         # Bare clone
```

Bare clones are shared across workspaces. Worktrees are cheap — they share the object store with the bare clone, so multiple workspaces pointing at the same repo don't duplicate data.

${\color{teal}\textbf{flow}}$ ships a built-in `flow` skill and consolidates skills from all repos into each workspace's `.claude/skills/` directory on render. Add your own skills to the shared directory or to individual repos. Run `flow reset skills` to update the built-in skill without touching your own.

See the [spec reference](docs/specs/) for YAML file schemas and the [command reference](docs/commands/) for usage, flags, and GIF demos.

## Customization

${\color{teal}\textbf{flow}}$ stores everything under `~/.flow` (override with `$FLOW_HOME`). Edit these files to customize your setup:

| Command | What it configures |
|---------|-------------------|
| `flow edit config` | Default agent, editor preferences |
| `flow edit status` | Status checks for tracking workstreams |
| `flow edit state <workspace>` | Repos and branches for a workspace |
| `flow reset skills` | Restore default agent skills to latest |

See the [spec reference](docs/specs/) for YAML schemas and the [command reference](docs/commands/) for all commands.

## Requirements

- Go 1.25+
- Git 2.20+ (worktree support)

## Support

I build and maintain projects like flow in my free time as personal hobbies. They're completely free and always will be. If you find this useful and want to show some support, feel free to buy me a coffee:

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/milldr)

## License

MIT
