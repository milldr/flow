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

	"github.com/milldr/flow/internal/agents"
	"github.com/milldr/flow/internal/config"
	"github.com/milldr/flow/internal/git"
	"github.com/milldr/flow/internal/state"
)

// Sentinel errors for workspace operations.
var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrWorkspaceExists   = errors.New("workspace already exists")
	ErrAmbiguousName     = errors.New("ambiguous workspace name")
)

// AmbiguousNameError is returned when a name matches multiple workspaces.
type AmbiguousNameError struct {
	Name    string
	Matches []Info
}

func (e *AmbiguousNameError) Error() string {
	return fmt.Sprintf("name %q matches %d workspaces", e.Name, len(e.Matches))
}

func (e *AmbiguousNameError) Unwrap() error { return ErrAmbiguousName }

// Info holds summary data for listing workspaces.
type Info struct {
	ID          string // directory name (unique identifier)
	Name        string // metadata.name (may be empty or duplicated)
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
// id is used as the directory name; st.Metadata.Name may differ from id or be empty.
func (s *Service) Create(id string, st *state.State) error {
	wsDir := s.Config.WorkspacePath(id)
	s.log().Debug("creating workspace", "id", id, "path", wsDir)

	if _, err := os.Stat(wsDir); err == nil {
		return fmt.Errorf("%w: %s", ErrWorkspaceExists, id)
	}

	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		return err
	}

	return state.Save(s.Config.StatePath(id), st)
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
			ID:          entry.Name(),
			Name:        st.Metadata.Name,
			Description: st.Metadata.Description,
			RepoCount:   len(st.Spec.Repos),
			Created:     created,
		})
	}

	s.log().Debug("found workspaces", "count", len(infos))
	return infos, nil
}

// Find loads a workspace state by ID (directory name).
func (s *Service) Find(id string) (*state.State, error) {
	stPath := s.Config.StatePath(id)
	s.log().Debug("finding workspace", "id", id, "path", stPath)

	st, err := state.Load(stPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, id)
		}
		return nil, err
	}
	return st, nil
}

// Resolve looks up workspaces by ID or name.
// It first tries an exact ID match, then falls back to scanning names.
// Returns 0 matches as ErrWorkspaceNotFound, N>1 matches as *AmbiguousNameError.
func (s *Service) Resolve(idOrName string) ([]Info, error) {
	// Try direct ID lookup first (O(1) filesystem check)
	if _, err := s.Find(idOrName); err == nil {
		// Load full info for the matched workspace
		stPath := s.Config.StatePath(idOrName)
		st, _ := state.Load(stPath)
		created, _ := time.Parse(time.RFC3339, st.Metadata.Created)
		return []Info{{
			ID:          idOrName,
			Name:        st.Metadata.Name,
			Description: st.Metadata.Description,
			RepoCount:   len(st.Spec.Repos),
			Created:     created,
		}}, nil
	}

	// Fall back to name scan
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var matches []Info
	for _, info := range all {
		if info.Name == idOrName {
			matches = append(matches, info)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, idOrName)
	}
	if len(matches) > 1 {
		return nil, &AmbiguousNameError{Name: idOrName, Matches: matches}
	}
	return matches, nil
}

// Render materializes a workspace: ensures bare clones and creates worktrees.
// progress is called with status messages for each repo.
func (s *Service) Render(ctx context.Context, id string, progress func(msg string)) error {
	st, err := s.Find(id)
	if err != nil {
		return err
	}

	if err := state.Validate(st); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	wsDir := s.Config.WorkspacePath(id)
	total := len(st.Spec.Repos)

	for i, repo := range st.Spec.Repos {
		repoPath := state.RepoPath(repo)
		barePath := s.Config.BareRepoPath(repo.URL)
		worktreePath := filepath.Join(wsDir, repoPath)

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
			exists, err := s.Git.BranchExists(ctx, barePath, repo.Branch)
			if err != nil {
				return fmt.Errorf("checking branch for %s: %w", repo.URL, err)
			}

			if exists {
				s.log().Debug("creating worktree from existing branch", "path", worktreePath, "branch", repo.Branch)
				if err := s.Git.AddWorktree(ctx, barePath, worktreePath, repo.Branch); err != nil {
					return fmt.Errorf("creating worktree for %s: %w", repo.URL, err)
				}
				progress(fmt.Sprintf("      └── %s (%s) ✓", repoPath, repo.Branch))
			} else {
				defaultBranch, err := s.Git.DefaultBranch(ctx, barePath)
				if err != nil {
					return fmt.Errorf("getting default branch for %s: %w", repo.URL, err)
				}
				s.log().Debug("creating worktree with new branch", "path", worktreePath, "branch", repo.Branch, "from", defaultBranch)
				if err := s.Git.AddWorktreeNewBranch(ctx, barePath, worktreePath, repo.Branch, defaultBranch); err != nil {
					return fmt.Errorf("creating worktree for %s: %w", repo.URL, err)
				}
				progress(fmt.Sprintf("      └── %s (%s, new branch from %s) ✓", repoPath, repo.Branch, defaultBranch))
			}
		} else {
			s.log().Debug("worktree already exists, skipping", "path", worktreePath)
			progress(fmt.Sprintf("      └── %s (%s) exists", repoPath, repo.Branch))
		}
	}

	// Set up Claude workspace files
	if err := agents.SetupWorkspaceClaude(wsDir, s.Config.AgentsDir, st, id); err != nil {
		return fmt.Errorf("setting up claude files: %w", err)
	}

	return nil
}

// Delete removes all worktrees and the workspace directory.
func (s *Service) Delete(ctx context.Context, id string) error {
	st, err := s.Find(id)
	if err != nil {
		return err
	}

	wsDir := s.Config.WorkspacePath(id)
	s.log().Debug("deleting workspace", "id", id, "path", wsDir)

	// Remove worktrees
	for _, repo := range st.Spec.Repos {
		barePath := s.Config.BareRepoPath(repo.URL)
		worktreePath := filepath.Join(wsDir, state.RepoPath(repo))

		if _, err := os.Stat(worktreePath); err == nil {
			s.log().Debug("removing worktree", "path", worktreePath)
			// Best effort — worktree may already be gone
			_ = s.Git.RemoveWorktree(ctx, barePath, worktreePath)
		}
	}

	return os.RemoveAll(wsDir)
}
