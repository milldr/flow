package cmd

import (
	"errors"

	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
)

// resolveWorkspace resolves a workspace by ID or name.
// If the name matches multiple workspaces, an interactive selector is shown.
func resolveWorkspace(svc *workspace.Service, idOrName string) (string, *state.State, error) {
	matches, err := svc.Resolve(idOrName)
	if err != nil {
		var ambErr *workspace.AmbiguousNameError
		if !errors.As(err, &ambErr) {
			return "", nil, err
		}

		// Multiple matches — prompt user to select
		options := make([]ui.WorkspaceOption, len(ambErr.Matches))
		for i, m := range ambErr.Matches {
			options[i] = ui.WorkspaceOption{
				ID:      m.ID,
				Name:    m.Name,
				Created: m.Created,
			}
		}

		selectedID, selectErr := ui.SelectWorkspace(options)
		if selectErr != nil {
			return "", nil, selectErr
		}

		st, findErr := svc.Find(selectedID)
		if findErr != nil {
			return "", nil, findErr
		}
		return selectedID, st, nil
	}

	id := matches[0].ID
	st, err := svc.Find(id)
	if err != nil {
		return "", nil, err
	}
	return id, st, nil
}

// workspaceDisplayName returns the name for user-facing output.
// Prefers metadata name if set, otherwise falls back to the ID.
func workspaceDisplayName(id string, st *state.State) string {
	if st.Metadata.Name != "" {
		return st.Metadata.Name
	}
	return id
}
