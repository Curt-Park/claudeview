package ui

import (
	"github.com/Curt-Park/claudeview/internal/model"
)

// MenuItem is a single key+description pair for the menu bar.
type MenuItem struct {
	Key  string
	Desc string
}

// MenuModel holds the menu bar state.
type MenuModel struct {
	NavItems  []MenuItem // col 2: navigation commands (up/down/page/top/bottom/enter)
	UtilItems []MenuItem // col 3: utility commands (filter/follow/detail/log)
}

// ResourceHasLog reports whether a resource type has a log view.
func ResourceHasLog(rt model.ResourceType) bool {
	return rt == model.ResourceSessions || rt == model.ResourceAgents
}

// TableNavItems returns navigation menu items for the table view.
func TableNavItems(rt model.ResourceType) []MenuItem {
	items := []MenuItem{
		{Key: "j/k", Desc: "up/down"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "ctrl+u/d", Desc: "page"},
	}
	if rt == model.ResourceProjects || rt == model.ResourceSessions || rt == model.ResourceAgents {
		items = append(items, MenuItem{Key: "enter", Desc: "drill-down"})
	}
	return items
}

// TableUtilItems returns utility menu items for the table view.
func TableUtilItems(rt model.ResourceType) []MenuItem {
	items := []MenuItem{
		{Key: "/", Desc: "filter"},
	}
	if ResourceHasLog(rt) {
		items = append(items, MenuItem{Key: "l", Desc: "logs"})
	}
	items = append(items, MenuItem{Key: "d", Desc: "detail"})
	items = append(items, MenuItem{Key: "esc", Desc: "back"})
	return items
}

// LogNavItems returns navigation menu items for the log view.
func LogNavItems() []MenuItem {
	return []MenuItem{
		{Key: "j/k", Desc: "up/down"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "ctrl+u/d", Desc: "page"},
	}
}

// LogUtilItems returns utility menu items for the log view.
func LogUtilItems() []MenuItem {
	return []MenuItem{
		{Key: "f", Desc: "follow"},
		{Key: "/", Desc: "filter"},
		{Key: "esc", Desc: "back"},
	}
}

// DetailNavItems returns navigation menu items for the detail/YAML view.
func DetailNavItems() []MenuItem {
	return []MenuItem{
		{Key: "j/k", Desc: "up/down"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "ctrl+u/d", Desc: "page"},
	}
}

// DetailUtilItems returns utility menu items for the detail/YAML view.
func DetailUtilItems() []MenuItem {
	return []MenuItem{
		{Key: "esc", Desc: "back"},
	}
}
