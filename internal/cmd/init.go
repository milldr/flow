package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/milldr/flow/internal/iterm"
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newInitCmd(svc *workspace.Service) *cobra.Command {
	var noExec bool

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create a new empty workspace",
		Long:  "Create a new empty workspace with a generated ID.\nOptionally provide a human-friendly name as the first argument.\nBy default, launches the configured agent in the new workspace.",
		Args:  cobra.MaximumNArgs(1),
		Example: `  flow init
  flow init my-project
  flow init my-project --no-exec`,
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

			// Try to launch the configured agent unless --no-exec was passed.
			if !noExec {
				cmdArgs, agentErr := resolveAgentCmd(svc)
				if agentErr == nil {
					wsDir := svc.Config.WorkspacePath(id)

					iterm.SetTabColor(id)
					iterm.SetTabTitle(workspaceDisplayName(id, st))

					c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
					c.Dir = wsDir
					c.Stdin = os.Stdin
					c.Stdout = os.Stdout
					c.Stderr = os.Stderr

					return c.Run()
				}

				// If the error is anything other than "no agents configured",
				// surface it so the user knows something went wrong.
				if !errors.Is(agentErr, ErrNoAgent) {
					return agentErr
				}
			}

			// Fallback: print manual next-step instructions.
			ref := id
			if name != "" {
				ref = name
			}
			ui.Printf("\n  Next: %s\n", ui.Code("flow edit state "+ref))

			return nil
		},
	}

	cmd.Flags().BoolVar(&noExec, "no-exec", false, "skip launching the agent after creation")

	return cmd
}
