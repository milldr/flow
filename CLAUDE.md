# Flow

Flow is a CLI tool that manages multi-repo development workspaces using git worktrees. Written in Go.

## Architecture

- `cmd/flow/main.go` — entry point
- `internal/cmd/` — cobra command definitions (root, init, list, render, state, exec, delete)
- `internal/config/` — path resolution (`~/.flow/` structure) and global config (`config.yaml`)
- `internal/state/` — workspace state file schema, load/save/validate
- `internal/workspace/` — core business logic (create, list, find, resolve, render, delete)
- `internal/git/` — git CLI wrapper with `Runner` interface for testability
- `internal/agents/` — AI agent file management (Claude instructions per workspace)
- `internal/ui/` — styled terminal output (lipgloss), spinners (bubbletea), prompts (huh)
- `internal/iterm/` — iTerm2 tab color/title escape sequences
- `docs/prd/` — product requirement documents

## Key Patterns

- **Dependency injection**: `workspace.Service` accepts a `git.Runner` interface. Tests use a mock runner — no real git calls.
- **ID vs Name**: workspaces have a stable random ID (directory name like `calm-delta`) and an optional human-friendly name in metadata.
- **State-driven**: everything flows from `state.yaml`. Render is idempotent.
- **Bare clone caching**: repos are bare-cloned once to `~/.flow/repos/`, shared across workspaces via worktrees.
- **Embedded defaults**: `internal/agents/defaults/` contains default Claude instructions, embedded via `go:embed` and written only if missing (preserves user edits).

## Development

```bash
make build       # build binary
make test        # run tests
make lint        # run golangci-lint
```

All three must pass before committing. The linter config (`.golangci.yml`) enforces `forbidigo` (no `fmt.Print*` outside `ui/`/`cmd/`) and `err113` (wrapped errors).

## Testing

Tests use `t.TempDir()` for isolation. The workspace tests use a `mockRunner` that implements `git.Runner` and creates directories to simulate worktrees. No real git operations in tests.

## Conventions

- Git: never force push. Create new commits instead of amending pushed commits.
- Commit messages: single line, no co-authored-by
- Errors: use sentinel errors with `fmt.Errorf("context: %w", err)` wrapping
- Output: all user-facing output goes through `internal/ui/` — never use `fmt.Print*` directly
- YAML: `gopkg.in/yaml.v3` with struct tags
- No over-engineering: only build what's needed now
