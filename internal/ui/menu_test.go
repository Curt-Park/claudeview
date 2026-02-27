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

	// Tools: esc present, no enter
	toolNav := ui.TableNavItems(model.ResourceTools)
	hasEscTools := false
	for _, item := range toolNav {
		if item.Key == "enter" {
			t.Error("TableNavItems(tools): unexpected 'enter' key")
		}
		if item.Key == "esc" {
			hasEscTools = true
		}
	}
	if !hasEscTools {
		t.Error("TableNavItems(tools): missing 'esc' key")
	}
}

func TestTableUtilItemsHasFilter(t *testing.T) {
	for _, rt := range []model.ResourceType{
		model.ResourceSessions, model.ResourceTools, model.ResourceProjects,
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
