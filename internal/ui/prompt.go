package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

// slugify converts a description into a workspace name.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
		s = strings.TrimRight(s, "-")
	}
	return s
}

// RepoEntry holds a repo's URL, branch, and path from the init prompt.
type RepoEntry struct {
	URL    string
	Branch string
	Path   string
}

// InitPrompt runs the interactive workspace init wizard.
func InitPrompt() (name, description string, repos []RepoEntry, err error) {
	// Step 1: Description
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What are you working on?").
				Placeholder("e.g., Add IPv6 support to VPC module").
				Value(&description),
		),
	).Run()
	if err != nil {
		return
	}

	// Step 2: Name (auto-generated from description, editable)
	name = slugify(description)
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Workspace name").
				Value(&name),
		),
	).Run()
	if err != nil {
		return
	}

	// Step 3: Repos (loop until empty)
	for {
		var repoURL string
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Add a repo (blank to finish)").
					Placeholder("e.g., github.com/org/repo").
					Value(&repoURL),
			),
		).Run()
		if err != nil {
			return
		}

		if repoURL == "" {
			break
		}

		var branch string
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Branch for " + repoURL + "?").
					Placeholder("e.g., main").
					Value(&branch),
			),
		).Run()
		if err != nil {
			return
		}

		// Derive path from repo URL (last segment)
		parts := strings.Split(repoURL, "/")
		defaultPath := "./" + parts[len(parts)-1]

		repos = append(repos, RepoEntry{
			URL:    repoURL,
			Branch: branch,
			Path:   defaultPath,
		})
	}

	return
}

// ConfirmDelete prompts the user to confirm workspace deletion.
func ConfirmDelete(name string, repos []string) (bool, error) {
	Warning("Delete workspace '" + name + "'?")
	Print("    This will remove:")
	for _, r := range repos {
		Printf("    - %s (worktree)\n", r)
	}
	Print("")

	var confirm bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirm?").
				Value(&confirm),
		),
	).Run()
	return confirm, err
}
