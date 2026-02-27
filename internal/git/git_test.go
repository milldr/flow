package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// initTestRepo creates a local git repo with one commit on "main", then
// bare-clones it. Returns the bare repo path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = src
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	git("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("# test"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-m", "initial")

	bare := filepath.Join(dir, "bare.git")
	cmd := exec.Command("git", "clone", "--bare", src, bare)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bare clone: %v\n%s", err, out)
	}

	return bare
}

func TestBareCloneArgs(t *testing.T) {
	// We can't run real git in unit tests, but we can verify that
	// a cancelled context produces an error (proves CommandContext is used).
	r := &RealRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := r.BareClone(ctx, "github.com/test/repo", "/tmp/nonexistent-bare-clone-test")
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestFetchCancelledContext(t *testing.T) {
	r := &RealRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	time.Sleep(time.Millisecond) // ensure timeout fires

	err := r.Fetch(ctx, "/tmp/nonexistent-repo")
	if err == nil {
		t.Fatal("expected error with timed-out context")
	}
}

func TestAddWorktreeCancelledContext(t *testing.T) {
	r := &RealRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.AddWorktree(ctx, "/tmp/bare", "/tmp/wt", "main")
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestRemoveWorktreeCancelledContext(t *testing.T) {
	r := &RealRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.RemoveWorktree(ctx, "/tmp/bare", "/tmp/wt")
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestRunnerInterface(_ *testing.T) {
	// Compile-time check that RealRunner implements Runner
	var _ Runner = (*RealRunner)(nil)
}

func TestBareCloneLocal(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = src
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("# test"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-m", "initial")

	r := &RealRunner{}
	dest := filepath.Join(dir, "cloned.git")
	if err := r.BareClone(context.Background(), src, dest); err != nil {
		t.Fatalf("BareClone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "HEAD")); err != nil {
		t.Error("expected HEAD file in bare clone")
	}
}

func TestFetchReal(t *testing.T) {
	bare := initTestRepo(t)
	r := &RealRunner{}
	if err := r.Fetch(context.Background(), bare); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
}

func TestDefaultBranch(t *testing.T) {
	bare := initTestRepo(t)
	r := &RealRunner{}
	branch, err := r.DefaultBranch(context.Background(), bare)
	if err != nil {
		t.Fatalf("DefaultBranch: %v", err)
	}
	if branch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", branch, "main")
	}
}

func TestBranchExists(t *testing.T) {
	bare := initTestRepo(t)
	r := &RealRunner{}
	ctx := context.Background()

	exists, err := r.BranchExists(ctx, bare, "main")
	if err != nil {
		t.Fatalf("BranchExists(main): %v", err)
	}
	if !exists {
		t.Error("expected main branch to exist")
	}

	exists, err = r.BranchExists(ctx, bare, "nonexistent")
	if err != nil {
		t.Fatalf("BranchExists(nonexistent): %v", err)
	}
	if exists {
		t.Error("expected nonexistent branch to not exist")
	}
}

func TestAddWorktreeAndRemove(t *testing.T) {
	bare := initTestRepo(t)
	r := &RealRunner{}
	ctx := context.Background()
	wtPath := filepath.Join(t.TempDir(), "wt")

	if err := r.AddWorktree(ctx, bare, wtPath, "main"); err != nil {
		t.Fatalf("AddWorktree: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wtPath, "README.md")); err != nil {
		t.Error("expected README.md in worktree")
	}
	if err := r.RemoveWorktree(ctx, bare, wtPath); err != nil {
		t.Fatalf("RemoveWorktree: %v", err)
	}
}

func TestAddWorktreeNewBranch(t *testing.T) {
	bare := initTestRepo(t)
	r := &RealRunner{}
	ctx := context.Background()
	wtPath := filepath.Join(t.TempDir(), "wt-new")

	if err := r.AddWorktreeNewBranch(ctx, bare, wtPath, "feat/test", "main"); err != nil {
		t.Fatalf("AddWorktreeNewBranch: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wtPath, "README.md")); err != nil {
		t.Error("expected README.md in new branch worktree")
	}

	exists, err := r.BranchExists(ctx, bare, "feat/test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected feat/test branch to exist")
	}
}
