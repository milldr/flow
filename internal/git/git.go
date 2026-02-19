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
	RemoveWorktree(ctx context.Context, bareRepo, worktreePath string) error
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
			r.log().Debug("git stderr output", "stderr", errMsg)
		}
		return fmt.Errorf("git %s: %w", args[0], err)
	}
	return nil
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

// Fetch fetches all refs in a bare repository.
func (r *RealRunner) Fetch(ctx context.Context, repoPath string) error {
	r.log().Debug("fetching repository", "path", repoPath)
	return r.run(ctx, "-C", repoPath, "fetch", "--all", "--prune")
}

// AddWorktree creates a worktree from a bare repo at the given path and branch.
func (r *RealRunner) AddWorktree(ctx context.Context, bareRepo, worktreePath, branch string) error {
	r.log().Debug("adding worktree", "bare_repo", bareRepo, "worktree", worktreePath, "branch", branch)
	return r.run(ctx, "-C", bareRepo, "worktree", "add", worktreePath, branch)
}

// RemoveWorktree removes a worktree from a bare repo.
func (r *RealRunner) RemoveWorktree(ctx context.Context, bareRepo, worktreePath string) error {
	r.log().Debug("removing worktree", "bare_repo", bareRepo, "worktree", worktreePath)
	return r.run(ctx, "-C", bareRepo, "worktree", "remove", "--force", worktreePath)
}
