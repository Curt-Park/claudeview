package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
)

// RenderPluginDetail renders plugin content as a plain string for the detail view.
func RenderPluginDetail(p *model.Plugin, _ int) string {
	if p == nil {
		return ""
	}

	var sb strings.Builder

	// Header line: Name (bold/colored) + Version and Scope (dimmed).
	header := StyleTitle.Render(p.Name)
	if p.Version != "" {
		header += "  " + StyleDim.Render(p.Version)
	}
	if p.Scope != "" {
		header += "  " + StyleDim.Render(p.Scope)
	}
	sb.WriteString(header)

	type section struct {
		label string
		items []string
	}

	sections := []section{
		{"Skills", model.ListSkills(p.CacheDir)},
		{"Commands", model.ListCommands(p.CacheDir)},
		{"Hooks", model.ListHooks(p.CacheDir)},
		{"Agents", model.ListAgents(p.CacheDir)},
		{"MCPs", model.ListMCPs(p.CacheDir)},
	}

	for _, s := range sections {
		if len(s.items) == 0 {
			continue
		}
		sb.WriteString("\n\n")
		sb.WriteString(StyleKey.Render(s.label + ":"))
		for _, item := range s.items {
			sb.WriteString("\n  " + item)
		}
	}

	return sb.String()
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
