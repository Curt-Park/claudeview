package ui

import (
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
	Width         int
}

// Height returns the number of terminal lines rendered by ViewWithMenu.
// This must stay in sync with the number of lines returned by ViewWithMenu.
func (info InfoModel) Height() int {
	// 1 project row + 4 data rows (Session, User, Claude Code, claudeview)
	return 5
}

// ViewWithMenu renders the info panel alongside key binding hints.
// Row 0 (Project) spans the full width. Rows 1-4 use the standard 3-column layout.
func (info InfoModel) ViewWithMenu(items []MenuItem) string {
	const labelW = 14 // visible chars reserved for label column
	const leftW = 46  // total visible chars for the left column (label + value) on rows 1-4

	dim := StyleDim

	val := func(s string) string {
		if s == "" {
			return dim.Render("--")
		}
		return s
	}

	// --- Row 0: Project — full width ---
	projectLabel := StyleKey.Render("Project:")
	projectLabelVis := lipgloss.Width(projectLabel)
	projectLabelPad := strings.Repeat(" ", max(labelW-projectLabelVis, 1))
	availW := max(info.Width-labelW-1, 10)
	projectLine := projectLabel + projectLabelPad + truncateLeft(info.Project, availW)

	// --- Rows 1-4: standard 3-column layout ---
	otherRows := []struct{ label, value string }{
		{"Session:", val(info.Session)},
		{"User:", val(info.User)},
		{"Claude Code:", val(info.ClaudeVersion)},
		{"claudeview:", val(info.AppVersion)},
	}

	// Right column: t/p/m shortcuts (fixed to rows 1-3)
	jumpHints := []string{
		StyleKey.Render("<t>") + StyleKeyDesc.Render(" tasks"),
		StyleKey.Render("<p>") + StyleKeyDesc.Render(" plugins"),
		StyleKey.Render("<m>") + StyleKeyDesc.Render(" mcps"),
		StyleKey.Render("<?>") + StyleKeyDesc.Render(" help"),
	}
	// Compute the widest hint so all hints start at the same column.
	rightColW := 0
	for _, h := range jumpHints {
		if w := lipgloss.Width(h); w > rightColW {
			rightColW = w
		}
	}
	// Compute widest center item to avoid overlap.
	maxCenterW := 0
	for _, item := range items {
		w := lipgloss.Width(StyleKey.Render("<"+item.Key+">") + StyleKeyDesc.Render(" "+item.Desc))
		if w > maxCenterW {
			maxCenterW = w
		}
	}
	// Place right column at leftW*2 so column-start spacing mirrors the left→center gap.
	rightColStart := max(leftW*2, leftW+maxCenterW+2)
	if rightColStart+rightColW > info.Width {
		rightColStart = info.Width - rightColW - 1
	}

	lines := []string{projectLine}
	for i, row := range otherRows {
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

		// Right column: t/p/m hint (rows 0-2 of otherRows), left-aligned at rightColStart
		if i < len(jumpHints) {
			used := lipgloss.Width(leftPart) + len(leftPadding) + lipgloss.Width(center)
			pad := strings.Repeat(" ", max(rightColStart-used, 1))
			lines = append(lines, leftPart+leftPadding+center+pad+jumpHints[i])
			continue
		}

		lines = append(lines, leftPart+leftPadding+center)
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
