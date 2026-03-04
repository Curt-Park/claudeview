package model_test

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

// TestTurnFields verifies all Turn fields are present and assignable.
// There are no methods to test — this is a compile-time structural check.
func TestTurnFields(t *testing.T) {
	_ = model.Turn{
		Role:         "assistant",
		Text:         "hello",
		Thinking:     "hmm",
		ToolCalls:    []*model.ToolCall{},
		ModelName:    "claude-sonnet",
		InputTokens:  100,
		OutputTokens: 50,
		Timestamp:    time.Now(),
	}
}
