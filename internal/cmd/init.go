package cmd

import (
	"errors"

	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

// ErrMissingWorkspaceName is returned when init is given an empty name.
var ErrMissingWorkspaceName = errors.New("workspace name is required")

func newInitCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a new workspace interactively",
		RunE: func(_ *cobra.Command, _ []string) error {
			name, description, repoEntries, err := ui.InitPrompt()
			if err != nil {
				return err
			}

			if name == "" {
				return ErrMissingWorkspaceName
			}

			var repos []state.Repo
			for _, r := range repoEntries {
				repos = append(repos, state.Repo{
					URL:    r.URL,
					Branch: r.Branch,
					Path:   r.Path,
				})
			}

			// Allow zero repos at init time — user can add via `flow state`
			st := state.NewState(name, description, repos)

			if err := svc.Create(st); err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Created workspace: " + name)
			ui.Printf("   Location: %s\n", svc.Config.WorkspacePath(name))
			ui.Print("")
			ui.Print("Next steps:")
			ui.Printf("  flow state %s     # Edit state file\n", name)
			ui.Printf("  flow render %s    # Create worktrees\n", name)

			return nil
		},
	}
}
