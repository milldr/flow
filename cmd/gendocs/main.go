// Package main generates CLI reference documentation using cobra/doc.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/milldr/flow/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "docs/commands"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	root := cmd.NewRootCmd()
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, dir); err != nil {
		log.Fatal(err)
	}

	// Inject GIF demos into generated docs where a matching GIF exists.
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || !strings.HasPrefix(e.Name(), "flow") {
			continue
		}
		// flow_init.md -> init
		name := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "flow_"), ".md")
		gifPath := filepath.Join("tapes", name+".gif")
		if _, err := os.Stat(filepath.Join(dir, gifPath)); err != nil {
			continue
		}
		mdPath := filepath.Join(dir, e.Name())
		content, err := os.ReadFile(mdPath)
		if err != nil {
			log.Fatal(err)
		}
		lines := strings.SplitN(string(content), "\n", 4)
		if len(lines) < 4 {
			continue
		}
		// Insert GIF after heading and short description (lines 0-2),
		// before Synopsis (line 3+).
		out := fmt.Sprintf("%s\n%s\n%s\n\n![flow %s](%s)\n\n%s",
			lines[0], lines[1], lines[2], name, gifPath, lines[3])
		if err := os.WriteFile(mdPath, []byte(out), 0o644); err != nil {
			log.Fatal(err)
		}
	}
}
