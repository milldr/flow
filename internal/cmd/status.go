package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/milldr/flow/internal/cache"
	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/state"
	"github.com/milldr/flow/internal/status"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newStatusCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [workspace]",
		Short: "Show workspace status",
		Long: `Show the resolved status of workspaces.

Without arguments, shows all workspaces with their statuses.
With a workspace argument, shows a detailed per-repo status breakdown.`,
		Args:    cobra.MaximumNArgs(1),
		Example: "  flow status                  # Show all workspace statuses\n  flow status vpc-ipv6          # Show per-repo breakdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runStatusAll(cmd.Context(), svc, cfg)
			}
			return runStatusWorkspace(cmd.Context(), svc, cfg, args[0])
		},
	}

	return cmd
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

	// Load cached statuses for initial sort ordering.
	cachePath := cfg.StatusCacheFile()
	statusCache := cache.LoadStatus(cachePath)

	// Build rows for the live table.
	rows := make([]ui.StatusRow, len(infos))
	for i, info := range infos {
		name := info.Name
		if name == "" {
			name = info.ID
		}
		rows[i] = ui.StatusRow{
			Name:         name,
			Repos:        fmt.Sprintf("%d", info.RepoCount),
			Created:      ui.RelativeTime(info.Created),
			CreatedAt:    info.Created,
			CachedStatus: statusCache[info.ID].Status,
		}
	}

	// Load the global spec for display config (sort order + colors).
	globalSpec, specErr := status.Load(cfg.StatusSpecFile)
	if specErr != nil {
		// Fall back to default spec if global spec can't be loaded.
		globalSpec = status.DefaultSpec()
	}
	display := ui.StatusDisplayConfig{
		Order:  globalSpec.DisplayOrder(),
		Colors: globalSpec.ColorMap(),
	}

	resolver := &status.Resolver{Runner: &status.ShellRunner{}}

	// Track errors from resolution goroutines.
	var resolveErr error
	var errOnce sync.Once

	resolved, err := ui.RunStatusTable(rows, display, func(send func(ui.StatusResolvedMsg)) {
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(4)

		for i, info := range infos {
			g.Go(func() error {
				spec, specErr := status.LoadWithFallback(
					cfg.WorkspaceStatusSpecPath(info.ID),
					cfg.StatusSpecFile,
				)
				if specErr != nil {
					if errors.Is(specErr, status.ErrSpecNotFound) {
						send(ui.StatusResolvedMsg{Index: i, Status: "-"})
						return nil
					}
					errOnce.Do(func() { resolveErr = fmt.Errorf("loading status spec for %s: %w", info.ID, specErr) })
					send(ui.StatusResolvedMsg{Index: i, Status: "error"})
					return nil
				}

				st, findErr := svc.Find(info.ID)
				if findErr != nil {
					send(ui.StatusResolvedMsg{Index: i, Status: "-"})
					return nil
				}

				repos := stateReposToInfo(st, cfg.WorkspacePath(info.ID))
				wsName := info.Name
				if wsName == "" {
					wsName = info.ID
				}
				result := resolver.ResolveWorkspace(gctx, spec, repos, info.ID, wsName)
				send(ui.StatusResolvedMsg{Index: i, Status: result.Status})
				return nil
			})
		}

		_ = g.Wait()

		if resolveErr != nil {
			ui.Errorf("%v", resolveErr)
		}
	})

	// Save resolved statuses to cache for next run.
	if resolved != nil {
		now := time.Now()
		for i, info := range infos {
			if resolved[i] != "" {
				statusCache[info.ID] = cache.StatusEntry{
					Status:     resolved[i],
					ResolvedAt: now,
				}
			}
		}
		// Best-effort save; don't fail the command if cache write fails.
		_ = cache.SaveStatus(cachePath, statusCache)
	}

	return err
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
			ui.Print("No status spec found. Run " + ui.Code("flow reset status") + " to create one.")
			return nil
		}
		return err
	}

	repos := stateReposToInfo(st, cfg.WorkspacePath(id))
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

	if st.Metadata.Description != "" {
		ui.Printf("%s\n\n", st.Metadata.Description)
	}

	colorMap := spec.ColorMap()
	ui.Printf("Status: %s  (%s)\n", ui.StatusStyle(result.Status, colorMap), ui.FormatDuration(result.Duration.Milliseconds()))

	if len(result.Repos) > 0 {
		headers := []string{"REPO", "BRANCH", "STATUS", "UPDATED", "TIME"}
		var rows [][]string
		for _, r := range result.Repos {
			updated := "-"
			if !r.LastCommit.IsZero() {
				updated = ui.RelativeTime(r.LastCommit)
			}
			rows = append(rows, []string{
				status.RepoSlug(r.URL),
				r.Branch,
				ui.StatusStyle(r.Status, colorMap),
				updated,
				ui.FormatDuration(r.Duration.Milliseconds()),
			})
		}
		fmt.Println(ui.Table(headers, rows))
	}

	return nil
}

// stateReposToInfo converts state repos to status RepoInfo slice.
// wsDir is the absolute workspace directory so FLOW_REPO_PATH is a full path.
func stateReposToInfo(st *state.State, wsDir string) []status.RepoInfo {
	repos := make([]status.RepoInfo, len(st.Spec.Repos))
	for i, r := range st.Spec.Repos {
		repos[i] = status.RepoInfo{
			URL:    r.URL,
			Branch: r.Branch,
			Path:   filepath.Join(wsDir, state.RepoPath(r)),
		}
	}
	return repos
}
