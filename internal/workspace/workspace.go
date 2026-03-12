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
	RepoNames   []string // short repo names (derived from URLs)
	Archived    bool
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

	if err := state.Save(s.Config.StatePath(id), st); err != nil {
		return err
	}

	// Set up Claude agent files immediately so skills are available before render
	if err := agents.SetupWorkspaceClaude(wsDir, s.Config.AgentsDir, st, id); err != nil {
		return fmt.Errorf("setting up claude files: %w", err)
	}

	return nil
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
		repoNames := make([]string, len(st.Spec.Repos))
		for i, r := range st.Spec.Repos {
			repoNames[i] = state.RepoPath(r)
		}
		infos = append(infos, Info{
			ID:          entry.Name(),
			Name:        st.Metadata.Name,
			Description: st.Metadata.Description,
			RepoCount:   len(st.Spec.Repos),
			RepoNames:   repoNames,
			Archived:    st.Metadata.Archived,
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
		repoNames := make([]string, len(st.Spec.Repos))
		for i, r := range st.Spec.Repos {
			repoNames[i] = state.RepoPath(r)
		}
		return []Info{{
			ID:          idOrName,
			Name:        st.Metadata.Name,
			Description: st.Metadata.Description,
			RepoCount:   len(st.Spec.Repos),
			RepoNames:   repoNames,
			Archived:    st.Metadata.Archived,
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

// BranchConflict controls what happens when a branch already exists during render.
type BranchConflict int

const (
	// BranchConflictPrompt asks the user interactively (default).
	BranchConflictPrompt BranchConflict = iota
	// BranchConflictReset deletes the existing branch and creates fresh from default.
	BranchConflictReset
	// BranchConflictUseExisting checks out the existing branch as-is.
	BranchConflictUseExisting
)

// RenderOptions configures render behavior.
type RenderOptions struct {
	// OnBranchConflict controls what to do when a branch already exists.
	OnBranchConflict BranchConflict
	// PromptBranchConflict is called when OnBranchConflict is BranchConflictPrompt
	// and a branch already exists. It should return true to reset (create fresh)
	// or false to use the existing branch.
	PromptBranchConflict func(repo, branch, defaultBranch string) (reset bool, err error)
}

// Render materializes a workspace: ensures bare clones and creates worktrees.
// progress is called with status messages for each repo.
func (s *Service) Render(ctx context.Context, id string, progress func(msg string), opts *RenderOptions) error {
	if opts == nil {
		opts = &RenderOptions{}
	}

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
				shouldReset, err := s.shouldResetBranch(ctx, opts, barePath, repoPath, repo.Branch)
				if err != nil {
					return err
				}

				if shouldReset {
					defaultBranch, err := s.Git.DefaultBranch(ctx, barePath)
					if err != nil {
						return fmt.Errorf("getting default branch for %s: %w", repo.URL, err)
					}
					s.log().Debug("resetting branch", "branch", repo.Branch, "from", defaultBranch)
					if err := s.Git.DeleteBranch(ctx, barePath, repo.Branch); err != nil {
						return fmt.Errorf("deleting branch %s in %s: %w", repo.Branch, repo.URL, err)
					}
					if err := s.Git.AddWorktreeNewBranch(ctx, barePath, worktreePath, repo.Branch, defaultBranch); err != nil {
						return fmt.Errorf("creating worktree for %s: %w", repo.URL, err)
					}
					progress(fmt.Sprintf("      └── %s (%s, reset from %s) ✓", repoPath, repo.Branch, defaultBranch))
				} else {
					s.log().Debug("creating worktree from existing branch", "path", worktreePath, "branch", repo.Branch)
					if err := s.Git.AddWorktree(ctx, barePath, worktreePath, repo.Branch); err != nil {
						return fmt.Errorf("creating worktree for %s: %w", repo.URL, err)
					}
					progress(fmt.Sprintf("      └── %s (%s) ✓", repoPath, repo.Branch))
				}
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

// shouldResetBranch determines whether to reset an existing branch based on options.
func (s *Service) shouldResetBranch(ctx context.Context, opts *RenderOptions, barePath, repoPath, branch string) (bool, error) {
	switch opts.OnBranchConflict {
	case BranchConflictReset:
		return true, nil
	case BranchConflictUseExisting:
		return false, nil
	default: // BranchConflictPrompt
		if opts.PromptBranchConflict == nil {
			// No prompt callback — default to using existing
			return false, nil
		}
		defaultBranch, err := s.Git.DefaultBranch(ctx, barePath)
		if err != nil {
			return false, fmt.Errorf("getting default branch: %w", err)
		}
		return opts.PromptBranchConflict(repoPath, branch, defaultBranch)
	}
}

// Archive removes worktrees (freeing branches) and marks the workspace as archived.
// The workspace directory and state file are preserved so it can still appear in listings.
func (s *Service) Archive(ctx context.Context, id string) error {
	st, err := s.Find(id)
	if err != nil {
		return err
	}

	wsDir := s.Config.WorkspacePath(id)
	s.log().Debug("archiving workspace", "id", id, "path", wsDir)

	// Remove worktrees to free branches
	for _, repo := range st.Spec.Repos {
		barePath := s.Config.BareRepoPath(repo.URL)
		worktreePath := filepath.Join(wsDir, state.RepoPath(repo))

		if _, err := os.Stat(worktreePath); err == nil {
			s.log().Debug("removing worktree", "path", worktreePath)
			_ = s.Git.RemoveWorktree(ctx, barePath, worktreePath)
		}
	}

	// Mark as archived in state
	st.Metadata.Archived = true
	return state.Save(s.Config.StatePath(id), st)
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
