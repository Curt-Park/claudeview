package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Plugin represents an installed Claude Code plugin.
type Plugin struct {
	Name         string
	Version      string
	Marketplace  string
	Scope        string // "user", "project", "local"
	Enabled      bool
	InstalledAt  string
	CacheDir     string
	SkillCount   int
	CommandCount int
	HookCount    int
	AgentCount   int
	MCPCount     int
}

// contentDir returns the effective content directory for a plugin.
// Some plugins (e.g. semgrep) nest their content under a "plugin/" subdirectory.
func contentDir(cacheDir string) string {
	sub := filepath.Join(cacheDir, "plugin")
	if info, err := os.Stat(sub); err == nil && info.IsDir() {
		return sub
	}
	return cacheDir
}

// CountSkills counts skill subdirectories in the plugin's skills/ directory.
func CountSkills(cacheDir string) int {
	return countDirs(filepath.Join(contentDir(cacheDir), "skills"))
}

// CountCommands counts .md files in the plugin's commands/ directory.
func CountCommands(cacheDir string) int {
	return countFiles(filepath.Join(contentDir(cacheDir), "commands"), ".md")
}

// CountHooks counts hook event entries for a plugin.
// If a hooks.json file is present it is parsed and the number of event types
// (top-level keys inside "hooks") is returned. Otherwise all files in the
// hooks/ directory are counted.
func CountHooks(cacheDir string) int {
	hooksDir := filepath.Join(contentDir(cacheDir), "hooks")
	jsonPath := filepath.Join(hooksDir, "hooks.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		var wrapper struct {
			Hooks map[string]json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil {
			return len(wrapper.Hooks)
		}
	}
	return countFiles(hooksDir, "")
}

// CountAgents counts .md files in the plugin's agents/ directory.
func CountAgents(cacheDir string) int {
	return countFiles(filepath.Join(contentDir(cacheDir), "agents"), ".md")
}

// mcpServers reads and returns the mcpServers map from .mcp.json or
// .claude-plugin/plugin.json, whichever is found first. Returns nil if neither exists.
func mcpServers(cacheDir string) map[string]json.RawMessage {
	type mcpWrapper struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	candidates := []string{
		filepath.Join(contentDir(cacheDir), ".mcp.json"),
		filepath.Join(cacheDir, ".claude-plugin", "plugin.json"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var w mcpWrapper
		if err := json.Unmarshal(data, &w); err == nil && len(w.MCPServers) > 0 {
			return w.MCPServers
		}
	}
	return nil
}

// CountMCPs counts MCP server entries for a plugin.
// It checks .mcp.json and .claude-plugin/plugin.json (in that order),
// returning the count from the first file that contains mcpServers.
func CountMCPs(cacheDir string) int {
	return len(mcpServers(cacheDir))
}

// ListSkills returns the names of skill subdirectories.
func ListSkills(cacheDir string) []string {
	return listDirNames(filepath.Join(contentDir(cacheDir), "skills"))
}

// ListCommands returns command names (filenames without .md extension).
func ListCommands(cacheDir string) []string {
	return listFileStems(filepath.Join(contentDir(cacheDir), "commands"), ".md")
}

// ListHooks returns hook event names from hooks.json or filenames.
func ListHooks(cacheDir string) []string {
	hooksDir := filepath.Join(contentDir(cacheDir), "hooks")
	jsonPath := filepath.Join(hooksDir, "hooks.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		var wrapper struct {
			Hooks map[string]json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil {
			names := make([]string, 0, len(wrapper.Hooks))
			for k := range wrapper.Hooks {
				names = append(names, k)
			}
			sort.Strings(names)
			return names
		}
	}
	return listFileStems(hooksDir, "")
}

// ListAgents returns agent names (filenames without .md extension).
func ListAgents(cacheDir string) []string {
	return listFileStems(filepath.Join(contentDir(cacheDir), "agents"), ".md")
}

// ListMCPs returns MCP server names from .mcp.json or plugin.json.
func ListMCPs(cacheDir string) []string {
	servers := mcpServers(cacheDir)
	if servers == nil {
		return nil
	}
	names := make([]string, 0, len(servers))
	for k := range servers {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// PluginItem represents a single skill, command, hook, agent, or MCP within a plugin.
type PluginItem struct {
	Name     string
	Category string // "skill", "command", "hook", "agent", "mcp"
	CacheDir string
}

// ListPluginItems aggregates all items from all categories for a plugin.
func ListPluginItems(cacheDir string) []*PluginItem {
	var items []*PluginItem
	for _, name := range ListSkills(cacheDir) {
		items = append(items, &PluginItem{Name: name, Category: "skill", CacheDir: cacheDir})
	}
	for _, name := range ListCommands(cacheDir) {
		items = append(items, &PluginItem{Name: name, Category: "command", CacheDir: cacheDir})
	}
	for _, name := range ListHooks(cacheDir) {
		items = append(items, &PluginItem{Name: name, Category: "hook", CacheDir: cacheDir})
	}
	for _, name := range ListAgents(cacheDir) {
		items = append(items, &PluginItem{Name: name, Category: "agent", CacheDir: cacheDir})
	}
	for _, name := range ListMCPs(cacheDir) {
		items = append(items, &PluginItem{Name: name, Category: "mcp", CacheDir: cacheDir})
	}
	return items
}

// ReadPluginItemContent returns the textual content for a plugin item.
func ReadPluginItemContent(item *PluginItem) string {
	cd := contentDir(item.CacheDir)
	switch item.Category {
	case "skill":
		// Find first .md file in skills/<name>/
		skillDir := filepath.Join(cd, "skills", item.Name)
		entries, err := os.ReadDir(skillDir)
		if err != nil {
			return fmt.Sprintf("error reading skill directory: %v", err)
		}
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
				data, err := os.ReadFile(filepath.Join(skillDir, e.Name()))
				if err != nil {
					return fmt.Sprintf("error reading %s: %v", e.Name(), err)
				}
				return string(data)
			}
		}
		return "(no content found)"
	case "command":
		data, err := os.ReadFile(filepath.Join(cd, "commands", item.Name+".md"))
		if err != nil {
			return fmt.Sprintf("error reading command: %v", err)
		}
		return string(data)
	case "hook":
		// Try hooks.json first
		hooksDir := filepath.Join(cd, "hooks")
		jsonPath := filepath.Join(hooksDir, "hooks.json")
		if data, err := os.ReadFile(jsonPath); err == nil {
			var wrapper struct {
				Hooks map[string]json.RawMessage `json:"hooks"`
			}
			if err := json.Unmarshal(data, &wrapper); err == nil {
				if raw, ok := wrapper.Hooks[item.Name]; ok {
					return normalizeJSON(raw)
				}
			}
		}
		// Fall back to file in hooks/
		entries, err := os.ReadDir(hooksDir)
		if err != nil {
			return fmt.Sprintf("error reading hooks directory: %v", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())) == item.Name {
				data, err := os.ReadFile(filepath.Join(hooksDir, e.Name()))
				if err != nil {
					return fmt.Sprintf("error reading hook file: %v", err)
				}
				return string(data)
			}
		}
		return "(no content found)"
	case "agent":
		data, err := os.ReadFile(filepath.Join(cd, "agents", item.Name+".md"))
		if err != nil {
			return fmt.Sprintf("error reading agent: %v", err)
		}
		return string(data)
	case "mcp":
		servers := mcpServers(item.CacheDir)
		if raw, ok := servers[item.Name]; ok {
			return normalizeJSON(raw)
		}
		return "(no content found)"
	default:
		return "(unknown category)"
	}
}

// listDirNames returns names of subdirectories in dir.
func listDirNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// listFileStems returns filenames in dir with the given extension stripped.
// If ext is empty, all filenames are returned as-is.
func listFileStems(dir, ext string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if ext != "" {
			if filepath.Ext(name) != ext {
				continue
			}
			name = strings.TrimSuffix(name, ext)
		}
		names = append(names, name)
	}
	return names
}

// countDirs counts subdirectories in dir.
func countDirs(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			count++
		}
	}
	return count
}

// countFiles counts files with the given extension in dir.
// If ext is empty, all files are counted.
func countFiles(dir, ext string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ext == "" || filepath.Ext(e.Name()) == ext {
			count++
		}
	}
	return count
}

// normalizeJSON re-indents raw JSON bytes with consistent 2-space indentation,
// stripping any original indentation carried over from the parent JSON structure.
func normalizeJSON(raw json.RawMessage) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return string(raw)
	}
	return buf.String()
}
