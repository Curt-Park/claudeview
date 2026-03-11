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
// navCount, actionCount, and utilCount are the number of items in each menu column.
// Minimum is 5 (1 project row + 4 data rows); expands if more items are needed.
func (info InfoModel) Height(navCount, actionCount, utilCount int) int {
	return max(5, 1+max(navCount, max(actionCount, utilCount)))
}

// ViewWithMenu renders the info panel with a 6-column layout:
//
//	Col 0 (leftW chars):    info labels + values
//	Col 1 (navColW chars):  nav menu items (movement commands)
//	Col 2 (actionColW):     action menu items (enter/space/esc)
//	Col 3 (utilColW chars): util menu items (filter)
//	Col 4 (rightW chars):   p/m jump shortcuts
//	Col 5:                  ctrl+c quit
//
// Row 0 (Project) spans the full width. Remaining rows use the 6-column layout.
func (info InfoModel) ViewWithMenu(menu MenuModel) string {
	navItems := menu.NavItems
	actionItems := menu.ActionItems
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
		{"Session Slug:", val(info.Session)},
		{"User:", val(info.User)},
		{"Claude Code:", val(info.ClaudeVersion)},
		{"claudeview:", val(info.AppVersion)},
	}

	// Within each column, pad keys to the column's max key width so descriptions align.
	maxNavKeyW := menuMaxKeyW(navItems)
	maxActionKeyW := menuMaxKeyW(actionItems)
	maxUtilKeyW := menuMaxKeyW(utilItems)

	// Each column gets its own natural width + trailing gap.
	navColW := menuColW(navItems, maxNavKeyW) + 2
	actionColW := menuColW(actionItems, maxActionKeyW) + 2
	utilColW := menuColW(utilItems, maxUtilKeyW) + 2

	// Col 4: p/m jump shortcuts.
	// Plugins and memories views cannot navigate to each other, so both hints
	// are hidden when either view is active.
	var jumpHints []string
	inPluginsOrMemories := info.Resource == model.ResourcePlugins || info.Resource == model.ResourceMemory ||
		isSubView(info.Resource)
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

	// Col 5: quit hint.
	quitHint := renderJumpHint(menu, "ctrl+c", "quit")

	// Total rows to render (at least len(otherRows)).
	totalRows := max(len(otherRows), max(len(navItems), max(len(actionItems), len(utilItems))))

	lines := []string{projectLine}
	for i := range totalRows {
		var leftPart, leftPadding string
		if i < len(otherRows) {
			row := otherRows[i]
			styledLabel := StyleKey.Render(row.label)
			labelVis := lipgloss.Width(styledLabel)
			labelPad := strings.Repeat(" ", max(labelW-labelVis, 1))
			// Truncate value so the left column never exceeds leftW.
			maxValW := leftW - labelVis - len(labelPad)
			valStr := row.value
			if valVisW := lipgloss.Width(valStr); valVisW > maxValW && maxValW > 1 {
				runes := []rune(valStr)
				if len(runes) > maxValW {
					valStr = string(runes[:maxValW-1]) + "…"
				}
			}
			leftPart = styledLabel + labelPad + valStr
			leftVis := lipgloss.Width(leftPart)
			// Always place nav at column leftW+2 so all rows align.
			leftPadding = strings.Repeat(" ", max(leftW+2-leftVis, 1))
		} else {
			leftPart = ""
			leftPadding = strings.Repeat(" ", leftW+2)
		}

		// Nav column (col 1) — movement keys.
		nav := renderMenuItem(menu, navItems, i, maxNavKeyW)
		navVis := lipgloss.Width(nav)
		navPad := strings.Repeat(" ", max(navColW-navVis, 2))

		// Action column (col 2) — enter/space/esc.
		action := renderMenuItem(menu, actionItems, i, maxActionKeyW)
		actionVis := lipgloss.Width(action)
		actionPad := strings.Repeat(" ", max(actionColW-actionVis, 2))

		// Util column (col 3) — filter.
		util := renderMenuItem(menu, utilItems, i, maxUtilKeyW)
		utilVis := lipgloss.Width(util)
		utilPad := strings.Repeat(" ", max(utilColW-utilVis, 2))

		// Col 4: jump hints (p/m), padded to rightColW.
		right := ""
		if i < len(jumpHints) {
			right = jumpHints[i]
		}
		rightVis := lipgloss.Width(right)
		rightPad := strings.Repeat(" ", max(rightColW-rightVis, 2))

		// Col 5: quit hint on the first row only.
		quit := ""
		if i == 0 {
			quit = quitHint
		}

		lines = append(lines, leftPart+leftPadding+nav+navPad+action+actionPad+util+utilPad+right+rightPad+quit)
	}

	return strings.Join(lines, "\n")
}

// menuMaxKeyW returns the max rendered width of "<key>" across all items.
func menuMaxKeyW(items []MenuItem) int {
	max := 0
	for _, item := range items {
		if w := lipgloss.Width(StyleKey.Render("<" + item.Key + ">")); w > max {
			max = w
		}
	}
	return max
}

// menuColW returns the natural content width for a column (max of key+desc widths).
func menuColW(items []MenuItem, maxKeyW int) int {
	w := 0
	for _, item := range items {
		if cw := maxKeyW + 1 + lipgloss.Width(StyleKeyDesc.Render(item.Desc)); cw > w {
			w = cw
		}
	}
	return w
}

// renderMenuItem renders item i from items, padding the key to maxKeyW.
// Returns empty string if i is out of range.
func renderMenuItem(menu MenuModel, items []MenuItem, i, maxKeyW int) string {
	if i >= len(items) {
		return ""
	}
	item := items[i]
	keyStyle := StyleKey
	descStyle := StyleKeyDesc
	if menu.IsHighlighted(item) {
		keyStyle = StyleKeyHighlight
		descStyle = StyleKeyHighlight
	}
	keyStr := keyStyle.Render("<" + item.Key + ">")
	keyPad := strings.Repeat(" ", max(maxKeyW-lipgloss.Width(StyleKey.Render("<"+item.Key+">")), 0))
	return keyStr + keyPad + " " + descStyle.Render(item.Desc)
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
