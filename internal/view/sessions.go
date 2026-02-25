package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var sessionColumnsBase = []ui.Column{
	{Title: "NAME", Width: 10},
	{Title: "MODEL", Width: 16, Flex: true, MaxPercent: 0.30},
	{Title: "STATUS", Width: 12},
	{Title: "AGENTS", Width: 6},
	{Title: "TOOLS", Width: 6},
	{Title: "TOKENS", Width: 8},
	{Title: "COST", Width: 8},
	{Title: "AGE", Width: 6},
}

var sessionColumnsFlat = []ui.Column{
	{Title: "PROJECT", Width: 20},
	{Title: "NAME", Width: 10},
	{Title: "MODEL", Width: 16, Flex: true, MaxPercent: 0.30},
	{Title: "STATUS", Width: 12},
	{Title: "AGENTS", Width: 6},
	{Title: "TOOLS", Width: 6},
	{Title: "TOKENS", Width: 8},
	{Title: "COST", Width: 8},
	{Title: "AGE", Width: 6},
}

// NewSessionsView creates a sessions view.
func NewSessionsView(width, height int) *ResourceView[*model.Session] {
	return NewResourceView(sessionColumnsBase, sessionColumnsFlat, sessionRow, width, height)
}

func sessionRow(items []*model.Session, i int, flatMode bool) ui.Row {
	s := items[i]
	statusStyle := ui.StatusStyle(string(s.Status))
	var cells []string
	if flatMode {
		cells = append(cells, truncateHash(s.ProjectHash))
	}
	cells = append(cells,
		s.ShortID(),
		s.Model,
		statusStyle.Render(string(s.Status)),
		fmt.Sprintf("%d", len(s.Agents)),
		fmt.Sprintf("%d", s.ToolCount()),
		s.TokenString(),
		s.CostString(),
		s.Age(),
	)
	return ui.Row{Cells: cells, Data: s}
}
