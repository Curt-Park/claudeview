package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Curt-Park/claudeview/internal/config"
)

func TestLoadSettings_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `{
		"model": "claude-opus-4-6",
		"mcpServers": {
			"fs": {"command": "npx", "args": ["@anthropic/mcp-fs"], "type": "stdio"}
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := config.LoadSettings(dir)
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if s.Model != "claude-opus-4-6" {
		t.Errorf("Model = %q, want %q", s.Model, "claude-opus-4-6")
	}
	if len(s.MCPServers) != 1 {
		t.Errorf("MCPServers len = %d, want 1", len(s.MCPServers))
	}
	srv := s.MCPServers["fs"]
	if srv.Command != "npx" {
		t.Errorf("Command = %q, want %q", srv.Command, "npx")
	}
}

func TestLoadSettings_MissingFile(t *testing.T) {
	s, err := config.LoadSettings(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil Settings")
	}
}

func TestLoadSettings_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{bad json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadSettings(dir)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestClaudeDir(t *testing.T) {
	dir := config.ClaudeDir()
	if !strings.HasSuffix(dir, "/.claude") {
		t.Errorf("ClaudeDir() = %q, want suffix /.claude", dir)
	}
}
