package model

import (
	"os"
	"path/filepath"
)

// Plugin represents an installed Claude Code plugin.
type Plugin struct {
	Name         string
	Version      string
	Marketplace  string
	Enabled      bool
	InstalledAt  string
	CacheDir     string
	MCPServers   []*MCPServer
	SkillCount   int
	CommandCount int
	HookCount    int
}

// HasMCP returns true if this plugin provides any MCP servers.
func (p *Plugin) HasMCP() bool {
	return len(p.MCPServers) > 0
}

// CountSkills counts .md files in the plugin's skills/ directory.
func CountSkills(cacheDir string) int {
	return countFiles(filepath.Join(cacheDir, "skills"), ".md")
}

// CountCommands counts .md files in the plugin's commands/ directory.
func CountCommands(cacheDir string) int {
	return countFiles(filepath.Join(cacheDir, "commands"), ".md")
}

// CountHooks counts .sh/.js files in the plugin's hooks/ directory.
func CountHooks(cacheDir string) int {
	return countFiles(filepath.Join(cacheDir, "hooks"), "")
}

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
