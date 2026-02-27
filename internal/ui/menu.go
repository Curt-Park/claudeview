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
	NavItems  []MenuItem // col 2: navigation commands
	UtilItems []MenuItem // col 3: utility commands (filter)
}

// TableNavItems returns navigation menu items for the table view with
// context-specific descriptions for enter and esc.
func TableNavItems(rt model.ResourceType) []MenuItem {
	items := []MenuItem{
		{Key: "j/k", Desc: "down/up"},
		{Key: "G/g", Desc: "bottom/top"},
		{Key: "ctrl+d/u", Desc: "page down/up"},
	}
	switch rt {
	case model.ResourceProjects:
		items = append(items, MenuItem{Key: "enter", Desc: "see sessions"})
	case model.ResourceSessions:
		items = append(items, MenuItem{Key: "enter", Desc: "see agents"})
		items = append(items, MenuItem{Key: "esc", Desc: "see projects"})
	case model.ResourceAgents:
		items = append(items, MenuItem{Key: "esc", Desc: "see sessions"})
	default:
		items = append(items, MenuItem{Key: "esc", Desc: "back"})
	}
	return items
}

// TableUtilItems returns utility menu items for the table view.
func TableUtilItems(_ model.ResourceType) []MenuItem {
	return []MenuItem{
		{Key: "/", Desc: "filter"},
	}
}
