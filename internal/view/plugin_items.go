package view

import (
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var pluginItemColumns = []ui.Column{
	{Title: "CATEGORY", Width: 10},
	{Title: "NAME", Width: 20, Flex: true},
}

// NewPluginItemsView creates a plugin-item navigable table view.
func NewPluginItemsView(width, height int) *ResourceView[*model.PluginItem] {
	return NewResourceView(pluginItemColumns, nil, pluginItemRow, width, height)
}

func pluginItemRow(items []*model.PluginItem, i int, _ bool) ui.Row {
	item := items[i]
	return ui.Row{
		Cells: []string{
			item.Category,
			item.Name,
		},
		Data: item,
	}
}
