package ui_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestRenderPluginItemDetail_Nil(t *testing.T) {
	got := ui.RenderPluginItemDetail(nil)
	if got != "" {
		t.Errorf("expected empty string for nil item, got %q", got)
	}
}

func TestRenderPluginItemDetail_ShowsSkillContent(t *testing.T) {
	cacheDir := t.TempDir()
	skillDir := filepath.Join(cacheDir, "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "my-skill.md"), []byte("# My Skill\nDoes things."), 0o644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	item := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: cacheDir}
	got := ui.RenderPluginItemDetail(item)
	if !strings.Contains(got, "my-skill") {
		t.Errorf("expected output to contain item name %q, got %q", "my-skill", got)
	}
	if !strings.Contains(got, "skill") {
		t.Errorf("expected output to contain category %q, got %q", "skill", got)
	}
	if !strings.Contains(got, "Does things.") {
		t.Errorf("expected output to contain skill content, got %q", got)
	}
}

func TestRenderPluginItemDetail_ErrorOnMissingContent(t *testing.T) {
	item := &model.PluginItem{Name: "missing-skill", Category: "skill", CacheDir: t.TempDir()}
	got := ui.RenderPluginItemDetail(item)
	if !strings.Contains(strings.ToLower(got), "error") && !strings.Contains(got, "no content") {
		t.Errorf("expected output to indicate error or missing content, got %q", got)
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

func TestRenderPluginItemDetail_Hook_ShowsCommandScripts(t *testing.T) {
	cacheDir := t.TempDir()
	hooksDir := filepath.Join(cacheDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("failed to create hooks dir: %v", err)
	}

	scriptContent := "#!/bin/bash\necho session-stop"
	scriptPath := filepath.Join(hooksDir, "stop-hook.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	hookJSON := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"` + scriptPath + `"}]}]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatalf("failed to write hooks.json: %v", err)
	}

	item := &model.PluginItem{Name: "Stop", Category: "hook", CacheDir: cacheDir}
	got := ui.RenderPluginItemDetail(item)

	if !strings.Contains(got, "command scripts below") {
		t.Errorf("expected 'command scripts below' section, got %q", got)
	}
	if !strings.Contains(got, "echo session-stop") {
		t.Errorf("expected script content in output, got %q", got)
	}
}

func TestRenderSessionChat_UserBubble(t *testing.T) {
	turns := []model.Turn{
		{Role: "user", Text: "Hello, Claude!", Timestamp: time.Date(2026, 3, 2, 9, 13, 0, 0, time.UTC)},
	}
	got := ui.RenderSessionChat(turns, nil, nil, 80)
	if !strings.Contains(got, "Hello, Claude!") {
		t.Errorf("expected user text in output, got:\n%s", got)
	}
	if !strings.Contains(got, "You") {
		t.Errorf("expected 'You' label in user bubble, got:\n%s", got)
	}
	if !strings.Contains(got, "09:13") {
		t.Errorf("expected timestamp in user bubble, got:\n%s", got)
	}
}

func TestRenderSessionChat_MultilineUserBubble(t *testing.T) {
	turns := []model.Turn{
		{Role: "user", Text: "Requirement:\n- req1\n- req2", Timestamp: time.Now()},
	}
	got := ui.RenderSessionChat(turns, nil, nil, 80)
	if !strings.Contains(got, "req1") || !strings.Contains(got, "req2") {
		t.Errorf("expected multiline text preserved, got:\n%s", got)
	}
}

func TestRenderSessionChat_NilTurnsReturnsEmpty(t *testing.T) {
	got := ui.RenderSessionChat(nil, nil, nil, 80)
	if strings.TrimSpace(got) != "" {
		t.Errorf("expected empty output for nil turns, got %q", got)
	}
}

func TestRenderSessionChat_ClaudeBubble(t *testing.T) {
	turns := []model.Turn{
		{
			Role:         "assistant",
			Text:         "네, 시작하겠습니다.",
			ModelName:    "claude-sonnet-4-6",
			InputTokens:  500,
			OutputTokens: 100,
			Timestamp:    time.Date(2026, 3, 2, 9, 14, 0, 0, time.UTC),
		},
	}
	got := ui.RenderSessionChat(turns, nil, nil, 120)
	if !strings.Contains(got, "네, 시작하겠습니다.") {
		t.Errorf("expected assistant text, got:\n%s", got)
	}
	if !strings.Contains(got, "Claude") {
		t.Errorf("expected 'Claude' label, got:\n%s", got)
	}
	if !strings.Contains(got, "09:14") {
		t.Errorf("expected timestamp, got:\n%s", got)
	}
	if !strings.Contains(got, "600") || !strings.Contains(got, "tok") {
		t.Errorf("expected token count (600 total), got:\n%s", got)
	}
}

func TestRenderSessionChat_ThinkingDim(t *testing.T) {
	turns := []model.Turn{
		{Role: "assistant", Text: "response", Thinking: "let me think about this carefully"},
	}
	got := ui.RenderSessionChat(turns, nil, nil, 120)
	if !strings.Contains(got, "think") {
		t.Errorf("expected thinking content in output, got:\n%s", got)
	}
}

func TestRenderSessionChat_ToolCallLines(t *testing.T) {
	input := json.RawMessage(`{"file_path":"internal/ui/app.go"}`)
	result := json.RawMessage(`"120 lines\nof content"`)
	turns := []model.Turn{
		{
			Role: "assistant",
			Text: "done",
			ToolCalls: []*model.ToolCall{
				{Name: "Read", Input: input, Result: result, IsError: false},
				{Name: "Bash", Input: json.RawMessage(`{"command":"make test"}`), IsError: true},
			},
		},
	}
	got := ui.RenderSessionChat(turns, nil, nil, 120)
	if !strings.Contains(got, "Read") {
		t.Errorf("expected 'Read' tool name, got:\n%s", got)
	}
	if !strings.Contains(got, "✓") {
		t.Errorf("expected ✓ for successful tool, got:\n%s", got)
	}
	if !strings.Contains(got, "✗") {
		t.Errorf("expected ✗ for failed tool, got:\n%s", got)
	}
}

func TestRenderPluginItemDetail_Hook_NoScriptsShowsOnlyJSON(t *testing.T) {
	cacheDir := t.TempDir()
	hooksDir := filepath.Join(cacheDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("failed to create hooks dir: %v", err)
	}

	// Hook command references inline shell, not a script file
	hookJSON := `{"hooks":{"PreToolUse":[{"hooks":[{"type":"command","command":"echo hello"}]}]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatalf("failed to write hooks.json: %v", err)
	}

	item := &model.PluginItem{Name: "PreToolUse", Category: "hook", CacheDir: cacheDir}
	got := ui.RenderPluginItemDetail(item)

	if strings.Contains(got, "command scripts below") {
		t.Errorf("expected no 'command scripts below' section for inline command, got %q", got)
	}
	if !strings.Contains(got, "PreToolUse") {
		t.Errorf("expected hook name in output, got %q", got)
	}
}
