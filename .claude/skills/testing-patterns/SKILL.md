---
name: testing-patterns
description: Testing conventions and patterns for this project. Use when writing tests, creating mocks, or understanding the test infrastructure.
---

# Testing Patterns

## General Rules

- Every new file gets a corresponding `_test.go`
- Tests use `t.TempDir()` for filesystem isolation — no shared state between tests
- No real git operations — use the mock runner
- Run `make test` (which uses `-race`) to verify

## Mock Git Runner

The workspace tests use a `mockRunner` that implements `git.Runner`. It records calls and simulates filesystem operations:

```go
type mockRunner struct {
    clones    []string  // URLs passed to BareClone
    fetches   []string  // paths passed to Fetch
    worktrees []string  // paths passed to AddWorktree
    removed   []string  // paths passed to RemoveWorktree

    cloneErr     error  // inject errors
    fetchErr     error
    addWTErr     error
    branchExists bool   // controls BranchExists return
}
```

Key behavior: `BareClone` and `AddWorktree` create the directories with `os.MkdirAll` so that subsequent `os.Stat` checks pass. This simulates git's side effects without actual git.

## testService Helper

Creates a fully isolated test environment:

```go
func testService(t *testing.T) (*Service, *mockRunner) {
    t.Helper()
    dir := t.TempDir()
    cfg := &config.Config{
        Home:          dir,
        WorkspacesDir: filepath.Join(dir, "workspaces"),
        ReposDir:      filepath.Join(dir, "repos"),
        AgentsDir:     filepath.Join(dir, "agents"),
        ConfigFile:    filepath.Join(dir, "config.yaml"),
    }
    if err := cfg.EnsureDirs(); err != nil {
        t.Fatal(err)
    }
    if err := agents.EnsureSharedAgent(cfg.AgentsDir); err != nil {
        t.Fatal(err)
    }
    mock := &mockRunner{}
    return &Service{Config: cfg, Git: mock}, mock
}
```

When adding new fields to `Config`, update this helper too.

## Test Structure

Tests follow a consistent pattern:

```go
func TestOperationName(t *testing.T) {
    svc, mock := testService(t)
    ctx := context.Background()

    // Setup: create workspace with state
    st := state.NewState("name", "description", []state.Repo{
        {URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
    })
    if err := svc.Create("ws-id", st); err != nil {
        t.Fatal(err)
    }

    // Act
    err := svc.Render(ctx, "ws-id", noop)
    if err != nil {
        t.Fatalf("Render: %v", err)
    }

    // Assert
    if len(mock.clones) != 1 {
        t.Errorf("clones = %d, want 1", len(mock.clones))
    }
}
```

## Error Injection

Set error fields on the mock to test failure paths:

```go
mock.cloneErr = errors.New("auth failed")
err := svc.Render(ctx, "ws-id", noop)
if !errors.Is(err, mock.cloneErr) {
    t.Errorf("expected wrapped auth error, got %v", err)
}
```

## Idempotency Tests

Many operations should be idempotent. Test by running twice:

```go
// First call
if err := svc.Render(ctx, "ws-id", noop); err != nil {
    t.Fatal(err)
}

// Second call — should not error, should skip work
mock.clones = nil
mock.fetches = nil
if err := svc.Render(ctx, "ws-id", noop); err != nil {
    t.Fatalf("second render: %v", err)
}
if len(mock.clones) != 0 {
    t.Error("second render should not clone")
}
```

## Config/State Tests

For packages without git dependencies (config, state, agents), test directly without mocks:

```go
func TestFlowConfigRoundTrip(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "config.yaml")

    fc := DefaultFlowConfig()
    if err := SaveFlowConfig(path, fc); err != nil {
        t.Fatalf("Save: %v", err)
    }

    loaded, err := LoadFlowConfig(path)
    if err != nil {
        t.Fatalf("Load: %v", err)
    }
    // assert fields match
}
```

## The `noop` Helper

Used as a progress callback when you don't care about messages:

```go
func noop(_ string) {}
```
