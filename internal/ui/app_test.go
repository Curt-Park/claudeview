package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestNavMoveDown(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(5))

	app = updateApp(app, keyMsg("j"))

	if app.Table.Selected != 1 {
		t.Errorf("expected Selected=1 after j, got %d", app.Table.Selected)
	}
}

func TestNavMoveUp(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(5))

	app = updateApp(app, keyMsg("j"))
	app = updateApp(app, keyMsg("j"))
	app = updateApp(app, keyMsg("k"))

	if app.Table.Selected != 1 {
		t.Errorf("expected Selected=1 after j/j/k, got %d", app.Table.Selected)
	}
}

func TestNavGotoTop(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(5))

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	app = updateApp(app, keyMsg("g"))

	if app.Table.Selected != 0 {
		t.Errorf("expected Selected=0 after G/g, got %d", app.Table.Selected)
	}
}

func TestNavPageDown(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(10))

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyCtrlD})

	if app.Table.Selected == 0 {
		t.Error("expected Selected>0 after ctrl+d")
	}
}

func TestDrilldownProjectsToSessions(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "3", "1", "2h"}, Data: p}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions after Enter, got %s", app.Resource)
	}
	if app.SelectedProjectHash != p.Hash {
		t.Errorf("expected SelectedProjectHash=%s, got %s", p.Hash, app.SelectedProjectHash)
	}
}

func TestDrilldownEscNavigatesBack(t *testing.T) {
	app := newApp(model.ResourceSessions)
	app.Table.SetRows(sessionRows(2))

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects after Esc from sessions, got %s", app.Resource)
	}
}

func TestFilterEscClears(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(3))

	app = updateApp(app, keyMsg("/"))
	app = updateApp(app, keyMsg("q"))
	app = updateApp(app, keyMsg("q"))
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Filter.Input != "" {
		t.Errorf("expected filter cleared, got %q", app.Filter.Input)
	}
	if app.Table.Filter != "" {
		t.Errorf("expected table filter cleared, got %q", app.Table.Filter)
	}
}

func TestResizeHandled(t *testing.T) {
	app := newApp(model.ResourceProjects)

	app = updateApp(app, tea.WindowSizeMsg{Width: 80, Height: 24})

	if app.Width != 80 {
		t.Errorf("expected Width=80, got %d", app.Width)
	}
	if app.Height != 24 {
		t.Errorf("expected Height=24, got %d", app.Height)
	}
}

func TestDrilldownClearsFilter(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.Table.Filter = "proj" // matches the row so drilldown can proceed

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Table.Filter != "" {
		t.Errorf("expected Table.Filter cleared after drilldown, got %q", app.Table.Filter)
	}
	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions, got %s", app.Resource)
	}
}

func TestNavigateBackRestoresFilter(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.Table.Filter = "proj" // matches row so drilldown proceeds

	// Drill down to sessions — filter should be saved and cleared
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	if app.Table.Filter != "" {
		t.Errorf("expected filter cleared after drilldown, got %q", app.Table.Filter)
	}

	// Navigate back — filter should be restored
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects after Esc, got %s", app.Resource)
	}
	if app.Table.Filter != "proj" {
		t.Errorf("expected restored filter=%q, got %q", "proj", app.Table.Filter)
	}
}

func TestEscClearsRestoredFilter(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.Table.Filter = "proj" // matches row so drilldown proceeds

	// Drill down then back — filter is restored
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Table.Filter != "proj" {
		t.Fatalf("expected restored filter=%q, got %q", "proj", app.Table.Filter)
	}

	// Esc again — should clear the restored filter (not navigate back further)
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Table.Filter != "" {
		t.Errorf("expected filter cleared by Esc, got %q", app.Table.Filter)
	}
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected still on projects, got %s", app.Resource)
	}
}

func TestJumpPreservesFilterStack(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	s := &model.Session{ID: "sess-xyz789"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.SelectedProjectHash = p.Hash
	app.Table.Filter = "proj"

	// Drill into sessions — filterStack=["proj"], filter=""
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	app.Table.SetRows([]ui.Row{{Cells: []string{s.ShortID(), "topic", "1", "1", "sonnet:1k", "1h"}, Data: s}})

	// Jump to plugins — saves {Filter:"", FilterStack:["proj"]}
	app = updateApp(app, keyMsg("p"))
	if app.Resource != model.ResourcePlugins {
		t.Fatalf("expected resource=plugins after p, got %s", app.Resource)
	}

	// Esc back to sessions — restores filter="" and filterStack=["proj"]
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceSessions {
		t.Fatalf("expected resource=sessions after Esc from plugins, got %s", app.Resource)
	}
	if app.Table.Filter != "" {
		t.Errorf("expected filter=%q after jump-back, got %q", "", app.Table.Filter)
	}

	// Esc from sessions — popFilter restores "proj", go to projects
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects, got %s", app.Resource)
	}
	if app.Table.Filter != "proj" {
		t.Errorf("expected filter=%q after esc from sessions, got %q", "proj", app.Table.Filter)
	}
}
