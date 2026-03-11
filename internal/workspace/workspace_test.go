package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/milldr/flow/internal/agents"
	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/state"
)

// mockRunner records calls without executing git.
type mockRunner struct {
	clones      []string
	fetches     []string
	worktrees   []string
	removed     []string
	startPoints []string
	remoteRefs  []string
	resets      []string
	rebases     []string
	aborts      []string

	cloneErr     error
	fetchErr     error
	addWTErr     error
	branchExists bool
	isClean      bool
	rebaseErr    error
	resetErr     error
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

func (m *mockRunner) AddWorktreeNewBranch(_ context.Context, _, worktreePath, _, startPoint string) error {
	m.startPoints = append(m.startPoints, startPoint)
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

func (m *mockRunner) BranchExists(_ context.Context, _, _ string) (bool, error) {
	return m.branchExists, nil
}

func (m *mockRunner) DefaultBranch(_ context.Context, _ string) (string, error) {
	return "main", nil
}

func (m *mockRunner) EnsureRemoteRef(_ context.Context, _, branch string) error {
	m.remoteRefs = append(m.remoteRefs, branch)
	return nil
}

func (m *mockRunner) ResetBranch(_ context.Context, _, ref string) error {
	m.resets = append(m.resets, ref)
	return m.resetErr
}

func (m *mockRunner) IsClean(_ context.Context, _ string) (bool, error) {
	return m.isClean, nil
}

func (m *mockRunner) Rebase(_ context.Context, _, onto string) error {
	m.rebases = append(m.rebases, onto)
	return m.rebaseErr
}

func (m *mockRunner) RebaseAbort(_ context.Context, worktreePath string) error {
	m.aborts = append(m.aborts, worktreePath)
	return nil
}

func testService(t *testing.T) (*Service, *mockRunner) {
	t.Helper()
	dir := t.TempDir()
	cfg := &config.Config{
		Home:           dir,
		WorkspacesDir:  filepath.Join(dir, "workspaces"),
		ReposDir:       filepath.Join(dir, "repos"),
		AgentsDir:      filepath.Join(dir, "agents"),
		CacheDir:       filepath.Join(dir, "cache"),
		ConfigFile:     filepath.Join(dir, "config.yaml"),
		StatusSpecFile: filepath.Join(dir, "status.yaml"),
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

func noop(_ string) {}

func TestCreateAndList(t *testing.T) {
	svc, _ := testService(t)

	st := state.NewState("test-ws", "Test workspace", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create("test-ws", st); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Duplicate ID should fail
	if err := svc.Create("test-ws", st); err == nil {
		t.Fatal("expected error for duplicate workspace")
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("List count = %d, want 1", len(infos))
	}
	if infos[0].ID != "test-ws" {
		t.Errorf("ID = %q, want test-ws", infos[0].ID)
	}
	if infos[0].Name != "test-ws" {
		t.Errorf("Name = %q, want test-ws", infos[0].Name)
	}
}

func TestCreateIDDiffersFromName(t *testing.T) {
	svc, _ := testService(t)

	st := state.NewState("my-project", "A project", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create("calm-delta", st); err != nil {
		t.Fatalf("Create: %v", err)
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("List count = %d, want 1", len(infos))
	}
	if infos[0].ID != "calm-delta" {
		t.Errorf("ID = %q, want calm-delta", infos[0].ID)
	}
	if infos[0].Name != "my-project" {
		t.Errorf("Name = %q, want my-project", infos[0].Name)
	}
}

func TestCreateEmptyName(t *testing.T) {
	svc, _ := testService(t)

	st := state.NewState("", "", nil)

	if err := svc.Create("bold-creek", st); err != nil {
		t.Fatalf("Create: %v", err)
	}

	infos, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("List count = %d, want 1", len(infos))
	}
	if infos[0].ID != "bold-creek" {
		t.Errorf("ID = %q, want bold-creek", infos[0].ID)
	}
	if infos[0].Name != "" {
		t.Errorf("Name = %q, want empty", infos[0].Name)
	}
}

func TestRender(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Render test", "Render test", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	})

	if err := svc.Create("render-ws", st); err != nil {
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

	// Second render should fetch (not clone) and update existing worktrees
	mock.clones = nil
	mock.fetches = nil
	mock.worktrees = nil
	mock.remoteRefs = nil
	mock.resets = nil
	mock.isClean = true
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
	// Existing worktrees should be updated via EnsureRemoteRef + ResetBranch
	if len(mock.remoteRefs) != 2 {
		t.Errorf("second render remoteRefs = %d, want 2", len(mock.remoteRefs))
	}
	if len(mock.resets) != 2 {
		t.Errorf("second render resets = %d, want 2", len(mock.resets))
	}
}

func TestDelete(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Delete test", "Delete test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create("del-ws", st); err != nil {
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
	if err := svc.Create("found-ws", st); err != nil {
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

	for _, id := range []string{"ws-a", "ws-b", "ws-c"} {
		st := state.NewState(id, "Desc "+id, []state.Repo{
			{URL: "github.com/org/" + id, Branch: "main", Path: "./" + id},
		})
		if err := svc.Create(id, st); err != nil {
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

func TestRenderExistingWorktreeDirtySkip(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Dirty skip test", "Dirty skip test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create("dirty-ws", st); err != nil {
		t.Fatal(err)
	}

	// First render creates worktree
	if err := svc.Render(ctx, "dirty-ws", noop); err != nil {
		t.Fatal(err)
	}

	// Second render with dirty worktree should skip reset
	mock.resets = nil
	mock.isClean = false
	var messages []string
	err := svc.Render(ctx, "dirty-ws", func(msg string) { messages = append(messages, msg) })
	if err != nil {
		t.Fatalf("Render (dirty): %v", err)
	}
	if len(mock.resets) != 0 {
		t.Errorf("resets = %d, want 0 (dirty worktree should skip)", len(mock.resets))
	}
	// Check progress message mentions dirty
	found := false
	for _, msg := range messages {
		if strings.Contains(msg, "dirty") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected progress message mentioning dirty worktree")
	}
}

func TestRenderParallelFetch(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Parallel fetch", "Parallel fetch test", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "main", Path: "./repo-b"},
		{URL: "github.com/org/repo-c", Branch: "main", Path: "./repo-c"},
	})
	if err := svc.Create("parallel-ws", st); err != nil {
		t.Fatal(err)
	}

	err := svc.Render(ctx, "parallel-ws", noop)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// All repos should have been cloned and fetched
	if len(mock.clones) != 3 {
		t.Errorf("clones = %d, want 3", len(mock.clones))
	}
	if len(mock.fetches) != 3 {
		t.Errorf("fetches = %d, want 3", len(mock.fetches))
	}
	if len(mock.worktrees) != 3 {
		t.Errorf("worktrees = %d, want 3", len(mock.worktrees))
	}
}

func TestRenderCloneError(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Clone fail test", "Clone fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create("clone-fail", st); err != nil {
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

	st := state.NewState("Fetch fail test", "Fetch fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create("fetch-fail", st); err != nil {
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

	st := state.NewState("Worktree fail test", "Worktree fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "nonexistent", Path: "./repo"},
	})
	if err := svc.Create("wt-fail", st); err != nil {
		t.Fatal(err)
	}

	mock.addWTErr = errors.New("branch not found")
	err := svc.Render(ctx, "wt-fail", noop)
	if err == nil {
		t.Fatal("expected error from AddWorktree failure")
	}
}

func TestRenderNewBranchUsesRemoteRef(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Remote ref test", "Test new branch uses origin/ prefix", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/new-feature", Path: "./repo"},
	})
	if err := svc.Create("remote-ref", st); err != nil {
		t.Fatal(err)
	}

	// branchExists=false triggers the new branch path
	mock.branchExists = false
	err := svc.Render(ctx, "remote-ref", noop)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(mock.startPoints) != 1 {
		t.Fatalf("startPoints = %d, want 1", len(mock.startPoints))
	}
	if mock.startPoints[0] != "origin/main" {
		t.Errorf("startPoint = %q, want origin/main", mock.startPoints[0])
	}
	// EnsureRemoteRef should be called for the base branch during render
	if len(mock.remoteRefs) != 1 || mock.remoteRefs[0] != "main" {
		t.Errorf("remoteRefs = %v, want [main]", mock.remoteRefs)
	}
}

func TestRenderNewBranchUsesBaseField(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("Base field test", "Test new branch uses configured base", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/new-feature", Base: "develop", Path: "./repo"},
	})
	if err := svc.Create("base-field", st); err != nil {
		t.Fatal(err)
	}

	// branchExists=false triggers the new branch path
	mock.branchExists = false
	err := svc.Render(ctx, "base-field", noop)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(mock.startPoints) != 1 {
		t.Fatalf("startPoints = %d, want 1", len(mock.startPoints))
	}
	if mock.startPoints[0] != "origin/develop" {
		t.Errorf("startPoint = %q, want origin/develop", mock.startPoints[0])
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

	st := state.NewState("Not rendered", "Not rendered", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create("no-render", st); err != nil {
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

	st := state.NewState("Multi repo delete", "Multi repo delete", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "main", Path: "./repo-b"},
		{URL: "github.com/org/repo-c", Branch: "main", Path: "./repo-c"},
	})
	if err := svc.Create("multi-del", st); err != nil {
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

	if err := svc.Create("dup", st); err != nil {
		t.Fatal(err)
	}

	err := svc.Create("dup", st)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkspaceExists) {
		t.Errorf("expected ErrWorkspaceExists, got %v", err)
	}
}

func TestResolveByID(t *testing.T) {
	svc, _ := testService(t)
	st := state.NewState("my-project", "A project", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create("calm-delta", st); err != nil {
		t.Fatal(err)
	}

	matches, err := svc.Resolve("calm-delta")
	if err != nil {
		t.Fatalf("Resolve by ID: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].ID != "calm-delta" {
		t.Errorf("ID = %q, want calm-delta", matches[0].ID)
	}
	if matches[0].Name != "my-project" {
		t.Errorf("Name = %q, want my-project", matches[0].Name)
	}
}

func TestResolveByName(t *testing.T) {
	svc, _ := testService(t)
	st := state.NewState("my-project", "A project", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	if err := svc.Create("calm-delta", st); err != nil {
		t.Fatal(err)
	}

	matches, err := svc.Resolve("my-project")
	if err != nil {
		t.Fatalf("Resolve by name: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].ID != "calm-delta" {
		t.Errorf("ID = %q, want calm-delta", matches[0].ID)
	}
}

func TestResolveNotFound(t *testing.T) {
	svc, _ := testService(t)

	_, err := svc.Resolve("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestResolveAmbiguous(t *testing.T) {
	svc, _ := testService(t)

	// Create two workspaces with the same name but different IDs
	st1 := state.NewState("vpc-ipv6", "First attempt", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})
	st2 := state.NewState("vpc-ipv6", "Second attempt", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main", Path: "./repo"},
	})

	if err := svc.Create("calm-delta", st1); err != nil {
		t.Fatal(err)
	}
	if err := svc.Create("bold-creek", st2); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Resolve("vpc-ipv6")
	if err == nil {
		t.Fatal("expected error for ambiguous name")
	}
	if !errors.Is(err, ErrAmbiguousName) {
		t.Errorf("expected ErrAmbiguousName, got %v", err)
	}

	var ambErr *AmbiguousNameError
	if !errors.As(err, &ambErr) {
		t.Fatal("expected *AmbiguousNameError")
	}
	if len(ambErr.Matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(ambErr.Matches))
	}
}

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if id == "" {
		t.Fatal("GenerateID returned empty string")
	}
	// Should be adjective-noun format
	parts := 0
	for _, c := range id {
		if c == '-' {
			parts++
		}
	}
	if parts < 1 {
		t.Errorf("expected at least one hyphen in ID %q", id)
	}
}

func TestGenerateUniqueIDNoCollision(t *testing.T) {
	id := GenerateUniqueID(nil)
	if id == "" {
		t.Fatal("GenerateUniqueID returned empty string")
	}
}

func TestGenerateUniqueIDAvoidsExisting(t *testing.T) {
	// Generate many IDs and verify uniqueness
	existing := make([]string, 0, 50)
	for range 50 {
		id := GenerateUniqueID(existing)
		for _, e := range existing {
			if id == e {
				t.Fatalf("GenerateUniqueID returned duplicate: %q", id)
			}
		}
		existing = append(existing, id)
	}
}

func TestCreateSetsUpClaudeFiles(t *testing.T) {
	svc, _ := testService(t)

	st := state.NewState("create-claude", "Test Claude setup on create", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	})

	if err := svc.Create("create-claude-ws", st); err != nil {
		t.Fatal(err)
	}

	wsDir := svc.Config.WorkspacePath("create-claude-ws")

	// Workspace CLAUDE.md should exist with repo info
	data, err := os.ReadFile(filepath.Join(wsDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal("workspace CLAUDE.md not created by Create()")
	}
	content := string(data)
	if !strings.Contains(content, "create-claude") {
		t.Error("workspace CLAUDE.md missing name")
	}
	if !strings.Contains(content, "repo-a") {
		t.Error("workspace CLAUDE.md missing repo-a")
	}
	if !strings.Contains(content, "repo-b") {
		t.Error("workspace CLAUDE.md missing repo-b")
	}

	// .claude/CLAUDE.md should be a symlink to shared
	target, err := os.Readlink(filepath.Join(wsDir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatal(".claude/CLAUDE.md symlink not created by Create()")
	}
	expected := filepath.Join(svc.Config.AgentsDir, "claude", "CLAUDE.md")
	if target != expected {
		t.Errorf(".claude/CLAUDE.md target = %q, want %q", target, expected)
	}

	// .claude/skills should be a symlink to shared
	target, err = os.Readlink(filepath.Join(wsDir, ".claude", "skills"))
	if err != nil {
		t.Fatal(".claude/skills symlink not created by Create()")
	}
	expected = filepath.Join(svc.Config.AgentsDir, "claude", "skills")
	if target != expected {
		t.Errorf(".claude/skills target = %q, want %q", target, expected)
	}
}

func TestRenderCreatesClaudeFiles(t *testing.T) {
	svc, _ := testService(t)
	ctx := context.Background()

	st := state.NewState("claude-test", "Test Claude setup", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	})

	if err := svc.Create("claude-ws", st); err != nil {
		t.Fatal(err)
	}
	if err := svc.Render(ctx, "claude-ws", noop); err != nil {
		t.Fatalf("Render: %v", err)
	}

	wsDir := svc.Config.WorkspacePath("claude-ws")

	// Workspace CLAUDE.md should exist with repo info
	data, err := os.ReadFile(filepath.Join(wsDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal("workspace CLAUDE.md not created")
	}
	content := string(data)
	if !strings.Contains(content, "claude-test") {
		t.Error("workspace CLAUDE.md missing name")
	}
	if !strings.Contains(content, "repo-a") {
		t.Error("workspace CLAUDE.md missing repo-a")
	}
	if !strings.Contains(content, "repo-b") {
		t.Error("workspace CLAUDE.md missing repo-b")
	}

	// .claude/CLAUDE.md should be a symlink to shared
	target, err := os.Readlink(filepath.Join(wsDir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatal(".claude/CLAUDE.md symlink not created")
	}
	expected := filepath.Join(svc.Config.AgentsDir, "claude", "CLAUDE.md")
	if target != expected {
		t.Errorf(".claude/CLAUDE.md target = %q, want %q", target, expected)
	}

	// .claude/skills should be a symlink to shared
	target, err = os.Readlink(filepath.Join(wsDir, ".claude", "skills"))
	if err != nil {
		t.Fatal(".claude/skills symlink not created")
	}
	expected = filepath.Join(svc.Config.AgentsDir, "claude", "skills")
	if target != expected {
		t.Errorf(".claude/skills target = %q, want %q", target, expected)
	}
}

func TestSyncCleanRebase(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("sync-clean", "Sync clean test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"},
	})
	if err := svc.Create("sync-clean", st); err != nil {
		t.Fatal(err)
	}
	// Render to create worktree directory
	if err := svc.Render(ctx, "sync-clean", noop); err != nil {
		t.Fatal(err)
	}

	mock.fetches = nil
	mock.remoteRefs = nil
	mock.isClean = true
	if err := svc.Sync(ctx, "sync-clean", noop); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if len(mock.fetches) != 1 {
		t.Errorf("fetches = %d, want 1", len(mock.fetches))
	}
	if len(mock.remoteRefs) != 1 || mock.remoteRefs[0] != "main" {
		t.Errorf("remoteRefs = %v, want [main]", mock.remoteRefs)
	}
	if len(mock.rebases) != 1 || mock.rebases[0] != "origin/main" {
		t.Errorf("rebases = %v, want [origin/main]", mock.rebases)
	}
}

func TestSyncDirtySkip(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("sync-dirty", "Sync dirty test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"},
	})
	if err := svc.Create("sync-dirty", st); err != nil {
		t.Fatal(err)
	}
	if err := svc.Render(ctx, "sync-dirty", noop); err != nil {
		t.Fatal(err)
	}

	mock.fetches = nil
	mock.isClean = false
	if err := svc.Sync(ctx, "sync-dirty", noop); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if len(mock.rebases) != 0 {
		t.Errorf("rebases = %d, want 0 (dirty worktree should skip)", len(mock.rebases))
	}
}

func TestSyncUsesBaseField(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("sync-base", "Sync base field test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/x", Base: "develop", Path: "./repo"},
	})
	if err := svc.Create("sync-base", st); err != nil {
		t.Fatal(err)
	}
	if err := svc.Render(ctx, "sync-base", noop); err != nil {
		t.Fatal(err)
	}

	mock.fetches = nil
	mock.remoteRefs = nil
	mock.isClean = true
	if err := svc.Sync(ctx, "sync-base", noop); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if len(mock.remoteRefs) != 1 || mock.remoteRefs[0] != "develop" {
		t.Errorf("remoteRefs = %v, want [develop]", mock.remoteRefs)
	}
	if len(mock.rebases) != 1 || mock.rebases[0] != "origin/develop" {
		t.Errorf("rebases = %v, want [origin/develop]", mock.rebases)
	}
}

func TestSyncNoWorktreeSkip(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("sync-nowt", "Sync no worktree test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"},
	})
	if err := svc.Create("sync-nowt", st); err != nil {
		t.Fatal(err)
	}
	// Don't render — worktree doesn't exist

	if err := svc.Sync(ctx, "sync-nowt", noop); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if len(mock.fetches) != 0 {
		t.Errorf("fetches = %d, want 0 (no worktree to sync)", len(mock.fetches))
	}
	if len(mock.rebases) != 0 {
		t.Errorf("rebases = %d, want 0", len(mock.rebases))
	}
}

func TestSyncRebaseError(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("sync-fail", "Sync rebase fail test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"},
	})
	if err := svc.Create("sync-fail", st); err != nil {
		t.Fatal(err)
	}
	if err := svc.Render(ctx, "sync-fail", noop); err != nil {
		t.Fatal(err)
	}

	mock.fetches = nil
	mock.isClean = true
	mock.rebaseErr = errors.New("conflict")

	err := svc.Sync(ctx, "sync-fail", noop)
	if err == nil {
		t.Fatal("expected error from rebase failure")
	}

	if len(mock.aborts) != 1 {
		t.Errorf("aborts = %d, want 1 (should abort failed rebase)", len(mock.aborts))
	}
}

func TestSyncMultiRepos(t *testing.T) {
	svc, mock := testService(t)
	ctx := context.Background()

	st := state.NewState("sync-multi", "Sync multi repos", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "feat/a", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/b", Path: "./repo-b"},
		{URL: "github.com/org/repo-c", Branch: "feat/c", Path: "./repo-c"},
	})
	if err := svc.Create("sync-multi", st); err != nil {
		t.Fatal(err)
	}
	if err := svc.Render(ctx, "sync-multi", noop); err != nil {
		t.Fatal(err)
	}

	mock.fetches = nil
	mock.isClean = true

	if err := svc.Sync(ctx, "sync-multi", noop); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if len(mock.fetches) != 3 {
		t.Errorf("fetches = %d, want 3", len(mock.fetches))
	}
	if len(mock.rebases) != 3 {
		t.Errorf("rebases = %d, want 3", len(mock.rebases))
	}
}
