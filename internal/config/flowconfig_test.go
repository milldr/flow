package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFlowConfig(t *testing.T) {
	fc := DefaultFlowConfig()
	if fc.APIVersion != "flow/v1" {
		t.Errorf("APIVersion = %q, want flow/v1", fc.APIVersion)
	}
	if fc.Kind != "Config" {
		t.Errorf("Kind = %q, want Config", fc.Kind)
	}
	if len(fc.Spec.Agents) != 1 {
		t.Fatalf("Spec.Agents length = %d, want 1", len(fc.Spec.Agents))
	}
	agent := fc.DefaultAgent()
	if agent == nil {
		t.Fatal("DefaultAgent() should not be nil for default config")
	}
	if agent.Name != "claude" {
		t.Errorf("DefaultAgent().Name = %q, want %q", agent.Name, "claude")
	}
	if agent.Exec != "claude" {
		t.Errorf("DefaultAgent().Exec = %q, want %q", agent.Exec, "claude")
	}
}

func TestFlowConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	fc := DefaultFlowConfig()
	if err := SaveFlowConfig(path, fc); err != nil {
		t.Fatalf("SaveFlowConfig: %v", err)
	}

	loaded, err := LoadFlowConfig(path)
	if err != nil {
		t.Fatalf("LoadFlowConfig: %v", err)
	}

	if loaded.APIVersion != fc.APIVersion {
		t.Errorf("APIVersion = %q, want %q", loaded.APIVersion, fc.APIVersion)
	}
	if loaded.Kind != fc.Kind {
		t.Errorf("Kind = %q, want %q", loaded.Kind, fc.Kind)
	}
}

func TestFlowConfigAgentsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	fc := DefaultFlowConfig()
	fc.Spec.Agents = []Agent{
		{Name: "claude", Exec: "claude", Default: true},
		{Name: "cursor", Exec: "cursor ."},
	}
	if err := SaveFlowConfig(path, fc); err != nil {
		t.Fatalf("SaveFlowConfig: %v", err)
	}

	loaded, err := LoadFlowConfig(path)
	if err != nil {
		t.Fatalf("LoadFlowConfig: %v", err)
	}

	if len(loaded.Spec.Agents) != 2 {
		t.Fatalf("Spec.Agents length = %d, want 2", len(loaded.Spec.Agents))
	}
	agent := loaded.DefaultAgent()
	if agent == nil {
		t.Fatal("DefaultAgent() returned nil")
	}
	if agent.Exec != "claude" {
		t.Errorf("DefaultAgent().Exec = %q, want %q", agent.Exec, "claude")
	}
}

func TestDefaultAgentNoDefault(t *testing.T) {
	fc := DefaultFlowConfig()
	fc.Spec.Agents = []Agent{
		{Name: "claude", Exec: "claude"},
	}
	if fc.DefaultAgent() != nil {
		t.Error("DefaultAgent() should be nil when no agent is marked default")
	}
}

func TestLoadFlowConfigNotFound(t *testing.T) {
	_, err := LoadFlowConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadFlowConfigMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(":\ninvalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Malformed YAML should still parse (yaml.v3 is lenient with this input)
	// but truly invalid YAML will error
	path2 := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path2, []byte("{{{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFlowConfig(path2)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}
