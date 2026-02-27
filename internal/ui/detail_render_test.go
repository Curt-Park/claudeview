package ui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestRenderPluginDetail_Nil(t *testing.T) {
	got := ui.RenderPluginDetail(nil)
	if got != "" {
		t.Errorf("expected empty string for nil plugin, got %q", got)
	}
}

func TestRenderPluginDetail_ShowsNameAndVersion(t *testing.T) {
	p := &model.Plugin{
		Name:     "superpowers",
		Version:  "1.0",
		Scope:    "user",
		CacheDir: t.TempDir(),
	}
	got := ui.RenderPluginDetail(p)
	if !strings.Contains(got, "superpowers") {
		t.Errorf("expected output to contain %q, got %q", "superpowers", got)
	}
	if !strings.Contains(got, "1.0") {
		t.Errorf("expected output to contain %q, got %q", "1.0", got)
	}
}

func TestRenderPluginDetail_ShowsSkills(t *testing.T) {
	cacheDir := t.TempDir()
	skillsDir := filepath.Join(cacheDir, "skills")
	if err := os.MkdirAll(filepath.Join(skillsDir, "foo"), 0o755); err != nil {
		t.Fatalf("failed to create skills/foo: %v", err)
	}

	p := &model.Plugin{
		Name:     "myplugin",
		CacheDir: cacheDir,
	}
	got := ui.RenderPluginDetail(p)
	if !strings.Contains(got, "foo") {
		t.Errorf("expected output to contain skill %q, got %q", "foo", got)
	}
}

func TestRenderMemoryDetail_Nil(t *testing.T) {
	got := ui.RenderMemoryDetail(nil)
	if got != "" {
		t.Errorf("expected empty string for nil memory, got %q", got)
	}
}

func TestRenderMemoryDetail_ReturnsFileContent(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "memory.md")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	m := &model.Memory{Path: tmpFile}
	got := ui.RenderMemoryDetail(m)
	if !strings.Contains(got, "hello world") {
		t.Errorf("expected output to contain %q, got %q", "hello world", got)
	}
}

func TestRenderMemoryDetail_ErrorOnBadPath(t *testing.T) {
	m := &model.Memory{Path: "/nonexistent/path.md"}
	got := ui.RenderMemoryDetail(m)
	if !strings.Contains(strings.ToLower(got), "error") {
		t.Errorf("expected output to contain %q for bad path, got %q", "error", got)
	}
}
