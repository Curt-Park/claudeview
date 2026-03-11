package transcript_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Curt-Park/claudeview/internal/transcript"
)

func TestParseFile(t *testing.T) {
	result, err := transcript.ParseFile("testdata/sample_transcript.jsonl")
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if result.NumTurns != 3 {
		t.Errorf("expected 3 turns, got %d", result.NumTurns)
	}

	if result.TotalCost != 0.0123 {
		t.Errorf("expected cost 0.0123, got %f", result.TotalCost)
	}

	// Count tool calls
	toolCount := 0
	for _, turn := range result.Turns {
		toolCount += len(turn.ToolCalls)
	}
	if toolCount != 2 {
		t.Errorf("expected 2 tool calls, got %d", toolCount)
	}

	// Verify tool call matching
	for _, turn := range result.Turns {
		for _, tc := range turn.ToolCalls {
			if tc.Result == nil {
				t.Errorf("tool call %s (%s) has no matched result", tc.ID, tc.Name)
			}
		}
	}

	// Verify topic is set from most recent user message
	// sample_transcript.jsonl has only one user text message
	wantTopic := "Hello! Can you help me explore this codebase?"
	if result.Topic != wantTopic {
		t.Errorf("expected topic %q, got %q", wantTopic, result.Topic)
	}

	// Verify tokens are accumulated per model
	usage, ok := result.TokensByModel["claude-opus-4-6"]
	if !ok {
		t.Error("expected token entry for claude-opus-4-6")
	} else {
		// After Task 3 refactor, TokensByModel accumulates NewInputTokens() per turn
		// (InputTokens + CacheCreationInputTokens, excluding CacheReadInputTokens).
		// turn1: 100+0=100, turn2: 200+0=200, turn3: 300+500=800 → total 1100
		if usage.InputTokens != 1100 {
			t.Errorf("expected InputTokens=1100, got %d", usage.InputTokens)
		}
		if usage.CacheReadInputTokens != 1000 {
			t.Errorf("expected CacheReadInputTokens=1000, got %d", usage.CacheReadInputTokens)
		}
		if usage.OutputTokens != 170 {
			t.Errorf("expected OutputTokens=170, got %d", usage.OutputTokens)
		}
	}

	// Verify total tool call count
	if result.TotalToolCalls != 2 {
		t.Errorf("expected TotalToolCalls=2, got %d", result.TotalToolCalls)
	}
}

func TestParseEmpty(t *testing.T) {
	f, err := os.CreateTemp("", "empty-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile empty file failed: %v", err)
	}
	if len(result.Turns) != 0 {
		t.Errorf("expected 0 turns for empty file, got %d", len(result.Turns))
	}
}

func TestParseNonExistent(t *testing.T) {
	_, err := transcript.ParseFile("/nonexistent/path/file.jsonl")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestTopicSkipsSkillPrefix(t *testing.T) {
	f, err := os.CreateTemp("", "skill-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	// Skill prefix first, then real user message — skill prefix must be skipped
	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Base directory for this skill: /some/path"}]}}` + "\n" +
		`{"type":"user","timestamp":"2025-01-01T10:00:01Z","message":{"role":"user","content":[{"type":"text","text":"How do I implement OAuth?"}]}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	want := "How do I implement OAuth?"
	if result.Topic != want {
		t.Errorf("expected topic %q, got %q", want, result.Topic)
	}
}

func TestTopicUsesFirstUserMessage(t *testing.T) {
	f, err := os.CreateTemp("", "firsttopic-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	// Topic should be the first user message, matching claude -r display
	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"How do I implement OAuth?"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"role":"assistant","content":[{"type":"text","text":"Here is how..."}],"model":"claude-sonnet-4-6","usage":{"input_tokens":10,"output_tokens":5}}}` + "\n" +
		`{"type":"user","timestamp":"2025-01-01T10:00:02Z","message":{"role":"user","content":[{"type":"text","text":"Now refactor the token refresh logic"}]}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	want := "How do I implement OAuth?"
	if result.Topic != want {
		t.Errorf("expected first user message as topic, got %q", result.Topic)
	}
}

func TestTopicSkipsInterruptedByUser(t *testing.T) {
	f, err := os.CreateTemp("", "interrupted-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"[Request interrupted by user for tool use]"}]}}` + "\n" +
		`{"type":"user","timestamp":"2025-01-01T10:00:01Z","message":{"role":"user","content":[{"type":"text","text":"Fix the login bug"}]}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	want := "Fix the login bug"
	if result.Topic != want {
		t.Errorf("expected topic %q, got %q", want, result.Topic)
	}
}

func TestTokensByModelMultiModel(t *testing.T) {
	f, err := os.CreateTemp("", "multimodel-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"role":"assistant","content":[{"type":"text","text":"Hi"}],"model":"claude-sonnet-4-6","usage":{"input_tokens":100,"output_tokens":20}}}` + "\n" +
		`{"type":"user","timestamp":"2025-01-01T10:00:02Z","message":{"role":"user","content":[{"type":"text","text":"Continue"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:03Z","message":{"role":"assistant","content":[{"type":"text","text":"OK"}],"model":"claude-opus-4-6","usage":{"input_tokens":200,"output_tokens":50}}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	sonnet, ok := result.TokensByModel["claude-sonnet-4-6"]
	if !ok {
		t.Fatal("expected entry for claude-sonnet-4-6")
	}
	if sonnet.InputTokens != 100 || sonnet.OutputTokens != 20 {
		t.Errorf("sonnet tokens: got in=%d out=%d, want in=100 out=20", sonnet.InputTokens, sonnet.OutputTokens)
	}

	opus, ok := result.TokensByModel["claude-opus-4-6"]
	if !ok {
		t.Fatal("expected entry for claude-opus-4-6")
	}
	if opus.InputTokens != 200 || opus.OutputTokens != 50 {
		t.Errorf("opus tokens: got in=%d out=%d, want in=200 out=50", opus.InputTokens, opus.OutputTokens)
	}
}

func TestParseAggregatesIncremental(t *testing.T) {
	f, err := os.CreateTemp("", "incremental-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	line1 := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}` + "\n"
	line2 := `{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"role":"assistant","content":[{"type":"text","text":"Hi"},{"type":"tool_use","id":"t1","name":"Read","input":{}}],"model":"claude-sonnet-4-6","usage":{"input_tokens":100,"output_tokens":20}}}` + "\n"

	if _, err := f.WriteString(line1 + line2); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	// First call: nil agg — full parse
	agg1, err := transcript.ParseAggregatesIncremental(f.Name(), nil)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if agg1.Topic != "Hello" {
		t.Errorf("expected topic 'Hello', got %q", agg1.Topic)
	}
	if agg1.TotalToolCalls != 1 {
		t.Errorf("expected 1 tool call, got %d", agg1.TotalToolCalls)
	}
	sonnet := agg1.TokensByModel["claude-sonnet-4-6"]
	if sonnet.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", sonnet.InputTokens)
	}
	if agg1.Offset == 0 {
		t.Error("expected non-zero offset after first call")
	}
	prevOffset := agg1.Offset

	// Second call: same agg, no new data
	agg2, err := transcript.ParseAggregatesIncremental(f.Name(), agg1)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if agg2.Offset != prevOffset {
		t.Errorf("offset should not change when no new data: got %d, want %d", agg2.Offset, prevOffset)
	}
	if agg2.TotalToolCalls != 1 {
		t.Errorf("tool calls should not increase: got %d", agg2.TotalToolCalls)
	}

	// Third call: append new line, incremental should pick it up
	newLine := `{"type":"assistant","timestamp":"2025-01-01T10:00:02Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"t2","name":"Write","input":{}}],"model":"claude-sonnet-4-6","usage":{"input_tokens":50,"output_tokens":10}}}` + "\n"
	appendFile, err := os.OpenFile(f.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := appendFile.WriteString(newLine); err != nil {
		_ = appendFile.Close()
		t.Fatal(err)
	}
	_ = appendFile.Close()

	agg3, err := transcript.ParseAggregatesIncremental(f.Name(), agg2)
	if err != nil {
		t.Fatalf("third call failed: %v", err)
	}
	if agg3.TotalToolCalls != 2 {
		t.Errorf("expected 2 tool calls after append, got %d", agg3.TotalToolCalls)
	}
	sonnet3 := agg3.TokensByModel["claude-sonnet-4-6"]
	if sonnet3.InputTokens != 150 {
		t.Errorf("expected 150 input tokens after append, got %d", sonnet3.InputTokens)
	}
}

func TestParseConsecutiveAssistantEntries(t *testing.T) {
	f, err := os.CreateTemp("", "consecutive-assistant-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	// User message, then two consecutive assistant entries:
	// first has text only, second has tool_use only.
	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Check the code"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"role":"assistant","content":[{"type":"text","text":"Let me check."}],"model":"claude-opus-4-6","usage":{"input_tokens":100,"output_tokens":20}}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:02Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"main.go"}}],"model":"claude-opus-4-6","usage":{"input_tokens":50,"output_tokens":10}}}` + "\n" +
		`{"type":"user","timestamp":"2025-01-01T10:00:03Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"package main"}]}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Should have 2 turns: 1 user + 1 merged assistant (not 2 separate assistant turns)
	if len(result.Turns) != 2 {
		t.Fatalf("expected 2 turns (1 user + 1 merged assistant), got %d", len(result.Turns))
	}

	assistant := result.Turns[1]
	if assistant.Role != "assistant" {
		t.Fatalf("expected assistant turn, got %q", assistant.Role)
	}
	if assistant.Text != "Let me check." {
		t.Errorf("expected merged text %q, got %q", "Let me check.", assistant.Text)
	}
	if len(assistant.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call in merged turn, got %d", len(assistant.ToolCalls))
	}
	if assistant.ToolCalls[0].Name != "Read" {
		t.Errorf("expected tool call name 'Read', got %q", assistant.ToolCalls[0].Name)
	}
	if assistant.ToolCalls[0].Result == nil {
		t.Error("expected tool result to be matched")
	}
	// Usage should be summed
	if assistant.Usage.InputTokens != 150 {
		t.Errorf("expected merged InputTokens=150, got %d", assistant.Usage.InputTokens)
	}
	if assistant.Usage.OutputTokens != 30 {
		t.Errorf("expected merged OutputTokens=30, got %d", assistant.Usage.OutputTokens)
	}
}

func TestParseCompactBoundary(t *testing.T) {
	f, err := os.CreateTemp("", "compact-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"role":"assistant","content":[{"type":"text","text":"Hi there"}],"model":"claude-opus-4-6","usage":{"input_tokens":100,"output_tokens":20}}}` + "\n" +
		`{"type":"system","subtype":"compact_boundary","content":"Conversation compacted","compactMetadata":{"trigger":"auto","preTokens":168000},"timestamp":"2025-01-01T10:05:00Z"}` + "\n" +
		`{"type":"user","timestamp":"2025-01-01T10:05:01Z","message":{"role":"user","content":[{"type":"text","text":"Continue please"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:05:02Z","message":{"role":"assistant","content":[{"type":"text","text":"Sure"}],"model":"claude-opus-4-6","usage":{"input_tokens":50,"output_tokens":10}}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	result, err := transcript.ParseFile(f.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Expect: user, assistant, system, user, assistant = 5 turns
	if len(result.Turns) != 5 {
		t.Fatalf("expected 5 turns, got %d", len(result.Turns))
	}

	sys := result.Turns[2]
	if sys.Role != "system" {
		t.Errorf("expected system turn at index 2, got role %q", sys.Role)
	}
	want := "Conversation compacted (168k tokens)"
	if sys.Text != want {
		t.Errorf("expected text %q, got %q", want, sys.Text)
	}
	if sys.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp on system turn")
	}
}

func TestParseFileIncremental(t *testing.T) {
	f, err := os.CreateTemp("", "incremental-turns-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	// Phase 1: user message + assistant with tool_use
	line1 := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Check the code"}]}}` + "\n"
	line2 := `{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","message":{"role":"assistant","content":[{"type":"text","text":"Let me check."},{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"main.go"}}],"model":"claude-opus-4-6","usage":{"input_tokens":100,"output_tokens":20}}}` + "\n"
	if _, err := f.WriteString(line1 + line2); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	// First call: nil cache
	cache1, err := transcript.ParseFileIncremental(f.Name(), nil)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	turns1 := cache1.Turns()
	// Should have 2 turns: user + pending assistant snapshot
	if len(turns1) != 2 {
		t.Fatalf("phase 1: expected 2 turns, got %d", len(turns1))
	}
	if turns1[0].Role != "user" || turns1[0].Text != "Check the code" {
		t.Errorf("phase 1: unexpected user turn: %+v", turns1[0])
	}
	if turns1[1].Role != "assistant" || turns1[1].Text != "Let me check." {
		t.Errorf("phase 1: unexpected assistant turn: %+v", turns1[1])
	}
	// Tool result not yet matched (no tool_result entry yet)
	if turns1[1].ToolCalls[0].Result != nil {
		t.Error("phase 1: tool result should not be matched yet")
	}

	prevOffset := cache1.Offset()

	// Phase 2: no new data — turns unchanged
	cache2, err := transcript.ParseFileIncremental(f.Name(), cache1)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	turns2 := cache2.Turns()
	if len(turns2) != 2 {
		t.Fatalf("phase 2: expected 2 turns, got %d", len(turns2))
	}
	if cache2.Offset() != prevOffset {
		t.Errorf("phase 2: offset should not change: got %d, want %d", cache2.Offset(), prevOffset)
	}

	// Phase 3: append user with tool_result + new assistant
	line3 := `{"type":"user","timestamp":"2025-01-01T10:00:02Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"package main"}]}}` + "\n"
	line4 := `{"type":"assistant","timestamp":"2025-01-01T10:00:03Z","message":{"role":"assistant","content":[{"type":"text","text":"The file contains a main package."}],"model":"claude-opus-4-6","usage":{"input_tokens":200,"output_tokens":50}}}` + "\n"
	af, err := os.OpenFile(f.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := af.WriteString(line3 + line4); err != nil {
		_ = af.Close()
		t.Fatal(err)
	}
	_ = af.Close()

	cache3, err := transcript.ParseFileIncremental(f.Name(), cache2)
	if err != nil {
		t.Fatalf("third call failed: %v", err)
	}
	turns3 := cache3.Turns()
	// Should have 3 turns: user, flushed assistant (with tool result), new pending assistant
	if len(turns3) != 3 {
		t.Fatalf("phase 3: expected 3 turns, got %d", len(turns3))
	}
	// First assistant turn should now have tool result matched
	if turns3[1].ToolCalls[0].Result == nil {
		t.Error("phase 3: tool result should be matched on flushed turn")
	}
	if turns3[2].Text != "The file contains a main package." {
		t.Errorf("phase 3: unexpected new assistant text: %q", turns3[2].Text)
	}
	if cache3.Offset() <= prevOffset {
		t.Error("phase 3: offset should have increased")
	}
}

func TestParseAggregatesIncremental_Slug(t *testing.T) {
	f, err := os.CreateTemp("", "slug-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","slug":"fizzy-napping-stallman","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}` + "\n" +
		`{"type":"assistant","timestamp":"2025-01-01T10:00:01Z","slug":"fizzy-napping-stallman","message":{"role":"assistant","content":[{"type":"text","text":"Hi"}],"model":"claude-sonnet-4-6","usage":{"input_tokens":100,"output_tokens":20}}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	agg, err := transcript.ParseAggregatesIncremental(f.Name(), nil)
	if err != nil {
		t.Fatalf("ParseAggregatesIncremental failed: %v", err)
	}
	if agg.Slug != "fizzy-napping-stallman" {
		t.Errorf("expected slug %q, got %q", "fizzy-napping-stallman", agg.Slug)
	}
}

func TestParseAggregatesIncremental_SlugEmpty(t *testing.T) {
	f, err := os.CreateTemp("", "no-slug-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

	content := `{"type":"user","timestamp":"2025-01-01T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}` + "\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	agg, err := transcript.ParseAggregatesIncremental(f.Name(), nil)
	if err != nil {
		t.Fatalf("ParseAggregatesIncremental failed: %v", err)
	}
	if agg.Slug != "" {
		t.Errorf("expected empty slug, got %q", agg.Slug)
	}
}

func TestParseConsecutiveAssistantEntriesWithCache(t *testing.T) {
	// Two consecutive assistant entries both with cache tokens
	// Should accumulate each field separately so NewInputTokens()
	// is not double-counted at flush time.
	const input = `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"hi"}],"model":"claude-opus-4-6","usage":{"input_tokens":100,"cache_creation_input_tokens":200,"cache_read_input_tokens":300,"output_tokens":50}},"uuid":"a1","parentUuid":""}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"there"}],"model":"claude-opus-4-6","usage":{"input_tokens":50,"cache_creation_input_tokens":100,"cache_read_input_tokens":150,"output_tokens":30}},"uuid":"a2","parentUuid":"a1"}
{"type":"user","message":{"role":"user","content":[]},"uuid":"u1","parentUuid":"a2"}
`
	result, err := transcript.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(result.Turns) != 1 {
		t.Fatalf("expected 1 merged turn, got %d", len(result.Turns))
	}
	u := result.Turns[0].Usage
	// Each field should be individually accumulated:
	// InputTokens = 100 + 50 = 150
	if u.InputTokens != 150 {
		t.Errorf("merged InputTokens = %d, want 150", u.InputTokens)
	}
	// CacheCreationInputTokens = 200 + 100 = 300
	if u.CacheCreationInputTokens != 300 {
		t.Errorf("merged CacheCreationInputTokens = %d, want 300", u.CacheCreationInputTokens)
	}
	// NewInputTokens = (100+200) + (50+100) = 450
	if got := u.NewInputTokens(); got != 450 {
		t.Errorf("merged NewInputTokens() = %d, want 450", got)
	}
	// CacheReadInputTokens should be 300+150=450
	if u.CacheReadInputTokens != 450 {
		t.Errorf("merged CacheReadInputTokens = %d, want 450", u.CacheReadInputTokens)
	}
	// TokensByModel should reflect the correct NewInputTokens total
	usage := result.TokensByModel["claude-opus-4-6"]
	if usage.InputTokens != 450 {
		t.Errorf("TokensByModel InputTokens = %d, want 450", usage.InputTokens)
	}
	if usage.CacheReadInputTokens != 450 {
		t.Errorf("TokensByModel CacheReadInputTokens = %d, want 450", usage.CacheReadInputTokens)
	}
}

func TestCountSubagents(t *testing.T) {
	// Empty directory
	dir, err := os.MkdirTemp("", "subagents-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	if got := transcript.CountSubagents(dir); got != 0 {
		t.Errorf("empty dir: expected 0, got %d", got)
	}

	// Non-existent directory
	if got := transcript.CountSubagents("/nonexistent/subagents"); got != 0 {
		t.Errorf("non-existent dir: expected 0, got %d", got)
	}

	// Empty string
	if got := transcript.CountSubagents(""); got != 0 {
		t.Errorf("empty string: expected 0, got %d", got)
	}

	// Mixed files
	for _, name := range []string{"agent1.jsonl", "agent2.jsonl", "notes.txt", "readme.md"} {
		f, err := os.Create(dir + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}
	if got := transcript.CountSubagents(dir); got != 2 {
		t.Errorf("mixed files: expected 2 jsonl files, got %d", got)
	}
}
