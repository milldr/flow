package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/state"
)

// mockRunner records calls without executing git.
type mockRunner struct {
	clones    []string
	fetches   []string
	worktrees []string
	removed   []string

	cloneErr error
	fetchErr error
	addWTErr error
}

func (m *mockRunner) BareClone(_ context.Context, url, dest string) error {
	m.clones = append(m.clones, url)
	if m.cloneErr != nil {
		return m.cloneErr
	}
	return os.MkdirAll(dest, 0o755) // create dir so stat checks pass
}

func (m *mockRunner) Fetch(_ context.Context, repoPath string) error {
	m.fetches = append(m.fetches, repoPath)
	return m.fetchErr
}

func (m *mockRunner) AddWorktree(_ context.Context, _, worktreePath, _ string) error {
	m.worktrees = append(m.worktrees, worktreePath)
	if m.addWTErr != nil {
		return m.addWTErr
	}
	return os.MkdirAll(worktreePath, 0o755)
}

func (m *mockRunner) RemoveWorktree(_ context.Context, _, worktreePath string) error {
	m.removed = append(m.removed, worktreePath)
	return os.RemoveAll(worktreePath)
}

func testService(t *testing.T) (*Service, *mockRunner) {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		Home:          dir,
		WorkspacesDir: filepath.Join(dir, "workspaces"),
		ReposDir:      filepath.Join(dir, "repos"),
	}
	if err := cfg.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	mock := &mockRunner{}
	return &Service{Config: cfg, Git: mock}, mock
}

func noop(_ string) {}

func TestCreateAndList(t *testing.T) {
	svc, _ := testService(t)

	st := state.NewState("test-ws", "Test workspace", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create(st); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Duplicate should fail
	if err := svc.Create(st); err == nil {
		t.Fatal("expected error for duplicate workspace")
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("List count = %d, want 1", len(infos))
	}
	if infos[0].Name != "test-ws" {
		t.Errorf("Name = %q, want test-ws", infos[0].Name)
	}
}

func TestRender(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("render-ws", "Render test", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	})

	if err := svc.Create(st); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var messages []string
	err := svc.Render(ctx, "render-ws", func(msg string) {
		messages = append(messages, msg)
	})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(mock.clones) != 2 {
		t.Errorf("clones = %d, want 2", len(mock.clones))
	}
	if len(mock.worktrees) != 2 {
		t.Errorf("worktrees = %d, want 2", len(mock.worktrees))
	}

	// Second render should fetch, not clone
	mock.clones = nil
	mock.fetches = nil
	mock.worktrees = nil
	err = svc.Render(ctx, "render-ws", noop)
	if err != nil {
		t.Fatalf("Render (2nd): %v", err)
	}
	if len(mock.clones) != 0 {
		t.Errorf("second render clones = %d, want 0", len(mock.clones))
	}
	if len(mock.fetches) != 2 {
		t.Errorf("second render fetches = %d, want 2", len(mock.fetches))
	}
	if len(mock.worktrees) != 0 {
		t.Errorf("second render worktrees = %d, want 0 (already exist)", len(mock.worktrees))
	}
}

func TestDelete(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("del-ws", "Delete test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	// Render first to create worktrees
	if err := svc.Render(ctx, "del-ws", noop); err != nil {
		t.Fatal(err)
	}

	if err := svc.Delete(ctx, "del-ws"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if len(mock.removed) != 1 {
		t.Errorf("removed worktrees = %d, want 1", len(mock.removed))
	}

	// Directory should be gone
	wsDir := svc.Config.WorkspacePath("del-ws")
	if _, err := os.Stat(wsDir); !os.IsNotExist(err) {
		t.Error("workspace directory still exists after delete")
	}
}

func TestFindNotFound(t *testing.T) {
	svc, _ := testService(t)
	_, err := svc.Find("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent workspace")
	}
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestFindSuccess(t *testing.T) {
	svc, _ := testService(t)
	st := state.NewState("found-ws", "Find test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	found, err := svc.Find("found-ws")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found.Metadata.Name != "found-ws" {
		t.Errorf("Name = %q, want found-ws", found.Metadata.Name)
	}
}

func TestListEmptyDir(t *testing.T) {
	svc, _ := testService(t)
	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if infos != nil {
		t.Errorf("expected nil, got %v", infos)
	}
}

func TestListSkipsNonDirs(t *testing.T) {
	svc, _ := testService(t)

	// Create a regular file in the workspaces dir
	if err := os.WriteFile(filepath.Join(svc.Config.WorkspacesDir, "not-a-dir"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 0 {
		t.Errorf("expected 0 infos, got %d", len(infos))
	}
}

func TestListSkipsMalformedState(t *testing.T) {
	svc, _ := testService(t)

	// Create a workspace dir with invalid state.yaml
	wsDir := filepath.Join(svc.Config.WorkspacesDir, "broken")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsDir, "state.yaml"), []byte(":\ninvalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Broken workspace should be skipped, not cause an error
	if len(infos) != 0 {
		t.Errorf("expected 0 infos (broken skipped), got %d", len(infos))
	}
}

func TestListMultipleWorkspaces(t *testing.T) {
	svc, _ := testService(t)

	for _, name := range []string{"ws-a", "ws-b", "ws-c"} {
		st := state.NewState(name, "Desc "+name, []state.Repo{
			{URL: "github.com/org/" + name, Branch: "main", Path: "./" + name},
		})
		if err := svc.Create(st); err != nil {
			t.Fatal(err)
		}
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 3 {
		t.Errorf("expected 3 infos, got %d", len(infos))
	}
}

func TestRenderCloneError(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("clone-fail", "Clone fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	mock.cloneErr = errors.New("auth failed")
	err := svc.Render(ctx, "clone-fail", noop)
	if err == nil {
		t.Fatal("expected error from clone failure")
	}
	if !errors.Is(err, mock.cloneErr) {
		t.Errorf("expected wrapped auth error, got %v", err)
	}
}

func TestRenderFetchError(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("fetch-fail", "Fetch fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	// First render succeeds (clones)
	if err := svc.Render(ctx, "fetch-fail", noop); err != nil {
		t.Fatal(err)
	}

	// Second render fails on fetch
	mock.fetchErr = errors.New("network down")
	err := svc.Render(ctx, "fetch-fail", noop)
	if err == nil {
		t.Fatal("expected error from fetch failure")
	}
}

func TestRenderAddWorktreeError(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("wt-fail", "Worktree fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "nonexistent", Path: "./repo"},
	})
	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	mock.addWTErr = errors.New("branch not found")
	err := svc.Render(ctx, "wt-fail", noop)
	if err == nil {
		t.Fatal("expected error from AddWorktree failure")
	}
}

func TestRenderNotFound(t *testing.T) {
	svc, _ := testService(t)
	ctx := context.Background()

	err := svc.Render(ctx, "nonexistent", noop)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	svc, _ := testService(t)
	ctx := context.Background()

	err := svc.Delete(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestDeleteWithoutRender(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("no-render", "Not rendered", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	// Delete without rendering — worktrees don't exist
	if err := svc.Delete(ctx, "no-render"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// RemoveWorktree should not have been called (worktree dir doesn't exist)
	if len(mock.removed) != 0 {
		t.Errorf("removed = %d, want 0 (no worktrees existed)", len(mock.removed))
	}
}

func TestDeleteMultipleRepos(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("multi-del", "Multi repo delete", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "main", Path: "./repo-b"},
		{URL: "github.com/org/repo-c", Branch: "main", Path: "./repo-c"},
	})
	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}
	if err := svc.Render(ctx, "multi-del", noop); err != nil {
		t.Fatal(err)
	}

	if err := svc.Delete(ctx, "multi-del"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(mock.removed) != 3 {
		t.Errorf("removed = %d, want 3", len(mock.removed))
	}
}

func TestCreateDuplicateError(t *testing.T) {
	svc, _ := testService(t)
	st := state.NewState("dup", "Dup test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create(st); err != nil {
		t.Fatal(err)
	}

	err := svc.Create(st)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkspaceExists) {
		t.Errorf("expected ErrWorkspaceExists, got %v", err)
	}
}
