// Package workspace provides the core business logic for managing workspaces.
package workspace

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
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

// repoRenderContext holds pre-computed paths for rendering a single repo.
type repoRenderContext struct {
	index        int
	repo         state.Repo
	repoPath     string
	barePath     string
	worktreePath string
}

// Render materializes a workspace: ensures bare clones and creates worktrees.
// Bare repos are fetched in parallel to ensure we always have the latest remote
// state before creating or updating worktrees.
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

	// Build render contexts for all repos
	repos := make([]repoRenderContext, total)
	for i, repo := range st.Spec.Repos {
		repos[i] = repoRenderContext{
			index:        i,
			repo:         repo,
			repoPath:     state.RepoPath(repo),
			barePath:     s.Config.BareRepoPath(repo.URL),
			worktreePath: filepath.Join(wsDir, state.RepoPath(repo)),
		}
	}

	// Phase 1: Clone and fetch all bare repos in parallel.
	// This ensures every bare clone has the latest remote refs before we
	// create or update any worktrees.
	fetchErrs := make([]error, total)
	var wg sync.WaitGroup
	for i := range repos {
		wg.Add(1)
		go func(rc *repoRenderContext) {
			defer wg.Done()
			fetchErrs[rc.index] = s.ensureBareRepo(ctx, rc)
		}(&repos[i])
	}
	wg.Wait()

	// Check for fetch errors — fail fast on any clone/fetch failure
	for i, err := range fetchErrs {
		if err != nil {
			return fmt.Errorf("%s: %w", repos[i].repo.URL, err)
		}
	}

	// Phase 2: Create or update worktrees (sequential — progress messages
	// are order-dependent and worktree operations are fast).
	for i := range repos {
		rc := &repos[i]
		progress(fmt.Sprintf("[%d/%d] %s", rc.index+1, total, rc.repo.URL))

		if err := s.ensureWorktree(ctx, rc, progress); err != nil {
			return err
		}
	}

	// Set up Claude workspace files
	if err := agents.SetupWorkspaceClaude(wsDir, s.Config.AgentsDir, st, id); err != nil {
		return fmt.Errorf("setting up claude files: %w", err)
	}

	return nil
}

// ensureBareRepo clones (if needed) and fetches a bare repository.
func (s *Service) ensureBareRepo(ctx context.Context, rc *repoRenderContext) error {
	if _, err := os.Stat(rc.barePath); os.IsNotExist(err) {
		s.log().Debug("bare clone not found, cloning", "url", rc.repo.URL, "dest", rc.barePath)
		if err := os.MkdirAll(filepath.Dir(rc.barePath), 0o755); err != nil {
			return err
		}
		if err := s.Git.BareClone(ctx, rc.repo.URL, rc.barePath); err != nil {
			return fmt.Errorf("cloning: %w", err)
		}
	}

	s.log().Debug("fetching bare repo", "url", rc.repo.URL, "path", rc.barePath)
	if err := s.Git.Fetch(ctx, rc.barePath); err != nil {
		return fmt.Errorf("fetching: %w", err)
	}
	return nil
}

// ensureWorktree creates a new worktree or updates an existing one to the
// latest remote state.
func (s *Service) ensureWorktree(ctx context.Context, rc *repoRenderContext, progress func(msg string)) error {
	if _, err := os.Stat(rc.worktreePath); os.IsNotExist(err) {
		return s.createWorktree(ctx, rc, progress)
	}
	return s.updateWorktree(ctx, rc, progress)
}

// createWorktree creates a new worktree, either from an existing branch or
// by creating a new branch from the base.
func (s *Service) createWorktree(ctx context.Context, rc *repoRenderContext, progress func(msg string)) error {
	exists, err := s.Git.BranchExists(ctx, rc.barePath, rc.repo.Branch)
	if err != nil {
		return fmt.Errorf("checking branch for %s: %w", rc.repo.URL, err)
	}

	if exists {
		s.log().Debug("creating worktree from existing branch", "path", rc.worktreePath, "branch", rc.repo.Branch)
		if err := s.Git.AddWorktree(ctx, rc.barePath, rc.worktreePath, rc.repo.Branch); err != nil {
			return fmt.Errorf("creating worktree for %s: %w", rc.repo.URL, err)
		}
		progress(fmt.Sprintf("      └── %s (%s) ✓", rc.repoPath, rc.repo.Branch))
		return nil
	}

	baseBranch, err := s.resolveBaseBranch(ctx, rc)
	if err != nil {
		return err
	}

	if err := s.Git.EnsureRemoteRef(ctx, rc.barePath, baseBranch); err != nil {
		return fmt.Errorf("ensuring remote ref for %s: %w", rc.repo.URL, err)
	}

	startPoint := "origin/" + baseBranch
	s.log().Debug("creating worktree with new branch", "path", rc.worktreePath, "branch", rc.repo.Branch, "from", startPoint)
	if err := s.Git.AddWorktreeNewBranch(ctx, rc.barePath, rc.worktreePath, rc.repo.Branch, startPoint); err != nil {
		return fmt.Errorf("creating worktree for %s: %w", rc.repo.URL, err)
	}
	progress(fmt.Sprintf("      └── %s (%s, new branch from %s) ✓", rc.repoPath, rc.repo.Branch, baseBranch))
	return nil
}

// updateWorktree resets an existing worktree to the latest remote ref for its
// branch so that re-rendering always picks up new upstream commits.
func (s *Service) updateWorktree(ctx context.Context, rc *repoRenderContext, progress func(msg string)) error {
	// Ensure the remote tracking ref exists for this branch so we can
	// check if the branch exists on the remote.
	if err := s.Git.EnsureRemoteRef(ctx, rc.barePath, rc.repo.Branch); err != nil {
		// Branch doesn't exist on remote — this is a local-only feature
		// branch. Leave it alone.
		s.log().Debug("worktree exists, no remote branch to update from", "path", rc.worktreePath, "branch", rc.repo.Branch)
		progress(fmt.Sprintf("      └── %s (%s) exists", rc.repoPath, rc.repo.Branch))
		return nil
	}

	// Check if the worktree is clean before resetting
	clean, err := s.Git.IsClean(ctx, rc.worktreePath)
	if err != nil {
		return fmt.Errorf("checking worktree status for %s: %w", rc.repo.URL, err)
	}

	if !clean {
		s.log().Debug("worktree is dirty, skipping update", "path", rc.worktreePath)
		progress(fmt.Sprintf("      └── %s (%s) exists (dirty, skipped update)", rc.repoPath, rc.repo.Branch))
		return nil
	}

	// Reset to the latest remote ref
	ref := "origin/" + rc.repo.Branch
	s.log().Debug("updating worktree to latest remote", "path", rc.worktreePath, "ref", ref)
	if err := s.Git.ResetBranch(ctx, rc.worktreePath, ref); err != nil {
		return fmt.Errorf("updating worktree for %s: %w", rc.repo.URL, err)
	}

	progress(fmt.Sprintf("      └── %s (%s) updated ✓", rc.repoPath, rc.repo.Branch))
	return nil
}

// resolveBaseBranch returns the base branch for creating new feature branches.
func (s *Service) resolveBaseBranch(ctx context.Context, rc *repoRenderContext) (string, error) {
	if rc.repo.Base != "" {
		return rc.repo.Base, nil
	}
	baseBranch, err := s.Git.DefaultBranch(ctx, rc.barePath)
	if err != nil {
		return "", fmt.Errorf("getting default branch for %s: %w", rc.repo.URL, err)
	}
	return baseBranch, nil
}

// Sync fetches and rebases worktrees onto their base branches.
// It continues through failures — one repo failing doesn't block others.
func (s *Service) Sync(ctx context.Context, id string, progress func(msg string)) error {
	st, err := s.Find(id)
	if err != nil {
		return err
	}

	if err := state.Validate(st); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	wsDir := s.Config.WorkspacePath(id)
	total := len(st.Spec.Repos)
	var errs []error

	for i, repo := range st.Spec.Repos {
		repoPath := state.RepoPath(repo)
		barePath := s.Config.BareRepoPath(repo.URL)
		worktreePath := filepath.Join(wsDir, repoPath)

		progress(fmt.Sprintf("[%d/%d] %s", i+1, total, repo.URL))

		// Skip if worktree doesn't exist
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			progress(fmt.Sprintf("      └── %s skipped (not rendered)", repoPath))
			continue
		}

		// Fetch bare repo
		if err := s.Git.Fetch(ctx, barePath); err != nil {
			errs = append(errs, fmt.Errorf("%s: fetch: %w", repoPath, err))
			progress(fmt.Sprintf("      └── %s fetch failed", repoPath))
			continue
		}

		// Resolve base branch
		var baseBranch string
		if repo.Base != "" {
			baseBranch = repo.Base
		} else {
			baseBranch, err = s.Git.DefaultBranch(ctx, barePath)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: default branch: %w", repoPath, err))
				progress(fmt.Sprintf("      └── %s failed to resolve base branch", repoPath))
				continue
			}
		}

		// Ensure remote ref so origin/{base} resolves from worktrees
		if err := s.Git.EnsureRemoteRef(ctx, barePath, baseBranch); err != nil {
			errs = append(errs, fmt.Errorf("%s: ensure remote ref: %w", repoPath, err))
			progress(fmt.Sprintf("      └── %s failed to ensure remote ref", repoPath))
			continue
		}

		// Check clean
		clean, err := s.Git.IsClean(ctx, worktreePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: checking clean: %w", repoPath, err))
			progress(fmt.Sprintf("      └── %s failed to check status", repoPath))
			continue
		}
		if !clean {
			progress(fmt.Sprintf("      └── %s skipped (dirty worktree)", repoPath))
			continue
		}

		// Rebase onto origin/{base}
		onto := "origin/" + baseBranch
		if err := s.Git.Rebase(ctx, worktreePath, onto); err != nil {
			_ = s.Git.RebaseAbort(ctx, worktreePath)
			errs = append(errs, fmt.Errorf("%s: rebase onto %s: %w", repoPath, onto, err))
			progress(fmt.Sprintf("      └── %s rebase failed (aborted)", repoPath))
			continue
		}

		progress(fmt.Sprintf("      └── %s rebased onto %s ✓", repoPath, onto))
	}

	return errors.Join(errs...)
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
