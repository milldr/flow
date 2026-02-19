// Package state handles workspace state file loading, saving, and validation.
package state

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Validation errors for state files.
var (
	ErrInvalidAPIVersion = errors.New("apiVersion must be flow/v1")
	ErrInvalidKind       = errors.New("kind must be State")
	ErrMissingName       = errors.New("metadata.name is required")
	ErrMissingRepos      = errors.New("spec.repos must not be empty")
	ErrMissingRepoURL    = errors.New("url is required")
	ErrMissingRepoBranch = errors.New("branch is required")
	ErrMissingRepoPath   = errors.New("path is required")
)

// Load reads and parses a state file from disk.
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	return &s, nil
}

// Save writes a state to disk as YAML.
func Save(path string, s *State) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate checks that a State has all required fields.
func Validate(s *State) error {
	if s.APIVersion != "flow/v1" {
		return ErrInvalidAPIVersion
	}
	if s.Kind != "State" {
		return ErrInvalidKind
	}
	if s.Metadata.Name == "" {
		return ErrMissingName
	}
	if len(s.Spec.Repos) == 0 {
		return ErrMissingRepos
	}

	for i, r := range s.Spec.Repos {
		if r.URL == "" {
			return fmt.Errorf("spec.repos[%d]: %w", i, ErrMissingRepoURL)
		}
		if r.Branch == "" {
			return fmt.Errorf("spec.repos[%d]: %w", i, ErrMissingRepoBranch)
		}
		if r.Path == "" {
			return fmt.Errorf("spec.repos[%d]: %w", i, ErrMissingRepoPath)
		}
	}

	return nil
}
