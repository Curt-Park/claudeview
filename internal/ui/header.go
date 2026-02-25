package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// InfoModel holds the top info panel data (k9s-style left column + right menu).
type InfoModel struct {
	Project       string // active project display name / path
	Session       string // active session short ID
	User          string // OS username
	ClaudeVersion string // Claude Code binary version
	AppVersion    string // claudeview binary version
	CPUPercent    float64
	MemMiB        uint64
	Width         int
}

// ViewWithMenu renders the 7-row info panel alongside key binding hints.
func (info InfoModel) ViewWithMenu(items []MenuItem) string {
	const labelW = 14 // visible chars reserved for label column
	const leftW = 46  // total visible chars for the left column (label + value)

	dim := StyleDim

	val := func(s string) string {
		if s == "" {
			return dim.Render("--")
		}
		return s
	}

	leftRows := []struct{ label, value string }{
		{"Project:", truncateLeft(info.Project, leftW-labelW-1)},
		{"Session:", val(info.Session)},
		{"User:", val(info.User)},
		{"Claude Code:", val(info.ClaudeVersion)},
		{"claudeview:", val(info.AppVersion)},
		{"CPU:", fmt.Sprintf("%.0f%%", info.CPUPercent)},
		{"MEM:", fmt.Sprintf("%d MiB", info.MemMiB)},
	}

	var lines []string
	for i, row := range leftRows {
		styledLabel := StyleKey.Render(row.label)
		labelVis := lipgloss.Width(styledLabel)
		labelPad := strings.Repeat(" ", max(labelW-labelVis, 1))

		leftPart := styledLabel + labelPad + row.value
		leftVis := lipgloss.Width(leftPart)
		leftPadding := strings.Repeat(" ", max(leftW-leftVis, 2))

		right := ""
		if i < len(items) {
			item := items[i]
			right = StyleKey.Render("<"+item.Key+">") + StyleKeyDesc.Render(" "+item.Desc)
		}

		lines = append(lines, leftPart+leftPadding+right)
	}

	return strings.Join(lines, "\n")
}

// truncateLeft returns at most maxW runes, prefixing "…" if truncated.
func truncateLeft(s string, maxW int) string {
	if s == "" {
		return StyleDim.Render("--")
	}
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	return "…" + string(runes[len(runes)-maxW+1:])
}
