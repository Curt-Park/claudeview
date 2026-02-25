package ui

import (
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
)

// MenuItem is a single key+description pair for the menu bar.
type MenuItem struct {
	Key  string
	Desc string
}

// MenuModel holds the menu bar state.
type MenuModel struct {
	Items []MenuItem
	Width int
}

// TableMenuItems returns menu items for the table view based on the current resource.
func TableMenuItems(rt model.ResourceType) []MenuItem {
	hasLog := rt == model.ResourceSessions || rt == model.ResourceAgents
	hasDrillDown := rt == model.ResourceProjects || rt == model.ResourceSessions || rt == model.ResourceAgents

	var items []MenuItem
	if hasDrillDown {
		items = append(items, MenuItem{Key: "enter", Desc: "view"})
	}
	if hasLog {
		items = append(items, MenuItem{Key: "l", Desc: "logs"})
	}
	items = append(items, MenuItem{Key: "d", Desc: "detail"})
	items = append(items, MenuItem{Key: "/", Desc: "filter"})
	return items
}

// LogMenuItems returns menu items for the log view.
func LogMenuItems() []MenuItem {
	return []MenuItem{
		{Key: "h/j/k/l", Desc: "scroll"},
		{Key: "f", Desc: "follow"},
		{Key: "/", Desc: "search"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
	}
}

// DetailMenuItems returns menu items for the detail view.
func DetailMenuItems() []MenuItem {
	return []MenuItem{
		{Key: "h/j/k/l", Desc: "scroll"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
	}
}

// View renders the menu bar.
func (m MenuModel) View() string {
	var parts []string
	for _, item := range m.Items {
		key := StyleKey.Render("<" + item.Key + ">")
		desc := StyleKeyDesc.Render(" " + item.Desc)
		parts = append(parts, key+desc)
	}
	line := strings.Join(parts, "  ")
	return StyleMenu.Width(m.Width).Render(line)
}
