package bdd

// TestResize — WindowSizeMsg is handled correctly.

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestResizeHandled(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	// Send a resize event
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	time.Sleep(100 * time.Millisecond)

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Width != 80 {
		t.Errorf("expected Width=80, got %d", app.Width)
	}
	if app.Height != 24 {
		t.Errorf("expected Height=24, got %d", app.Height)
	}
}
