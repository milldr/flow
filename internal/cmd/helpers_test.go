package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/milldr/flow/internal/ui"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world", 8, "hello..."},
		{"empty", "", 10, ""},
		{"small max clamps to 4", "hello world", 2, "h..."},
		{"long string", strings.Repeat("a", 100), 10, "aaaaaaa..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ui.Truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
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
			got := ui.RelativeTime(tt.t)
			if got != tt.want {
				t.Errorf("RelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}
