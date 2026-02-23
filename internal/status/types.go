// Package status handles status spec loading, validation, and resolution.
package status

// Spec represents a status spec file.
type Spec struct {
	APIVersion string  `yaml:"apiVersion"`
	Kind       string  `yaml:"kind"`
	Statuses   []Entry `yaml:"statuses"`
}

// Entry defines a single status in the spec.
type Entry struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Check       string `yaml:"check,omitempty"`
	Default     bool   `yaml:"default,omitempty"`
}

// RepoInfo provides context for running status checks against a repo.
type RepoInfo struct {
	URL    string
	Branch string
	Path   string
}

// RepoResult holds the resolved status for a single repo.
type RepoResult struct {
	URL    string
	Branch string
	Status string
}

// WorkspaceResult holds the resolved status for a workspace.
type WorkspaceResult struct {
	WorkspaceID   string
	WorkspaceName string
	Status        string
	Repos         []RepoResult
}
