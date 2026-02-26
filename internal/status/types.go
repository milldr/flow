// Package status handles status spec loading, validation, and resolution.
package status

import "time"

// SpecBody holds the statuses list nested under spec.
type SpecBody struct {
	Statuses []Entry `yaml:"statuses,omitempty"`
}

// Spec represents a status spec file.
type Spec struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Spec       SpecBody `yaml:"spec,omitempty"`
}

// Entry defines a single status in the spec.
type Entry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Color       string `yaml:"color,omitempty"`
	Check       string `yaml:"check,omitempty"`
	Default     bool   `yaml:"default,omitempty"`
}

// namedColors maps human-readable color names to ANSI color codes.
var namedColors = map[string]string{
	"black":   "0",
	"red":     "1",
	"green":   "2",
	"yellow":  "3",
	"blue":    "4",
	"magenta": "5",
	"cyan":    "6",
	"white":   "7",
	"gray":    "8",
	"purple":  "93",
}

// ColorCode returns the ANSI color code for an Entry's color.
// Supports named colors (green, yellow, etc.) and raw ANSI codes.
// Returns empty string if no color is set.
func (e Entry) ColorCode() string {
	if e.Color == "" {
		return ""
	}
	if code, ok := namedColors[e.Color]; ok {
		return code
	}
	return e.Color
}

// DisplayOrder returns a map of status name to display sort index.
// The default (least-advanced) status gets index 0, most-advanced gets the highest.
func (s *Spec) DisplayOrder() map[string]int {
	n := len(s.Spec.Statuses)
	order := make(map[string]int, n)
	for i, e := range s.Spec.Statuses {
		order[e.Name] = n - 1 - i
	}
	return order
}

// ColorMap returns a map of status name to ANSI color code.
func (s *Spec) ColorMap() map[string]string {
	colors := make(map[string]string, len(s.Spec.Statuses))
	for _, e := range s.Spec.Statuses {
		if code := e.ColorCode(); code != "" {
			colors[e.Name] = code
		}
	}
	return colors
}

// RepoInfo provides context for running status checks against a repo.
type RepoInfo struct {
	URL    string
	Branch string
	Path   string
}

// RepoResult holds the resolved status for a single repo.
type RepoResult struct {
	URL      string
	Branch   string
	Status   string
	Duration time.Duration
}

// WorkspaceResult holds the resolved status for a workspace.
type WorkspaceResult struct {
	WorkspaceID   string
	WorkspaceName string
	Status        string
	Duration      time.Duration
	Repos         []RepoResult
}
