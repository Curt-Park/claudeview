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
	got := ui.RenderPluginItemDetail(nil, 80)
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
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill\nDoes things."), 0o644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	item := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: cacheDir}
	got := ui.RenderPluginItemDetail(item, 80)
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
	got := ui.RenderPluginItemDetail(item, 80)
	if !strings.Contains(strings.ToLower(got), "error") && !strings.Contains(got, "no content") {
		t.Errorf("expected output to indicate error or missing content, got %q", got)
	}
}

func TestRenderMemoryDetail_Nil(t *testing.T) {
	got := ui.RenderMemoryDetail(nil, 80)
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
	got := ui.RenderMemoryDetail(m, 80)
	if !strings.Contains(got, "hello world") {
		t.Errorf("expected output to contain %q, got %q", "hello world", got)
	}
}

func TestRenderMemoryDetail_ErrorOnBadPath(t *testing.T) {
	m := &model.Memory{Path: "/nonexistent/path.md"}
	got := ui.RenderMemoryDetail(m, 80)
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
	got := ui.RenderPluginItemDetail(item, 80)

	if !strings.Contains(got, "command scripts below") {
		t.Errorf("expected 'command scripts below' section, got %q", got)
	}
	if !strings.Contains(got, "echo session-stop") {
		t.Errorf("expected script content in output, got %q", got)
	}
}

func TestRenderChatItemDetail_UserBubble(t *testing.T) {
	item := ui.ChatItem{
		Turn:        model.Turn{Role: "user", Text: "Hello, Claude!", Timestamp: time.Date(2026, 3, 2, 9, 13, 0, 0, time.UTC)},
		SubagentIdx: -1,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 80)
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

func TestRenderChatItemDetail_MultilineUserBubble(t *testing.T) {
	item := ui.ChatItem{
		Turn:        model.Turn{Role: "user", Text: "Requirement:\n- req1\n- req2", Timestamp: time.Now()},
		SubagentIdx: -1,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 80)
	if !strings.Contains(got, "req1") || !strings.Contains(got, "req2") {
		t.Errorf("expected multiline text preserved, got:\n%s", got)
	}
}

func TestRenderChatItemDetail_ClaudeBubble(t *testing.T) {
	item := ui.ChatItem{
		Turn: model.Turn{
			Role:         "assistant",
			Text:         "네, 시작하겠습니다.",
			ModelName:    "claude-sonnet-4-6",
			InputTokens:  500,
			OutputTokens: 100,
			Timestamp:    time.Date(2026, 3, 2, 9, 14, 0, 0, time.UTC),
		},
		SubagentIdx: -1,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 120)
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

func TestRenderChatItemDetail_ThinkingDim(t *testing.T) {
	item := ui.ChatItem{
		Turn:        model.Turn{Role: "assistant", Text: "response", Thinking: "let me think about this carefully"},
		SubagentIdx: -1,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 120)
	if !strings.Contains(got, "think") {
		t.Errorf("expected thinking content in output, got:\n%s", got)
	}
}

func TestRenderChatItemDetail_ToolCallLines(t *testing.T) {
	input := json.RawMessage(`{"file_path":"internal/ui/app.go"}`)
	result := json.RawMessage(`"120 lines\nof content"`)
	item := ui.ChatItem{
		Turn: model.Turn{
			Role: "assistant",
			Text: "done",
			ToolCalls: []*model.ToolCall{
				{Name: "Read", Input: input, Result: result, IsError: false},
				{Name: "Bash", Input: json.RawMessage(`{"command":"make test"}`), IsError: true},
			},
		},
		SubagentIdx: -1,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 120)
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

func TestRenderChatItemDetail_SubagentBubble(t *testing.T) {
	item := ui.ChatItem{
		Turn:        model.Turn{Text: "Found 42 files.", ModelName: "claude-sonnet-4-6", Role: "assistant"},
		IsSubagent:  true,
		AgentType:   model.AgentTypeExplore,
		SubagentIdx: 0,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 120)
	if !strings.Contains(got, "Found 42 files.") {
		t.Errorf("expected subagent text in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Explorer") || !strings.Contains(got, "🔍") {
		t.Errorf("expected subagent label with icon, got:\n%s", got)
	}
}

func TestRenderChatItemDetail_GroupedWithExtraTurns(t *testing.T) {
	item := ui.ChatItem{
		Turn: model.Turn{
			Role:         "assistant",
			Text:         "Let me check the code.",
			ModelName:    "claude-opus-4-6",
			InputTokens:  500,
			OutputTokens: 100,
			Timestamp:    time.Date(2026, 3, 2, 9, 15, 0, 0, time.UTC),
			ToolCalls: []*model.ToolCall{
				{Name: "Grep", Input: json.RawMessage(`{"pattern":"KeyUp"}`)},
			},
		},
		ExtraTurns: []model.Turn{
			{
				Role:         "assistant",
				InputTokens:  300,
				OutputTokens: 50,
				ToolCalls: []*model.ToolCall{
					{Name: "Read", Input: json.RawMessage(`{"file_path":"app.go"}`)},
					{Name: "Edit", Input: json.RawMessage(`{"file_path":"app.go"}`)},
				},
			},
		},
		SubagentIdx: -1,
	}
	got := ui.RenderChatItemDetail([]ui.ChatItem{item}, 0, 120)

	// Primary text should be visible
	if !strings.Contains(got, "Let me check the code.") {
		t.Errorf("expected primary text in output, got:\n%s", got)
	}
	// Primary tool call
	if !strings.Contains(got, "Grep") {
		t.Errorf("expected primary tool call 'Grep', got:\n%s", got)
	}
	// ExtraTurn tool calls
	if !strings.Contains(got, "Read") {
		t.Errorf("expected extra turn tool call 'Read', got:\n%s", got)
	}
	if !strings.Contains(got, "Edit") {
		t.Errorf("expected extra turn tool call 'Edit', got:\n%s", got)
	}
	// No separator — ExtraTurns flow naturally after primary turn
	// Aggregated tokens: 500+100+300+50 = 950
	if !strings.Contains(got, "950") {
		t.Errorf("expected aggregated token count 950, got:\n%s", got)
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
	got := ui.RenderPluginItemDetail(item, 80)

	if strings.Contains(got, "command scripts below") {
		t.Errorf("expected no 'command scripts below' section for inline command, got %q", got)
	}
	if !strings.Contains(got, "PreToolUse") {
		t.Errorf("expected hook name in output, got %q", got)
	}
}

func TestRenderChatItemDetail_SubagentFullTranscript(t *testing.T) {
	// Subagent is now a single collapsed ChatItem with ExtraTurns.
	items := []ui.ChatItem{
		{Turn: model.Turn{Role: "user", Text: "Do something"}, SubagentIdx: -1},
		{Turn: model.Turn{Role: "assistant", Text: "Sure, let me delegate."}, SubagentIdx: -1},
		{
			Turn:        model.Turn{Role: "assistant", Text: "Step 1: searching files"},
			ExtraTurns:  []model.Turn{{Role: "assistant", Text: "Step 2: found results"}},
			IsSubagent:  true,
			AgentType:   model.AgentTypeExplore,
			SubagentIdx: 0,
		},
		{Turn: model.Turn{Role: "assistant", Text: "Done!"}, SubagentIdx: -1},
	}

	// Select the collapsed subagent item (index 2) — should render both turns.
	got := ui.RenderChatItemDetail(items, 2, 120)
	if !strings.Contains(got, "Step 1: searching files") {
		t.Errorf("expected first subagent turn in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Step 2: found results") {
		t.Errorf("expected second subagent turn in output, got:\n%s", got)
	}
	// Should NOT contain non-subagent turns.
	if strings.Contains(got, "Do something") {
		t.Errorf("expected user turn excluded from subagent transcript, got:\n%s", got)
	}
	if strings.Contains(got, "Done!") {
		t.Errorf("expected regular assistant turn excluded from subagent transcript, got:\n%s", got)
	}
}

func TestRenderChatItemDetail_RegularAssistantSingleTurn(t *testing.T) {
	items := []ui.ChatItem{
		{Turn: model.Turn{Role: "user", Text: "Hello"}, SubagentIdx: -1},
		{Turn: model.Turn{Role: "assistant", Text: "Hi there"}, SubagentIdx: -1},
		{Turn: model.Turn{Role: "assistant", Text: "Subagent work"}, IsSubagent: true, SubagentIdx: 0},
	}

	// Select regular assistant (index 1) — should render only that single turn.
	got := ui.RenderChatItemDetail(items, 1, 120)
	if !strings.Contains(got, "Hi there") {
		t.Errorf("expected assistant text, got:\n%s", got)
	}
	if strings.Contains(got, "Subagent work") {
		t.Errorf("expected subagent turn excluded from regular assistant detail, got:\n%s", got)
	}
}

func TestRenderChatItemDetail_AgentCallCompact(t *testing.T) {
	// Agent/Task tool calls in a Claude turn should render as icon+name + full result.
	taskInput, _ := json.Marshal(map[string]string{"subagent_type": "Explore"})
	result, _ := json.Marshal("Found 42 files matching the pattern.")
	claudeItem := ui.ChatItem{
		Turn: model.Turn{
			Role: "assistant",
			Text: "delegating",
			ToolCalls: []*model.ToolCall{
				{Name: "Agent", Input: taskInput, Result: result},
				{Name: "Read", Input: json.RawMessage(`{"file_path":"app.go"}`), Result: json.RawMessage(`"120 lines"`)},
			},
		},
		SubagentIdx: -1,
	}
	items := []ui.ChatItem{claudeItem}
	got := ui.RenderChatItemDetail(items, 0, 120)

	if !strings.Contains(got, "Explorer") {
		t.Errorf("expected agent name 'Explorer' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Found 42 files matching the pattern.") {
		t.Errorf("expected full result in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Read") {
		t.Errorf("expected 'Read' tool in output, got:\n%s", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Explorer") && (strings.Contains(line, "✓") || strings.Contains(line, "✗")) {
			t.Errorf("Agent call line should not contain ✓/✗, got: %q", line)
		}
	}
}

func TestRenderChatItemDetail_AgentCallFallbackToSubItem(t *testing.T) {
	// When tc.Result is nil, the sub-agent ChatItem's last turn text is shown.
	taskInput, _ := json.Marshal(map[string]string{"subagent_type": "Explore"})
	claudeItem := ui.ChatItem{
		Turn: model.Turn{
			Role: "assistant",
			Text: "delegating",
			ToolCalls: []*model.ToolCall{
				// No Result set (agent in progress).
				{Name: "Agent", Input: taskInput},
			},
		},
		SubagentIdx: -1,
	}
	subItem := ui.ChatItem{
		Turn:        model.Turn{Role: "assistant", Text: "still searching..."},
		ExtraTurns:  []model.Turn{{Role: "assistant", Text: "found the answer"}},
		IsSubagent:  true,
		AgentType:   model.AgentTypeExplore,
		SubagentIdx: 0,
	}
	items := []ui.ChatItem{claudeItem, subItem}
	got := ui.RenderChatItemDetail(items, 0, 120)

	// Should show the last ExtraTurn text as fallback.
	if !strings.Contains(got, "found the answer") {
		t.Errorf("expected fallback sub-agent text in output, got:\n%s", got)
	}
}

func TestRenderChatItemDetail_OutOfBounds(t *testing.T) {
	items := []ui.ChatItem{
		{Turn: model.Turn{Role: "user", Text: "Hello"}, SubagentIdx: -1},
	}
	if got := ui.RenderChatItemDetail(items, -1, 80); got != "" {
		t.Errorf("expected empty for negative index, got %q", got)
	}
	if got := ui.RenderChatItemDetail(items, 5, 80); got != "" {
		t.Errorf("expected empty for out-of-range index, got %q", got)
	}
}
