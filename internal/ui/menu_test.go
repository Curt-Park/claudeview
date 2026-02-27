package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestTableNavItems(t *testing.T) {
	// Sessions: enter + esc with specific descriptions
	nav := ui.TableNavItems(model.ResourceSessions, false)
	if len(nav) == 0 {
		t.Fatal("TableNavItems returned empty slice")
	}
	enterDesc, escDesc := "", ""
	for _, item := range nav {
		if item.Key == "enter" {
			enterDesc = item.Desc
		}
		if item.Key == "esc" {
			escDesc = item.Desc
		}
	}
	if enterDesc == "" {
		t.Error("TableNavItems(sessions): missing 'enter' key")
	}
	if escDesc == "" {
		t.Error("TableNavItems(sessions): missing 'esc' key")
	}

	// Projects: enter present, no esc (root level)
	projNav := ui.TableNavItems(model.ResourceProjects, false)
	hasEnter := false
	for _, item := range projNav {
		if item.Key == "esc" {
			t.Error("TableNavItems(projects): unexpected 'esc' key (root level)")
		}
		if item.Key == "enter" {
			hasEnter = true
		}
	}
	if !hasEnter {
		t.Error("TableNavItems(projects): missing 'enter' key")
	}

	// Agents: esc present, no enter (leaf node — no tools drill-down)
	agentNav := ui.TableNavItems(model.ResourceAgents, false)
	hasEscAgents := false
	for _, item := range agentNav {
		if item.Key == "enter" {
			t.Error("TableNavItems(agents): unexpected 'enter' key")
		}
		if item.Key == "esc" {
			hasEscAgents = true
		}
	}
	if !hasEscAgents {
		t.Error("TableNavItems(agents): missing 'esc' key")
	}
}

func TestTableUtilItemsHasFilter(t *testing.T) {
	for _, rt := range []model.ResourceType{
		model.ResourceSessions, model.ResourceAgents, model.ResourceProjects,
	} {
		util := ui.TableUtilItems(rt)
		hasFilter := false
		for _, item := range util {
			if item.Key == "/" {
				hasFilter = true
			}
		}
		if !hasFilter {
			t.Errorf("TableUtilItems(%s): missing '/' (filter) key", rt)
		}
	}
}

func TestTableNavItemsWithFilter(t *testing.T) {
	// When hasFilter=true, esc should show "clear filter" for all resource types
	for _, rt := range []model.ResourceType{
		model.ResourceProjects, model.ResourceSessions, model.ResourceAgents,
	} {
		nav := ui.TableNavItems(rt, true)
		escDesc := ""
		for _, item := range nav {
			if item.Key == "esc" {
				escDesc = item.Desc
			}
		}
		if escDesc != "clear filter" {
			t.Errorf("TableNavItems(%s, hasFilter=true): expected esc desc %q, got %q", rt, "clear filter", escDesc)
		}
	}
}

func TestSetHighlightMatchesCompoundKey(t *testing.T) {
	menu := ui.MenuModel{
		NavItems: []ui.MenuItem{
			{Key: "j/k", Desc: "down/up"},
			{Key: "enter", Desc: "drill"},
		},
	}

	menu.SetHighlight("j")

	jk := ui.MenuItem{Key: "j/k", Desc: "down/up"}
	if !menu.IsHighlighted(jk) {
		t.Error("expected j/k to be highlighted after SetHighlight('j')")
	}

	enter := ui.MenuItem{Key: "enter", Desc: "drill"}
	if menu.IsHighlighted(enter) {
		t.Error("expected enter NOT to be highlighted")
	}
}

func TestSetHighlightNoMatchIsNoOp(t *testing.T) {
	menu := ui.MenuModel{
		NavItems: []ui.MenuItem{{Key: "j/k", Desc: "down/up"}},
	}

	menu.SetHighlight("z") // no match

	item := ui.MenuItem{Key: "j/k", Desc: "down/up"}
	if menu.IsHighlighted(item) {
		t.Error("expected no highlight when key does not match any item")
	}
}

func TestClearHighlight(t *testing.T) {
	menu := ui.MenuModel{
		NavItems: []ui.MenuItem{{Key: "j/k", Desc: "down/up"}},
	}
	menu.SetHighlight("j")

	item := ui.MenuItem{Key: "j/k", Desc: "down/up"}
	if !menu.IsHighlighted(item) {
		t.Fatal("expected highlight active before clear")
	}

	menu.ClearHighlight()
	if menu.IsHighlighted(item) {
		t.Error("expected highlight cleared after ClearHighlight")
	}
}
