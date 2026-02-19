// Package main is the entry point for the flow CLI.
package main

import (
	"os"

	"github.com/milldr/flow/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
