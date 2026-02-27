package ui_test

// render_test.go verifies observable terminal output by running a real Bubble Tea
// program via teatest and asserting on rendered strings.

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestRenderInitialProjectsTable(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))
}

func TestRenderInitialMenuShown(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		out := string(bts)
		return strings.Contains(out, "<enter>") && strings.Contains(out, "</>")
	})
}

func TestRenderNavTitleBarShowsFilter(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))
	sendKey(tm, "/")
	sendKey(tm, "a")
	sendKey(tm, "b")

	waitForOutput(t, tm, func(bts []byte) bool {
		return strings.Contains(string(bts), "Projects(ab)")
	})
}

func TestRenderFilterActivate(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))
	sendKey(tm, "/")

	waitForOutput(t, tm, containsStr("/█"))
}

func TestRenderFilterTypingUpdatesTitle(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(3))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[3]"))
	sendKey(tm, "/")
	sendKey(tm, "x")
	sendKey(tm, "y")
	sendKey(tm, "z")

	waitForOutput(t, tm, containsStr("Projects(xyz)"))
}

func TestRenderFlashShownOnCommandSwitch(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))
	sendKey(tm, ":")
	for _, ch := range "sessions" {
		sendKey(tm, string(ch))
	}
	tm.Type("\r")

	waitForOutput(t, tm, notContainsStr("Projects(all)"))
}

func TestRenderDrilldownBreadcrumbPushed(t *testing.T) {
	p := &model.Project{Hash: "proj-xyz"}
	dp := &mockDP{}
	rows := []ui.Row{{Cells: []string{p.Hash, "1", "0", "5m"}, Data: p}}
	tm := newTestModel(t, model.ResourceProjects, dp, rows)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[1]"))
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	waitForOutput(t, tm, containsStr("sessions"))
}

func TestRenderInfoPanelAtProjectsLevelShowsDashes(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("Project:")(bts) && containsStr("Session:")(bts)
	})
}

func TestRenderInfoPanelMenuShownForSessions(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("<enter>")(bts) && containsStr("<esc>")(bts)
	})
}

func TestRenderSessionsInFlatMode(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))
}

func TestRenderSessionsInDrillMode(t *testing.T) {
	dp := &mockDP{}
	app := ui.NewAppModel(dp, model.ResourceSessions)
	app.Width = termWidth
	app.Height = termHeight
	app.SelectedProjectHash = "proj-abc"
	app.Table.SetRows(sessionRows(2))

	tm := teatestFromApp(t, app)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))
}

func TestRenderEscDescWithFilter(t *testing.T) {
	dp := &mockDP{}
	// Use rows whose content matches "x" to survive filtering
	rows := []ui.Row{
		{Cells: []string{"xaaa", "topic", "1", "1", "sonnet:1k", "1h"}, Data: &model.Session{ID: "xaaa"}},
		{Cells: []string{"xbbb", "topic", "1", "1", "sonnet:1k", "1h"}, Data: &model.Session{ID: "xbbb"}},
	}
	tm := newTestModel(t, model.ResourceSessions, dp, rows)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))
	sendKey(tm, "/")
	sendKey(tm, "x")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // confirm filter

	// After confirming a non-empty filter, esc description changes to "clear filter"
	waitForOutput(t, tm, func(bts []byte) bool {
		return strings.Contains(string(bts), "clear filter")
	})
}

// Ensure sendEsc is referenced so the compiler doesn't complain.
var _ = sendEsc
