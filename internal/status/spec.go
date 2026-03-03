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
//
// Status checks are evaluated top-to-bottom per repo; first match wins.
// The workspace-level status is the least-advanced (highest index) across repos.
//
// Default statuses:
//
//	closed       — a merged PR exists and no open PRs remain (requires gh)
//	stale        — no commits on the branch in the last 14 days (local git only)
//	in-review    — a non-draft open PR exists on the branch (requires gh)
//	in-progress  — uncommitted changes, unpushed commits, or a draft PR
//	open         — default; none of the above matched
func DefaultSpec() *Spec {
	return &Spec{
		APIVersion: "flow/v1",
		Kind:       "Status",
		Spec: SpecBody{
			Skip: `_default=$(git -C "$FLOW_REPO_PATH" symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's|refs/remotes/origin/||')
[ -z "$_default" ] && _default=$(git -C "$FLOW_REPO_PATH" ls-remote --symref origin HEAD 2>/dev/null | awk '/^ref:/ { sub("refs/heads/", "", $2); print $2 }')
[ -n "$_default" ] && [ "$FLOW_REPO_BRANCH" = "$_default" ]`,
			Statuses: []Entry{
				{
					Name:        "closed",
					Description: "Merged PR, no open PRs remaining",
					Color:       "131",
					Check:       `gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state merged --json number | jq -e 'length > 0' > /dev/null 2>&1 && gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state open --json number | jq -e 'length == 0' > /dev/null 2>&1`,
				},
				{
					Name:        "stale",
					Description: "No commits on branch in last 14 days",
					Color:       "magenta",
					Check:       `_now=$(date +%s) && _last=$(git -C "$FLOW_REPO_PATH" log -1 --format=%ct 2>/dev/null) && [ -n "$_last" ] && [ $((_now - _last)) -gt 1209600 ]`,
				},
				{
					Name:        "in-review",
					Description: "Non-draft open PR on branch",
					Color:       "purple",
					Check:       `gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state open --json isDraft | jq -e 'map(select(.isDraft == false)) | length > 0' > /dev/null 2>&1`,
				},
				{
					Name:        "in-progress",
					Description: "Uncommitted changes, unpushed commits, or draft PR",
					Color:       "yellow",
					Check:       `git -C "$FLOW_REPO_PATH" status --porcelain 2>/dev/null | grep -q . || git -C "$FLOW_REPO_PATH" log --oneline "origin/$FLOW_REPO_BRANCH..HEAD" 2>/dev/null | grep -q . || gh pr list --repo "$FLOW_REPO_SLUG" --head "$FLOW_REPO_BRANCH" --state open --json isDraft | jq -e 'map(select(.isDraft)) | length > 0' > /dev/null 2>&1`,
				},
				{
					Name:        "open",
					Description: "No changes detected",
					Color:       "green",
					Default:     true,
				},
			},
		},
	}
}
