package cmd

import (
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newRenderCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "render <workspace>",
		Short:   "Create worktrees from workspace state file",
		Args:    cobra.ExactArgs(1),
		Example: `  flow render calm-delta`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			ui.Info("Rendering workspace: " + workspaceDisplayName(id, st))
			ui.Print("")

			err = svc.Render(cmd.Context(), id, func(msg string) {
				ui.Printf("  %s\n", msg)
			})
			if err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Workspace ready")
			ui.Print("")
			name := workspaceDisplayName(id, st)
			ui.Printf("  flow exec %s -- cursor .   # Open in editor\n", name)

			return nil
		},
	}
}
