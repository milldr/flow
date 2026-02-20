// Package cmd defines all cobra commands for the flow CLI.
package cmd

import (
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newDeleteCmd(svc *workspace.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <workspace>",
		Short: "Delete a workspace and its worktrees",
		Args:  cobra.ExactArgs(1),
		Example: `  flow delete calm-delta
  flow delete calm-delta --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			name := workspaceDisplayName(id, st)

			if !force {
				repos := make([]ui.DeleteRepo, len(st.Spec.Repos))
				for i, r := range st.Spec.Repos {
					repos[i] = ui.DeleteRepo{
						Path:   state.RepoPath(r),
						Branch: r.Branch,
					}
				}

				confirmed, err := ui.ConfirmDelete(name, id, repos)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Print("Cancelled.")
					return nil
				}
			}

			err = ui.RunWithSpinner("Deleting workspace: "+name, func(_ func(string)) error {
				return svc.Delete(cmd.Context(), id)
			})
			if err != nil {
				return err
			}

			ui.Success("Deleted workspace: " + name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
