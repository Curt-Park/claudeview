package bdd

// TestCommand — `:command` resource switching and breadcrumb reset.

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestCommandSwitchToSessions(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	// Enter command mode and type "sessions"
	sendKey(tm, ":")
	for _, ch := range "sessions" {
		sendKey(tm, string(ch))
	}

	// Wait for autocomplete to appear
	waitForOutput(t, tm, containsStr(":sessions"))

	// Submit
	sendEnter(tm)
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions, got %s", app.Resource)
	}
}

func TestCommandSwitchResetsNavContext(t *testing.T) {
	dp := &mockDP{}
	// Start with sessions, simulating having drilled down
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))

	// Switch to agents via command
	sendKey(tm, ":")
	for _, ch := range "agents" {
		sendKey(tm, string(ch))
	}
	sendEnter(tm)
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	// After :command, navigation context should be reset
	if app.SelectedProjectHash != "" {
		t.Errorf("expected SelectedProjectHash reset, got %s", app.SelectedProjectHash)
	}
	if app.SelectedSessionID != "" {
		t.Errorf("expected SelectedSessionID reset, got %s", app.SelectedSessionID)
	}
}

func TestCommandEscCancels(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(1))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[1]"))

	sendKey(tm, ":")
	waitForOutput(t, tm, containsStr(":"))

	// Esc should cancel command mode
	sendEsc(tm) // ESC
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects after Esc cancel, got %s", app.Resource)
	}
}
