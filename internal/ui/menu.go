package ui

import (
	"strings"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

// HighlightDuration is how long a key highlight lasts before it is cleared by timer.
const HighlightDuration = 150 * time.Millisecond

// MenuItem is a single key+description pair for the menu bar.
type MenuItem struct {
	Key  string
	Desc string
}

// MenuModel holds the menu bar state.
type MenuModel struct {
	NavItems     []MenuItem // col 2: navigation commands
	UtilItems    []MenuItem // col 3: utility commands (filter)
	HighlightKey string     // currently highlighted key (e.g. "j"), cleared by HighlightClearMsg
}

// SetHighlight records key as the currently highlighted key.
// Any rendered item whose Key matches (exactly or as a compound part) will light up.
func (m *MenuModel) SetHighlight(key string) {
	m.HighlightKey = key
}

// ClearHighlight removes any active highlight.
func (m *MenuModel) ClearHighlight() {
	m.HighlightKey = ""
}

// IsHighlighted returns true if item is currently highlighted.
// Checks exact match first (handles single-char keys like "/"), then compound parts.
func (m *MenuModel) IsHighlighted(item MenuItem) bool {
	if m.HighlightKey == "" {
		return false
	}
	if item.Key == m.HighlightKey {
		return true
	}
	for _, part := range splitKey(item.Key) {
		if part == m.HighlightKey {
			return true
		}
	}
	return false
}

// splitKey splits a compound key into its individual parts, inheriting any
// modifier prefix from the first segment.
//
//	"j/k"      → ["j", "k"]
//	"G/g"      → ["G", "g"]
//	"ctrl+d/u" → ["ctrl+d", "ctrl+u"]
func splitKey(key string) []string {
	slashIdx := strings.Index(key, "/")
	if slashIdx < 0 {
		return []string{key}
	}
	first := key[:slashIdx]
	// Derive modifier prefix, e.g. "ctrl+" from "ctrl+d".
	prefix := ""
	if plusIdx := strings.LastIndex(first, "+"); plusIdx >= 0 {
		prefix = first[:plusIdx+1]
	}
	parts := []string{first}
	for _, p := range strings.Split(key[slashIdx+1:], "/") {
		if p == "" {
			continue
		}
		if prefix != "" && !strings.Contains(p, "+") {
			parts = append(parts, prefix+p)
		} else {
			parts = append(parts, p)
		}
	}
	return parts
}

// TableNavItems returns navigation menu items for the table view with
// context-specific descriptions for enter and esc.
// When hasFilter is true, esc is shown as "clear filter" regardless of resource.
func TableNavItems(rt model.ResourceType, hasFilter bool) []MenuItem {
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
	case model.ResourceAgents:
		// no enter hint
	case model.ResourcePlugins:
		items = append(items, MenuItem{Key: "enter", Desc: "detail"})
	case model.ResourceMemory:
		items = append(items, MenuItem{Key: "enter", Desc: "detail"})
	}
	if hasFilter {
		items = append(items, MenuItem{Key: "esc", Desc: "clear filter"})
	} else {
		switch rt {
		case model.ResourceSessions:
			items = append(items, MenuItem{Key: "esc", Desc: "see projects"})
		case model.ResourceAgents:
			items = append(items, MenuItem{Key: "esc", Desc: "see sessions"})
		case model.ResourcePlugins, model.ResourceMemory:
			items = append(items, MenuItem{Key: "esc", Desc: "back"})
		case model.ResourcePluginDetail, model.ResourceMemoryDetail:
			items = append(items, MenuItem{Key: "esc", Desc: "back"})
			// ResourceProjects: no esc (root level)
		}
	}
	return items
}

// TableUtilItems returns utility menu items for the table view.
func TableUtilItems(_ model.ResourceType) []MenuItem {
	return []MenuItem{
		{Key: "/", Desc: "filter"},
	}
}
