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
		filepath.Join(agentsDir, "claude", "skills", "flow", "SKILL.md"),
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

func TestEnsureSharedAgentCleansUpLegacySkills(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	// Create legacy skill directories
	for _, name := range []string{"flow-cli", "workspace-structure"} {
		legacyDir := filepath.Join(agentsDir, "claude", "skills", name)
		if err := os.MkdirAll(legacyDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(legacyDir, "SKILL.md"), []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatalf("EnsureSharedAgent: %v", err)
	}

	// Legacy dirs should be removed
	for _, name := range []string{"flow-cli", "workspace-structure"} {
		legacyDir := filepath.Join(agentsDir, "claude", "skills", name)
		if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
			t.Errorf("legacy skill directory %q should have been removed", name)
		}
	}

	// New skill should exist
	if _, err := os.Stat(filepath.Join(agentsDir, "claude", "skills", "flow", "SKILL.md")); err != nil {
		t.Error("new flow skill not created")
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

	// Verify .claude/skills/ is a real directory with symlinked skills
	skillsDir := filepath.Join(wsDir, ".claude", "skills")
	info, err := os.Lstat(skillsDir)
	if err != nil {
		t.Fatalf("skills dir: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error(".claude/skills should be a real directory, not a symlink")
	}
	// Should contain the shared flow skill as a symlink
	flowTarget, err := os.Readlink(filepath.Join(skillsDir, "flow"))
	if err != nil {
		t.Fatal("flow skill symlink not found in .claude/skills/")
	}
	expectedFlow := filepath.Join(agentsDir, "claude", "skills", "flow")
	if flowTarget != expectedFlow {
		t.Errorf("flow skill target = %q, want %q", flowTarget, expectedFlow)
	}
}

func TestConsolidateSkillsFromRepos(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "test-ws")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	// Create fake repo with skills
	repoSkills := filepath.Join(wsDir, "repo-a", ".claude", "skills", "creating-prs")
	if err := os.MkdirAll(repoSkills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoSkills, "SKILL.md"), []byte("# Creating PRs"), 0o644); err != nil {
		t.Fatal(err)
	}

	st := state.NewState("my-project", "Test", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
	})

	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "test-ws"); err != nil {
		t.Fatalf("SetupWorkspaceClaude: %v", err)
	}

	skillsDir := filepath.Join(wsDir, ".claude", "skills")

	// Shared flow skill should be present
	if _, err := os.Readlink(filepath.Join(skillsDir, "flow")); err != nil {
		t.Error("shared flow skill not symlinked")
	}

	// Repo skill should be present
	target, err := os.Readlink(filepath.Join(skillsDir, "creating-prs"))
	if err != nil {
		t.Fatal("repo skill creating-prs not symlinked")
	}
	if target != repoSkills {
		t.Errorf("creating-prs target = %q, want %q", target, repoSkills)
	}
}

func TestConsolidateSkillsPrecedence(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "test-ws")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	// Create a workspace-level skill (real directory, highest precedence)
	wsSkill := filepath.Join(wsDir, ".claude", "skills", "flow")
	if err := os.MkdirAll(wsSkill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wsSkill, "SKILL.md"), []byte("# Custom flow"), 0o644); err != nil {
		t.Fatal(err)
	}

	st := state.NewState("my-project", "Test", []state.Repo{})

	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "test-ws"); err != nil {
		t.Fatalf("SetupWorkspaceClaude: %v", err)
	}

	// The workspace-level flow/ should remain as a real directory, not replaced by shared symlink
	skillsDir := filepath.Join(wsDir, ".claude", "skills")
	info, err := os.Lstat(filepath.Join(skillsDir, "flow"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("workspace-level flow skill should remain a real directory, not be replaced by symlink")
	}
	data, err := os.ReadFile(filepath.Join(skillsDir, "flow", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# Custom flow" {
		t.Error("workspace-level flow skill content was overwritten")
	}
}

func TestConsolidateSkillsRepoOverridesShared(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	wsDir := filepath.Join(dir, "workspaces", "test-ws")

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	// Create a repo with a "flow" skill (same name as shared — repo wins)
	repoFlowSkill := filepath.Join(wsDir, "repo-a", ".claude", "skills", "flow")
	if err := os.MkdirAll(repoFlowSkill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repoFlowSkill, "SKILL.md"), []byte("# Repo flow"), 0o644); err != nil {
		t.Fatal(err)
	}

	st := state.NewState("my-project", "Test", []state.Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
	})

	if err := SetupWorkspaceClaude(wsDir, agentsDir, st, "test-ws"); err != nil {
		t.Fatalf("SetupWorkspaceClaude: %v", err)
	}

	// flow/ should be a symlink pointing to the repo, not the shared dir
	target, err := os.Readlink(filepath.Join(wsDir, ".claude", "skills", "flow"))
	if err != nil {
		t.Fatal("flow skill should be a symlink")
	}
	if target != repoFlowSkill {
		t.Errorf("flow target = %q, want %q (repo should override shared)", target, repoFlowSkill)
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

func TestResetSharedAgent(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")

	// First create defaults
	if err := EnsureSharedAgent(agentsDir); err != nil {
		t.Fatal(err)
	}

	// Modify a file to simulate user edits
	claudeMD := filepath.Join(agentsDir, "claude", "CLAUDE.md")
	if err := os.WriteFile(claudeMD, []byte("custom"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Reset should overwrite user edits
	if err := ResetSharedAgent(agentsDir); err != nil {
		t.Fatalf("ResetSharedAgent: %v", err)
	}

	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "custom" {
		t.Error("ResetSharedAgent should have overwritten user-edited file")
	}
}

func TestEnsureSymlinkReplacesRegularFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")

	if err := os.WriteFile(target, []byte("target"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a regular file where the symlink should be
	if err := os.WriteFile(link, []byte("not a symlink"), 0o644); err != nil {
		t.Fatal(err)
	}

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
