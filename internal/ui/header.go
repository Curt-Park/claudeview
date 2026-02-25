package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ParentShortcut is a numbered parent context entry shown in the info panel right column.
type ParentShortcut struct {
	Number int    // 0-9 (0 = all)
	Label  string // display name, e.g. "my-project"
	Active bool   // true if this parent is currently selected
}

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

	// ParentShortcuts is the list of numbered parent shortcuts (1-9).
	// Index 0 = shortcut 1, index 1 = shortcut 2, etc.
	ParentShortcuts []ParentShortcut
}

// ViewWithMenu renders the 7-row info panel alongside key binding hints (3 columns).
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

	// Build shortcut entries: row 0 is always "0: all", rows 1+ are ParentShortcuts
	shortcuts := make([]string, len(leftRows))
	shortcuts[0] = StyleKey.Render("<0>") + StyleKeyDesc.Render(" all")
	for i, sc := range info.ParentShortcuts {
		row := i + 1
		if row >= len(shortcuts) {
			break
		}
		label := sc.Label
		if sc.Active {
			label = StyleActive.Render(label)
		} else {
			label = StyleKeyDesc.Render(label)
		}
		shortcuts[row] = StyleKey.Render(fmt.Sprintf("<%d>", sc.Number)) + " " + label
	}

	var lines []string
	for i, row := range leftRows {
		styledLabel := StyleKey.Render(row.label)
		labelVis := lipgloss.Width(styledLabel)
		labelPad := strings.Repeat(" ", max(labelW-labelVis, 1))

		leftPart := styledLabel + labelPad + row.value
		leftVis := lipgloss.Width(leftPart)
		leftPadding := strings.Repeat(" ", max(leftW-leftVis, 2))

		// Center column: keybinding hint
		center := ""
		if i < len(items) {
			item := items[i]
			center = StyleKey.Render("<"+item.Key+">") + StyleKeyDesc.Render(" "+item.Desc)
		}

		// Right column: parent shortcut
		right := ""
		if i < len(shortcuts) && shortcuts[i] != "" {
			centerVis := lipgloss.Width(center)
			const centerW = 22
			centerPad := strings.Repeat(" ", max(centerW-centerVis, 2))
			right = centerPad + shortcuts[i]
		}

		lines = append(lines, leftPart+leftPadding+center+right)
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
