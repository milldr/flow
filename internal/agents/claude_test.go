package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/milldr/flow/internal/state"
)

func TestEnsureSharedAgent(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatalf("EnsureSharedAgent: %v", err)
	}

	// Verify files exist
	files := []string{
		filepath.Join(agentsDir, "claude", "CLAUDE.md"),
		filepath.Join(agentsDir, "claude", "skills", "flow-cli", "SKILL.md"),
		filepath.Join(agentsDir, "claude", "skills", "workspace-structure", "SKILL.md"),
	}
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("file not created: %s", f)
		}
	}
}

func TestEnsureSharedAgentIdempotent(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	// Second call should not error
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatalf("second call: %v", err)
	}
}

func TestEnsureSharedAgentPreservesEdits(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	// Modify a file
	claudeMD := filepath.Join(agentsDir, "claude", "CLAUDE.md")
	customContent := []byte("# My Custom Instructions\n")
	if err := os.WriteFile(claudeMD, customContent, 0o644); err != nil {
		t.Fatal(err)
	}

	// Re-run — should NOT overwrite
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(customContent) {
		t.Error("EnsureSharedAgent overwrote user-edited file")
	}
}

func TestSetupWorkspaceClaude(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "test-ws")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	st := state.NewState("my-project", "A test project", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x"},
	})

	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "test-ws"); err != nil {
		t.Fatalf("SetupWorkspaceClaude: %v", err)
	}

	// Verify workspace CLAUDE.md
	data, err := os.ReadFile(filepath.Join(wsDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "my-project") {
		t.Error("workspace CLAUDE.md missing name")
	}
	if !strings.Contains(content, "A test project") {
		t.Error("workspace CLAUDE.md missing description")
	}
	if !strings.Contains(content, "repo-a") {
		t.Error("workspace CLAUDE.md missing repo-a")
	}
	if !strings.Contains(content, "repo-b") {
		t.Error("workspace CLAUDE.md missing repo-b")
	}

	// Verify .claude/CLAUDE.md symlink
	target, err := os.Readlink(filepath.Join(wsDir, ".claude", "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md symlink: %v", err)
	}
	expected := filepath.Join(agentsDir, "claude", "CLAUDE.md")
	if target != expected {
		t.Errorf("CLAUDE.md symlink target = %q, want %q", target, expected)
	}

	// Verify .claude/skills symlink
	target, err = os.Readlink(filepath.Join(wsDir, ".claude", "skills"))
	if err != nil {
		t.Fatalf("skills symlink: %v", err)
	}
	expected = filepath.Join(agentsDir, "claude", "skills")
	if target != expected {
		t.Errorf("skills symlink target = %q, want %q", target, expected)
	}
}

func TestSetupWorkspaceClaudeIdempotent(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "test-ws")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	st := state.NewState("my-project", "Test", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main"},
	})

	// Run twice — should not error
	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "test-ws"); err != nil {
		t.Fatal(err)
	}
	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "test-ws"); err != nil {
		t.Fatalf("second call: %v", err)
	}
}

func TestSetupWorkspaceClaudeEmptyNameDescription(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "bold-creek")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	st := state.NewState("", "", []state.Repo{
		{URL: "github.com/org/repo", Branch: "main"},
	})

	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "bold-creek"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(wsDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	// Should use ID as fallback for name
	if !strings.Contains(content, "bold-creek") {
		t.Error("workspace CLAUDE.md should use ID when name is empty")
	}
}

func TestSetupWorkspaceClaudeOverwritesGenerated(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "test-ws")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	st1 := state.NewState("v1", "First version", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main"},
	})
	if err := SetupWorkspaceClaude(wsDir, agentsDir, st1, "test-ws"); err != nil {
		t.Fatal(err)
	}

	// Update state with new repo
	st2 := state.NewState("v2", "Second version", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main"},
		{URL: "github.com/org/repo-b", Branch: "feat/x"},
	})
	if err := SetupWorkspaceClaude(wsDir, agentsDir, st2, "test-ws"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(wsDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "v2") {
		t.Error("workspace CLAUDE.md should reflect updated state")
	}
	if !strings.Contains(content, "repo-b") {
		t.Error("workspace CLAUDE.md should include new repo")
	}
}

func TestEnsureSymlinkRecreatesWrongTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")

	if err := os.WriteFile(target, []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create symlink to wrong target
	wrongTarget := filepath.Join(dir, "wrong")
	if err := os.WriteFile(wrongTarget, []byte("wrong"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(wrongTarget, link); err != nil {
		t.Fatal(err)
	}

	// ensureSymlink should fix it
	if err := ensureSymlink(target, link); err != nil {
		t.Fatalf("ensureSymlink: %v", err)
	}

	got, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Errorf("symlink target = %q, want %q", got, target)
	}
}
