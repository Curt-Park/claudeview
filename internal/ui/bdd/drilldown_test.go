package bdd

// TestDrilldown — Enter drill-down, Esc back, breadcrumb changes.

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestDrilldownProjectsToSessions(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	dp := &mockDP{}
	rows := []ui.Row{
		{Cells: []string{p.Hash, "3", "1", "2h"}, Data: p},
	}
	tm := newTestModel(t, model.ResourceProjects, dp, rows)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[1]"))

	// Press Enter to drill down into sessions
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
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
		t.Errorf("expected resource=sessions after Enter, got %s", app.Resource)
	}
	if app.SelectedProjectHash != p.Hash {
		t.Errorf("expected SelectedProjectHash=%s, got %s", p.Hash, app.SelectedProjectHash)
	}
}

func TestDrilldownBreadcrumbPushed(t *testing.T) {
	p := &model.Project{Hash: "proj-xyz"}
	dp := &mockDP{}
	rows := []ui.Row{
		{Cells: []string{p.Hash, "1", "0", "5m"}, Data: p},
	}
	tm := newTestModel(t, model.ResourceProjects, dp, rows)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[1]"))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Breadcrumbs should now show "sessions"
	waitForOutput(t, tm, containsStr("sessions"))
}

func TestDrilldownEscNavigatesBack(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))

	// Press Esc to go back to projects
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
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
		t.Errorf("expected resource=projects after Esc from sessions, got %s", app.Resource)
	}
}
