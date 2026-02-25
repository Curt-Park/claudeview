package bdd

// TestFilter — `/` filter mode: live filtering and clear.

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestFilterActivate(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))

	sendKey(tm, "/")

	// Filter input should appear
	waitForOutput(t, tm, containsStr("/"))
}

func TestFilterTypingUpdatesTitle(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))

	sendKey(tm, "/")
	sendKey(tm, "x")
	sendKey(tm, "y")
	sendKey(tm, "z")

	// Title bar should show filter string
	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("Projects(xyz)")(bts)
	})
}

func TestFilterEscClears(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))

	sendKey(tm, "/")
	sendKey(tm, "q")
	sendKey(tm, "q")
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
	// After Esc, filter should be cleared
	if app.Filter.Input != "" {
		t.Errorf("expected filter cleared, got %q", app.Filter.Input)
	}
	if app.Table.Filter != "" {
		t.Errorf("expected table filter cleared, got %q", app.Table.Filter)
	}
}
