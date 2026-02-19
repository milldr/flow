package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoundTrip(t *testing.T) {
	s := NewState("test-ws", "A test workspace", []Repo{
		{URL: "github.com/org/repo-a", Branch: "main", Path: "./repo-a"},
		{URL: "github.com/org/repo-b", Branch: "feat/x", Path: "./repo-b"},
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	if err := Save(path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.APIVersion != "flow/v1" {
		t.Errorf("APIVersion = %q, want flow/v1", loaded.APIVersion)
	}
	if loaded.Kind != "State" {
		t.Errorf("Kind = %q, want State", loaded.Kind)
	}
	if loaded.Metadata.Name != "test-ws" {
		t.Errorf("Name = %q, want test-ws", loaded.Metadata.Name)
	}
	if loaded.Metadata.Description != "A test workspace" {
		t.Errorf("Description = %q, want 'A test workspace'", loaded.Metadata.Description)
	}
	if loaded.Metadata.Created == "" {
		t.Error("Created should not be empty")
	}
	if _, err := time.Parse(time.RFC3339, loaded.Metadata.Created); err != nil {
		t.Errorf("Created is not valid RFC3339: %q", loaded.Metadata.Created)
	}
	if len(loaded.Spec.Repos) != 2 {
		t.Fatalf("Repos count = %d, want 2", len(loaded.Spec.Repos))
	}
	if loaded.Spec.Repos[0].URL != "github.com/org/repo-a" {
		t.Errorf("Repos[0].URL = %q, want github.com/org/repo-a", loaded.Spec.Repos[0].URL)
	}
	if loaded.Spec.Repos[1].Branch != "feat/x" {
		t.Errorf("Repos[1].Branch = %q, want feat/x", loaded.Spec.Repos[1].Branch)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		state   *State
		wantErr bool
	}{
		{
			name: "valid",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{Repos: []Repo{{URL: "u", Branch: "b", Path: "p"}}},
			},
		},
		{
			name: "bad api version",
			state: &State{
				APIVersion: "flow/v2",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{Repos: []Repo{{URL: "u", Branch: "b", Path: "p"}}},
			},
			wantErr: true,
		},
		{
			name: "bad kind",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "Workspace",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{Repos: []Repo{{URL: "u", Branch: "b", Path: "p"}}},
			},
			wantErr: true,
		},
		{
			name: "empty name is valid",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{},
				Spec:       Spec{Repos: []Repo{{URL: "u", Branch: "b", Path: "p"}}},
			},
			wantErr: false,
		},
		{
			name: "no repos",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{},
			},
			wantErr: true,
		},
		{
			name: "repo missing url",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{Repos: []Repo{{Branch: "b", Path: "p"}}},
			},
			wantErr: true,
		},
		{
			name: "repo missing branch",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{Repos: []Repo{{URL: "u", Path: "p"}}},
			},
			wantErr: true,
		},
		{
			name: "repo missing path",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec:       Spec{Repos: []Repo{{URL: "u", Branch: "b"}}},
			},
			wantErr: true,
		},
		{
			name: "second repo invalid",
			state: &State{
				APIVersion: "flow/v1",
				Kind:       "State",
				Metadata:   Metadata{Name: "ws"},
				Spec: Spec{Repos: []Repo{
					{URL: "u", Branch: "b", Path: "p"},
					{URL: "", Branch: "b", Path: "p"},
				}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/state.yaml")
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist, got %v", err)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(path, []byte(":\n\t invalid: [yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := Load(path)
	if err != nil {
		t.Fatalf("Load empty file: %v", err)
	}
	// Empty file produces zero-value State
	if s.APIVersion != "" {
		t.Errorf("APIVersion = %q, want empty", s.APIVersion)
	}
}

func TestNewStateCreatedTimestamp(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	s := NewState("ts-test", "Test", nil)
	after := time.Now().UTC().Add(time.Second)

	created, err := time.Parse(time.RFC3339, s.Metadata.Created)
	if err != nil {
		t.Fatalf("Created is not valid RFC3339: %q", s.Metadata.Created)
	}
	if created.Before(before) || created.After(after) {
		t.Errorf("Created %v not within expected range [%v, %v]", created, before, after)
	}
}
