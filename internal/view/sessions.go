package view

import (
	"fmt"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var sessionColumnsBase = []ui.Column{
	{Title: "SLUG", Width: 16},
	{Title: "SESSION_IDs", Width: 19},
	{Title: "TOPIC", Width: 20, Flex: true, MaxPercent: 0.35},
	{Title: "TURNS", Width: 6},
	{Title: "AGENTS", Width: 6},
	{Title: "MODEL:TOKEN", Width: 20, Flex: true, MaxPercent: 0.25},
	{Title: "LAST ACTIVE", Width: 11},
}

var sessionColumnsFlat = []ui.Column{
	{Title: "PROJECT", Width: 20},
	{Title: "SLUG", Width: 16},
	{Title: "SESSION_IDs", Width: 19},
	{Title: "TOPIC", Width: 20, Flex: true, MaxPercent: 0.35},
	{Title: "TURNS", Width: 6},
	{Title: "AGENTS", Width: 6},
	{Title: "MODEL:TOKEN", Width: 20, Flex: true, MaxPercent: 0.25},
	{Title: "LAST ACTIVE", Width: 11},
}

// NewSessionsView creates a sessions view.
func NewSessionsView(width, height int) *ResourceView[*model.Session] {
	return NewResourceView(sessionColumnsBase, sessionColumnsFlat, sessionRow, width, height)
}

func sessionRow(items []*model.Session, i int, flatMode bool) ui.Row {
	s := items[i]
	var cells []string
	// subtitleIndent = sum of fixed column widths + separating spaces before the TOPIC column
	subtitleIndent := 16 + 1 + 19 + 1 // SLUG(16) + space + NAME(19) + space
	if flatMode {
		cells = append(cells, truncateHash(s.ProjectHash))
		subtitleIndent = 20 + 1 + 16 + 1 + 19 + 1 // PROJECT(20) + space + SLUG(16) + space + NAME(19) + space
	}
	cells = append(cells,
		s.Slug,
		s.GroupNameCell(),
		s.TopicShort(120),
		fmt.Sprintf("%d", s.NumTurns),
		fmt.Sprintf("%d", s.AgentCount),
		s.TokenString(),
		s.LastActive(),
	)
	row := ui.Row{
		Cells:          cells,
		Subtitle:       s.MetaLine(),
		SubtitleIndent: subtitleIndent,
		Data:           s,
		Hot:            time.Since(s.ModTime) <= 5*time.Second,
	}
	return row
}
