package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestTableNavItems(t *testing.T) {
	// Sessions: enter + esc with specific descriptions
	nav := ui.TableNavItems(model.ResourceSessions)
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
	projNav := ui.TableNavItems(model.ResourceProjects)
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
	agentNav := ui.TableNavItems(model.ResourceAgents)
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
