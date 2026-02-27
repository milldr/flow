package status

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const checkTimeout = 10 * time.Second

// CheckRunner executes a status check command and returns whether it matched.
type CheckRunner interface {
	RunCheck(ctx context.Context, command string, env []string) bool
}

// ShellRunner executes check commands via sh -c.
type ShellRunner struct{}

// RunCheck executes the command in a shell. Returns true if exit code is 0.
func (r *ShellRunner) RunCheck(ctx context.Context, command string, env []string) bool {
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = append(os.Environ(), env...)
	return cmd.Run() == nil
}

// Resolver resolves workspace statuses using a CheckRunner.
type Resolver struct {
	Runner CheckRunner
}

// RepoSlug converts a repo URL to owner/repo format for gh CLI.
// Handles SSH (git@github.com:org/repo.git) and HTTPS (github.com/org/repo) URLs.
func RepoSlug(url string) string {
	s := url
	// SSH: git@github.com:org/repo.git -> org/repo.git
	if i := strings.Index(s, ":"); strings.Contains(s, "@") && i > 0 {
		s = s[i+1:]
	} else {
		// HTTPS: github.com/org/repo -> org/repo
		parts := strings.SplitN(s, "/", 2)
		if len(parts) == 2 {
			s = parts[1]
		}
	}
	return strings.TrimSuffix(s, ".git")
}

// buildEnv creates the environment variables for a status check.
func buildEnv(repo RepoInfo, wsID, wsName string) []string {
	return []string{
		"FLOW_REPO_URL=" + repo.URL,
		"FLOW_REPO_BRANCH=" + repo.Branch,
		"FLOW_REPO_PATH=" + repo.Path,
		"FLOW_REPO_SLUG=" + RepoSlug(repo.URL),
		"FLOW_WORKSPACE_ID=" + wsID,
		"FLOW_WORKSPACE_NAME=" + wsName,
	}
}

// ResolveRepo determines the status of a single repo by evaluating checks
// in order. Returns the name of the first matching status or the default.
func (r *Resolver) ResolveRepo(ctx context.Context, spec *Spec, repo RepoInfo, wsID, wsName string) string {
	env := buildEnv(repo, wsID, wsName)

	for _, entry := range spec.Spec.Statuses {
		if entry.Default {
			return entry.Name
		}
		if r.Runner.RunCheck(ctx, entry.Check, env) {
			return entry.Name
		}
	}

	return ""
}

// ResolveWorkspace resolves the status for each repo concurrently and
// returns the aggregated workspace result. The workspace status is the
// least-advanced status (highest index in spec order) across all repos,
// excluding skipped repos and repos at the default status.
func (r *Resolver) ResolveWorkspace(ctx context.Context, spec *Spec, repos []RepoInfo, wsID, wsName string) *WorkspaceResult {
	wsStart := time.Now()
	result := &WorkspaceResult{
		WorkspaceID:   wsID,
		WorkspaceName: wsName,
		Repos:         make([]RepoResult, len(repos)),
	}

	// Find the default status index.
	defaultIdx := -1
	for i, e := range spec.Spec.Statuses {
		if e.Default {
			defaultIdx = i
			break
		}
	}

	if len(repos) == 0 {
		if defaultIdx >= 0 {
			result.Status = spec.Spec.Statuses[defaultIdx].Name
		}
		result.Duration = time.Since(wsStart)
		return result
	}

	skipped := make([]bool, len(repos))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(idx int, rp RepoInfo) {
			defer wg.Done()
			repoStart := time.Now()
			env := buildEnv(rp, wsID, wsName)

			// Run skip check if configured.
			skip := false
			if spec.Spec.Skip != "" {
				skip = r.Runner.RunCheck(ctx, spec.Spec.Skip, env)
			}

			status := r.ResolveRepo(ctx, spec, rp, wsID, wsName)
			mu.Lock()
			skipped[idx] = skip
			result.Repos[idx] = RepoResult{
				URL:      rp.URL,
				Branch:   rp.Branch,
				Status:   status,
				Duration: time.Since(repoStart),
			}
			mu.Unlock()
		}(i, repo)
	}

	wg.Wait()
	result.Duration = time.Since(wsStart)

	// Build index map for ordering.
	orderIndex := make(map[string]int, len(spec.Spec.Statuses))
	for i, e := range spec.Spec.Statuses {
		orderIndex[e.Name] = i
	}

	// Workspace status = least advanced (highest index) across repos,
	// excluding skipped repos and repos at the default status.
	worstIdx := -1
	for i, rr := range result.Repos {
		if skipped[i] {
			continue
		}
		idx, ok := orderIndex[rr.Status]
		if !ok {
			continue
		}
		if idx == defaultIdx {
			continue
		}
		if idx > worstIdx {
			worstIdx = idx
		}
	}

	if worstIdx >= 0 {
		result.Status = spec.Spec.Statuses[worstIdx].Name
	} else if defaultIdx >= 0 {
		result.Status = spec.Spec.Statuses[defaultIdx].Name
	}

	return result
}
