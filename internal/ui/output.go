// Package ui provides styled terminal output and interactive prompts.
package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	successPrefix = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).Render("✓")
	warningPrefix = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")).Render("!")
	errorPrefix   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render("✗")
	infoPrefix    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Render("●")
	codeStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
)

// Success prints a success message with a green prefix.
func Success(msg string) {
	fmt.Printf("%s %s\n", successPrefix, msg)
}

// Warning prints a warning message with a yellow prefix.
func Warning(msg string) {
	fmt.Printf("%s %s\n", warningPrefix, msg)
}

// Error prints an error message with a red prefix to stderr.
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", errorPrefix, msg)
}

// Errorf prints a formatted error message with a red prefix to stderr.
func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "%s %s\n", errorPrefix, fmt.Sprintf(format, a...))
}

// Info prints an info message with a blue prefix.
func Info(msg string) {
	fmt.Printf("%s %s\n", infoPrefix, msg)
}

// Code returns the string styled as an inline command (bold cyan).
func Code(s string) string {
	return codeStyle.Render(s)
}

// Print writes a line to stdout.
func Print(msg string) {
	fmt.Println(msg)
}

// Printf writes formatted output to stdout.
func Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}

// Truncate shortens a string to maxLen characters, appending "..." if needed.
func Truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RelativeTime returns a human-friendly relative time string.
func RelativeTime(t time.Time) string {
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
