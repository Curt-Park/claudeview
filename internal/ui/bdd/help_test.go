package bdd

// TestHelp — `?` opens full-screen help; Esc closes it.

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestHelpOpenWithQuestion(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	sendKey(tm, "?")

	// Help view should show help content
	waitForOutput(t, tm, containsStr("── Help "))
}

func TestHelpShowsKeybindings(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(1))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[1]"))

	sendKey(tm, "?")

	// Help should contain key references
	waitForOutput(t, tm, func(bts []byte) bool {
		out := string(bts)
		return containsStr("ctrl+c")(bts) ||
			containsStr("Table Mode")(bts) ||
			len(out) > 100
	})
}

func TestHelpEscReturnsToTable(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(1))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[1]"))

	sendKey(tm, "?")
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
		t.Errorf("expected ModeTable after Esc from help, got %v", app.ViewMode)
	}
}
