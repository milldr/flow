package status

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "status.yaml")

	spec := DefaultSpec()
	if err := Save(path, spec); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.APIVersion != "flow/v1" {
		t.Errorf("APIVersion = %q, want flow/v1", loaded.APIVersion)
	}
	if loaded.Kind != "Status" {
		t.Errorf("Kind = %q, want Status", loaded.Kind)
	}
	if len(loaded.Statuses) != 3 {
		t.Errorf("Statuses count = %d, want 3", len(loaded.Statuses))
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/status.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestValidateValid(t *testing.T) {
	spec := DefaultSpec()
	if err := Validate(spec); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateInvalidAPIVersion(t *testing.T) {
	spec := DefaultSpec()
	spec.APIVersion = "wrong"
	err := Validate(spec)
	if !errors.Is(err, ErrInvalidAPIVersion) {
		t.Errorf("expected ErrInvalidAPIVersion, got %v", err)
	}
}

func TestValidateInvalidKind(t *testing.T) {
	spec := DefaultSpec()
	spec.Kind = "Wrong"
	err := Validate(spec)
	if !errors.Is(err, ErrInvalidKind) {
		t.Errorf("expected ErrInvalidKind, got %v", err)
	}
}

func TestValidateNoStatuses(t *testing.T) {
	spec := &Spec{APIVersion: "flow/v1", Kind: "Status"}
	err := Validate(spec)
	if !errors.Is(err, ErrNoStatuses) {
		t.Errorf("expected ErrNoStatuses, got %v", err)
	}
}

func TestValidateNoDefault(t *testing.T) {
	spec := &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Name: "open", Check: "true"},
		},
	}
	err := Validate(spec)
	if !errors.Is(err, ErrNoDefault) {
		t.Errorf("expected ErrNoDefault, got %v", err)
	}
}

func TestValidateMultipleDefaults(t *testing.T) {
	spec := &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Name: "a", Default: true},
			{Name: "b", Default: true},
		},
	}
	err := Validate(spec)
	if !errors.Is(err, ErrMultipleDefaults) {
		t.Errorf("expected ErrMultipleDefaults, got %v", err)
	}
}

func TestValidateDuplicateName(t *testing.T) {
	spec := &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Name: "open", Check: "true"},
			{Name: "open", Default: true},
		},
	}
	err := Validate(spec)
	if !errors.Is(err, ErrDuplicateName) {
		t.Errorf("expected ErrDuplicateName, got %v", err)
	}
}

func TestValidateMissingName(t *testing.T) {
	spec := &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Check: "true"},
		},
	}
	err := Validate(spec)
	if !errors.Is(err, ErrMissingName) {
		t.Errorf("expected ErrMissingName, got %v", err)
	}
}

func TestValidateMissingCheck(t *testing.T) {
	spec := &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Name: "open"},
			{Name: "default", Default: true},
		},
	}
	err := Validate(spec)
	if !errors.Is(err, ErrMissingCheck) {
		t.Errorf("expected ErrMissingCheck, got %v", err)
	}
}

func TestLoadWithFallbackWorkspaceLevel(t *testing.T) {
	dir := t.TempDir()
	wsPath := filepath.Join(dir, "ws-status.yaml")
	globalPath := filepath.Join(dir, "global-status.yaml")

	wsSpec := &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{Name: "ws-only", Default: true},
		},
	}
	if err := Save(wsPath, wsSpec); err != nil {
		t.Fatal(err)
	}
	if err := Save(globalPath, DefaultSpec()); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadWithFallback(wsPath, globalPath)
	if err != nil {
		t.Fatalf("LoadWithFallback: %v", err)
	}
	if loaded.Statuses[0].Name != "ws-only" {
		t.Errorf("expected workspace-level spec, got %q", loaded.Statuses[0].Name)
	}
}

func TestLoadWithFallbackGlobal(t *testing.T) {
	dir := t.TempDir()
	wsPath := filepath.Join(dir, "ws-status.yaml") // does not exist
	globalPath := filepath.Join(dir, "global-status.yaml")

	if err := Save(globalPath, DefaultSpec()); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadWithFallback(wsPath, globalPath)
	if err != nil {
		t.Fatalf("LoadWithFallback: %v", err)
	}
	if loaded.Statuses[0].Name != "closed" {
		t.Errorf("expected global spec, got %q", loaded.Statuses[0].Name)
	}
}

func TestLoadWithFallbackNotFound(t *testing.T) {
	_, err := LoadWithFallback("/nonexistent/ws.yaml", "/nonexistent/global.yaml")
	if !errors.Is(err, ErrSpecNotFound) {
		t.Errorf("expected ErrSpecNotFound, got %v", err)
	}
}

func TestDefaultSpec(t *testing.T) {
	spec := DefaultSpec()
	if err := Validate(spec); err != nil {
		t.Fatalf("DefaultSpec should be valid: %v", err)
	}
}

func TestSaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new-status.yaml")

	if err := Save(path, DefaultSpec()); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}
