package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var testDisplay = StatusDisplayConfig{
	Order:  map[string]int{"open": 0, "in-progress": 1, "closed": 2},
	Colors: map[string]string{"closed": "2", "open": "8"},
}

func TestStatusTableModel_ResolvedMsgUpdatesRow(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "2", Created: "1d ago"},
		{Name: "ws-2", Repos: "1", Created: "3d ago"},
	}
	m := newStatusTableModel(rows, testDisplay)

	updated, _ := m.Update(StatusResolvedMsg{Index: 0, Status: "closed"})
	m = updated.(statusTableModel)

	if m.statuses[0] != "closed" {
		t.Errorf("statuses[0] = %q, want \"closed\"", m.statuses[0])
	}
	if m.resolved != 1 {
		t.Errorf("resolved = %d, want 1", m.resolved)
	}
	if m.done {
		t.Error("should not be done with 1 of 2 resolved")
	}
}

func TestStatusTableModel_AllResolvedQuits(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "1", Created: "1d ago"},
		{Name: "ws-2", Repos: "2", Created: "2d ago"},
	}
	m := newStatusTableModel(rows, testDisplay)

	updated, _ := m.Update(StatusResolvedMsg{Index: 0, Status: "closed"})
	m = updated.(statusTableModel)

	updated, cmd := m.Update(StatusResolvedMsg{Index: 1, Status: "open"})
	m = updated.(statusTableModel)

	if !m.done {
		t.Error("expected done after all rows resolved")
	}
	if m.statuses[0] != "closed" || m.statuses[1] != "open" {
		t.Errorf("unexpected statuses: %v", m.statuses)
	}
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestStatusTableModel_CtrlCQuits(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "1", Created: "1d ago"},
	}
	m := newStatusTableModel(rows, testDisplay)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(statusTableModel)

	if !m.done {
		t.Error("expected done after ctrl+c")
	}
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
}

func TestStatusTableModel_OutOfBoundsIgnored(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "1", Created: "1d ago"},
	}
	m := newStatusTableModel(rows, testDisplay)

	updated, _ := m.Update(StatusResolvedMsg{Index: 5, Status: "closed"})
	m = updated.(statusTableModel)

	if m.resolved != 0 {
		t.Errorf("resolved = %d, want 0 (out of bounds should be ignored)", m.resolved)
	}
}

func TestStatusTableModel_ViewShowsSpinnerForPending(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "1", Created: "1d ago"},
		{Name: "ws-2", Repos: "2", Created: "2d ago"},
	}
	m := newStatusTableModel(rows, testDisplay)

	// Resolve one row.
	updated, _ := m.Update(StatusResolvedMsg{Index: 0, Status: "closed"})
	m = updated.(statusTableModel)

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
	// The resolved row should contain the status text.
	if !containsSubstring(view, "closed") {
		t.Error("view should contain resolved status \"closed\"")
	}
	// The pending row should show the placeholder.
	if !containsSubstring(view, "...") {
		t.Error("view should contain \"...\" for pending row")
	}
}

func TestStatusTableModel_ViewAfterDone(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "1", Created: "1d ago"},
	}
	m := newStatusTableModel(rows, testDisplay)

	updated, _ := m.Update(StatusResolvedMsg{Index: 0, Status: "in-progress"})
	m = updated.(statusTableModel)

	view := m.View()
	if !containsSubstring(view, "in-progress") {
		t.Error("final view should contain the resolved status")
	}
}

func TestRunStatusTablePlain(t *testing.T) {
	rows := []StatusRow{
		{Name: "ws-1", Repos: "2", Created: "1d ago"},
		{Name: "ws-2", Repos: "1", Created: "3d ago"},
	}

	var called bool
	resolved, err := runStatusTablePlain(rows, testDisplay, func(send func(StatusResolvedMsg)) {
		called = true
		send(StatusResolvedMsg{Index: 0, Status: "closed"})
		send(StatusResolvedMsg{Index: 1, Status: "open"})
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected resolve function to be called")
	}
	if resolved[0] != "closed" || resolved[1] != "open" {
		t.Errorf("unexpected resolved statuses: %v", resolved)
	}
}

func containsSubstring(s, sub string) bool {
	return len(sub) == 0 || len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
