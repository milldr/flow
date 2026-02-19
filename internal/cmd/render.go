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

			name := workspaceDisplayName(id, st)

			err = ui.RunWithSpinner("Rendering workspace: "+name, func(report func(string)) error {
				return svc.Render(cmd.Context(), id, report)
			})
			if err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Workspace ready")
			ui.Print("")
			ui.Printf("  %s\n", ui.Code("flow exec "+name+" -- <command>"))
			ui.Printf("  %s\n", ui.Code("flow exec "+name+" -- cursor ."))
			ui.Printf("  %s\n", ui.Code("flow exec "+name+" -- claude"))

			return nil
		},
	}
}
