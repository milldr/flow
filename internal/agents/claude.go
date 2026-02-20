package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/milldr/flow/internal/state"
)

// EnsureSharedAgent creates the shared Claude agent directory and writes
// default files only if they don't already exist (preserves user edits).
func EnsureSharedAgent(agentsDir string) error {
	claudeDir := filepath.Join(agentsDir, "claude")
	flowCLIDir := filepath.Join(claudeDir, "skills", "flow-cli")
	wsStructDir := filepath.Join(claudeDir, "skills", "workspace-structure")

	for _, dir := range []string{flowCLIDir, wsStructDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	defaults := []struct {
		path    string
		content []byte
	}{
		{filepath.Join(claudeDir, "CLAUDE.md"), defaultClaudeMD},
		{filepath.Join(flowCLIDir, "SKILL.md"), defaultFlowCLI},
		{filepath.Join(wsStructDir, "SKILL.md"), defaultWorkspaceStructure},
	}

	for _, d := range defaults {
		if _, err := os.Stat(d.path); os.IsNotExist(err) {
			if err := os.WriteFile(d.path, d.content, 0o644); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetupWorkspaceClaude generates workspace-specific Claude files and creates
// symlinks to the shared agent directory.
func SetupWorkspaceClaude(wsDir, agentsDir string, st *state.State, id string) error {
	claudeDir := filepath.Join(agentsDir, "claude")

	// Generate workspace-specific CLAUDE.md (always overwritten)
	content := generateWorkspaceClaude(st, id)
	if err := os.WriteFile(filepath.Join(wsDir, "CLAUDE.md"), []byte(content), 0o644); err != nil {
		return err
	}

	// Create .claude/ directory
	dotClaudeDir := filepath.Join(wsDir, ".claude")
	if err := os.MkdirAll(dotClaudeDir, 0o755); err != nil {
		return err
	}

	// Symlink .claude/CLAUDE.md → shared CLAUDE.md
	if err := ensureSymlink(
		filepath.Join(claudeDir, "CLAUDE.md"),
		filepath.Join(dotClaudeDir, "CLAUDE.md"),
	); err != nil {
		return err
	}

	// Symlink .claude/skills/ → shared skills/
	if err := ensureSymlink(
		filepath.Join(claudeDir, "skills"),
		filepath.Join(dotClaudeDir, "skills"),
	); err != nil {
		return err
	}

	return nil
}

// generateWorkspaceClaude produces the workspace-specific CLAUDE.md content.
func generateWorkspaceClaude(st *state.State, id string) string {
	var b strings.Builder

	b.WriteString("# Workspace: ")
	if st.Metadata.Name != "" {
		b.WriteString(st.Metadata.Name)
	} else {
		b.WriteString(id)
	}
	b.WriteString("\n")

	if st.Metadata.Description != "" {
		b.WriteString("\n")
		b.WriteString(st.Metadata.Description)
		b.WriteString("\n")
	}

	b.WriteString("\n## Workspace ID\n\n")
	b.WriteString(id)
	b.WriteString("\n")

	if len(st.Spec.Repos) > 0 {
		b.WriteString("\n## Repositories\n\n")
		b.WriteString("| Path | URL | Branch |\n")
		b.WriteString("|------|-----|--------|\n")
		for _, r := range st.Spec.Repos {
			p := state.RepoPath(r)
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", p, r.URL, r.Branch))
		}
	}

	b.WriteString("\n## Quick Reference\n\n")
	b.WriteString(fmt.Sprintf("- View state: `flow state %s`\n", id))
	b.WriteString(fmt.Sprintf("- Re-render: `flow render %s`\n", id))
	b.WriteString(fmt.Sprintf("- Run command: `flow exec %s -- <cmd>`\n", id))

	return b.String()
}

// ensureSymlink creates a symlink at linkPath pointing to target.
// It's idempotent: no-op if correct, recreated if wrong target.
func ensureSymlink(target, linkPath string) error {
	existing, err := os.Readlink(linkPath)
	if err == nil {
		if existing == target {
			return nil // already correct
		}
		// Wrong target — remove and recreate
		if err := os.Remove(linkPath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		// Path exists but is not a symlink — remove it
		if err := os.Remove(linkPath); err != nil {
			return err
		}
	}

	return os.Symlink(target, linkPath)
}
