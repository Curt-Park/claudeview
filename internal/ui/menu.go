package ui

import (
	"strings"
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

// TableMenuItems returns menu items for the table view.
func TableMenuItems() []MenuItem {
	return []MenuItem{
		{Key: "enter", Desc: "view"},
		{Key: "l", Desc: "logs"},
		{Key: "d", Desc: "detail"},
		{Key: "/", Desc: "filter"},
		{Key: ":", Desc: "cmd"},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}
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
