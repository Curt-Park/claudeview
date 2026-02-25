package config

import (
	"os"
	"path/filepath"
)

// Settings represents ~/.claude/settings.json.
type Settings struct {
	Model           string               `json:"model"`
	EnabledMCPJSONs []string             `json:"enabledMcpjsons"`
	MCPServers      map[string]MCPServer `json:"mcpServers"`
	Hooks           map[string]any       `json:"hooks"`
	Permissions     map[string]any       `json:"permissions"`
}

// MCPServer represents a single MCP server config.
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	Type    string            `json:"type"`
	URL     string            `json:"url"`
}

// LoadSettings loads ~/.claude/settings.json.
func LoadSettings(claudeDir string) (*Settings, error) {
	s, err := loadJSON[Settings](filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ClaudeDir returns the default ~/.claude directory.
func ClaudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}
