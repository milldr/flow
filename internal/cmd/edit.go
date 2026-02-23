package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/status"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

// openInEditor opens a file in the user's preferred editor.
func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	parts := strings.Fields(editor)
	editorArgs := make([]string, len(parts)-1, len(parts))
	copy(editorArgs, parts[1:])
	editorArgs = append(editorArgs, path)

	c := exec.Command(parts[0], editorArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}
	return nil
}

func newEditCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Open flow configuration files in editor",
	}

	cmd.AddCommand(newEditStateCmd(svc))
	cmd.AddCommand(newEditConfigCmd(cfg))
	cmd.AddCommand(newEditStatusCmd(svc, cfg))

	return cmd
}

func newEditStateCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "state <workspace>",
		Short:   "Open workspace state file in editor",
		Args:    cobra.ExactArgs(1),
		Example: `  flow edit state calm-delta    # Opens state.yaml in $EDITOR`,
		RunE: func(_ *cobra.Command, args []string) error {
			id, _, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			return openInEditor(svc.Config.StatePath(id))
		},
	}
}

func newEditConfigCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "config",
		Short:   "Open global config file in editor",
		Args:    cobra.NoArgs,
		Example: `  flow edit config    # Opens ~/.flow/config.yaml in $EDITOR`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return openInEditor(cfg.ConfigFile)
		},
	}
}

func newEditStatusCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status [workspace]",
		Short: "Open status spec file in editor",
		Long: `Open a status spec file in your editor.

Without arguments, opens the global status spec (~/.flow/status.yaml).
With a workspace argument, opens the workspace-specific status spec.`,
		Args:    cobra.MaximumNArgs(1),
		Example: "  flow edit status             # Edit global status spec\n  flow edit status vpc-ipv6     # Edit workspace status spec",
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				return editGlobalStatus(cfg)
			}
			return editWorkspaceStatus(svc, cfg, args[0])
		},
	}
}

func editGlobalStatus(cfg *config.Config) error {
	path := cfg.StatusSpecFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := status.Save(path, status.DefaultSpec()); err != nil {
			return fmt.Errorf("creating status spec: %w", err)
		}
		ui.Success("Created status spec at " + path)
	}

	return openInEditor(path)
}

func editWorkspaceStatus(svc *workspace.Service, cfg *config.Config, idOrName string) error {
	id, _, err := resolveWorkspace(svc, idOrName)
	if err != nil {
		return err
	}

	path := cfg.WorkspaceStatusSpecPath(id)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Copy global spec as starting point, or create default.
		var spec *status.Spec
		if globalSpec, loadErr := status.Load(cfg.StatusSpecFile); loadErr == nil {
			spec = globalSpec
		} else {
			spec = status.DefaultSpec()
		}

		if err := status.Save(path, spec); err != nil {
			return fmt.Errorf("creating workspace status spec: %w", err)
		}
		ui.Success("Created workspace status spec at " + path)
	}

	return openInEditor(path)
}
