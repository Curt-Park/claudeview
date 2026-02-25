package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestTableNavItems(t *testing.T) {
	// Sessions: enter + esc in nav items
	nav := ui.TableNavItems(model.ResourceSessions)
	if len(nav) == 0 {
		t.Fatal("TableNavItems returned empty slice")
	}
	hasEnter, hasEsc := false, false
	for _, item := range nav {
		if item.Key == "enter" {
			hasEnter = true
		}
		if item.Key == "esc" {
			hasEsc = true
		}
	}
	if !hasEnter {
		t.Error("TableNavItems(sessions): missing 'enter' key")
	}
	if !hasEsc {
		t.Error("TableNavItems(sessions): missing 'esc' key")
	}

	// Tools: no enter, but has esc
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

	// Projects: no esc (root level, nothing to go back to)
	projNav := ui.TableNavItems(model.ResourceProjects)
	for _, item := range projNav {
		if item.Key == "esc" {
			t.Error("TableNavItems(projects): unexpected 'esc' key (root level)")
		}
	}
}

func TestTableUtilItems(t *testing.T) {
	// Sessions: logs key present, no esc (moved to nav)
	util := ui.TableUtilItems(model.ResourceSessions)
	if len(util) == 0 {
		t.Fatal("TableUtilItems returned empty slice")
	}
	hasLogs := false
	for _, item := range util {
		if item.Key == "l" {
			hasLogs = true
		}
		if item.Key == "esc" {
			t.Error("TableUtilItems: unexpected 'esc' key (should be in nav)")
		}
	}
	if !hasLogs {
		t.Error("TableUtilItems(sessions): missing 'l' (logs) key")
	}

	// Tools: no logs
	toolUtil := ui.TableUtilItems(model.ResourceTools)
	for _, item := range toolUtil {
		if item.Key == "l" {
			t.Error("TableUtilItems(tools): unexpected 'l' key")
		}
	}
}

func TestLogNavItems(t *testing.T) {
	items := ui.LogNavItems()
	if len(items) == 0 {
		t.Fatal("LogNavItems returned empty slice")
	}
}

func TestLogUtilItems(t *testing.T) {
	items := ui.LogUtilItems()
	if len(items) == 0 {
		t.Fatal("LogUtilItems returned empty slice")
	}
	hasFollow, hasFilter := false, false
	for _, item := range items {
		if item.Key == "f" {
			hasFollow = true
		}
		if item.Key == "/" {
			hasFilter = true
		}
	}
	if !hasFollow {
		t.Error("LogUtilItems: missing 'f' (follow) key")
	}
	if !hasFilter {
		t.Error("LogUtilItems: missing '/' (filter) key")
	}
}

func TestDetailNavItems(t *testing.T) {
	items := ui.DetailNavItems()
	if len(items) == 0 {
		t.Fatal("DetailNavItems returned empty slice")
	}
	hasEsc := false
	for _, item := range items {
		if item.Key == "esc" {
			hasEsc = true
		}
	}
	if !hasEsc {
		t.Error("DetailNavItems: missing 'esc' key")
	}
}
