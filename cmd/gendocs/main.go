// Package main generates CLI reference documentation using cobra/doc.
package main

import (
	"log"
	"os"

	"github.com/milldr/flow/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	dir := "docs/reference"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	root := cmd.NewRootCmd()
	root.DisableAutoGenTag = true
	if err := doc.GenMarkdownTree(root, dir); err != nil {
		log.Fatal(err)
	}
}
