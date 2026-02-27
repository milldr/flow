package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short", "abc", 10, "abc"},
		{"exact", "abcde", 5, "abcde"},
		{"truncated", "abcdefghij", 7, "abcd..."},
		{"empty", "", 10, ""},
		{"small max clamps to 4", "abcdef", 2, "a..."},
		{"long string", "the quick brown fox jumps", 10, "the qui..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"zero", time.Time{}, "unknown"},
		{"just now", time.Now(), "just now"},
		{"30 seconds ago", time.Now().Add(-30 * time.Second), "just now"},
		{"5 minutes ago", time.Now().Add(-5 * time.Minute), "5m ago"},
		{"3 hours ago", time.Now().Add(-3 * time.Hour), "3h ago"},
		{"2 days ago", time.Now().Add(-48 * time.Hour), "2d ago"},
		{"30 days ago", time.Now().Add(-30 * 24 * time.Hour), "30d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RelativeTime(tt.time)
			if got != tt.want {
				t.Errorf("RelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatusStyle(t *testing.T) {
	colorMap := map[string]string{"closed": "2", "open": "8"}

	tests := []struct {
		status string
		want   string // substring that must appear in styled output
	}{
		{"closed", "closed"},
		{"in-review", "in-review"},
		{"in-progress", "in-progress"},
		{"open", "open"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusStyle(tt.status, colorMap)
			if !strings.Contains(got, tt.want) {
				t.Errorf("StatusStyle(%q) = %q, does not contain %q", tt.status, got, tt.want)
			}
		})
	}

	// Unknown status should be returned unstyled (no ANSI escape codes).
	got := StatusStyle("unknown", colorMap)
	if got != "unknown" {
		t.Errorf("StatusStyle(\"unknown\") = %q, want plain \"unknown\"", got)
	}
}

func TestCode(t *testing.T) {
	got := Code("flow render")
	if !strings.Contains(got, "flow render") {
		t.Errorf("Code() = %q, does not contain input text", got)
	}
}

// captureStdout captures stdout during fn execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// captureStderr captures stderr during fn execution.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stderr
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestSuccess(t *testing.T) {
	out := captureStdout(t, func() { Success("it worked") })
	if !strings.Contains(out, "it worked") {
		t.Errorf("Success() output = %q, want it to contain 'it worked'", out)
	}
}

func TestWarning(t *testing.T) {
	out := captureStdout(t, func() { Warning("watch out") })
	if !strings.Contains(out, "watch out") {
		t.Errorf("Warning() output = %q, want it to contain 'watch out'", out)
	}
}

func TestError(t *testing.T) {
	out := captureStderr(t, func() { Error("bad thing") })
	if !strings.Contains(out, "bad thing") {
		t.Errorf("Error() output = %q, want it to contain 'bad thing'", out)
	}
}

func TestErrorf(t *testing.T) {
	out := captureStderr(t, func() { Errorf("code %d", 42) })
	if !strings.Contains(out, "code 42") {
		t.Errorf("Errorf() output = %q, want it to contain 'code 42'", out)
	}
}

func TestPrint(t *testing.T) {
	out := captureStdout(t, func() { Print("hello") })
	if !strings.Contains(out, "hello") {
		t.Errorf("Print() output = %q, want it to contain 'hello'", out)
	}
}

func TestPrintf(t *testing.T) {
	out := captureStdout(t, func() { Printf("num=%d", 7) })
	if !strings.Contains(out, "num=7") {
		t.Errorf("Printf() output = %q, want it to contain 'num=7'", out)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{"zero", 0, "0ms"},
		{"sub-second", 450, "450ms"},
		{"exactly one second", 1000, "1.0s"},
		{"fractional seconds", 1500, "1.5s"},
		{"multi-second", 3200, "3.2s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.ms)
			if got != tt.want {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
			}
		})
	}
}
