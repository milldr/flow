package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newStateCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "state <workspace>",
		Short:   "Open workspace state file in editor",
		Args:    cobra.ExactArgs(1),
		Example: `  flow state calm-delta    # Opens state.yaml in $EDITOR`,
		RunE: func(_ *cobra.Command, args []string) error {
			id, _, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			statePath := svc.Config.StatePath(id)
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}

			// Split EDITOR to handle values with arguments (e.g. "code --wait")
			parts := strings.Fields(editor)
			editorArgs := make([]string, len(parts)-1, len(parts))
			copy(editorArgs, parts[1:])
			editorArgs = append(editorArgs, statePath)
			c := exec.Command(parts[0], editorArgs...)
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
