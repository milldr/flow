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
	if len(s.Spec.Statuses) == 0 {
		return ErrNoStatuses
	}

	seen := make(map[string]bool)
	defaults := 0

	for i, e := range s.Spec.Statuses {
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
		Spec: SpecBody{
			Statuses: []Entry{
				{
					Name:        "closed",
					Description: "All PRs merged or closed",
					Check:       `gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state merged --json number | jq -e 'length > 0' > /dev/null 2>&1 && gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length == 0' > /dev/null 2>&1`,
				},
				{
					Name:        "stale",
					Description: "Workspace inactive",
					Check:       "false",
				},
				{
					Name:        "in-review",
					Description: "Non-draft PR open",
					Check:       `gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state open --json isDraft | jq -e 'map(select(.isDraft == false)) | length > 0' > /dev/null 2>&1`,
				},
				{
					Name:        "in-progress",
					Description: "Local diffs or draft PR",
					Check:       `git -C "$FLOW_REPO_PATH" status --porcelain 2>/dev/null | grep -q . || { _r=$(git ls-remote "$(git -C "$FLOW_REPO_PATH" remote get-url origin 2>/dev/null)" "$FLOW_REPO_BRANCH" 2>/dev/null | cut -f1) && [ -n "$_r" ] && [ "$(git -C "$FLOW_REPO_PATH" rev-parse HEAD 2>/dev/null)" != "$_r" ] && git -C "$FLOW_REPO_PATH" cat-file -e "$_r" 2>/dev/null && if git -C "$FLOW_REPO_PATH" merge-base --is-ancestor HEAD "$_r" 2>/dev/null; then false; fi; } || gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state open --json isDraft | jq -e 'map(select(.isDraft)) | length > 0' > /dev/null 2>&1`,
				},
				{
					Name:        "open",
					Description: "Workspace created, no changes yet",
					Default:     true,
				},
			},
		},
	}
}
