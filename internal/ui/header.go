package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Curt-Park/claudeview/internal/model"
)

// InfoModel holds the top info panel data (k9s-style left column + right menu).
type InfoModel struct {
	Project        string // active project display name / path
	Session        string // active session short ID
	User           string // OS username
	ClaudeVersion  string // Claude Code binary version
	AppVersion     string // claudeview binary version
	Width          int
	MemoriesActive bool               // whether <m> memories jump is available
	Resource       model.ResourceType // current active resource (hides its own jump hint)
}

// Height returns the number of terminal lines rendered by ViewWithMenu.
// navCount and utilCount are the number of items in each menu column.
// Minimum is 5 (1 project row + 4 data rows); expands if more items are needed.
func (info InfoModel) Height(navCount, utilCount int) int {
	return max(5, 1+max(navCount, utilCount))
}

// ViewWithMenu renders the info panel with a 5-column layout:
//
//	Col 0 (leftW chars):    info labels + values
//	Col 1 (menuColW chars): nav menu items (movement commands)
//	Col 2 (menuColW chars): util menu items (filter/follow/detail/back)
//	Col 3 (rightW chars):   p/m jump shortcuts
//	Col 4:                  ctrl+c quit
//
// Row 0 (Project) spans the full width. Remaining rows use the 5-column layout.
func (info InfoModel) ViewWithMenu(menu MenuModel) string {
	navItems := menu.NavItems
	utilItems := menu.UtilItems
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

	// navColW / utilColW: each column gets its own natural width + trailing gap.
	navColW := 0
	for _, item := range navItems {
		w := maxNavKeyW + 1 + lipgloss.Width(StyleKeyDesc.Render(item.Desc))
		if w > navColW {
			navColW = w
		}
	}
	navColW += 2

	utilColW := 0
	for _, item := range utilItems {
		w := maxUtilKeyW + 1 + lipgloss.Width(StyleKeyDesc.Render(item.Desc))
		if w > utilColW {
			utilColW = w
		}
	}
	utilColW += 2

	// Col 3: p/m jump shortcuts.
	// Plugins and memories views cannot navigate to each other, so both hints
	// are hidden when either view is active.
	var jumpHints []string
	inPluginsOrMemories := info.Resource == model.ResourcePlugins || info.Resource == model.ResourceMemory
	if !inPluginsOrMemories {
		jumpHints = append(jumpHints, renderJumpHint(menu, "p", "plugins"))
	}
	if info.MemoriesActive && !inPluginsOrMemories {
		jumpHints = append(jumpHints, renderJumpHint(menu, "m", "memories"))
	}
	rightColW := 0
	for _, h := range jumpHints {
		if w := lipgloss.Width(h); w > rightColW {
			rightColW = w
		}
	}
	rightColW += 2 // trailing gap before quit column

	// Col 4: quit hint.
	quitHint := renderJumpHint(menu, "ctrl+c", "quit")

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
			keyStyle := StyleKey
			descStyle := StyleKeyDesc
			if menu.IsHighlighted(item) {
				keyStyle = StyleKeyHighlight
				descStyle = StyleKeyHighlight
			}
			keyStr := keyStyle.Render("<" + item.Key + ">")
			keyPad := strings.Repeat(" ", max(maxNavKeyW-lipgloss.Width(StyleKey.Render("<"+item.Key+">")), 0))
			nav = keyStr + keyPad + " " + descStyle.Render(item.Desc)
		}
		navVis := lipgloss.Width(nav)
		navPad := strings.Repeat(" ", max(navColW-navVis, 2))

		// Util column (col 2) — key padded to maxUtilKeyW so descriptions align.
		util := ""
		if i < len(utilItems) {
			item := utilItems[i]
			keyStyle := StyleKey
			descStyle := StyleKeyDesc
			if menu.IsHighlighted(item) {
				keyStyle = StyleKeyHighlight
				descStyle = StyleKeyHighlight
			}
			keyStr := keyStyle.Render("<" + item.Key + ">")
			keyPad := strings.Repeat(" ", max(maxUtilKeyW-lipgloss.Width(StyleKey.Render("<"+item.Key+">")), 0))
			util = keyStr + keyPad + " " + descStyle.Render(item.Desc)
		}
		utilVis := lipgloss.Width(util)
		utilPad := strings.Repeat(" ", max(utilColW-utilVis, 2))

		// Col 3: jump hints (p/m), padded to rightColW.
		right := ""
		if i < len(jumpHints) {
			right = jumpHints[i]
		}
		rightVis := lipgloss.Width(right)
		rightPad := strings.Repeat(" ", max(rightColW-rightVis, 2))

		// Col 4: quit hint on the first row only.
		quit := ""
		if i == 0 {
			quit = quitHint
		}

		lines = append(lines, leftPart+leftPadding+nav+navPad+util+utilPad+right+rightPad+quit)
	}

	return strings.Join(lines, "\n")
}

// renderJumpHint renders a single jump shortcut (e.g. "<p> plugins") with
// optional highlight styling when the key is currently active.
func renderJumpHint(menu MenuModel, key, desc string) string {
	item := MenuItem{Key: key}
	if menu.IsHighlighted(item) {
		return StyleKeyHighlight.Render("<"+key+">") + StyleKeyHighlight.Render(" "+desc)
	}
	return StyleKey.Render("<"+key+">") + StyleKeyDesc.Render(" "+desc)
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
