package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/milldr/flow/internal/agents"
	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/git"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	verbose bool
)

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
}

func newLogger() *slog.Logger {
	level := slog.LevelWarn
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:               "flow",
		Short:             "Multi-repo workspace manager using git worktrees",
		Long:              "Flow manages multi-repo development workspaces using git worktrees.\nA YAML manifest defines which repos/branches belong together, and `flow render` materializes the workspace.",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug output")

	// Build dependencies eagerly. The logger is created lazily via
	// PersistentPreRunE so that --verbose is parsed first.
	cfg, err := config.New()
	if err != nil {
		ui.Errorf("%v", err)
		os.Exit(1)
	}

	if err := cfg.EnsureDirs(); err != nil {
		ui.Errorf("creating directories: %v", err)
		os.Exit(1)
	}

	if err := agents.EnsureSharedAgent(cfg.AgentsDir); err != nil {
		ui.Errorf("setting up agent files: %v", err)
		os.Exit(1)
	}

	gitRunner := &git.RealRunner{}
	svc := &workspace.Service{
		Config: cfg,
		Git:    gitRunner,
	}

	// Wire the logger after flags are parsed.
	root.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		if verbose {
			ui.SetPlain(true)
		}
		log := newLogger()
		gitRunner.Log = log
		svc.Log = log
		log.Debug("flow starting", "version", version, "flow_home", cfg.Home)
	}

	// Register commands
	root.AddCommand(newVersionCmd())
	root.AddCommand(newInitCmd(svc))
	root.AddCommand(newListCmd(svc))
	root.AddCommand(newRenderCmd(svc))
	root.AddCommand(newStateCmd(svc))
	root.AddCommand(newExecCmd(svc))
	root.AddCommand(newDeleteCmd(svc))

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("flow " + version)
		},
	}
}

// NewRootCmd returns the root command for use by doc generation tools.
func NewRootCmd() *cobra.Command {
	return newRootCmd()
}

// Execute runs the root command.
func Execute() error {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		ui.Errorf("%v", err)
		return err
	}
	return nil
}
