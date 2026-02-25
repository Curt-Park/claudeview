package bdd

// TestDetailYAML — detail and YAML view mode entry and content.

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestDetailViewEntered(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	sendKey(tm, "d")
	time.Sleep(50 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.ViewMode != ui.ModeDetail {
		t.Errorf("expected ModeDetail, got %v", app.ViewMode)
	}
}

func TestYAMLViewEntered(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	sendKey(tm, "y")
	time.Sleep(50 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.ViewMode != ui.ModeYAML {
		t.Errorf("expected ModeYAML, got %v", app.ViewMode)
	}
}
