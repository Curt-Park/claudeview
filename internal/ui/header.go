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
// navCount and utilCount are the number of items in each menu column.
// Minimum is 5 (1 project row + 4 data rows); expands if more items are needed.
func (info InfoModel) Height(navCount, utilCount int) int {
	return max(5, 1+max(navCount, utilCount))
}

// ViewWithMenu renders the info panel with a 4-column layout:
//
//	Col 0 (leftW chars):    info labels + values
//	Col 1 (menuColW chars): nav menu items (movement commands)
//	Col 2 (menuColW chars): util menu items (filter/follow/detail/back)
//	Col 3 (rightW chars):   t/p/m jump shortcuts
//
// Row 0 (Project) spans the full width. Remaining rows use the 4-column layout.
func (info InfoModel) ViewWithMenu(navItems, utilItems []MenuItem) string {
	const labelW = 14 // visible chars reserved for label column
	const leftW = 32  // total visible chars for the left info column

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

	// --- Data rows ---
	otherRows := []struct{ label, value string }{
		{"Session:", val(info.Session)},
		{"User:", val(info.User)},
		{"Claude Code:", val(info.ClaudeVersion)},
		{"claudeview:", val(info.AppVersion)},
	}

	// Within each column, pad keys to the column's max key width so descriptions align.
	maxNavKeyW := 0
	for _, item := range navItems {
		if w := lipgloss.Width(StyleKey.Render("<" + item.Key + ">")); w > maxNavKeyW {
			maxNavKeyW = w
		}
	}
	maxUtilKeyW := 0
	for _, item := range utilItems {
		if w := lipgloss.Width(StyleKey.Render("<" + item.Key + ">")); w > maxUtilKeyW {
			maxUtilKeyW = w
		}
	}

	// menuColW = max aligned item width across both columns, plus trailing gap.
	menuColW := 0
	for _, item := range navItems {
		w := maxNavKeyW + 1 + lipgloss.Width(StyleKeyDesc.Render(item.Desc))
		if w > menuColW {
			menuColW = w
		}
	}
	for _, item := range utilItems {
		w := maxUtilKeyW + 1 + lipgloss.Width(StyleKeyDesc.Render(item.Desc))
		if w > menuColW {
			menuColW = w
		}
	}
	menuColW += 2 // trailing gap between menu columns

	// Right column: t/p/m shortcuts.
	jumpHints := []string{
		StyleKey.Render("<t>") + StyleKeyDesc.Render(" tasks"),
		StyleKey.Render("<p>") + StyleKeyDesc.Render(" plugins"),
		StyleKey.Render("<m>") + StyleKeyDesc.Render(" mcps"),
	}
	rightColW := 0
	for _, h := range jumpHints {
		if w := lipgloss.Width(h); w > rightColW {
			rightColW = w
		}
	}

	// Total rows to render (at least len(otherRows)).
	totalRows := max(len(otherRows), max(len(navItems), len(utilItems)))

	lines := []string{projectLine}
	for i := range totalRows {
		var leftPart, leftPadding string
		if i < len(otherRows) {
			row := otherRows[i]
			styledLabel := StyleKey.Render(row.label)
			labelVis := lipgloss.Width(styledLabel)
			labelPad := strings.Repeat(" ", max(labelW-labelVis, 1))
			leftPart = styledLabel + labelPad + row.value
			leftVis := lipgloss.Width(leftPart)
			// Always place nav at column leftW+2 so all rows align.
			leftPadding = strings.Repeat(" ", max(leftW+2-leftVis, 1))
		} else {
			leftPart = ""
			leftPadding = strings.Repeat(" ", leftW+2)
		}

		// Nav column (col 1) — key padded to maxNavKeyW so descriptions align.
		nav := ""
		if i < len(navItems) {
			item := navItems[i]
			keyStr := StyleKey.Render("<" + item.Key + ">")
			keyPad := strings.Repeat(" ", max(maxNavKeyW-lipgloss.Width(keyStr), 0))
			nav = keyStr + keyPad + " " + StyleKeyDesc.Render(item.Desc)
		}
		navVis := lipgloss.Width(nav)
		navPad := strings.Repeat(" ", max(menuColW-navVis, 2))

		// Util column (col 2) — key padded to maxUtilKeyW so descriptions align.
		util := ""
		if i < len(utilItems) {
			item := utilItems[i]
			keyStr := StyleKey.Render("<" + item.Key + ">")
			keyPad := strings.Repeat(" ", max(maxUtilKeyW-lipgloss.Width(keyStr), 0))
			util = keyStr + keyPad + " " + StyleKeyDesc.Render(item.Desc)
		}
		utilVis := lipgloss.Width(util)
		utilPad := strings.Repeat(" ", max(menuColW-utilVis, 2))

		// Right column (col 3): jump hints
		right := ""
		if i < len(jumpHints) {
			right = jumpHints[i]
		}

		lines = append(lines, leftPart+leftPadding+nav+navPad+util+utilPad+right)
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
