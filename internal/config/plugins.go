package config

import (
	"encoding/json"
	"os"
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
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var plugins []InstalledPlugin
	if err := json.Unmarshal(data, &plugins); err != nil {
		// Try map format
		var m map[string]InstalledPlugin
		if err2 := json.Unmarshal(data, &m); err2 != nil {
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
	settings, err := LoadSettings(claudeDir)
	if err != nil {
		return nil, err
	}
	// settings may include enabledPlugins field
	path := filepath.Join(claudeDir, "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	_ = settings
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
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
