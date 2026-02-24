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

// ContentText returns the header content as plain text (no styling or padding).
// Used to embed the header in the top border line.
func (h HeaderModel) ContentText() string {
	text := "claudeview"
	if h.ProjectName != "" {
		text += "  │  " + h.ProjectName
	}
	if h.Model != "" {
		text += "  │  " + h.Model
	}
	if h.MCPCount > 0 {
		text += fmt.Sprintf("  │  MCP: %d", h.MCPCount)
	}
	if h.Mode != "" {
		text += "  │  " + h.Mode
	}
	return text
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
