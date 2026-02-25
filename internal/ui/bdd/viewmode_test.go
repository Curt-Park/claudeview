package bdd

// TestViewMode — d/l/y key switches view mode; Esc returns to table.

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestViewModeSwitchToDetail(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	sendKey(tm, "d")
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.ViewMode != ui.ModeDetail {
		t.Errorf("expected ModeDetail after d, got %v", app.ViewMode)
	}
}

func TestViewModeSwitchToYAML(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	sendKey(tm, "y")
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.ViewMode != ui.ModeYAML {
		t.Errorf("expected ModeYAML after y, got %v", app.ViewMode)
	}
}

func TestViewModeEscReturnsToTable(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	// Go to detail then return
	sendKey(tm, "d")
	time.Sleep(50 * time.Millisecond)
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
	if app.ViewMode != ui.ModeTable {
		t.Errorf("expected ModeTable after Esc from detail, got %v", app.ViewMode)
	}
}
