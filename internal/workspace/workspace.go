// Package workspace provides the core business logic for managing workspaces.
package workspace

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/git"
	"github.com/milldr/flow/internal/state"
)

// Sentinel errors for workspace operations.
var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrWorkspaceExists   = errors.New("workspace already exists")
)

// Info holds summary data for listing workspaces.
type Info struct {
	Name        string
	Description string
	RepoCount   int
	Created     time.Time
}

// Service orchestrates workspace operations.
type Service struct {
	Config *config.Config
	Git    git.Runner
	Log    *slog.Logger
}

func (s *Service) log() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}

// Create initializes a new workspace directory and writes the state file.
func (s *Service) Create(st *state.State) error {
	wsDir := s.Config.WorkspacePath(st.Metadata.Name)
	s.log().Debug("creating workspace", "name", st.Metadata.Name, "path", wsDir)

	if _, err := os.Stat(wsDir); err == nil {
		return fmt.Errorf("%w: %s", ErrWorkspaceExists, st.Metadata.Name)
	}

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		return err
	}

	return state.Save(s.Config.StatePath(st.Metadata.Name), st)
}

// List returns info for all workspaces.
func (s *Service) List() ([]Info, error) {
	s.log().Debug("listing workspaces", "dir", s.Config.WorkspacesDir)

	entries, err := os.ReadDir(s.Config.WorkspacesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var infos []Info
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		stPath := s.Config.StatePath(entry.Name())
		st, err := state.Load(stPath)
		if err != nil {
			s.log().Debug("skipping directory", "name", entry.Name(), "error", err)
			continue
		}

		created, _ := time.Parse(time.RFC3339, st.Metadata.Created)
		infos = append(infos, Info{
			Name:        st.Metadata.Name,
			Description: st.Metadata.Description,
			RepoCount:   len(st.Spec.Repos),
			Created:     created,
		})
	}

	s.log().Debug("found workspaces", "count", len(infos))
	return infos, nil
}

// Find loads a workspace state by name.
func (s *Service) Find(name string) (*state.State, error) {
	stPath := s.Config.StatePath(name)
	s.log().Debug("finding workspace", "name", name, "path", stPath)

	st, err := state.Load(stPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, name)
		}
		return nil, err
	}
	return st, nil
}

// Render materializes a workspace: ensures bare clones and creates worktrees.
// progress is called with status messages for each repo.
func (s *Service) Render(ctx context.Context, name string, progress func(msg string)) error {
	st, err := s.Find(name)
	if err != nil {
		return err
	}

	wsDir := s.Config.WorkspacePath(name)
	total := len(st.Spec.Repos)

	for i, repo := range st.Spec.Repos {
		barePath := s.Config.BareRepoPath(repo.URL)
		worktreePath := filepath.Join(wsDir, repo.Path)

		progress(fmt.Sprintf("[%d/%d] %s", i+1, total, repo.URL))

		// Ensure bare clone exists
		if _, err := os.Stat(barePath); os.IsNotExist(err) {
			s.log().Debug("bare clone not found, cloning", "url", repo.URL, "dest", barePath)
			if err := os.MkdirAll(filepath.Dir(barePath), 0o755); err != nil {
				return err
			}
			if err := s.Git.BareClone(ctx, repo.URL, barePath); err != nil {
				return fmt.Errorf("cloning %s: %w", repo.URL, err)
			}
		} else {
			s.log().Debug("bare clone exists, fetching", "url", repo.URL, "path", barePath)
			if err := s.Git.Fetch(ctx, barePath); err != nil {
				return fmt.Errorf("fetching %s: %w", repo.URL, err)
			}
		}

		// Create worktree if it doesn't exist
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			s.log().Debug("creating worktree", "path", worktreePath, "branch", repo.Branch)
			if err := s.Git.AddWorktree(ctx, barePath, worktreePath, repo.Branch); err != nil {
				return fmt.Errorf("creating worktree for %s: %w", repo.URL, err)
			}
		} else {
			s.log().Debug("worktree already exists, skipping", "path", worktreePath)
		}

		progress(fmt.Sprintf("      └── Worktree: %s (%s) ✓", repo.Path, repo.Branch))
	}

	return nil
}

// Delete removes all worktrees and the workspace directory.
func (s *Service) Delete(ctx context.Context, name string) error {
	st, err := s.Find(name)
	if err != nil {
		return err
	}

	wsDir := s.Config.WorkspacePath(name)
	s.log().Debug("deleting workspace", "name", name, "path", wsDir)

	// Remove worktrees
	for _, repo := range st.Spec.Repos {
		barePath := s.Config.BareRepoPath(repo.URL)
		worktreePath := filepath.Join(wsDir, repo.Path)

		if _, err := os.Stat(worktreePath); err == nil {
			s.log().Debug("removing worktree", "path", worktreePath)
			// Best effort — worktree may already be gone
			_ = s.Git.RemoveWorktree(ctx, barePath, worktreePath)
		}
	}

	return os.RemoveAll(wsDir)
}
