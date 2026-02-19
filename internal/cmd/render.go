package cmd

import (
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newRenderCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "render <name>",
		Short:   "Create worktrees from workspace state file",
		Args:    cobra.ExactArgs(1),
		Example: `  flow render vpc-ipv6`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Verify workspace exists before printing header
			if _, err := svc.Find(name); err != nil {
				return err
			}

			ui.Info("Rendering workspace: " + name)
			ui.Print("")

			err := svc.Render(cmd.Context(), name, func(msg string) {
				ui.Printf("  %s\n", msg)
			})
			if err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Workspace ready")
			ui.Print("")
			ui.Printf("  flow exec %s -- cursor .   # Open in editor\n", name)

			return nil
		},
	}
}
