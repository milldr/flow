package cmd

import (
	"os"
	"os/exec"

	"github.com/milldr/flow/internal/iterm"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newOpenCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "open <workspace>",
		Short:   "Open a shell in the workspace directory",
		Args:    cobra.ExactArgs(1),
		Example: "  flow open calm-delta",
		RunE: func(_ *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			wsDir := svc.Config.WorkspacePath(id)

			iterm.SetTabColor(id)
			iterm.SetTabTitle(workspaceDisplayName(id, st))

			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}

			c := exec.Command(shell)
			c.Dir = wsDir
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			return c.Run()
		},
	}
}
