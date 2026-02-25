package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Curt-Park/claudeview/internal/config"
)

func TestLoadInstalledPlugins_ArrayFormat(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `[
		{"name":"superpowers","version":"4.3.1","marketplace":"official","enabled":true,"installedAt":"2025-12-15"},
		{"name":"Notion","version":"1.2.0","marketplace":"official","enabled":false,"installedAt":"2025-11-20"}
	]`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	plugins, err := config.LoadInstalledPlugins(dir)
	if err != nil {
		t.Fatalf("LoadInstalledPlugins: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("len = %d, want 2", len(plugins))
	}
	if plugins[0].Name != "superpowers" {
		t.Errorf("Name = %q, want %q", plugins[0].Name, "superpowers")
	}
	if !plugins[0].Enabled {
		t.Error("expected plugins[0].Enabled = true")
	}
}

func TestLoadInstalledPlugins_MissingFile(t *testing.T) {
	plugins, err := config.LoadInstalledPlugins(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plugins != nil {
		t.Errorf("expected nil plugins, got %v", plugins)
	}
}

func TestPluginCacheDir(t *testing.T) {
	dir := config.PluginCacheDir("/home/user/.claude", "official", "superpowers", "4.3.1")
	want := "/home/user/.claude/plugins/cache/official/superpowers/4.3.1"
	if dir != want {
		t.Errorf("PluginCacheDir = %q, want %q", dir, want)
	}
}

func TestEnabledPlugins_WithField(t *testing.T) {
	dir := t.TempDir()
	content := `{"enabledPlugins":["superpowers","Notion"]}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	enabled, err := config.EnabledPlugins(dir)
	if err != nil {
		t.Fatalf("EnabledPlugins: %v", err)
	}
	if !enabled["superpowers"] {
		t.Error("expected superpowers to be enabled")
	}
	if !enabled["Notion"] {
		t.Error("expected Notion to be enabled")
	}
	if enabled["other"] {
		t.Error("expected other to not be enabled")
	}
}

func TestEnabledPlugins_MissingFile(t *testing.T) {
	enabled, err := config.EnabledPlugins(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error for missing settings file, got %v", err)
	}
	if len(enabled) != 0 {
		t.Errorf("expected empty map for missing file, got %v", enabled)
	}
}

func TestEnabledPlugins_NoField(t *testing.T) {
	dir := t.TempDir()
	content := `{"model":"claude-opus-4-6"}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	enabled, err := config.EnabledPlugins(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(enabled) != 0 {
		t.Errorf("expected empty map, got %v", enabled)
	}
}
