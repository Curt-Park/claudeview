package view

import (
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TasksView renders the tasks list.
type TasksView struct {
	Tasks    []*model.Task
	Table    ui.TableView
	FlatMode bool // when true, SESSION column is shown (flat :command access)
}

var taskColumnsBase = []ui.Column{
	{Title: "ID", Width: 4},
	{Title: "STATUS", Width: 12},
	{Title: "SUBJECT", Width: 40, Flex: true},
	{Title: "BLOCKED BY", Width: 14},
}

var taskColumnsFlat = []ui.Column{
	{Title: "SESSION", Width: 12},
	{Title: "ID", Width: 4},
	{Title: "STATUS", Width: 12},
	{Title: "SUBJECT", Width: 40, Flex: true},
	{Title: "BLOCKED BY", Width: 14},
}

// NewTasksView creates a tasks view.
func NewTasksView(width, height int) *TasksView {
	return &TasksView{
		Table: ui.NewTableView(taskColumnsBase, width, height),
	}
}

// SetTasks updates the tasks list.
func (v *TasksView) SetTasks(tasks []*model.Task) {
	v.Tasks = tasks
	if v.FlatMode {
		v.Table.Columns = taskColumnsFlat
	} else {
		v.Table.Columns = taskColumnsBase
	}
	rows := make([]ui.Row, len(tasks))
	for i, t := range tasks {
		statusStyle := ui.StatusStyle(string(t.Status))
		blockedBy := strings.Join(t.BlockedBy, ", ")
		var cells []string
		if v.FlatMode {
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
			blockedBy,
		)
		rows[i] = ui.Row{Cells: cells, Data: t}
	}
	v.Table.SetRows(rows)
}

// SelectedTask returns the currently selected task.
func (v *TasksView) SelectedTask() *model.Task {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if t, ok := row.Data.(*model.Task); ok {
		return t
	}
	return nil
}

// View renders the tasks table.
func (v *TasksView) View() string {
	return v.Table.View()
}
