package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// ProjectsView renders the projects list.
type ProjectsView struct {
	Projects []*model.Project
	Table    ui.TableView
}

var projectColumns = []ui.Column{
	{Title: "NAME", Width: 50, Flex: true},
	{Title: "SESSIONS", Width: 8},
	{Title: "ACTIVE", Width: 6},
	{Title: "LAST SEEN", Width: 10},
}

// NewProjectsView creates a projects view.
func NewProjectsView(width, height int) *ProjectsView {
	return &ProjectsView{
		Table: ui.NewTableView(projectColumns, width, height),
	}
}

// SetProjects updates the projects list.
func (v *ProjectsView) SetProjects(projects []*model.Project) {
	v.Projects = projects
	rows := make([]ui.Row, len(projects))
	for i, p := range projects {
		active := len(p.ActiveSessions())
		rows[i] = ui.Row{
			Cells: []string{
				truncateHash(p.Hash),
				fmt.Sprintf("%d", p.SessionCount()),
				fmt.Sprintf("%d", active),
				formatAge(p.LastSeen),
			},
			Data: p,
		}
	}
	v.Table.SetRows(rows)
}

// SelectedProject returns the currently selected project.
func (v *ProjectsView) SelectedProject() *model.Project {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if p, ok := row.Data.(*model.Project); ok {
		return p
	}
	return nil
}

// View renders the projects table.
func (v *ProjectsView) View() string {
	return v.Table.View()
}
