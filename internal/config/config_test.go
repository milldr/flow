package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDefaultPath(t *testing.T) {
	t.Setenv("FLOW_HOME", "")

	cfg, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".flow")
	if cfg.Home != expected {
		t.Errorf("Home = %q, want %q", cfg.Home, expected)
	}
	if cfg.WorkspacesDir != filepath.Join(expected, "workspaces") {
		t.Errorf("WorkspacesDir = %q", cfg.WorkspacesDir)
	}
	if cfg.ReposDir != filepath.Join(expected, "repos") {
		t.Errorf("ReposDir = %q", cfg.ReposDir)
	}
	if cfg.AgentsDir != filepath.Join(expected, "agents") {
		t.Errorf("AgentsDir = %q", cfg.AgentsDir)
	}
	if cfg.ConfigFile != filepath.Join(expected, "config.yaml") {
		t.Errorf("ConfigFile = %q", cfg.ConfigFile)
	}
}

func TestNewFlowHomeOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("FLOW_HOME", dir)

	cfg, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if cfg.Home != dir {
		t.Errorf("Home = %q, want %q", cfg.Home, dir)
	}
}

func TestWorkspacePath(t *testing.T) {
	cfg := &Config{Home: "/test", WorkspacesDir: "/test/workspaces", ReposDir: "/test/repos"}

	got := cfg.WorkspacePath("my-ws")
	if got != "/test/workspaces/my-ws" {
		t.Errorf("WorkspacePath = %q", got)
	}
}

func TestStatePath(t *testing.T) {
	cfg := &Config{Home: "/test", WorkspacesDir: "/test/workspaces", ReposDir: "/test/repos"}

	got := cfg.StatePath("my-ws")
	if got != "/test/workspaces/my-ws/state.yaml" {
		t.Errorf("StatePath = %q", got)
	}
}

func TestBareRepoPath(t *testing.T) {
	cfg := &Config{Home: "/test", WorkspacesDir: "/test/workspaces", ReposDir: "/test/repos"}

	tests := []struct {
		url  string
		want string
	}{
		{"github.com/org/repo", "/test/repos/github.com/org/repo.git"},
		{"gitlab.com/deep/nested/repo", "/test/repos/gitlab.com/deep/nested/repo.git"},
		{"git@github.com:org/repo.git", "/test/repos/git@github.com:org/repo.git"},
	}

	for _, tt := range tests {
		got := cfg.BareRepoPath(tt.url)
		if got != tt.want {
			t.Errorf("BareRepoPath(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestClaudeAgentDir(t *testing.T) {
	cfg := &Config{Home: "/test", AgentsDir: "/test/agents"}
	got := cfg.ClaudeAgentDir()
	if got != "/test/agents/claude" {
		t.Errorf("ClaudeAgentDir = %q, want /test/agents/claude", got)
	}
}

func TestEnsureDirs(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Home:          dir,
		WorkspacesDir: filepath.Join(dir, "workspaces"),
		ReposDir:      filepath.Join(dir, "repos"),
		AgentsDir:     filepath.Join(dir, "agents"),
		ConfigFile:    filepath.Join(dir, "config.yaml"),
	}

	if err := cfg.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	for _, d := range []string{cfg.WorkspacesDir, cfg.ReposDir, cfg.AgentsDir} {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("directory %q not created: %v", d, err)
		} else if !info.IsDir() {
			t.Errorf("%q is not a directory", d)
		}
	}

	// Config file should be created
	if _, err := os.Stat(cfg.ConfigFile); err != nil {
		t.Errorf("config file not created: %v", err)
	}

	// FlowConfig should be loaded
	if cfg.FlowConfig == nil {
		t.Fatal("FlowConfig not loaded")
	}
	if cfg.FlowConfig.APIVersion != "flow/v1" {
		t.Errorf("FlowConfig.APIVersion = %q, want flow/v1", cfg.FlowConfig.APIVersion)
	}

	// Idempotent — calling again should not error
	if err := cfg.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs (2nd call): %v", err)
	}
}

func TestEnsureDirsLoadsExistingConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Home:          dir,
		WorkspacesDir: filepath.Join(dir, "workspaces"),
		ReposDir:      filepath.Join(dir, "repos"),
		AgentsDir:     filepath.Join(dir, "agents"),
		ConfigFile:    filepath.Join(dir, "config.yaml"),
	}

	// Pre-create a config file
	fc := &FlowConfig{APIVersion: "flow/v1", Kind: "Config"}
	if err := SaveFlowConfig(cfg.ConfigFile, fc); err != nil {
		t.Fatal(err)
	}

	if err := cfg.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	if cfg.FlowConfig == nil {
		t.Fatal("FlowConfig not loaded from existing file")
	}
	if cfg.FlowConfig.Kind != "Config" {
		t.Errorf("Kind = %q, want Config", cfg.FlowConfig.Kind)
	}
}
