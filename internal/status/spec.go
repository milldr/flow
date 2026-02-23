package status

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Validation errors for status spec files.
var (
	ErrInvalidAPIVersion = errors.New("apiVersion must be flow/v1")
	ErrInvalidKind       = errors.New("kind must be Status")
	ErrNoStatuses        = errors.New("statuses must not be empty")
	ErrNoDefault         = errors.New("exactly one status must have default: true")
	ErrMultipleDefaults  = errors.New("only one status may have default: true")
	ErrDuplicateName     = errors.New("status names must be unique")
	ErrMissingName       = errors.New("status name is required")
	ErrMissingCheck      = errors.New("non-default status must have a check command")
	ErrSpecNotFound      = errors.New("no status spec found")
)

// Load reads and parses a status spec file from disk.
func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing status spec: %w", err)
	}

	return &s, nil
}

// Save writes a Spec to disk as YAML.
func Save(path string, s *Spec) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling status spec: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// Validate checks that a Spec has all required fields and is well-formed.
func Validate(s *Spec) error {
	if s.APIVersion != "flow/v1" {
		return ErrInvalidAPIVersion
	}
	if s.Kind != "Status" {
		return ErrInvalidKind
	}
	if len(s.Statuses) == 0 {
		return ErrNoStatuses
	}

	seen := make(map[string]bool)
	defaults := 0

	for i, e := range s.Statuses {
		if e.Name == "" {
			return fmt.Errorf("statuses[%d]: %w", i, ErrMissingName)
		}
		if seen[e.Name] {
			return fmt.Errorf("statuses[%d] %q: %w", i, e.Name, ErrDuplicateName)
		}
		seen[e.Name] = true

		if e.Default {
			defaults++
		} else if e.Check == "" {
			return fmt.Errorf("statuses[%d] %q: %w", i, e.Name, ErrMissingCheck)
		}
	}

	if defaults == 0 {
		return ErrNoDefault
	}
	if defaults > 1 {
		return ErrMultipleDefaults
	}

	return nil
}

// LoadWithFallback loads a workspace-level spec if it exists, otherwise
// falls back to the global spec. Returns ErrSpecNotFound if neither exists.
func LoadWithFallback(workspacePath, globalPath string) (*Spec, error) {
	if _, err := os.Stat(workspacePath); err == nil {
		return Load(workspacePath)
	}

	if _, err := os.Stat(globalPath); err == nil {
		return Load(globalPath)
	}

	return nil, ErrSpecNotFound
}

// DefaultSpec returns a starter status spec with a basic PR workflow.
func DefaultSpec() *Spec {
	return &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Statuses: []Entry{
			{
				Name:        "closed",
				Description: "PR merged",
				Check:       `gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state merged --json number | jq -e 'length > 0' > /dev/null 2>&1`,
			},
			{
				Name:        "in-review",
				Description: "PR open",
				Check:       `gh pr list --repo "https://$FLOW_REPO_URL" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length > 0' > /dev/null 2>&1`,
			},
			{
				Name:        "in-progress",
				Description: "Active development",
				Default:     true,
			},
		},
	}
}
