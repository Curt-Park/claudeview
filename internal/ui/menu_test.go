package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestTableMenuItems(t *testing.T) {
	items := ui.TableMenuItems()
	if len(items) == 0 {
		t.Fatal("TableMenuItems returned empty slice")
	}
	hasEnter, hasLogs := false, false
	for _, item := range items {
		if item.Key == "enter" {
			hasEnter = true
		}
		if item.Key == "l" {
			hasLogs = true
		}
	}
	if !hasEnter {
		t.Error("TableMenuItems: missing 'enter' key")
	}
	if !hasLogs {
		t.Error("TableMenuItems: missing 'l' (logs) key")
	}
}

func TestLogMenuItems(t *testing.T) {
	items := ui.LogMenuItems()
	if len(items) == 0 {
		t.Fatal("LogMenuItems returned empty slice")
	}
	hasFollow := false
	for _, item := range items {
		if item.Key == "f" {
			hasFollow = true
		}
	}
	if !hasFollow {
		t.Error("LogMenuItems: missing 'f' (follow) key")
	}
}

func TestDetailMenuItems(t *testing.T) {
	items := ui.DetailMenuItems()
	if len(items) == 0 {
		t.Fatal("DetailMenuItems returned empty slice")
	}
}
