package view

import (
	"fmt"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var projectColumns = []ui.Column{
	{Title: "NAME", Width: 20, Flex: true, MaxPercent: 0.55},
	{Title: "SESSIONS", Width: 8},
	{Title: "ACTIVE", Width: 6},
	{Title: "LAST ACTIVE", Width: 11},
}

// NewProjectsView creates a projects view.
func NewProjectsView(width, height int) *ResourceView[*model.Project] {
	return NewResourceView(projectColumns, nil, projectRow, width, height)
}

func projectRow(items []*model.Project, i int, _ bool) ui.Row {
	p := items[i]
	active := len(p.ActiveSessions())
	return ui.Row{
		Cells: []string{
			truncateHash(p.Hash),
			fmt.Sprintf("%d", p.SessionCount()),
			fmt.Sprintf("%d", active),
			model.FormatAge(time.Since(p.LastSeen)),
		},
		Data: p,
	}
}
