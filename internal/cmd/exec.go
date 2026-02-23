package cmd

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/milldr/flow/internal/iterm"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

// ErrNoAgent is returned when no agents are configured and no command is given.
var ErrNoAgent = errors.New("no agents configured; specify a command with -- or add agents to config")

func newExecCmd(svc *workspace.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec <workspace> [-- <command>]",
		Short: "Run a command from the workspace directory",
		Args:  cobra.MinimumNArgs(1),
		Example: `  flow exec calm-delta
  flow exec calm-delta -- cursor .
  flow exec calm-delta -- git status`,
		RunE: func(_ *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			// Args after "--" are the explicit command.
			cmdArgs := args[1:]

			// No explicit command — resolve from configured agents.
			if len(cmdArgs) == 0 {
				cmdArgs, err = resolveAgentCmd(svc)
				if err != nil {
					return err
				}
			}

			wsDir := svc.Config.WorkspacePath(id)

			iterm.SetTabColor(id)
			iterm.SetTabTitle(workspaceDisplayName(id, st))

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

// resolveAgentCmd picks the command to run from configured agents.
// One agent → use it directly. Multiple → prompt the user to choose.
func resolveAgentCmd(svc *workspace.Service) ([]string, error) {
	agents := svc.Config.FlowConfig.Spec.Agents
	if len(agents) == 0 {
		return nil, ErrNoAgent
	}

	var execStr string
	if len(agents) == 1 {
		execStr = agents[0].Exec
	} else {
		options := make([]ui.AgentOption, len(agents))
		for i, a := range agents {
			options[i] = ui.AgentOption{Name: a.Name, Exec: a.Exec}
		}
		selected, err := ui.SelectAgent(options)
		if err != nil {
			return nil, err
		}
		execStr = selected
	}

	return strings.Fields(execStr), nil
}
