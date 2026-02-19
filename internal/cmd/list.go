package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

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

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 3, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tDESCRIPTION\tREPOS\tCREATED")
			for _, info := range infos {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
					info.Name,
					truncate(info.Description, 40),
					info.RepoCount,
					relativeTime(info.Created),
				)
			}
			return w.Flush()
		},
	}
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
