package view

import (
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var taskColumnsBase = []ui.Column{
	{Title: "ID", Width: 4},
	{Title: "STATUS", Width: 12},
	{Title: "SUBJECT", Width: 40, Flex: true},
	{Title: "OWNER", Width: 10},
	{Title: "BLOCKED BY", Width: 14},
	{Title: "BLOCKS", Width: 12},
}

var taskColumnsFlat = []ui.Column{
	{Title: "SESSION", Width: 12},
	{Title: "ID", Width: 4},
	{Title: "STATUS", Width: 12},
	{Title: "SUBJECT", Width: 40, Flex: true},
	{Title: "OWNER", Width: 10},
	{Title: "BLOCKED BY", Width: 14},
	{Title: "BLOCKS", Width: 12},
}

// NewTasksView creates a tasks view.
func NewTasksView(width, height int) *ResourceView[*model.Task] {
	return NewResourceView(taskColumnsBase, taskColumnsFlat, taskRow, width, height)
}

func taskRow(items []*model.Task, i int, flatMode bool) ui.Row {
	t := items[i]
	statusStyle := ui.StatusStyle(string(t.Status))
	blockedBy := strings.Join(t.BlockedBy, ", ")
	blocks := strings.Join(t.Blocks, ", ")
	var cells []string
	if flatMode {
		sessionID := t.SessionID
		if len(sessionID) > 8 {
			sessionID = sessionID[:8]
		}
		cells = append(cells, sessionID)
	}
	cells = append(cells,
		t.ID,
		statusStyle.Render(t.StatusIcon()+" "+string(t.Status)),
		t.Subject,
		t.Owner,
		blockedBy,
		blocks,
	)
	return ui.Row{Cells: cells, Data: t}
}
