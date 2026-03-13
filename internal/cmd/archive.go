package cmd

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/status"
	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var errWorkspaceArgRequired = errors.New("workspace argument required (or use --closed)")

func newArchiveCmd(svc *workspace.Service, cfg *config.Config) *cobra.Command {
	var closed bool
	var force bool

	cmd := &cobra.Command{
		Use:   "archive [workspace]",
		Short: "Archive a workspace (remove worktrees, keep state)",
		Long: `Archive a workspace by removing its worktrees to free up branches,
while preserving the state file. Archived workspaces are hidden from
flow status by default (use --all to see them).

Use --closed to archive all workspaces with "closed" status at once.`,
		Args:    cobra.MaximumNArgs(1),
		Example: "  flow archive my-workspace    # Archive a single workspace\n  flow archive --closed        # Archive all closed workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			if closed {
				return runArchiveClosed(cmd.Context(), svc, cfg, force)
			}
			if len(args) == 0 {
				return errWorkspaceArgRequired
			}
			return runArchiveOne(cmd.Context(), svc, args[0], force)
		},
	}

	cmd.Flags().BoolVar(&closed, "closed", false, "Archive all workspaces with closed status")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func runArchiveOne(ctx context.Context, svc *workspace.Service, idOrName string, force bool) error {
	id, st, err := resolveWorkspace(svc, idOrName)
	if err != nil {
		return err
	}

	if st.Metadata.Archived {
		ui.Print("Workspace is already archived.")
		return nil
	}

	name := workspaceDisplayName(id, st)

	if !force {
		confirmed, err := ui.Confirm(fmt.Sprintf("Archive workspace %q? This will remove worktrees and free branches.", name))
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Print("Cancelled.")
			return nil
		}
	}

	err = ui.RunWithSpinner("Archiving workspace: "+name, func(_ func(string)) error {
		return svc.Archive(ctx, id)
	})
	if err != nil {
		return err
	}

	ui.Success("Archived workspace: " + name)
	return nil
}

func runArchiveClosed(ctx context.Context, svc *workspace.Service, cfg *config.Config, force bool) error {
	infos, err := svc.List()
	if err != nil {
		return err
	}

	// Filter to non-archived workspaces only.
	var candidates []workspace.Info
	for _, info := range infos {
		if !info.Archived {
			candidates = append(candidates, info)
		}
	}

	if len(candidates) == 0 {
		ui.Print("No active workspaces to check.")
		return nil
	}

	// Resolve statuses to find closed workspaces.
	type closedWS struct {
		id   string
		name string
	}
	var closedList []closedWS
	var mu sync.Mutex

	resolver := &status.Resolver{Runner: &status.ShellRunner{}}

	err = ui.RunWithSpinner("Resolving workspace statuses...", func(_ func(string)) error {
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(4)

		for _, info := range candidates {
			g.Go(func() error {
				spec, specErr := status.LoadWithFallback(
					cfg.WorkspaceStatusSpecPath(info.ID),
					cfg.StatusSpecFile,
				)
				if specErr != nil {
					return nil // skip workspaces without status spec
				}

				st, findErr := svc.Find(info.ID)
				if findErr != nil {
					return nil
				}

				repos := stateReposToInfo(st, cfg.WorkspacePath(info.ID))
				wsName := info.Name
				if wsName == "" {
					wsName = info.ID
				}
				result := resolver.ResolveWorkspace(gctx, spec, repos, info.ID, wsName)
				if result.Status == "closed" {
					mu.Lock()
					closedList = append(closedList, closedWS{id: info.ID, name: wsName})
					mu.Unlock()
				}
				return nil
			})
		}

		return g.Wait()
	})
	if err != nil {
		return err
	}

	if len(closedList) == 0 {
		ui.Print("No closed workspaces to archive.")
		return nil
	}

	if !force {
		ui.Printf("Found %d closed workspace(s) to archive:\n", len(closedList))
		for _, ws := range closedList {
			ui.Printf("  - %s\n", ws.name)
		}
		confirmed, err := ui.Confirm("Archive all?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Print("Cancelled.")
			return nil
		}
	}

	var archiveErrors []error
	for _, ws := range closedList {
		if err := svc.Archive(ctx, ws.id); err != nil {
			archiveErrors = append(archiveErrors, fmt.Errorf("archiving %s: %w", ws.name, err))
			continue
		}
		ui.Success("Archived: " + ws.name)
	}

	if len(archiveErrors) > 0 {
		return errors.Join(archiveErrors...)
	}
	return nil
}
