package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestTableNavItems(t *testing.T) {
	// Sessions: enter in nav items
	nav := ui.TableNavItems(model.ResourceSessions)
	if len(nav) == 0 {
		t.Fatal("TableNavItems returned empty slice")
	}
	hasEnter := false
	for _, item := range nav {
		if item.Key == "enter" {
			hasEnter = true
		}
	}
	if !hasEnter {
		t.Error("TableNavItems(sessions): missing 'enter' key")
	}

	// Tools: no enter
	toolNav := ui.TableNavItems(model.ResourceTools)
	for _, item := range toolNav {
		if item.Key == "enter" {
			t.Error("TableNavItems(tools): unexpected 'enter' key")
		}
	}
}

func TestTableUtilItems(t *testing.T) {
	// Sessions: logs key present
	util := ui.TableUtilItems(model.ResourceSessions)
	if len(util) == 0 {
		t.Fatal("TableUtilItems returned empty slice")
	}
	hasLogs := false
	for _, item := range util {
		if item.Key == "l" {
			hasLogs = true
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
}

func TestDetailUtilItems(t *testing.T) {
	items := ui.DetailUtilItems()
	if len(items) == 0 {
		t.Fatal("DetailUtilItems returned empty slice")
	}
}
