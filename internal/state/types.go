package state

import "time"

// State represents a workspace state file.
type State struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata contains workspace identification.
type Metadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Created     string `yaml:"created"`
}

// Spec defines the workspace contents.
type Spec struct {
	Repos []Repo `yaml:"repos"`
}

// Repo defines a single repository in the workspace.
type Repo struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
	Path   string `yaml:"path"`
}

// NewState creates a State with defaults filled in.
func NewState(name, description string, repos []Repo) *State {
	return &State{
		APIVersion: "flow/v1",
		Kind:       "State",
		Metadata: Metadata{
			Name:        name,
			Description: description,
			Created:     time.Now().UTC().Format(time.RFC3339),
		},
		Spec: Spec{
			Repos: repos,
		},
	}
}
