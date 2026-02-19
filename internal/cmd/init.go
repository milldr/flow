package cmd

import (
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newInitCmd(svc *workspace.Service) *cobra.Command {
	var nameFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new empty workspace",
		Long:  "Create a new empty workspace with a generated ID. Use --name to set a human-friendly name.",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			// Collect existing IDs to avoid collisions
			infos, err := svc.List()
			if err != nil {
				return err
			}
			existingIDs := make([]string, len(infos))
			for i, info := range infos {
				existingIDs[i] = info.ID
			}

			id := workspace.GenerateUniqueID(existingIDs)
			st := state.NewState(nameFlag, "", nil)

			if err := svc.Create(id, st); err != nil {
				return err
			}

			ui.Print("")
			ui.Success("Created workspace: " + id)
			if nameFlag != "" {
				ui.Printf("   Name: %s\n", nameFlag)
			}
			ui.Printf("   Location: %s\n", svc.Config.WorkspacePath(id))
			ui.Print("")
			ui.Print("Next steps:")
			ui.Printf("  flow state %s     # Add repos to state file\n", id)
			ui.Printf("  flow render %s    # Create worktrees\n", id)

			return nil
		},
	}

	cmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Human-friendly workspace name")
	return cmd
}
