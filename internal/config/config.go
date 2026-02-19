// Package config provides path resolution for Flow's directory structure.
package config

import (
	"os"
	"path/filepath"
)

// Config holds resolved paths for Flow's directory structure.
type Config struct {
	Home          string // Root directory (~/.flow or $FLOW_HOME)
	WorkspacesDir string // ~/.flow/workspaces/
	ReposDir      string // ~/.flow/repos/
}

// New creates a Config with resolved paths.
// Respects $FLOW_HOME if set, otherwise defaults to ~/.flow.
func New() (*Config, error) {
	home := os.Getenv("FLOW_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		home = filepath.Join(userHome, ".flow")
	}

	return &Config{
		Home:          home,
		WorkspacesDir: filepath.Join(home, "workspaces"),
		ReposDir:      filepath.Join(home, "repos"),
	}, nil
}

// WorkspacePath returns the path for a named workspace.
func (c *Config) WorkspacePath(name string) string {
	return filepath.Join(c.WorkspacesDir, name)
}

// StatePath returns the state.yaml path for a named workspace.
func (c *Config) StatePath(name string) string {
	return filepath.Join(c.WorkspacesDir, name, "state.yaml")
}

// BareRepoPath returns the bare clone path for a repo URL.
// e.g., github.com/org/repo → ~/.flow/repos/github.com/org/repo.git
func (c *Config) BareRepoPath(repoURL string) string {
	return filepath.Join(c.ReposDir, repoURL+".git")
}

// EnsureDirs creates the top-level directories if they don't exist.
func (c *Config) EnsureDirs() error {
	for _, dir := range []string{c.WorkspacesDir, c.ReposDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
