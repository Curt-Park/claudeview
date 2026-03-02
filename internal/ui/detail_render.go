package ui

import (
	"fmt"
	"os"

	"github.com/Curt-Park/claudeview/internal/model"
)

// RenderSessionChat renders a session's conversation as a chat timeline.
// Stub — implemented in Tasks 7-9.
func RenderSessionChat(_ []model.Turn, _ [][]model.Turn, _ []model.AgentType, _ int) string {
	return ""
}

// RenderPluginItemDetail renders the content of a selected plugin item.
func RenderPluginItemDetail(item *model.PluginItem) string {
	if item == nil {
		return ""
	}
	header := StyleTitle.Render(item.Name) + "  " + StyleDim.Render(item.Category)
	content := model.ReadPluginItemContent(item)
	result := header + "\n\n" + content
	if item.Category == "hook" {
		scripts := model.ReadHookCommandScripts(item)
		if len(scripts) > 0 {
			result += "\n\ncommand scripts below:\n"
			for _, s := range scripts {
				result += "\n" + StyleDim.Render("--- "+s.Path+" ---") + "\n" + s.Content
			}
		}
	}
	return result
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
