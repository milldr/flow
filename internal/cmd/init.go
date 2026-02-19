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

			err = ui.RunWithSpinner("Creating workspace", func(_ func(string)) error {
				return svc.Create(id, st)
			})
			if err != nil {
				return err
			}

			label := id
			if nameFlag != "" {
				label = id + " (" + nameFlag + ")"
			}
			ui.Print("")
			ui.Success("Created workspace " + label)
			ref := id
			if nameFlag != "" {
				ref = nameFlag
			}
			ui.Printf("\n  Next: %s\n", ui.Code("flow state "+ref))

			return nil
		},
	}

	cmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Human-friendly workspace name")
	return cmd
}
