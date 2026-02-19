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

// ConfirmDelete prompts the user to confirm workspace deletion.
func ConfirmDelete(name string, repos []string) (bool, error) {
	Warning("Delete workspace '" + name + "'?")
	Print("    This will remove:")
	for _, r := range repos {
		Printf("    - %s (worktree)\n", r)
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
