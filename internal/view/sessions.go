package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var sessionColumnsBase = []ui.Column{
	{Title: "NAME", Width: 10},
	{Title: "TOPIC", Width: 20, Flex: true, MaxPercent: 0.35},
	{Title: "AGENTS", Width: 6},
	{Title: "TOOLS", Width: 6},
	{Title: "TOKENS", Width: 20, Flex: true, MaxPercent: 0.25},
	{Title: "LAST ACTIVE", Width: 11},
}

var sessionColumnsFlat = []ui.Column{
	{Title: "PROJECT", Width: 20},
	{Title: "NAME", Width: 10},
	{Title: "TOPIC", Width: 20, Flex: true, MaxPercent: 0.35},
	{Title: "AGENTS", Width: 6},
	{Title: "TOOLS", Width: 6},
	{Title: "TOKENS", Width: 20, Flex: true, MaxPercent: 0.25},
	{Title: "LAST ACTIVE", Width: 11},
}

// NewSessionsView creates a sessions view.
func NewSessionsView(width, height int) *ResourceView[*model.Session] {
	return NewResourceView(sessionColumnsBase, sessionColumnsFlat, sessionRow, width, height)
}

func sessionRow(items []*model.Session, i int, flatMode bool) ui.Row {
	s := items[i]
	var cells []string
	if flatMode {
		cells = append(cells, truncateHash(s.ProjectHash))
	}
	cells = append(cells,
		s.ShortID(),
		s.TopicShort(120),
		fmt.Sprintf("%d", s.AgentCount),
		fmt.Sprintf("%d", s.ToolCallCount),
		s.TokenString(),
		s.LastActive(),
	)
	return ui.Row{Cells: cells, Subtitle: s.MetaLine(), Data: s}
}
