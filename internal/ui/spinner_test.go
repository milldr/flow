package ui

import (
	"errors"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSpinnerModel_ProgressAccumulatesLines(t *testing.T) {
	m := newSpinnerModel("test")

	updated, _ := m.Update(progressMsg("line 1"))
	m = updated.(spinnerModel)

	updated, _ = m.Update(progressMsg("line 2"))
	m = updated.(spinnerModel)

	if len(m.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(m.lines))
	}
	if m.lines[0] != "line 1" || m.lines[1] != "line 2" {
		t.Fatalf("unexpected lines: %v", m.lines)
	}
}

func TestSpinnerModel_DoneNilSetsQuit(t *testing.T) {
	m := newSpinnerModel("test")

	updated, cmd := m.Update(doneMsg{nil})
	m = updated.(spinnerModel)

	if !m.done {
		t.Fatal("expected done to be true")
	}
	if m.err != nil {
		t.Fatalf("expected nil error, got %v", m.err)
	}
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	// Verify it produces a QuitMsg
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestSpinnerModel_DoneErrorPropagates(t *testing.T) {
	m := newSpinnerModel("test")
	testErr := fmt.Errorf("test error")

	updated, cmd := m.Update(doneMsg{err: testErr})
	m = updated.(spinnerModel)

	if !m.done {
		t.Fatal("expected done to be true")
	}
	if !errors.Is(m.err, testErr) {
		t.Fatalf("expected test error, got %v", m.err)
	}
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
}

func TestSpinnerModel_ViewEmptyWhenDone(t *testing.T) {
	m := newSpinnerModel("test")

	updated, _ := m.Update(doneMsg{nil})
	m = updated.(spinnerModel)

	if v := m.View(); v != "" {
		t.Fatalf("expected empty view when done, got %q", v)
	}
}

func TestRunPlain(t *testing.T) {
	var called bool
	var reportedMsgs []string

	err := runPlain("test title", func(report func(string)) error {
		called = true
		report("msg1")
		report("msg2")
		reportedMsgs = append(reportedMsgs, "msg1", "msg2")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected op to be called")
	}
	if len(reportedMsgs) != 2 {
		t.Fatalf("expected 2 reported messages, got %d", len(reportedMsgs))
	}
}

func TestRunPlain_Error(t *testing.T) {
	testErr := fmt.Errorf("op failed")

	err := runPlain("test title", func(_ func(string)) error {
		return testErr
	})

	if !errors.Is(err, testErr) {
		t.Fatalf("expected test error, got %v", err)
	}
}
