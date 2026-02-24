package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestToolCallInputSummary(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    json.RawMessage
		wantSub  string // substring that must appear in result
	}{
		{
			name:     "Read tool uses file_path",
			toolName: "Read",
			input:    mustJSON(map[string]any{"file_path": "/path/to/file.go"}),
			wantSub:  "/path/to/file.go",
		},
		{
			name:     "Grep tool uses pattern and path",
			toolName: "Grep",
			input:    mustJSON(map[string]any{"pattern": "func main", "path": "./cmd"}),
			wantSub:  "func main",
		},
		{
			name:     "Bash tool uses command",
			toolName: "Bash",
			input:    mustJSON(map[string]any{"command": "go test ./..."}),
			wantSub:  "go test",
		},
		{
			name:     "Edit tool uses file_path",
			toolName: "Edit",
			input:    mustJSON(map[string]any{"file_path": "/src/main.go"}),
			wantSub:  "/src/main.go",
		},
		{
			name:     "Glob tool uses pattern",
			toolName: "Glob",
			input:    mustJSON(map[string]any{"pattern": "**/*.go"}),
			wantSub:  "**/*.go",
		},
		{
			name:     "Task tool uses description",
			toolName: "Task",
			input:    mustJSON(map[string]any{"description": "analyze the codebase"}),
			wantSub:  "analyze",
		},
		{
			name:     "WebFetch tool uses url",
			toolName: "WebFetch",
			input:    mustJSON(map[string]any{"url": "https://example.com"}),
			wantSub:  "https://example.com",
		},
		{
			name:     "nil input returns empty",
			toolName: "Read",
			input:    nil,
			wantSub:  "",
		},
		{
			name:     "unknown tool falls back to first string value",
			toolName: "UnknownTool",
			input:    mustJSON(map[string]any{"some_field": "some_value"}),
			wantSub:  "some_value",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			call := &model.ToolCall{Name: tc.toolName, Input: tc.input}
			got := call.InputSummary()
			if tc.wantSub == "" {
				if got != "" {
					t.Errorf("InputSummary() = %q, want empty", got)
				}
				return
			}
			if len(got) == 0 {
				t.Errorf("InputSummary() returned empty, want substring %q", tc.wantSub)
				return
			}
			found := false
			for i := 0; i <= len(got)-len(tc.wantSub); i++ {
				if got[i:i+len(tc.wantSub)] == tc.wantSub {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("InputSummary() = %q, want it to contain %q", got, tc.wantSub)
			}
		})
	}
}

func TestToolCallResultSummary(t *testing.T) {
	t.Run("nil result returns dash", func(t *testing.T) {
		call := &model.ToolCall{Result: nil}
		if got := call.ResultSummary(); got != "-" {
			t.Errorf("ResultSummary() = %q, want %q", got, "-")
		}
	})

	t.Run("IsError returns error", func(t *testing.T) {
		call := &model.ToolCall{IsError: true, Result: mustJSON("something failed")}
		if got := call.ResultSummary(); got != "error" {
			t.Errorf("ResultSummary() = %q, want %q", got, "error")
		}
	})

	t.Run("string result shows line count", func(t *testing.T) {
		call := &model.ToolCall{Result: mustJSON("line1\nline2\nline3")}
		got := call.ResultSummary()
		if got != "3 lines" {
			t.Errorf("ResultSummary() = %q, want %q", got, "3 lines")
		}
	})

	t.Run("array result shows block count", func(t *testing.T) {
		arr := []map[string]any{{"type": "text"}, {"type": "text"}}
		call := &model.ToolCall{Result: mustJSON(arr)}
		got := call.ResultSummary()
		if got != "2 blocks" {
			t.Errorf("ResultSummary() = %q, want %q", got, "2 blocks")
		}
	})

	t.Run("string result truncation at 20 chars", func(t *testing.T) {
		// A string with many lines so "N lines" would exceed 20 chars
		call := &model.ToolCall{Result: mustJSON("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm")}
		got := call.ResultSummary()
		if len(got) > 20 {
			t.Errorf("ResultSummary() = %q (len %d), want len <= 20", got, len(got))
		}
	})
}

func TestToolCallDurationString(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero duration returns dash", 0, "-"},
		{"500ms", 500 * time.Millisecond, "500.0ms"},
		{"2.5s", 2500 * time.Millisecond, "2.5s"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			call := &model.ToolCall{Duration: tc.duration}
			if got := call.DurationString(); got != tc.want {
				t.Errorf("DurationString() = %q, want %q", got, tc.want)
			}
		})
	}
}
