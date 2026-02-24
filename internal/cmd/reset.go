package cmd

import (
	"fmt"

	"github.com/milldr/flow/internal/agents"
	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/status"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newResetCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset a config file to its default value",
		Long:  "Reset status, config, or state files to their default values.",
	}

	cmd.AddCommand(newResetStatusCmd(cfg))
	cmd.AddCommand(newResetConfigCmd(cfg))
	cmd.AddCommand(newResetStateCmd(svc, cfg))
	cmd.AddCommand(newResetSkillsCmd(cfg))

	return cmd
}

func newResetStatusCmd(cfg *config.Config) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Reset the global status spec to its default",
		Example: "  flow reset status\n  flow reset status --force",
		RunE: func(_ *cobra.Command, _ []string) error {
			path := cfg.StatusSpecFile

			if !force {
				confirmed, err := ui.ConfirmReset(path)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Print("Cancelled.")
					return nil
				}
			}

			if err := status.Save(path, status.DefaultSpec()); err != nil {
				return fmt.Errorf("resetting status spec: %w", err)
			}

			ui.Success("Reset status spec at " + path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func newResetConfigCmd(cfg *config.Config) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Reset the global config to its default",
		Example: "  flow reset config\n  flow reset config --force",
		RunE: func(_ *cobra.Command, _ []string) error {
			path := cfg.ConfigFile

			if !force {
				confirmed, err := ui.ConfirmReset(path)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Print("Cancelled.")
					return nil
				}
			}

			if err := config.SaveFlowConfig(path, config.DefaultFlowConfig()); err != nil {
				return fmt.Errorf("resetting config: %w", err)
			}

			ui.Success("Reset config at " + path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func newResetStateCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "state <workspace>",
		Short:   "Reset a workspace state file to its default",
		Args:    cobra.ExactArgs(1),
		Example: "  flow reset state vpc-ipv6\n  flow reset state vpc-ipv6 --force",
		RunE: func(_ *cobra.Command, args []string) error {
			id, st, err := resolveWorkspace(svc, args[0])
			if err != nil {
				return err
			}

			path := cfg.StatePath(id)

			if !force {
				confirmed, err := ui.ConfirmReset(path)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Print("Cancelled.")
					return nil
				}
			}

			newSt := state.NewState(st.Metadata.Name, st.Metadata.Description, nil)
			if err := state.Save(path, newSt); err != nil {
				return fmt.Errorf("resetting state: %w", err)
			}

			ui.Success("Reset state at " + path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func newResetSkillsCmd(cfg *config.Config) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "skills",
		Short:   "Reset shared agent skills to their defaults",
		Example: "  flow reset skills\n  flow reset skills --force",
		RunE: func(_ *cobra.Command, _ []string) error {
			path := cfg.AgentsDir

			if !force {
				confirmed, err := ui.ConfirmReset(path)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Print("Cancelled.")
					return nil
				}
			}

			if err := agents.ResetSharedAgent(path); err != nil {
				return fmt.Errorf("resetting skills: %w", err)
			}

			ui.Success("Reset shared agent skills at " + path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
