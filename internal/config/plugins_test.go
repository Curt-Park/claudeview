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
		{"name":"superpowers","version":"4.3.1","marketplace":"official","installedAt":"2025-12-15"},
		{"name":"Notion","version":"1.2.0","marketplace":"official","installedAt":"2025-11-20"}
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
}

func TestLoadInstalledPluginsV2(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{
		"version": 2,
		"plugins": {
			"superpowers@claude-plugins-official": [
				{
					"scope": "user",
					"installPath": "/home/user/.claude/plugins/cache/claude-plugins-official/superpowers/4.3.1",
					"version": "4.3.1",
					"installedAt": "2025-12-15T00:00:00Z"
				}
			],
			"Notion@claude-plugins-official": [
				{
					"scope": "project",
					"installPath": "/home/user/.claude/plugins/cache/claude-plugins-official/Notion/1.2.0",
					"version": "1.2.0",
					"installedAt": "2025-11-20T00:00:00Z"
				}
			]
		}
	}`
	if err := os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	plugins, err := config.LoadInstalledPlugins(dir)
	if err != nil {
		t.Fatalf("LoadInstalledPlugins v2: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("len = %d, want 2", len(plugins))
	}

	byName := make(map[string]config.InstalledPlugin, len(plugins))
	for _, p := range plugins {
		byName[p.Name] = p
	}

	sp, ok := byName["superpowers"]
	if !ok {
		t.Fatal("superpowers not found in parsed plugins")
	}
	if sp.Marketplace != "claude-plugins-official" {
		t.Errorf("Marketplace = %q, want %q", sp.Marketplace, "claude-plugins-official")
	}
	if sp.Scope != "user" {
		t.Errorf("Scope = %q, want %q", sp.Scope, "user")
	}
	if sp.Version != "4.3.1" {
		t.Errorf("Version = %q, want %q", sp.Version, "4.3.1")
	}
	wantCacheDir := "/home/user/.claude/plugins/cache/claude-plugins-official/superpowers/4.3.1"
	if sp.CacheDir != wantCacheDir {
		t.Errorf("CacheDir = %q, want %q", sp.CacheDir, wantCacheDir)
	}

	no, ok := byName["Notion"]
	if !ok {
		t.Fatal("Notion not found in parsed plugins")
	}
	if no.Scope != "project" {
		t.Errorf("Scope = %q, want %q", no.Scope, "project")
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

func TestEnabledPluginsMapFormat(t *testing.T) {
	dir := t.TempDir()
	content := `{"enabledPlugins":{"superpowers@claude-plugins-official":true,"Notion@claude-plugins-official":false}}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	enabled, err := config.EnabledPlugins(dir)
	if err != nil {
		t.Fatalf("EnabledPlugins: %v", err)
	}
	if !enabled["superpowers@claude-plugins-official"] {
		t.Error("expected superpowers@claude-plugins-official to be enabled")
	}
	if enabled["Notion@claude-plugins-official"] {
		t.Error("expected Notion@claude-plugins-official to be disabled")
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


func TestProjectEnabledPlugins(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"enabledPlugins":{"superpowers@official":true,"Notion@official":false}}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	localContent := `{"enabledPlugins":{"autology@official":true}}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.local.json"), []byte(localContent), 0644); err != nil {
		t.Fatal(err)
	}

	enabled := config.ProjectEnabledPlugins(dir)
	if !enabled["superpowers@official"] {
		t.Error("expected superpowers@official to be enabled")
	}
	if enabled["Notion@official"] {
		t.Error("expected Notion@official to be disabled")
	}
	if !enabled["autology@official"] {
		t.Error("expected autology@official to be enabled (from settings.local.json)")
	}
}

func TestProjectEnabledPluginsMissing(t *testing.T) {
	enabled := config.ProjectEnabledPlugins(t.TempDir())
	if len(enabled) != 0 {
		t.Errorf("expected empty map for missing .claude dir, got %v", enabled)
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
