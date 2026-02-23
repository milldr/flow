package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
)

// WorkspaceOption represents a workspace choice for interactive selection.
type WorkspaceOption struct {
	ID      string
	Name    string
	Created time.Time
}

// SelectWorkspace prompts the user to choose among multiple matching workspaces.
// Returns the selected workspace's ID.
func SelectWorkspace(matches []WorkspaceOption) (string, error) {
	options := make([]huh.Option[string], len(matches))
	for i, m := range matches {
		label := fmt.Sprintf("%s (%s) — created %s", m.Name, m.ID, RelativeTime(m.Created))
		options[i] = huh.NewOption(label, m.ID)
	}

	var selected string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Multiple workspaces match. Select one:").
				Options(options...).
				Value(&selected),
		),
	).Run()
	return selected, err
}

// AgentOption represents an agent choice for interactive selection.
type AgentOption struct {
	Name string
	Exec string
}

// SelectAgent prompts the user to choose among multiple configured agents.
// Returns the selected agent's exec command.
func SelectAgent(agents []AgentOption) (string, error) {
	options := make([]huh.Option[string], len(agents))
	for i, a := range agents {
		options[i] = huh.NewOption(a.Name, a.Exec)
	}

	var selected string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an agent:").
				Options(options...).
				Value(&selected),
		),
	).Run()
	return selected, err
}

// DeleteRepo holds repo display info for the delete confirmation prompt.
type DeleteRepo struct {
	Path   string
	Branch string
}

// ConfirmDelete prompts the user to confirm workspace deletion.
func ConfirmDelete(name, id string, repos []DeleteRepo) (bool, error) {
	title := fmt.Sprintf("Delete workspace '%s'", name)
	if name != id {
		title += fmt.Sprintf(" (%s)", id)
	}
	title += "?"

	Warning(title)
	if len(repos) > 0 {
		Print("")
		Print("  Repos:")
		for _, r := range repos {
			Printf("    %s (%s)\n", r.Path, r.Branch)
		}
	}
	Print("")

	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirm?").
				Value(&confirm),
		),
	).Run()
	return confirm, err
}
