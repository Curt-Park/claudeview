package transcript_test

import (
	"os"
	"testing"

	"github.com/Curt-Park/claudeview/internal/transcript"
)

func TestParseFile(t *testing.T) {
	result, err := transcript.ParseFile("../../testdata/sample_transcript.jsonl")
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
		if usage.InputTokens != 600 {
			t.Errorf("expected InputTokens=600, got %d", usage.InputTokens)
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

func TestTopicUsesLastUserMessage(t *testing.T) {
	f, err := os.CreateTemp("", "lasttopic-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(f.Name()) }()

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
	want := "Now refactor the token refresh logic"
	if result.Topic != want {
		t.Errorf("expected last user message as topic, got %q", result.Topic)
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
