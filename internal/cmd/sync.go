package cmd

import (
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newSyncCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "sync <workspace>",
		Short:   "Fetch and rebase worktrees onto their base branches",
		Args:    cobra.ExactArgs(1),
		Example: `  flow sync calm-delta`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			name := workspaceDisplayName(id, st)

			err = ui.RunWithSpinner("Syncing workspace: "+name, func(report func(string)) error {
				return svc.Sync(cmd.Context(), id, report)
			})
			if err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Workspace synced")
			return nil
		},
	}
}
