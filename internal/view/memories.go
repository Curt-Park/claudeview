package view

import (
	"fmt"
	"os"

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

// RenderMemoryDetail reads and returns the raw content of a memory file.
func RenderMemoryDetail(m *model.Memory) string {
	if m == nil {
		return ""
	}
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Sprintf("error reading %s: %v", m.Path, err)
	}
	return string(data)
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
