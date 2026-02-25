// Package bdd contains BDD-style integration tests for the claudeview TUI
// using the teatest library to run real Bubble Tea programs and verify output.
package bdd

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
	projects   []*model.Project
	sessions   []*model.Session
	agents     []*model.Agent
	toolCalls  []*model.ToolCall
	tasks      []*model.Task
	plugins    []*model.Plugin
	mcpServers []*model.MCPServer
}

func (m *mockDP) GetProjects() []*model.Project         { return m.projects }
func (m *mockDP) GetSessions(_ string) []*model.Session { return m.sessions }
func (m *mockDP) GetAgents(_ string) []*model.Agent     { return m.agents }
func (m *mockDP) GetTools(_ string) []*model.ToolCall   { return m.toolCalls }
func (m *mockDP) GetTasks(_ string) []*model.Task       { return m.tasks }
func (m *mockDP) GetPlugins() []*model.Plugin           { return m.plugins }
func (m *mockDP) GetMCPServers() []*model.MCPServer     { return m.mcpServers }

// newTestModel creates an AppModel pre-populated with test data for the given resource.
// It wraps the AppModel in a thin coordinating model to populate Table rows from dp.
func newTestModel(t *testing.T, resource model.ResourceType, dp *mockDP, rows []ui.Row) *teatest.TestModel {
	t.Helper()
	app := ui.NewAppModel(dp, resource)
	app.Width = termWidth
	app.Height = termHeight
	if rows != nil {
		app.Table.SetRows(rows)
	}
	return teatest.NewTestModel(
		t, app,
		teatest.WithInitialTermSize(termWidth, termHeight),
	)
}

// projectRows creates test project rows.
func projectRows(n int) []ui.Row {
	rows := make([]ui.Row, n)
	for i := range rows {
		p := &model.Project{Hash: strings.Repeat("a", 8) + strings.Repeat("b", i+1)}
		rows[i] = ui.Row{
			Cells: []string{p.Hash, "2", "1", "1h"},
			Data:  p,
		}
	}
	return rows
}

// sessionRows creates test session rows.
func sessionRows(n int) []ui.Row {
	rows := make([]ui.Row, n)
	for i := range rows {
		s := &model.Session{ID: strings.Repeat("c", 8) + strings.Repeat("d", i+1), Status: model.StatusDone}
		rows[i] = ui.Row{
			Cells: []string{s.ShortID(), "claude-opus-4", "done", "2", "10", "5k", "$0.01", "1h"},
			Data:  s,
		}
	}
	return rows
}

// sendKey sends a single key rune to the test model.
// For special keys like Esc/Enter, use sendEsc/sendEnter.
func sendKey(tm *teatest.TestModel, key string) {
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

// sendEsc sends the Escape key.
func sendEsc(tm *teatest.TestModel) {
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
}

// waitForOutput polls the test model output until condition is true or timeout.
func waitForOutput(t *testing.T, tm *teatest.TestModel, condition func([]byte) bool) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), condition,
		teatest.WithDuration(waitDur),
		teatest.WithCheckInterval(checkEvery),
	)
}

// containsStr returns true if the output contains the given string.
func containsStr(s string) func([]byte) bool {
	return func(bts []byte) bool {
		return strings.Contains(string(bts), s)
	}
}

// notContainsStr returns true if the output does NOT contain the given string.
func notContainsStr(s string) func([]byte) bool {
	return func(bts []byte) bool {
		return !strings.Contains(string(bts), s)
	}
}

// teatestNewTestModelFromApp creates a TestModel from a pre-configured AppModel.
func teatestNewTestModelFromApp(t *testing.T, app ui.AppModel) *teatest.TestModel {
	t.Helper()
	return teatest.NewTestModel(
		t, app,
		teatest.WithInitialTermSize(termWidth, termHeight),
	)
}
