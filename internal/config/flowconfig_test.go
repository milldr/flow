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
