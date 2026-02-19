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
		Use:   "exec <name> -- <command>",
		Short: "Run a command from the workspace directory",
		Args:  cobra.MinimumNArgs(1),
		Example: `  flow exec vpc-ipv6 -- cursor .
  flow exec vpc-ipv6 -- git status
  flow exec vpc-ipv6 -- ls -la`,
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]

			// Verify workspace exists
			if _, err := svc.Find(name); err != nil {
				return err
			}

			// Cobra places args after "--" starting at args[1:]
			cmdArgs := args[1:]
			if len(cmdArgs) == 0 {
				return ErrNoCommand
			}

			wsDir := svc.Config.WorkspacePath(name)

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
