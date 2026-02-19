// Package ui provides styled terminal output and interactive prompts.
package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	successPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✅")
	warningPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⚠️ ")
	errorPrefix   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("❌")
	infoPrefix    = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("📦")
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// Success prints a success message with a green prefix.
func Success(msg string) {
	fmt.Printf("%s %s\n", successPrefix, msg)
}

// Warning prints a warning message with a yellow prefix.
func Warning(msg string) {
	fmt.Printf("%s %s\n", warningPrefix, msg)
}

// Error prints an error message with a red prefix.
func Error(msg string) {
	fmt.Printf("%s %s\n", errorPrefix, msg)
}

// Info prints an info message with a blue prefix.
func Info(msg string) {
	fmt.Printf("%s %s\n", infoPrefix, msg)
}

// Dim returns the message styled in a dim color.
func Dim(msg string) string {
	return dimStyle.Render(msg)
}

// Print writes a line to stdout.
func Print(msg string) {
	fmt.Println(msg)
}

// Printf writes formatted output to stdout.
func Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}
