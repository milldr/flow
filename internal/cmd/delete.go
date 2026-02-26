// Package cmd defines all cobra commands for the flow CLI.
package cmd

import (
	"fmt"

	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newDeleteCmd(svc *workspace.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <workspace> [workspace...]",
		Short: "Delete one or more workspaces and their worktrees",
		Args:  cobra.MinimumNArgs(1),
		Example: `  flow delete calm-delta
  flow delete calm-delta warm-brook --force
  flow delete ws1 ws2 ws3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				id, st, err := resolveWorkspace(svc, arg)
				if err != nil {
					return fmt.Errorf("%s: %w", arg, err)
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
						ui.Print("Skipped: " + name)
						continue
					}
				}

				err = ui.RunWithSpinner("Deleting workspace: "+name, func(_ func(string)) error {
					return svc.Delete(cmd.Context(), id)
				})
				if err != nil {
					return fmt.Errorf("deleting %s: %w", name, err)
				}

				ui.Success("Deleted workspace: " + name)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
