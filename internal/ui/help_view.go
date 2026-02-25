package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// HelpView renders a full-screen scrollable help overlay.
type HelpView struct {
	Offset int
	Width  int
	Height int
}

// NewHelpView creates a HelpView with the given dimensions.
func NewHelpView(width, height int) HelpView {
	return HelpView{Width: width, Height: height}
}

// Update handles key events for scrolling the help view.
func (h *HelpView) Update(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			h.Offset++
		case "k", "up":
			if h.Offset > 0 {
				h.Offset--
			}
		case "g":
			h.Offset = 0
		case "G":
			lines := helpLines()
			if len(lines) > h.Height {
				h.Offset = len(lines) - h.Height
			}
		}
	}
}

// View renders the help content.
func (h HelpView) View() string {
	lines := helpLines()
	var sb strings.Builder

	// Title
	sb.WriteString(StyleTitle.Render("── Help "))
	sb.WriteString("\n")

	visible := h.Height - 2 // title + blank
	if visible <= 0 {
		visible = 20
	}
	offset := h.Offset
	if offset >= len(lines) {
		offset = max(0, len(lines)-1)
	}

	for i := offset; i < len(lines) && i < offset+visible; i++ {
		sb.WriteString(lines[i])
		sb.WriteString("\n")
	}

	// Fill remaining
	rendered := min(len(lines)-offset, visible)
	for i := rendered; i < visible; i++ {
		sb.WriteString("\n")
	}

	// Scroll indicator
	total := len(lines)
	status := StyleDim.Render("j/k: scroll  esc/q: close")
	if total > visible {
		shown := min(offset+visible, total)
		status = StyleDim.Render("j/k: scroll  g/G: top/bottom  esc/q: close") +
			StyleDim.Render("  "+styleCount(shown, total))
	}
	sb.WriteString(status)

	return sb.String()
}

func styleCount(shown, total int) string {
	return strings.Repeat(" ", 0) + "[" + itoa(shown) + "/" + itoa(total) + "]"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// helpLines returns the complete help text as a slice of pre-formatted strings.
func helpLines() []string {
	key := func(k string) string { return StyleKey.Render(k) }
	desc := func(d string) string { return StyleKeyDesc.Render(d) }
	section := func(s string) string { return StyleTitle.Render("── " + s + " ") }

	lines := []string{
		"",
		section("Global"),
		key("ctrl+c") + "    " + desc("quit immediately"),
		key(":") + "         " + desc("enter command mode (switch resource)"),
		key("/") + "         " + desc("enter filter mode"),
		key("?") + "         " + desc("show this help screen"),
		"",
		section("Table Mode"),
		key("j") + " / " + key("↓") + "    " + desc("move selection down"),
		key("k") + " / " + key("↑") + "    " + desc("move selection up"),
		key("g") + "         " + desc("go to top"),
		key("G") + "         " + desc("go to bottom"),
		key("ctrl+d") + " / " + key("pgdn") + "  " + desc("page down (half page)"),
		key("ctrl+u") + " / " + key("pgup") + "  " + desc("page up (half page)"),
		key("enter") + "     " + desc("drill down (projects→sessions→agents→tools)"),
		key("l") + "         " + desc("log view"),
		key("d") + "         " + desc("detail view"),
		key("y") + "         " + desc("YAML/JSON dump view"),
		key("0") + "         " + desc("clear parent filter (show all)"),
		key("1-9") + "       " + desc("filter by Nth parent shortcut"),
		key("esc") + " / " + key("q") + "  " + desc("navigate back"),
		"",
		section("Log Mode"),
		key("j/k") + "       " + desc("scroll down/up"),
		key("h/l") + "       " + desc("scroll left/right"),
		key("g/G") + "       " + desc("top/bottom"),
		key("f") + "         " + desc("toggle follow mode"),
		key("/") + "         " + desc("search within log"),
		key("esc") + "       " + desc("return to table"),
		"",
		section("Detail / YAML Mode"),
		key("j/k") + "       " + desc("scroll down/up"),
		key("h/l") + "       " + desc("scroll left/right"),
		key("g/G") + "       " + desc("top/bottom"),
		key("esc") + "       " + desc("return to table"),
		"",
		section("Command Mode (:)"),
		key("typing") + "    " + desc("input resource name"),
		key("tab") + "       " + desc("accept autocomplete suggestion"),
		key("enter") + "     " + desc("execute (switch resource)"),
		key("esc") + "       " + desc("cancel"),
		"",
		desc("Resources: projects, sessions, agents, tools, tasks, plugins, mcp"),
		"",
		section("Filter Mode (/)"),
		key("typing") + "    " + desc("live filter table rows (substring, case-insensitive)"),
		key("enter") + "     " + desc("confirm filter"),
		key("esc") + "       " + desc("clear filter and exit"),
		"",
		section("Info Panel Context Rules"),
		desc("At projects level:  all fields show --"),
		desc("After drill-down to sessions: Project field filled"),
		desc("After drill-down to agents:   Session field filled"),
		desc("CPU/MEM: claudeview process stats"),
		"",
		section("Navigation Hierarchy"),
		desc("projects → sessions → agents → tools"),
		desc("Use :command for flat (unfiltered) access to any resource"),
		desc("Parent columns (PROJECT/SESSION/AGENT) shown in flat mode"),
		"",
	}
	return lines
}
