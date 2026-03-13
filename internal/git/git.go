// Package git wraps git CLI operations behind a testable interface.
package git

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner abstracts git operations for testability.
type Runner interface {
	BareClone(ctx context.Context, url, dest string) error
	Fetch(ctx context.Context, repoPath string) error
	AddWorktree(ctx context.Context, bareRepo, worktreePath, branch string) error
	AddWorktreeNewBranch(ctx context.Context, bareRepo, worktreePath, newBranch, startPoint string) error
	RemoveWorktree(ctx context.Context, bareRepo, worktreePath string) error
	BranchExists(ctx context.Context, bareRepo, branch string) (bool, error)
	DeleteBranch(ctx context.Context, bareRepo, branch string) error
	DefaultBranch(ctx context.Context, bareRepo string) (string, error)
	EnsureRemoteRef(ctx context.Context, bareRepo, branch string) error
	ResetBranch(ctx context.Context, worktreePath, ref string) error
	IsClean(ctx context.Context, worktreePath string) (bool, error)
	CurrentBranch(ctx context.Context, worktreePath string) (string, error)
	CheckoutBranch(ctx context.Context, worktreePath, branch string) error
	CheckoutNewBranch(ctx context.Context, worktreePath, newBranch, startPoint string) error
	Rebase(ctx context.Context, worktreePath, onto string) error
	RebaseAbort(ctx context.Context, worktreePath string) error
}

// RealRunner shells out to the git binary.
type RealRunner struct {
	Log *slog.Logger
}

func (r *RealRunner) log() *slog.Logger {
	if r.Log != nil {
		return r.Log
	}
	return slog.Default()
}

func (r *RealRunner) run(ctx context.Context, args ...string) error {
	r.log().Debug("executing git command", "args", args)

	cmd := exec.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return fmt.Errorf("git %s: %w: %s", args[0], err, errMsg)
		}
		return fmt.Errorf("git %s: %w", args[0], err)
	}
	return nil
}

func (r *RealRunner) output(ctx context.Context, args ...string) (string, error) {
	r.log().Debug("executing git command", "args", args)

	cmd := exec.CommandContext(ctx, "git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, errMsg)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// BareClone creates a bare clone of a repository.
func (r *RealRunner) BareClone(ctx context.Context, url, dest string) error {
	r.log().Debug("bare cloning repository", "url", url, "dest", dest)
	cloneURL := url
	if !strings.Contains(url, "://") && !strings.HasPrefix(url, "git@") && !filepath.IsAbs(url) && !strings.HasPrefix(url, ".") {
		cloneURL = "https://" + url
	}
	return r.run(ctx, "clone", "--bare", cloneURL, dest)
}

// Fetch fetches all refs in a bare repository and ensures the default branch
// has a remote tracking ref so status checks can reference origin/main.
func (r *RealRunner) Fetch(ctx context.Context, repoPath string) error {
	r.log().Debug("fetching repository", "path", repoPath)

	if err := r.run(ctx, "-C", repoPath, "fetch", "--all", "--prune"); err != nil {
		return err
	}

	// Create a remote tracking ref for the default branch so status checks
	// can compare against origin/main (or origin/master). Bare clones store
	// refs in refs/heads/ and don't create refs/remotes/origin/* by default.
	// We only track the default branch to avoid case-conflict errors on
	// macOS for repos with branches that differ only by casing.
	defaultBranch, err := r.output(ctx, "-C", repoPath, "symbolic-ref", "--short", "HEAD")
	if err == nil && defaultBranch != "" {
		_ = r.run(ctx, "-C", repoPath, "fetch", "origin",
			"+refs/heads/"+defaultBranch+":refs/remotes/origin/"+defaultBranch)
		_ = r.run(ctx, "-C", repoPath, "symbolic-ref",
			"refs/remotes/origin/HEAD", "refs/remotes/origin/"+defaultBranch)
	}

	return nil
}

// AddWorktree creates a worktree from a bare repo at the given path and branch.
// Uses --force to allow the same branch to be checked out in multiple worktrees
// across different workspaces.
func (r *RealRunner) AddWorktree(ctx context.Context, bareRepo, worktreePath, branch string) error {
	r.log().Debug("adding worktree", "bare_repo", bareRepo, "worktree", worktreePath, "branch", branch)
	return r.run(ctx, "-C", bareRepo, "worktree", "add", "--force", worktreePath, branch)
}

// AddWorktreeNewBranch creates a worktree with a new branch starting from startPoint.
func (r *RealRunner) AddWorktreeNewBranch(ctx context.Context, bareRepo, worktreePath, newBranch, startPoint string) error {
	r.log().Debug("adding worktree with new branch", "bare_repo", bareRepo, "worktree", worktreePath, "branch", newBranch, "start_point", startPoint)
	return r.run(ctx, "-C", bareRepo, "worktree", "add", "-b", newBranch, worktreePath, startPoint)
}

// RemoveWorktree removes a worktree from a bare repo.
func (r *RealRunner) RemoveWorktree(ctx context.Context, bareRepo, worktreePath string) error {
	r.log().Debug("removing worktree", "bare_repo", bareRepo, "worktree", worktreePath)
	return r.run(ctx, "-C", bareRepo, "worktree", "remove", "--force", worktreePath)
}

// BranchExists checks if a branch exists in the bare repo (local or remote ref).
func (r *RealRunner) BranchExists(ctx context.Context, bareRepo, branch string) (bool, error) {
	// Check for exact ref match: heads/<branch> or remotes/origin/<branch>
	out, err := r.output(ctx, "-C", bareRepo, "branch", "-a", "--list", branch, "origin/"+branch)
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// DeleteBranch force-deletes a branch in the bare repo.
func (r *RealRunner) DeleteBranch(ctx context.Context, bareRepo, branch string) error {
	r.log().Debug("deleting branch", "bare_repo", bareRepo, "branch", branch)
	return r.run(ctx, "-C", bareRepo, "branch", "-D", branch)
}

// DefaultBranch returns the default branch name (e.g. "main" or "master") for a bare repo.
func (r *RealRunner) DefaultBranch(ctx context.Context, bareRepo string) (string, error) {
	// In a bare clone, HEAD points to the default branch
	out, err := r.output(ctx, "-C", bareRepo, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("determining default branch: %w", err)
	}
	return out, nil
}

// EnsureRemoteRef creates refs/remotes/origin/{branch} in a bare repo so
// origin/{branch} resolves from worktrees.
func (r *RealRunner) EnsureRemoteRef(ctx context.Context, bareRepo, branch string) error {
	r.log().Debug("ensuring remote ref", "bare_repo", bareRepo, "branch", branch)
	return r.run(ctx, "-C", bareRepo, "fetch", "origin",
		"+refs/heads/"+branch+":refs/remotes/origin/"+branch)
}

// ResetBranch resets the current branch in a worktree to the given ref.
// This is used to fast-forward existing worktrees to the latest remote state.
func (r *RealRunner) ResetBranch(ctx context.Context, worktreePath, ref string) error {
	r.log().Debug("resetting branch", "path", worktreePath, "ref", ref)
	return r.run(ctx, "-C", worktreePath, "reset", "--hard", ref)
}

// IsClean returns true if the worktree has no uncommitted changes.
func (r *RealRunner) IsClean(ctx context.Context, worktreePath string) (bool, error) {
	r.log().Debug("checking worktree cleanliness", "path", worktreePath)
	out, err := r.output(ctx, "-C", worktreePath, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

// CurrentBranch returns the currently checked-out branch in a worktree.
func (r *RealRunner) CurrentBranch(ctx context.Context, worktreePath string) (string, error) {
	r.log().Debug("getting current branch", "path", worktreePath)
	return r.output(ctx, "-C", worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
}

// CheckoutBranch switches to an existing branch in a worktree.
func (r *RealRunner) CheckoutBranch(ctx context.Context, worktreePath, branch string) error {
	r.log().Debug("checking out branch", "path", worktreePath, "branch", branch)
	return r.run(ctx, "-C", worktreePath, "checkout", branch)
}

// CheckoutNewBranch creates and switches to a new branch from a start point.
func (r *RealRunner) CheckoutNewBranch(ctx context.Context, worktreePath, newBranch, startPoint string) error {
	r.log().Debug("checking out new branch", "path", worktreePath, "branch", newBranch, "start_point", startPoint)
	return r.run(ctx, "-C", worktreePath, "checkout", "-b", newBranch, startPoint)
}

// Rebase rebases the current branch onto the given ref.
func (r *RealRunner) Rebase(ctx context.Context, worktreePath, onto string) error {
	r.log().Debug("rebasing", "path", worktreePath, "onto", onto)
	return r.run(ctx, "-C", worktreePath, "rebase", onto)
}

// RebaseAbort aborts a rebase in progress.
func (r *RealRunner) RebaseAbort(ctx context.Context, worktreePath string) error {
	r.log().Debug("aborting rebase", "path", worktreePath)
	return r.run(ctx, "-C", worktreePath, "rebase", "--abort")
}
