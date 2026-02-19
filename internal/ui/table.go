package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Table renders a styled table with the given headers and rows.
func Table(headers []string, rows [][]string) string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).PaddingRight(2)
	cellStyle := lipgloss.NewStyle().PaddingRight(2)

	t := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderHeader(false).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		})

	return t.Render()
}
