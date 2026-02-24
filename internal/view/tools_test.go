package view_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/view"
)

func TestToolCallDetailLines_Full(t *testing.T) {
	tc := &model.ToolCall{
		ID:        "tool-1",
		AgentID:   "agent-abc",
		Name:      "Read",
		Input:     []byte(`{"file_path":"src/main.go"}`),
		Result:    []byte(`"142 lines"`),
		IsError:   false,
		Timestamp: time.Date(2025, 1, 1, 17, 42, 16, 0, time.UTC),
		Duration:  300 * time.Millisecond,
	}

	lines := view.ToolCallDetailLines(tc)
	joined := strings.Join(lines, "\n")

	for _, want := range []string{"Tool:", "Read", "Agent:", "Time:", "Input:", "Output:"} {
		if !strings.Contains(joined, want) {
			t.Errorf("detail lines missing %q\nGot:\n%s", want, joined)
		}
	}
}

func TestToolCallDetailLines_Error(t *testing.T) {
	tc := &model.ToolCall{
		Name:    "Bash",
		IsError: true,
		Input:   []byte(`{"command":"bad cmd"}`),
	}

	lines := view.ToolCallDetailLines(tc)
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "[ERROR]") {
		t.Errorf("expected [ERROR] in detail lines\nGot:\n%s", joined)
	}
}

func TestToolCallDetailLines_NilFields(t *testing.T) {
	tc := &model.ToolCall{
		Name: "Read",
	}
	// Should not panic with nil Input/Result and zero Timestamp
	lines := view.ToolCallDetailLines(tc)
	if len(lines) == 0 {
		t.Error("expected non-empty detail lines")
	}
}
