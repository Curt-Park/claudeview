package config

import (
	"encoding/json"
	"path/filepath"
)

// InstalledPlugin represents an entry in installed_plugins.json.
type InstalledPlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Marketplace string `json:"marketplace"`
	Enabled     bool   `json:"enabled"`
	InstalledAt string `json:"installedAt"`
}

// LoadInstalledPlugins reads ~/.claude/plugins/installed_plugins.json.
func LoadInstalledPlugins(claudeDir string) ([]InstalledPlugin, error) {
	path := filepath.Join(claudeDir, "plugins", "installed_plugins.json")
	plugins, err := loadJSON[[]InstalledPlugin](path)
	if err != nil {
		// Try map format
		m, err2 := loadJSON[map[string]InstalledPlugin](path)
		if err2 != nil {
			return nil, err
		}
		for name, p := range m {
			p.Name = name
			plugins = append(plugins, p)
		}
	}
	return plugins, nil
}

// PluginCacheDir returns the path where a plugin's files are cached.
func PluginCacheDir(claudeDir, marketplace, name, version string) string {
	return filepath.Join(claudeDir, "plugins", "cache", marketplace, name, version)
}

// EnabledPlugins reads the list of enabled plugin names from settings.
func EnabledPlugins(claudeDir string) (map[string]bool, error) {
	raw, err := loadJSON[map[string]json.RawMessage](filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		return nil, nil
	}
	enabled := make(map[string]bool)
	if ep, ok := raw["enabledPlugins"]; ok {
		var names []string
		if err := json.Unmarshal(ep, &names); err == nil {
			for _, n := range names {
				enabled[n] = true
			}
		}
	}
	return enabled, nil
}
