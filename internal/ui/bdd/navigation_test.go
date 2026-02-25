package bdd

// TestNav — verify j/k cursor movement and g/G top/bottom in table mode.

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestNavMoveDown(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(5))
	t.Cleanup(func() { _ = tm.Quit() })

	// Wait for initial render
	waitForOutput(t, tm, containsStr("Projects(all)[5]"))

	// Move down — selection should advance
	sendKey(tm, "j")
	time.Sleep(100 * time.Millisecond)

	// Check the model updated (via FinalModel after quit)
	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Table.Selected != 1 {
		t.Errorf("expected Selected=1 after j, got %d", app.Table.Selected)
	}
}

func TestNavMoveUp(t *testing.T) {
	dp := &mockDP{}
	rows := projectRows(5)
	tm := newTestModel(t, model.ResourceProjects, dp, rows)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[5]"))

	// Move down twice then up once → Selected should be 1
	sendKey(tm, "j")
	sendKey(tm, "j")
	sendKey(tm, "k")

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Table.Selected != 1 {
		t.Errorf("expected Selected=1 after j/j/k, got %d", app.Table.Selected)
	}
}

func TestNavGotoTop(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(5))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[5]"))

	// Move to bottom then go to top
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	sendKey(tm, "g")

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	if app.Table.Selected != 0 {
		t.Errorf("expected Selected=0 after G/g, got %d", app.Table.Selected)
	}
}

func TestNavPageDown(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(10))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[10]"))

	// ctrl+d = page down
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlD})

	if err := tm.Quit(); err != nil {
		t.Fatal(err)
	}
	fm := tm.FinalModel(t)
	app, ok := fm.(ui.AppModel)
	if !ok {
		t.Fatal("expected AppModel")
	}
	// After page down, selected should be > 0
	if app.Table.Selected == 0 {
		t.Error("expected Selected>0 after ctrl+d")
	}
}

func TestNavTitleBarShowsFilter(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))

	// Enter filter mode and type
	sendKey(tm, "/")
	sendKey(tm, "a")
	sendKey(tm, "b")

	// Title bar should now show the filter
	waitForOutput(t, tm, func(bts []byte) bool {
		return strings.Contains(string(bts), "Projects(ab)")
	})
}
