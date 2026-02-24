package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// SessionsView renders the sessions list.
type SessionsView struct {
	Sessions []*model.Session
	Table    ui.TableView
}

var sessionColumns = []ui.Column{
	{Title: "NAME", Width: 10},
	{Title: "MODEL", Width: 16},
	{Title: "STATUS", Width: 12},
	{Title: "AGENTS", Width: 6},
	{Title: "TOOLS", Width: 6},
	{Title: "TOKENS", Width: 8},
	{Title: "COST", Width: 8},
	{Title: "AGE", Width: 6},
}

// NewSessionsView creates a sessions view.
func NewSessionsView(width, height int) *SessionsView {
	return &SessionsView{
		Table: ui.NewTableView(sessionColumns, width, height),
	}
}

// SetSessions updates the sessions list.
func (v *SessionsView) SetSessions(sessions []*model.Session) {
	v.Sessions = sessions
	rows := make([]ui.Row, len(sessions))
	for i, s := range sessions {
		statusStyle := ui.StatusStyle(string(s.Status))
		rows[i] = ui.Row{
			Cells: []string{
				s.ShortID(),
				s.Model,
				statusStyle.Render(string(s.Status)),
				fmt.Sprintf("%d", len(s.Agents)),
				fmt.Sprintf("%d", s.ToolCount()),
				s.TokenString(),
				s.CostString(),
				s.Age(),
			},
			Data: s,
		}
	}
	v.Table.SetRows(rows)
}

// SelectedSession returns the currently selected session.
func (v *SessionsView) SelectedSession() *model.Session {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if s, ok := row.Data.(*model.Session); ok {
		return s
	}
	return nil
}

// View renders the sessions table.
func (v *SessionsView) View() string {
	return v.Table.View()
}
