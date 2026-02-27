package view

import (
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var agentColumnsBase = []ui.Column{
	{Title: "NAME", Width: 10, Flex: true, MaxPercent: 0.20},
	{Title: "TYPE", Width: 16},
	{Title: "STATUS", Width: 10},
	{Title: "LAST ACTIVITY", Width: 20, Flex: true, MaxPercent: 0.35},
}

var agentColumnsFlat = []ui.Column{
	{Title: "SESSION", Width: 12},
	{Title: "NAME", Width: 10, Flex: true, MaxPercent: 0.20},
	{Title: "TYPE", Width: 16},
	{Title: "STATUS", Width: 10},
	{Title: "LAST ACTIVITY", Width: 20, Flex: true, MaxPercent: 0.35},
}

// NewAgentsView creates an agents view.
func NewAgentsView(width, height int) *ResourceView[*model.Agent] {
	return NewResourceView(agentColumnsBase, agentColumnsFlat, agentRow, width, height)
}

func agentRow(items []*model.Agent, i int, flatMode bool) ui.Row {
	a := items[i]
	isLast := isLastSubagent(items, i)
	prefix := a.TreePrefix(isLast)
	name := prefix + a.DisplayName()

	statusStyle := ui.StatusStyle(a.Status)
	var cells []string
	if flatMode {
		cells = append(cells, ShortID(a.SessionID, 8))
	}
	cells = append(cells,
		name,
		string(a.Type),
		statusStyle.Render(string(a.Status)),
		a.LastActivity,
	)
	return ui.Row{Cells: cells, Data: a}
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
