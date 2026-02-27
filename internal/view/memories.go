package view

import (
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var memoryColumns = []ui.Column{
	{Title: "NAME", Width: 18},
	{Title: "TITLE", Width: 20, Flex: true, MaxPercent: 0.45},
	{Title: "SIZE", Width: 8},
	{Title: "MODIFIED", Width: 11},
}

// NewMemoriesView creates a memories view.
func NewMemoriesView(width, height int) *ResourceView[*model.Memory] {
	return NewResourceView(memoryColumns, nil, memoryRow, width, height)
}


func memoryRow(items []*model.Memory, i int, _ bool) ui.Row {
	m := items[i]
	return ui.Row{
		Cells: []string{
			m.Name,
			m.Title,
			m.SizeStr(),
			m.LastModified(),
		},
		Data: m,
	}
}
