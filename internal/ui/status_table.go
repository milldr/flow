package ui

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
)

// statusOrder defines the display sort order for statuses.
var statusOrder = map[string]int{
	"open":        0,
	"in-progress": 1,
	"in-review":   2,
	"stale":       3,
	"closed":      4,
}

// StatusRow holds the static columns for one workspace row.
type StatusRow struct {
	Name         string
	Repos        string
	Created      string
	CreatedAt    time.Time // used for sorting within status groups
	CachedStatus string    // last-known status for initial sort order
}

// StatusResolvedMsg signals that a row's status has been resolved.
type StatusResolvedMsg struct {
	Index  int
	Status string
}

type statusTableModel struct {
	rows     []StatusRow
	statuses []string // resolved status per row; empty = pending
	spinner  spinner.Model
	done     bool
	total    int
	resolved int
}

func newStatusTableModel(rows []StatusRow) statusTableModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	statuses := make([]string, len(rows))
	return statusTableModel{
		rows:     rows,
		statuses: statuses,
		spinner:  s,
		total:    len(rows),
	}
}

func (m statusTableModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m statusTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StatusResolvedMsg:
		if msg.Index >= 0 && msg.Index < len(m.statuses) {
			m.statuses[msg.Index] = msg.Status
			m.resolved++
		}
		if m.resolved >= m.total {
			m.done = true
			return m, tea.Quit
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// sortedOrder returns indices sorted by status group then created time.
func sortedOrder(rows []StatusRow, statuses []string) []int {
	indices := make([]int, len(rows))
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(a, b int) bool {
		ia, ib := indices[a], indices[b]
		oa, oka := statusOrder[statuses[ia]]
		ob, okb := statusOrder[statuses[ib]]
		if !oka {
			oa = len(statusOrder)
		}
		if !okb {
			ob = len(statusOrder)
		}
		if oa != ob {
			return oa < ob
		}
		return rows[ia].CreatedAt.Before(rows[ib].CreatedAt)
	})
	return indices
}

func (m statusTableModel) View() string {
	if m.done {
		// Final render: sort by resolved statuses.
		return m.renderTable(false, m.statuses) + "\n"
	}
	// Live render: sort by cached statuses for stable initial ordering.
	cached := make([]string, len(m.rows))
	for i, row := range m.rows {
		cached[i] = row.CachedStatus
	}
	return m.renderTable(true, cached) + "\n"
}

func (m statusTableModel) renderTable(showSpinner bool, sortStatuses []string) string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).PaddingRight(2)
	cellStyle := lipgloss.NewStyle().PaddingRight(2)

	headers := []string{"NAME", "STATUS", "REPOS", "CREATED"}
	order := sortedOrder(m.rows, sortStatuses)

	var rows [][]string
	for _, i := range order {
		row := m.rows[i]
		var statusCell string
		if m.statuses[i] == "" {
			if showSpinner {
				statusCell = m.spinner.View() + " ..."
			} else {
				statusCell = "..."
			}
		} else {
			statusCell = StatusStyle(m.statuses[i])
		}
		rows = append(rows, []string{row.Name, statusCell, row.Repos, row.Created})
	}

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

// RunStatusTable displays a live-updating table in TTY mode. The resolve
// function runs concurrently and sends StatusResolvedMsg via send as each
// workspace status is determined. In non-TTY mode, resolve runs to completion
// and a static table is printed. Returns the resolved statuses indexed by row.
func RunStatusTable(rows []StatusRow, resolve func(send func(StatusResolvedMsg))) ([]string, error) {
	if plain || !isatty.IsTerminal(os.Stdout.Fd()) {
		return runStatusTablePlain(rows, resolve)
	}

	m := newStatusTableModel(rows)
	p := tea.NewProgram(m)

	go resolve(func(msg StatusResolvedMsg) {
		p.Send(msg)
	})

	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := result.(statusTableModel)
	if final.resolved < final.total {
		return final.statuses, ErrInterrupted
	}
	return final.statuses, nil
}

func runStatusTablePlain(rows []StatusRow, resolve func(send func(StatusResolvedMsg))) ([]string, error) {
	statuses := make([]string, len(rows))
	resolve(func(msg StatusResolvedMsg) {
		if msg.Index >= 0 && msg.Index < len(statuses) {
			statuses[msg.Index] = msg.Status
		}
	})

	order := sortedOrder(rows, statuses)

	headers := []string{"NAME", "STATUS", "REPOS", "CREATED"}
	var tableRows [][]string
	for _, i := range order {
		s := statuses[i]
		if s == "" {
			s = "-"
		}
		tableRows = append(tableRows, []string{rows[i].Name, s, rows[i].Repos, rows[i].Created})
	}

	fmt.Println(Table(headers, tableRows))
	return statuses, nil
}

// FormatDuration formats a duration as a short human-readable string.
func FormatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}
