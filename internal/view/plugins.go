package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// PluginsView renders the plugins list.
type PluginsView struct {
	Plugins []*model.Plugin
	Table   ui.TableView
}

var pluginColumns = []ui.Column{
	{Title: "NAME", Width: 20, Flex: true},
	{Title: "VERSION", Width: 10},
	{Title: "STATUS", Width: 10},
	{Title: "MCP SERVERS", Width: 11},
	{Title: "INSTALLED", Width: 12},
}

// NewPluginsView creates a plugins view.
func NewPluginsView(width, height int) *PluginsView {
	return &PluginsView{
		Table: ui.NewTableView(pluginColumns, width, height),
	}
}

// SetPlugins updates the plugins list.
func (v *PluginsView) SetPlugins(plugins []*model.Plugin) {
	v.Plugins = plugins
	rows := make([]ui.Row, len(plugins))
	for i, p := range plugins {
		statusStr := "disabled"
		statusStyle := ui.StyleDone
		if p.Enabled {
			statusStr = "enabled"
			statusStyle = ui.StyleRunning
		}
		rows[i] = ui.Row{
			Cells: []string{
				p.Name,
				p.Version,
				statusStyle.Render(statusStr),
				fmt.Sprintf("%d", len(p.MCPServers)),
				p.InstalledAt,
			},
			Data: p,
		}
	}
	v.Table.SetRows(rows)
}

// SelectedPlugin returns the currently selected plugin.
func (v *PluginsView) SelectedPlugin() *model.Plugin {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if p, ok := row.Data.(*model.Plugin); ok {
		return p
	}
	return nil
}

// View renders the plugins table.
func (v *PluginsView) View() string {
	return v.Table.View()
}
