package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/milldr/flow/internal/state"
)

// legacySkillDirs are old skill directories that should be cleaned up.
var legacySkillDirs = []string{"flow-cli", "workspace-structure"}

// EnsureSharedAgent creates the shared Claude agent directory and writes
// default files only if they don't already exist (preserves user edits).
func EnsureSharedAgent(agentsDir string) error {
	claudeDir := filepath.Join(agentsDir, "claude")
	flowDir := filepath.Join(claudeDir, "skills", "flow")

	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		return err
	}

	defaults := []struct {
		path    string
		content []byte
	}{
		{filepath.Join(claudeDir, "CLAUDE.md"), defaultClaudeMD},
		{filepath.Join(flowDir, "SKILL.md"), defaultFlowSkill},
	}

	for _, d := range defaults {
		if _, err := os.Stat(d.path); os.IsNotExist(err) {
			if err := os.WriteFile(d.path, d.content, 0o644); err != nil {
				return err
			}
		}
	}

	// Clean up legacy skill directories
	cleanupLegacySkills(filepath.Join(claudeDir, "skills"))

	return nil
}

// ResetSharedAgent overwrites the shared Claude agent files with their defaults,
// regardless of whether they already exist.
func ResetSharedAgent(agentsDir string) error {
	claudeDir := filepath.Join(agentsDir, "claude")
	flowDir := filepath.Join(claudeDir, "skills", "flow")

	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		return err
	}

	defaults := []struct {
		path    string
		content []byte
	}{
		{filepath.Join(claudeDir, "CLAUDE.md"), defaultClaudeMD},
		{filepath.Join(flowDir, "SKILL.md"), defaultFlowSkill},
	}

	for _, d := range defaults {
		if err := os.WriteFile(d.path, d.content, 0o644); err != nil {
			return err
		}
	}

	// Clean up legacy skill directories
	cleanupLegacySkills(filepath.Join(claudeDir, "skills"))

	return nil
}

// cleanupLegacySkills removes old skill directories that have been consolidated.
func cleanupLegacySkills(skillsDir string) {
	for _, name := range legacySkillDirs {
		_ = os.RemoveAll(filepath.Join(skillsDir, name))
	}
}

// SetupWorkspaceClaude generates workspace-specific Claude files and consolidates
// skills from all sources into the workspace's .claude/skills/ directory.
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

	// Build consolidated skills directory
	if err := consolidateSkills(wsDir, agentsDir, st); err != nil {
		return err
	}

	return nil
}

// consolidateSkills rebuilds .claude/skills/ as a real directory containing
// symlinks to skills from all sources. On each call, stale symlinks are removed
// and the full set is rebuilt. Real directories (user-created skills at the
// workspace level) are preserved.
//
// Precedence (highest first):
//  1. Workspace root — real directories in .claude/skills/ (never touched)
//  2. Nested repos — <workspace>/<repo>/.claude/skills/*
//  3. Shared flow skills — ~/.flow/agents/claude/skills/*
func consolidateSkills(wsDir, agentsDir string, st *state.State) error {
	skillsDir := filepath.Join(wsDir, ".claude", "skills")

	// If .claude/skills is a symlink (legacy), remove it so we can create a real dir
	if target, err := os.Readlink(skillsDir); err == nil {
		_ = target
		if err := os.Remove(skillsDir); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return err
	}

	// Remove existing symlinks (rebuilt below). Preserve real directories.
	if err := removeSkillSymlinks(skillsDir); err != nil {
		return err
	}

	// Track which skill names are claimed (real dirs already win)
	claimed := make(map[string]bool)
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		claimed[e.Name()] = true
	}

	// Source 1: Shared flow skills (lowest precedence, added first so repos can override)
	// We add these first, then repos override by removing and re-creating symlinks.
	// Actually — repos have higher precedence, so add shared first, then let repos
	// skip if already claimed. Since we cleared all symlinks above, we just need to
	// not add shared skills if a repo claims the same name.

	// Collect repo skills first to know what they claim
	repoClaimed := make(map[string]string) // skill name → absolute path
	for _, repo := range st.Spec.Repos {
		repoSkillsDir := filepath.Join(wsDir, state.RepoPath(repo), ".claude", "skills")
		repoEntries, err := os.ReadDir(repoSkillsDir)
		if err != nil {
			continue // repo may not have skills
		}
		for _, e := range repoEntries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if claimed[name] {
				continue // workspace root (real dir) takes precedence
			}
			if _, exists := repoClaimed[name]; !exists {
				// First repo to claim wins (state.yaml order)
				repoClaimed[name] = filepath.Join(repoSkillsDir, name)
			}
		}
	}

	// Add shared skills (skip if claimed by workspace root or repo)
	sharedSkillsDir := filepath.Join(agentsDir, "claude", "skills")
	sharedEntries, err := os.ReadDir(sharedSkillsDir)
	if err == nil {
		for _, e := range sharedEntries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if claimed[name] || repoClaimed[name] != "" {
				continue
			}
			if err := os.Symlink(filepath.Join(sharedSkillsDir, name), filepath.Join(skillsDir, name)); err != nil {
				return err
			}
			claimed[name] = true
		}
	}

	// Add repo skills (skip if claimed by workspace root)
	for name, target := range repoClaimed {
		if claimed[name] {
			continue
		}
		if err := os.Symlink(target, filepath.Join(skillsDir, name)); err != nil {
			return err
		}
		claimed[name] = true
	}

	return nil
}

// removeSkillSymlinks removes all symlinks in a directory, preserving real
// directories and files (which represent user-created workspace-level skills).
func removeSkillSymlinks(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.Type()&os.ModeSymlink != 0 {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
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
	b.WriteString(fmt.Sprintf("- Edit state: `flow edit state %s`\n", id))
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
