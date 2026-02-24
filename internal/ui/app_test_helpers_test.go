package ui_test

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// mockDP satisfies ui.DataProvider with empty stubs.
type mockDP struct{}

func (m *mockDP) GetProjects() any       { return []*model.Project{} }
func (m *mockDP) GetSessions(string) any { return []*model.Session{} }
func (m *mockDP) GetAgents(string) any   { return []*model.Agent{} }
func (m *mockDP) GetTools(string) any    { return []*model.ToolCall{} }
func (m *mockDP) GetTasks(string) any    { return []*model.Task{} }
func (m *mockDP) GetPlugins() any        { return []*model.Plugin{} }
func (m *mockDP) GetMCPServers() any     { return []*model.MCPServer{} }
func (m *mockDP) CurrentProject() string { return "" }
func (m *mockDP) CurrentSession() string { return "" }
func (m *mockDP) CurrentAgent() string   { return "" }

// makeModel creates an AppModel pre-populated with rows, sized 120×40.
func makeModel(resource model.ResourceType, rows []ui.Row) ui.AppModel {
	app := ui.NewAppModel(&mockDP{}, resource)
	app.Width = 120
	app.Height = 40
	app.Table.SetRows(rows)
	return app
}

// key sends a rune key to the model and returns the updated AppModel.
func key(m ui.AppModel, k string) ui.AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	return next.(ui.AppModel)
}

// specialKey sends a special key (Enter, Esc, etc.) and returns updated AppModel.
func specialKey(m ui.AppModel, t tea.KeyType) ui.AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: t})
	return next.(ui.AppModel)
}

// projectRow builds a Row whose Data is a *model.Project.
func projectRow(hash string) ui.Row {
	return ui.Row{
		Cells: []string{hash, "1", "0", "1d"},
		Data:  &model.Project{Hash: hash},
	}
}

// sessionRow builds a Row whose Data is a *model.Session.
func sessionRow(id string) ui.Row {
	return ui.Row{
		Cells: []string{id[:8], "claude-3", "active", "1", "5", "1.2k", "$0.01", "5m"},
		Data:  &model.Session{ID: id},
	}
}

// agentRow builds a Row whose Data is a *model.Agent.
func agentRow(id string) ui.Row {
	return ui.Row{
		Cells: []string{id, "main", "done", "3", "read file"},
		Data:  &model.Agent{ID: id},
	}
}

// nRows returns n generic rows (no meaningful Data).
func nRows(n int) []ui.Row {
	rows := make([]ui.Row, n)
	for i := range rows {
		rows[i] = ui.Row{Cells: []string{fmt.Sprintf("row-%d", i)}}
	}
	return rows
}
