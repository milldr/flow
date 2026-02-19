package cmd

import (
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newInitCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "init [name]",
		Short: "Create a new empty workspace",
		Long:  "Create a new empty workspace with a generated ID.\nOptionally provide a human-friendly name as the first argument.",
		Args:  cobra.MaximumNArgs(1),
		Example: `  flow init
  flow init my-project`,
		RunE: func(_ *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}

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
			st := state.NewState(name, "", nil)

			err = ui.RunWithSpinner("Creating workspace", func(_ func(string)) error {
				return svc.Create(id, st)
			})
			if err != nil {
				return err
			}

			label := id
			if name != "" {
				label = id + " (" + name + ")"
			}
			ui.Print("")
			ui.Success("Created workspace " + label)
			ref := id
			if name != "" {
				ref = name
			}
			ui.Printf("\n  Next: %s\n", ui.Code("flow state "+ref))

			return nil
		},
	}
}
