package ui

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Add IPv6 support", "add-ipv6-support"},
		{"special chars", "Fix bug #123!", "fix-bug-123"},
		{"already slug", "my-workspace", "my-workspace"},
		{"mixed case", "Hello World", "hello-world"},
		{"leading trailing special", "---hello---", "hello"},
		{"empty", "", ""},
		{"all special", "!!!", ""},
		{"long string truncated", "this is a very long description that should be truncated at forty characters exactly", "this-is-a-very-long-description-that-sho"},
		{"spaces and dashes", "  hello - world  ", "hello-world"},
		{"numbers", "v2 release", "v2-release"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
