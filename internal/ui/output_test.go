package ui

import (
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
			got := StatusStyle(tt.status)
			if !strings.Contains(got, tt.want) {
				t.Errorf("StatusStyle(%q) = %q, does not contain %q", tt.status, got, tt.want)
			}
		})
	}

	// Unknown status should be returned unstyled (no ANSI escape codes).
	got := StatusStyle("unknown")
	if got != "unknown" {
		t.Errorf("StatusStyle(\"unknown\") = %q, want plain \"unknown\"", got)
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
