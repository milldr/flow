package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/status"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newStatusCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	var initFlag bool

	cmd := &cobra.Command{
		Use:   "status [workspace]",
		Short: "Show workspace status",
		Long: `Show the resolved status of workspaces.

Without arguments, shows all workspaces with their statuses.
With a workspace argument, shows a detailed per-repo status breakdown.`,
		Args:    cobra.MaximumNArgs(1),
		Example: "  flow status                  # Show all workspace statuses\n  flow status vpc-ipv6          # Show per-repo breakdown\n  flow status --init            # Create starter status spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			if initFlag {
				return runStatusInit(cfg)
			}

			if len(args) == 0 {
				return runStatusAll(cmd.Context(), svc, cfg)
			}
			return runStatusWorkspace(cmd.Context(), svc, cfg, args[0])
		},
	}

	cmd.Flags().BoolVar(&initFlag, "init", false, "Create a starter status spec file")

	return cmd
}

func runStatusInit(cfg *config.Config) error {
	path := cfg.StatusSpecFile

	if err := status.Save(path, status.DefaultSpec()); err != nil {
		return fmt.Errorf("creating status spec: %w", err)
	}

	ui.Success("Created status spec at " + path)
	return openInEditor(path)
}

func runStatusAll(ctx context.Context, svc *workspace.Service, cfg *config.Config) error {
	infos, err := svc.List()
	if err != nil {
		return err
	}

	if len(infos) == 0 {
		ui.Print("No workspaces found. Run " + ui.Code("flow init") + " to create one.")
		return nil
	}

	resolver := &status.Resolver{Runner: &status.ShellRunner{}}

	type wsResult struct {
		info   workspace.Info
		status string
	}

	results := make([]wsResult, len(infos))

	err = ui.RunWithSpinner("Resolving workspace statuses...", func(report func(string)) error {
		for i, info := range infos {
			report(workspaceDisplayName(info.ID, stateFromInfo(info)))

			spec, specErr := status.LoadWithFallback(
				cfg.WorkspaceStatusSpecPath(info.ID),
				cfg.StatusSpecFile,
			)
			if specErr != nil {
				if errors.Is(specErr, status.ErrSpecNotFound) {
					results[i] = wsResult{info: info, status: "-"}
					continue
				}
				return fmt.Errorf("loading status spec for %s: %w", info.ID, specErr)
			}

			st, findErr := svc.Find(info.ID)
			if findErr != nil {
				results[i] = wsResult{info: info, status: "-"}
				continue
			}

			repos := stateReposToInfo(st)
			wsName := info.Name
			if wsName == "" {
				wsName = info.ID
			}
			result := resolver.ResolveWorkspace(ctx, spec, repos, info.ID, wsName)
			results[i] = wsResult{info: info, status: result.Status}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, status.ErrSpecNotFound) {
			ui.Print("No status spec found. Run " + ui.Code("flow status --init") + " to create one.")
			return nil
		}
		return err
	}

	headers := []string{"NAME", "STATUS", "REPOS", "CREATED"}
	var rows [][]string
	for _, r := range results {
		name := r.info.Name
		if name == "" {
			name = r.info.ID
		}
		rows = append(rows, []string{
			name,
			r.status,
			fmt.Sprintf("%d", r.info.RepoCount),
			ui.RelativeTime(r.info.Created),
		})
	}

	fmt.Println(ui.Table(headers, rows))
	return nil
}

func runStatusWorkspace(ctx context.Context, svc *workspace.Service, cfg *config.Config, idOrName string) error {
	id, st, err := resolveWorkspace(svc, idOrName)
	if err != nil {
		return err
	}

	spec, err := status.LoadWithFallback(
		cfg.WorkspaceStatusSpecPath(id),
		cfg.StatusSpecFile,
	)
	if err != nil {
		if errors.Is(err, status.ErrSpecNotFound) {
			ui.Print("No status spec found. Run " + ui.Code("flow status --init") + " to create one.")
			return nil
		}
		return err
	}

	repos := stateReposToInfo(st)
	wsName := workspaceDisplayName(id, st)

	var result *status.WorkspaceResult
	err = ui.RunWithSpinner("Resolving status for "+wsName+"...", func(_ func(string)) error {
		resolver := &status.Resolver{Runner: &status.ShellRunner{}}
		result = resolver.ResolveWorkspace(ctx, spec, repos, id, wsName)
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Status: %s\n\n", result.Status)

	if len(result.Repos) > 0 {
		headers := []string{"REPO", "BRANCH", "STATUS"}
		var rows [][]string
		for _, r := range result.Repos {
			rows = append(rows, []string{r.URL, r.Branch, r.Status})
		}
		fmt.Println(ui.Table(headers, rows))
	}

	return nil
}

// stateReposToInfo converts state repos to status RepoInfo slice.
func stateReposToInfo(st *state.State) []status.RepoInfo {
	repos := make([]status.RepoInfo, len(st.Spec.Repos))
	for i, r := range st.Spec.Repos {
		repos[i] = status.RepoInfo{
			URL:    r.URL,
			Branch: r.Branch,
			Path:   state.RepoPath(r),
		}
	}
	return repos
}

// stateFromInfo creates a minimal State from workspace.Info for display helpers.
func stateFromInfo(info workspace.Info) *state.State {
	return &state.State{
		Metadata: state.Metadata{
			Name: info.Name,
		},
	}
}
