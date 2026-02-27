package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var mcpColumns = []ui.Column{
	{Title: "NAME", Width: 16},
	{Title: "PLUGIN", Width: 14},
	{Title: "TRANSPORT", Width: 10},
	{Title: "STATUS", Width: 10},
	{Title: "TOOLS", Width: 6},
	{Title: "COMMAND", Width: 40, Flex: true},
}

// NewMCPView creates an MCP view.
func NewMCPView(width, height int) *ResourceView[*model.MCPServer] {
	return NewResourceView(mcpColumns, nil, mcpRow, width, height)
}

func mcpRow(items []*model.MCPServer, i int, _ bool) ui.Row {
	s := items[i]
	statusStyle := ui.StatusStyle(string(s.Status))
	return ui.Row{
		Cells: []string{
			s.Name,
			s.Plugin,
			s.Transport,
			statusStyle.Render(string(s.Status)),
			fmt.Sprintf("%d", s.ToolCount),
			s.CommandString(),
		},
		Data: s,
	}
}
