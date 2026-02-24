package view

import (
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TasksView renders the tasks list.
type TasksView struct {
	Tasks []*model.Task
	Table ui.TableView
}

var taskColumns = []ui.Column{
	{Title: "ID", Width: 4},
	{Title: "STATUS", Width: 12},
	{Title: "SUBJECT", Width: 40, Flex: true},
	{Title: "BLOCKED BY", Width: 14},
}

// NewTasksView creates a tasks view.
func NewTasksView(width, height int) *TasksView {
	return &TasksView{
		Table: ui.NewTableView(taskColumns, width, height),
	}
}

// SetTasks updates the tasks list.
func (v *TasksView) SetTasks(tasks []*model.Task) {
	v.Tasks = tasks
	rows := make([]ui.Row, len(tasks))
	for i, t := range tasks {
		statusStyle := ui.StatusStyle(string(t.Status))
		blockedBy := strings.Join(t.BlockedBy, ", ")
		rows[i] = ui.Row{
			Cells: []string{
				t.ID,
				statusStyle.Render(t.StatusIcon() + " " + string(t.Status)),
				t.Subject,
				blockedBy,
			},
			Data: t,
		}
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
