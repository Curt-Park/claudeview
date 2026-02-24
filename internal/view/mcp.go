package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// MCPView renders the MCP servers list.
type MCPView struct {
	Servers []*model.MCPServer
	Table   ui.TableView
}

var mcpColumns = []ui.Column{
	{Title: "NAME", Width: 16},
	{Title: "PLUGIN", Width: 14},
	{Title: "TRANSPORT", Width: 10},
	{Title: "TOOLS", Width: 6},
	{Title: "COMMAND", Width: 40, Flex: true},
}

// NewMCPView creates an MCP view.
func NewMCPView(width, height int) *MCPView {
	return &MCPView{
		Table: ui.NewTableView(mcpColumns, width, height),
	}
}

// SetServers updates the MCP servers list.
func (v *MCPView) SetServers(servers []*model.MCPServer) {
	v.Servers = servers
	rows := make([]ui.Row, len(servers))
	for i, s := range servers {
		statusStyle := ui.StatusStyle(string(s.Status))
		rows[i] = ui.Row{
			Cells: []string{
				statusStyle.Render(s.Name),
				s.Plugin,
				s.Transport,
				fmt.Sprintf("%d", s.ToolCount),
				s.CommandString(),
			},
			Data: s,
		}
	}
	v.Table.SetRows(rows)
}

// SelectedServer returns the currently selected MCP server.
func (v *MCPView) SelectedServer() *model.MCPServer {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if s, ok := row.Data.(*model.MCPServer); ok {
		return s
	}
	return nil
}

// View renders the MCP servers table.
func (v *MCPView) View() string {
	return v.Table.View()
}
