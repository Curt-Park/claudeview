package config

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
)

// InstalledPlugin represents an entry from installed_plugins.json.
type InstalledPlugin struct {
	Name        string
	Version     string
	Marketplace string
	Scope       string // "user", "project", "local"
	ProjectPath string // set for project/local scope plugins
	InstalledAt string
	CacheDir    string // full path to the plugin cache directory
}

// installedPluginsV2 is the v2 format of installed_plugins.json.
type installedPluginsV2 struct {
	Version int                                 `json:"version"`
	Plugins map[string][]installedPluginV2Entry `json:"plugins"`
}

type installedPluginV2Entry struct {
	Scope        string `json:"scope"`
	ProjectPath  string `json:"projectPath"`
	InstallPath  string `json:"installPath"`
	Version      string `json:"version"`
	InstalledAt  string `json:"installedAt"`
	LastUpdated  string `json:"lastUpdated"`
	GitCommitSha string `json:"gitCommitSha"`
}

// LoadInstalledPlugins reads ~/.claude/plugins/installed_plugins.json.
// It supports v2 ({"version":2,"plugins":{...}}) and v1 ([{...}] or {...}) formats.
func LoadInstalledPlugins(claudeDir string) ([]InstalledPlugin, error) {
	path := filepath.Join(claudeDir, "plugins", "installed_plugins.json")

	// Try v2 format first
	v2, err := loadJSON[installedPluginsV2](path)
	if err == nil && v2.Version == 2 && v2.Plugins != nil {
		var plugins []InstalledPlugin
		for key, entries := range v2.Plugins {
			name, marketplace := splitPluginKey(key)
			for _, e := range entries {
				plugins = append(plugins, InstalledPlugin{
					Name:        name,
					Version:     e.Version,
					Marketplace: marketplace,
					Scope:       e.Scope,
					ProjectPath: e.ProjectPath,
					InstalledAt: e.InstalledAt,
					CacheDir:    e.InstallPath,
				})
			}
		}
		sort.Slice(plugins, func(i, j int) bool { return plugins[i].InstalledAt > plugins[j].InstalledAt })
		return plugins, nil
	}

	// v1 fallback: try array format
	type v1Plugin struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Marketplace string `json:"marketplace"`
		InstalledAt string `json:"installedAt"`
	}
	arr, err := loadJSON[[]v1Plugin](path)
	if err == nil {
		var plugins []InstalledPlugin
		for _, p := range arr {
			plugins = append(plugins, InstalledPlugin{
				Name:        p.Name,
				Version:     p.Version,
				Marketplace: p.Marketplace,
				InstalledAt: p.InstalledAt,
				CacheDir:    PluginCacheDir(claudeDir, p.Marketplace, p.Name, p.Version),
			})
		}
		return plugins, nil
	}

	// v1 fallback: try map format
	m, err2 := loadJSON[map[string]v1Plugin](path)
	if err2 != nil {
		return nil, err
	}
	var plugins []InstalledPlugin
	for name, p := range m {
		p.Name = name
		plugins = append(plugins, InstalledPlugin{
			Name:        p.Name,
			Version:     p.Version,
			Marketplace: p.Marketplace,
			InstalledAt: p.InstalledAt,
			CacheDir:    PluginCacheDir(claudeDir, p.Marketplace, p.Name, p.Version),
		})
	}
	return plugins, nil
}

// splitPluginKey splits "name@marketplace" into (name, marketplace).
// If no "@" is found, the whole string is returned as the name.
func splitPluginKey(key string) (name, marketplace string) {
	idx := strings.LastIndex(key, "@")
	if idx < 0 {
		return key, ""
	}
	return key[:idx], key[idx+1:]
}

// PluginCacheDir returns the path where a plugin's files are cached.
func PluginCacheDir(claudeDir, marketplace, name, version string) string {
	return filepath.Join(claudeDir, "plugins", "cache", marketplace, name, version)
}

// ProjectEnabledPlugins reads enabledPlugins from a project's .claude/settings.json
// and .claude/settings.local.json, merging both into a single map.
func ProjectEnabledPlugins(projectRoot string) map[string]bool {
	merged := make(map[string]bool)
	for _, name := range []string{"settings.json", "settings.local.json"} {
		path := filepath.Join(projectRoot, ".claude", name)
		raw, err := loadJSON[map[string]json.RawMessage](path)
		if err != nil {
			continue
		}
		ep, ok := raw["enabledPlugins"]
		if !ok {
			continue
		}
		var m map[string]bool
		if json.Unmarshal(ep, &m) == nil {
			for k, v := range m {
				merged[k] = v
			}
		}
	}
	return merged
}

// EnabledPlugins reads the enabled plugin map from settings.json.
// The actual format is {"enabledPlugins": {"name@marketplace": true, ...}}.
func EnabledPlugins(claudeDir string) (map[string]bool, error) {
	raw, err := loadJSON[map[string]json.RawMessage](filepath.Join(claudeDir, "settings.json"))
	if err != nil {
		return nil, nil
	}
	enabled := make(map[string]bool)
	ep, ok := raw["enabledPlugins"]
	if !ok {
		return enabled, nil
	}
	// Try map[string]bool first (actual format)
	var m map[string]bool
	if err := json.Unmarshal(ep, &m); err == nil {
		return m, nil
	}
	// Fallback: []string
	var names []string
	if err := json.Unmarshal(ep, &names); err == nil {
		for _, n := range names {
			enabled[n] = true
		}
	}
	return enabled, nil
}
