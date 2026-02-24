package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// AgentsView renders the agents tree.
type AgentsView struct {
	Agents []*model.Agent
	Table  ui.TableView
}

var agentColumns = []ui.Column{
	{Title: "NAME", Width: 26, Flex: true},
	{Title: "TYPE", Width: 12},
	{Title: "STATUS", Width: 14},
	{Title: "TOOLS", Width: 6},
	{Title: "LAST ACTIVITY", Width: 30},
}

// NewAgentsView creates an agents view.
func NewAgentsView(width, height int) *AgentsView {
	return &AgentsView{
		Table: ui.NewTableView(agentColumns, width, height),
	}
}

// SetAgents updates the agents list.
func (v *AgentsView) SetAgents(agents []*model.Agent) {
	v.Agents = agents
	rows := make([]ui.Row, len(agents))
	for i, a := range agents {
		isLast := isLastSubagent(agents, i)
		prefix := a.TreePrefix(isLast)
		name := prefix + a.DisplayName()

		statusStyle := ui.StatusStyle(string(a.Status))
		rows[i] = ui.Row{
			Cells: []string{
				name,
				string(a.Type),
				statusStyle.Render(string(a.Status)),
				fmt.Sprintf("%d", len(a.ToolCalls)),
				a.LastActivity,
			},
			Data: a,
		}
	}
	v.Table.SetRows(rows)
}

// SelectedAgent returns the currently selected agent.
func (v *AgentsView) SelectedAgent() *model.Agent {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if a, ok := row.Data.(*model.Agent); ok {
		return a
	}
	return nil
}

// View renders the agents table.
func (v *AgentsView) View() string {
	return v.Table.View()
}

// isLastSubagent returns true if this is the last subagent in sequence.
func isLastSubagent(agents []*model.Agent, idx int) bool {
	if !agents[idx].IsSubagent {
		return false
	}
	for i := idx + 1; i < len(agents); i++ {
		if agents[i].IsSubagent {
			return false
		}
	}
	return true
}
