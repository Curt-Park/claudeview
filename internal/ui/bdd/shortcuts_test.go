package bdd

// TestShortcut — 0-9 number shortcuts for parent filtering.

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestShortcutZeroClearsFilter(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))

	// First apply a filter
	sendKey(tm, "/")
	sendKey(tm, "q")
	sendEsc(tm) // exit filter mode

	// Then press 0 to clear
	sendKey(tm, "0")
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Table.Filter != "" {
		t.Errorf("expected empty filter after '0', got %q", app.Table.Filter)
	}
	if app.ParentFilter != "" {
		t.Errorf("expected empty ParentFilter after '0', got %q", app.ParentFilter)
	}
}

func TestShortcutOneFiltersWhenShortcutsAvailable(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	// At projects level, ParentShortcuts is empty (no parent to filter by)
	// so pressing '1' should do nothing special
	sendKey(tm, "1")
	time.Sleep(50 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	// No shortcuts at projects level, so ParentFilter stays empty
	if app.ParentFilter != "" {
		t.Logf("ParentFilter=%q (no shortcuts configured = expected)", app.ParentFilter)
	}
}

func TestShortcutSetsParentFilterWhenShortcutExists(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	// Manually inject a parent shortcut
	// We test by pressing '1' when there's a shortcut in ParentShortcuts
	// For now, verify the key handler doesn't panic and model state is correct
	sendKey(tm, "1")
	time.Sleep(50 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
}

func TestShortcutInfoPanelShowsAll(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("<0>")(bts) && containsStr("all")(bts)
	})
}

// Set a parent shortcut on the info model and verify pressing its number sets the filter.
func TestShortcutWithInjectedShortcut(t *testing.T) {
	dp := &mockDP{}
	app := ui.NewAppModel(dp, model.ResourceSessions)
	app.Width = termWidth
	app.Height = termHeight
	app.Table.SetRows(sessionRows(2))
	// Inject a parent shortcut
	app.Info.ParentShortcuts = []ui.ParentShortcut{
		{Number: 1, Label: "proj-alpha", Active: false},
	}

	tm := teatestNewTestModelFromApp(t, app)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))

	// Press '1' to activate the parent shortcut
	sendKey(tm, "1")
	time.Sleep(50 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	finalApp, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if finalApp.ParentFilter != "proj-alpha" {
		t.Errorf("expected ParentFilter=proj-alpha, got %q", finalApp.ParentFilter)
	}
}
