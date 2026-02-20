---
name: adding-commands
description: How to add new CLI commands to flow. Use when creating a new subcommand, adding flags, or modifying the command tree.
---

# Adding Commands

All commands live in `internal/cmd/` and use cobra. Follow the existing pattern exactly.

## File Structure

Each command gets its own file: `internal/cmd/<name>.go`.

## Command Pattern

Every command follows this structure:

```go
package cmd

import (
    "github.com/milldr/flow/internal/ui"
    "github.com/milldr/flow/internal/workspace"
    "github.com/spf13/cobra"
)

func newXxxCmd(svc *workspace.Service) *cobra.Command {
    return &cobra.Command{
        Use:     "xxx <workspace>",
        Short:   "One-line description",
        Args:    cobra.ExactArgs(1),
        Example: `  flow xxx calm-delta`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Resolve workspace
            id, st, err := resolveWorkspace(svc, args[0])
            if err != nil {
                return err
            }

            // 2. Business logic (often wrapped in spinner)
            name := workspaceDisplayName(id, st)
            err = ui.RunWithSpinner("Doing thing: "+name, func(report func(string)) error {
                // call svc methods here
                return nil
            })
            if err != nil {
                return err
            }

            // 3. Success output
            ui.Success("Done: " + name)
            return nil
        },
    }
}
```

## Wiring Up

Register the command in `internal/cmd/root.go` inside `newRootCmd()`:

```go
root.AddCommand(newXxxCmd(svc))
```

The `svc` (*workspace.Service) is the only dependency passed to commands. It provides access to config, git runner, and all workspace operations.

## Key Helpers

These are defined in `internal/cmd/resolve.go`:

- `resolveWorkspace(svc, idOrName)` — resolves by ID or name, handles ambiguity with interactive prompt. Returns `(id, *state.State, error)`.
- `workspaceDisplayName(id, st)` — returns name if set, otherwise ID.

## Flags

Add flags to the command, not globally:

```go
func newXxxCmd(svc *workspace.Service) *cobra.Command {
    var myFlag bool

    cmd := &cobra.Command{
        // ...
    }

    cmd.Flags().BoolVarP(&myFlag, "flag-name", "f", false, "Description")
    return cmd
}
```

## Output Rules

- All user-facing output goes through `internal/ui/` — never `fmt.Print*` directly
- Use `ui.Success()`, `ui.Warning()`, `ui.Error()`, `ui.Info()` for prefixed messages
- Use `ui.Code()` to style inline commands
- Long operations get `ui.RunWithSpinner()`
- Errors are returned, not printed — the root command handles display
