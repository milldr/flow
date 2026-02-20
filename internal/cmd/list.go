package cmd

import (
	"fmt"

	"github.com/milldr/flow/internal/ui"
	"github.com/milldr/flow/internal/workspace"
	"github.com/spf13/cobra"
)

func newListCmd(svc *workspace.Service) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all workspaces",
		Aliases: []string{"ls"},
		RunE: func(_ *cobra.Command, _ []string) error {
			infos, err := svc.List()
			if err != nil {
				return err
			}

			if len(infos) == 0 {
				ui.Print("No workspaces found. Run `flow init` to create one.")
				return nil
			}

			// Count name occurrences to detect duplicates
			nameCounts := make(map[string]int)
			for _, info := range infos {
				if info.Name != "" {
					nameCounts[info.Name]++
				}
			}

			// Track per-name index for duplicate labeling
			nameIndex := make(map[string]int)

			headers := []string{"ID", "NAME", "DESCRIPTION", "REPOS", "CREATED"}
			var rows [][]string
			for _, info := range infos {
				displayName := "-"
				if info.Name != "" {
					displayName = info.Name
				}
				if info.Name != "" && nameCounts[info.Name] > 1 {
					nameIndex[info.Name]++
					displayName = fmt.Sprintf("%s (%d)", info.Name, nameIndex[info.Name])
				}

				desc := ui.Truncate(info.Description, 40)
				if desc == "" {
					desc = "-"
				}

				rows = append(rows, []string{
					info.ID,
					displayName,
					desc,
					fmt.Sprintf("%d", info.RepoCount),
					ui.RelativeTime(info.Created),
				})
			}

			fmt.Println(ui.Table(headers, rows))
			return nil
		},
	}
}
