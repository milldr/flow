// Package cmd defines all cobra commands for the flow CLI.
package cmd

import (
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newDeleteCmd(svc *workspace.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a workspace and its worktrees",
		Args:  cobra.ExactArgs(1),
		Example: `  flow delete vpc-ipv6
  flow delete vpc-ipv6 --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			st, err := svc.Find(name)
			if err != nil {
				return err
			}

			if !force {
				var repoPaths []string
				for _, r := range st.Spec.Repos {
					repoPaths = append(repoPaths, r.Path)
				}

				confirmed, err := ui.ConfirmDelete(name, repoPaths)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Print("Cancelled.")
					return nil
				}
			}

			if err := svc.Delete(cmd.Context(), name); err != nil {
				return err
			}

			ui.Success("Deleted workspace: " + name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
