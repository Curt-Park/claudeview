package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var pluginColumns = []ui.Column{
	{Title: "NAME", Width: 20, Flex: true},
	{Title: "VERSION", Width: 10},
	{Title: "STATUS", Width: 10},
	{Title: "SKILLS", Width: 7},
	{Title: "COMMANDS", Width: 9},
	{Title: "HOOKS", Width: 6},
	{Title: "INSTALLED", Width: 12},
}

// NewPluginsView creates a plugins view.
func NewPluginsView(width, height int) *ResourceView[*model.Plugin] {
	return NewResourceView(pluginColumns, nil, pluginRow, width, height)
}

func pluginRow(items []*model.Plugin, i int, _ bool) ui.Row {
	p := items[i]
	statusStr := "disabled"
	statusStyle := ui.StyleDone
	if p.Enabled {
		statusStr = "enabled"
		statusStyle = ui.StyleRunning
	}
	return ui.Row{
		Cells: []string{
			p.Name,
			p.Version,
			statusStyle.Render(statusStr),
			fmt.Sprintf("%d", p.SkillCount),
			fmt.Sprintf("%d", p.CommandCount),
			fmt.Sprintf("%d", p.HookCount),
			p.InstalledAt,
		},
		Data: p,
	}
}
