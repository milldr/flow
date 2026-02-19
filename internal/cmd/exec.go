package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

// ErrNoCommand is returned when no command is specified after --.
var ErrNoCommand = errors.New("no command specified after '--'")

func newExecCmd(svc *workspace.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <workspace> -- <command>",
		Short: "Run a command from the workspace directory",
		Args:  cobra.MinimumNArgs(1),
		Example: `  flow exec calm-delta -- cursor .
  flow exec calm-delta -- git status
  flow exec calm-delta -- ls -la`,
		RunE: func(_ *cobra.Command, args []string) error {
			id, _, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			// Cobra places args after "--" starting at args[1:]
			cmdArgs := args[1:]
			if len(cmdArgs) == 0 {
				return ErrNoCommand
			}

			wsDir := svc.Config.WorkspacePath(id)

			c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			c.Dir = wsDir
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			return c.Run()
		},
	}

	return cmd
}
