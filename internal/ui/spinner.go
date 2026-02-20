package ui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// ErrInterrupted is returned when the spinner is cancelled via ctrl+c.
var ErrInterrupted = errors.New("interrupted")

// plain forces plain text output even in TTY mode (e.g. when verbose logging is active).
var plain bool

// SetPlain forces all spinner output to use plain text mode.
func SetPlain(v bool) {
	plain = v
}

// SpinnerOp is the function executed while the spinner is displayed.
// Call report to emit progress lines shown below the spinner.
type SpinnerOp func(report func(msg string)) error

type progressMsg string
type doneMsg struct{ err error }

type spinnerModel struct {
	spinner spinner.Model
	title   string
	lines   []string
	done    bool
	err     error
}

func newSpinnerModel(title string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	return spinnerModel{spinner: s, title: title}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressMsg:
		m.lines = append(m.lines, string(msg))
		return m, nil
	case doneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			m.err = ErrInterrupted
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n", m.spinner.View(), m.title)
	for _, line := range m.lines {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	return b.String()
}

// RunWithSpinner runs op while displaying an animated spinner in TTY mode.
// In non-TTY mode it falls back to plain text output.
func RunWithSpinner(title string, op SpinnerOp) error {
	if plain || !isatty.IsTerminal(os.Stdout.Fd()) {
		return runPlain(title, op)
	}

	m := newSpinnerModel(title)
	p := tea.NewProgram(m)

	go func() {
		err := op(func(msg string) {
			p.Send(progressMsg(msg))
		})
		p.Send(doneMsg{err: err})
	}()

	result, err := p.Run()
	if err != nil {
		return err
	}

	final := result.(spinnerModel)
	return final.err
}

func runPlain(title string, op SpinnerOp) error {
	Info(title)
	return op(func(msg string) {
		Printf("  %s\n", msg)
	})
}
