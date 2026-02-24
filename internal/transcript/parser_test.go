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
}

func TestParseEmpty(t *testing.T) {
	f, err := os.CreateTemp("", "empty-*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

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
