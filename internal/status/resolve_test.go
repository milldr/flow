package status

import (
	"context"
	"testing"
)

// mockRunner implements CheckRunner for testing.
type mockRunner struct {
	// results maps command strings to their return value.
	results map[string]bool
}

func (m *mockRunner) RunCheck(_ context.Context, command string, _ []string) bool {
	if result, ok := m.results[command]; ok {
		return result
	}
	return false
}

func testSpec() *Spec {
	return &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Name: "closed", Check: "check-closed"},
			{Name: "in-review", Check: "check-review"},
			{Name: "in-progress", Default: true},
		},
	}
}

func TestResolveRepoFirstMatch(t *testing.T) {
	mock := &mockRunner{results: map[string]bool{
		"check-closed": true,
		"check-review": true,
	}}
	resolver := &Resolver{Runner: mock}

	repo := RepoInfo{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"}
	status := resolver.ResolveRepo(context.Background(), testSpec(), repo, "ws-1", "my-ws")

	if status != "closed" {
		t.Errorf("expected closed (first match), got %q", status)
	}
}

func TestResolveRepoSecondMatch(t *testing.T) {
	mock := &mockRunner{results: map[string]bool{
		"check-closed": false,
		"check-review": true,
	}}
	resolver := &Resolver{Runner: mock}

	repo := RepoInfo{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"}
	status := resolver.ResolveRepo(context.Background(), testSpec(), repo, "ws-1", "my-ws")

	if status != "in-review" {
		t.Errorf("expected in-review, got %q", status)
	}
}

func TestResolveRepoDefault(t *testing.T) {
	mock := &mockRunner{results: map[string]bool{
		"check-closed": false,
		"check-review": false,
	}}
	resolver := &Resolver{Runner: mock}

	repo := RepoInfo{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"}
	status := resolver.ResolveRepo(context.Background(), testSpec(), repo, "ws-1", "my-ws")

	if status != "in-progress" {
		t.Errorf("expected in-progress (default), got %q", status)
	}
}

func TestResolveWorkspaceSingleRepo(t *testing.T) {
	mock := &mockRunner{results: map[string]bool{
		"check-closed": false,
		"check-review": true,
	}}
	resolver := &Resolver{Runner: mock}

	repos := []RepoInfo{
		{URL: "github.com/org/repo", Branch: "feat/x", Path: "./repo"},
	}
	result := resolver.ResolveWorkspace(context.Background(), testSpec(), repos, "ws-1", "my-ws")

	if result.Status != "in-review" {
		t.Errorf("workspace status = %q, want in-review", result.Status)
	}
	if len(result.Repos) != 1 {
		t.Fatalf("repo count = %d, want 1", len(result.Repos))
	}
	if result.Repos[0].Status != "in-review" {
		t.Errorf("repo status = %q, want in-review", result.Repos[0].Status)
	}
}

func TestResolveWorkspaceMultiRepoLeastAdvanced(t *testing.T) {
	// Two repos with different statuses — repo-a is closed, repo-b is in-review.
	// Workspace status should be the least advanced (in-review).
	perRepoMock := &perRepoRunner{
		results: map[string]map[string]bool{
			"github.com/org/repo-a": {"check-closed": true},
			"github.com/org/repo-b": {"check-closed": false, "check-review": true},
		},
	}
	resolver := &Resolver{Runner: perRepoMock}

	repos := []RepoInfo{
		{URL: "github.com/org/repo-a", Branch: "feat/x", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	}
	result := resolver.ResolveWorkspace(context.Background(), testSpec(), repos, "ws-1", "my-ws")

	if result.Status != "in-review" {
		t.Errorf("workspace status = %q, want in-review (least advanced)", result.Status)
	}
	if result.Repos[0].Status != "closed" {
		t.Errorf("repo-a status = %q, want closed", result.Repos[0].Status)
	}
	if result.Repos[1].Status != "in-review" {
		t.Errorf("repo-b status = %q, want in-review", result.Repos[1].Status)
	}
}

func TestResolveWorkspaceNoRepos(t *testing.T) {
	mock := &mockRunner{results: map[string]bool{}}
	resolver := &Resolver{Runner: mock}

	result := resolver.ResolveWorkspace(context.Background(), testSpec(), nil, "ws-1", "my-ws")

	if result.Status != "in-progress" {
		t.Errorf("workspace status = %q, want in-progress (default)", result.Status)
	}
}

func TestResolveWorkspaceAllClosed(t *testing.T) {
	mock := &mockRunner{results: map[string]bool{
		"check-closed": true,
	}}
	resolver := &Resolver{Runner: mock}

	repos := []RepoInfo{
		{URL: "github.com/org/repo-a", Branch: "feat/x", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	}
	result := resolver.ResolveWorkspace(context.Background(), testSpec(), repos, "ws-1", "my-ws")

	if result.Status != "closed" {
		t.Errorf("workspace status = %q, want closed", result.Status)
	}
}

// perRepoRunner distinguishes check results by repo URL (extracted from env).
type perRepoRunner struct {
	results map[string]map[string]bool // repoURL -> command -> result
}

func (m *perRepoRunner) RunCheck(_ context.Context, command string, env []string) bool {
	var repoURL string
	for _, e := range env {
		if len(e) > 14 && e[:14] == "FLOW_REPO_URL=" {
			repoURL = e[14:]
			break
		}
	}
	if repoResults, ok := m.results[repoURL]; ok {
		if result, ok := repoResults[command]; ok {
			return result
		}
	}
	return false
}
