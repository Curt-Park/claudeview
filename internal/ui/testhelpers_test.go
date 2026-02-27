package ui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

const (
	termWidth  = 120
	termHeight = 40
	waitDur    = 2 * time.Second
	checkEvery = 50 * time.Millisecond
)

// mockDP is a minimal DataProvider for testing.
type mockDP struct {
	projects []*model.Project
	sessions []*model.Session
	agents   []*model.Agent
	plugins  []*model.Plugin
	memories []*model.Memory
}

func (m *mockDP) GetProjects() []*model.Project         { return m.projects }
func (m *mockDP) GetSessions(_ string) []*model.Session { return m.sessions }
func (m *mockDP) GetAgents(_ string) []*model.Agent     { return m.agents }
func (m *mockDP) GetPlugins(_ string) []*model.Plugin   { return m.plugins }
func (m *mockDP) GetMemories(_ string) []*model.Memory  { return m.memories }

// newApp creates an AppModel pre-sized for tests.
func newApp(resource model.ResourceType) ui.AppModel {
	dp := &mockDP{}
	app := ui.NewAppModel(dp, resource)
	app.Width = termWidth
	app.Height = termHeight
	return app
}

// updateApp sends a message to the app and returns the updated model.
func updateApp(app ui.AppModel, msg tea.Msg) ui.AppModel {
	m, _ := app.Update(msg)
	return m.(ui.AppModel)
}

// keyMsg creates a key rune message.
func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

// projectRows creates test project rows.
func projectRows(n int) []ui.Row {
	rows := make([]ui.Row, n)
	for i := range rows {
		p := &model.Project{Hash: strings.Repeat("a", 8) + strings.Repeat("b", i+1)}
		rows[i] = ui.Row{Cells: []string{p.Hash, "2", "1h"}, Data: p}
	}
	return rows
}

// sessionRows creates test session rows.
func sessionRows(n int) []ui.Row {
	rows := make([]ui.Row, n)
	for i := range rows {
		s := &model.Session{ID: strings.Repeat("c", 8) + strings.Repeat("d", i+1)}
		rows[i] = ui.Row{
			Cells: []string{s.ShortID(), "Refactor auth module", "2", "10", "sonnet:5k", "1h"},
			Data:  s,
		}
	}
	return rows
}

// newTestModel creates a teatest TestModel pre-populated with rows.
func newTestModel(t *testing.T, resource model.ResourceType, dp *mockDP, rows []ui.Row) *teatest.TestModel {
	t.Helper()
	app := ui.NewAppModel(dp, resource)
	app.Width = termWidth
	app.Height = termHeight
	if rows != nil {
		app.Table.SetRows(rows)
	}
	return teatest.NewTestModel(t, app, teatest.WithInitialTermSize(termWidth, termHeight))
}

// teatestFromApp creates a teatest TestModel from a pre-configured AppModel.
func teatestFromApp(t *testing.T, app ui.AppModel) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(t, app, teatest.WithInitialTermSize(termWidth, termHeight))
}

// sendKey sends a single key rune to the test model.
func sendKey(tm *teatest.TestModel, key string) {
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

// sendEsc sends the Escape key.
func sendEsc(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
}

// waitForOutput polls output until condition is true or timeout.
func waitForOutput(t *testing.T, tm *teatest.TestModel, condition func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), condition,
		teatest.WithDuration(waitDur),
		teatest.WithCheckInterval(checkEvery),
	)
}

// containsStr returns true if output contains s.
func containsStr(s string) func([]byte) bool {
	return func(bts []byte) bool { return strings.Contains(string(bts), s) }
}

// notContainsStr returns true if output does NOT contain s.
func notContainsStr(s string) func([]byte) bool {
	return func(bts []byte) bool { return !strings.Contains(string(bts), s) }
}
