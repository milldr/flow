package cmd

import (
	"fmt"

	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newRenderCmd(svc *workspace.Service) *cobra.Command {
	var reset bool

	cmd := &cobra.Command{
		Use:     "render <workspace>",
		Short:   "Create worktrees from workspace state file",
		Args:    cobra.ExactArgs(1),
		Example: "  flow render calm-delta\n  flow render calm-delta --reset=false   # Use existing remote branches instead of creating fresh",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			name := workspaceDisplayName(id, st)

			opts := &workspace.RenderOptions{}
			if reset {
				opts.OnBranchConflict = workspace.BranchConflictReset
			} else {
				opts.OnBranchConflict = workspace.BranchConflictUseExisting
			}

			err = ui.RunWithSpinner("Rendering workspace: "+name, func(report func(string)) error {
				return svc.Render(cmd.Context(), id, report, opts)
			})
			if err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Workspace ready")
			ui.Print("")
			if len(svc.Config.FlowConfig.Spec.Agents) > 0 {
				ui.Printf("  %s\n", ui.Code("flow exec "+name))
			} else {
				ui.Printf("  %s\n", ui.Code(fmt.Sprintf("flow exec %s -- <command>", name)))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&reset, "reset", true, "Reset existing branches to fresh state from default branch")
	return cmd
}
