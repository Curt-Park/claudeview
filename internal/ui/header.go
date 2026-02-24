package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HeaderModel holds state for the header bar.
type HeaderModel struct {
	ProjectName string
	Model       string
	MCPCount    int
	Mode        string // normal, plan, auto, etc.
	Width       int
}

// View renders the header bar.
func (h HeaderModel) View() string {
	left := " claudeview"
	if h.ProjectName != "" {
		left += fmt.Sprintf("  |  Project: %s", h.ProjectName)
	}
	if h.Model != "" {
		left += fmt.Sprintf("  |  Model: %s", h.Model)
	}
	if h.MCPCount > 0 {
		left += fmt.Sprintf("  |  MCP: %d", h.MCPCount)
	}

	right := ""
	if h.Mode != "" {
		right = fmt.Sprintf("Mode: %s ", h.Mode)
	}

	// Pad between left and right
	padding := max(h.Width-lipgloss.Width(left)-lipgloss.Width(right), 0)

	line := left + strings.Repeat(" ", padding) + right
	return StyleHeader.Width(h.Width).Render(line)
}
