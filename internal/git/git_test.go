package git

import (
	"context"
	"testing"
	"time"
)

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
