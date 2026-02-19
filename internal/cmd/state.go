package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newStateCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "state <name>",
		Short:   "Open workspace state file in editor",
		Args:    cobra.ExactArgs(1),
		Example: `  flow state vpc-ipv6    # Opens state.yaml in $EDITOR`,
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]

			// Verify workspace exists
			if _, err := svc.Find(name); err != nil {
				return err
			}

			statePath := svc.Config.StatePath(name)
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}

			c := exec.Command(editor, statePath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			if err := c.Run(); err != nil {
				return fmt.Errorf("editor exited with error: %w", err)
			}
			return nil
		},
	}
}
