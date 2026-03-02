package model_test

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestTurn_HasExpectedFields(t *testing.T) {
	turn := model.Turn{
		Role:         "assistant",
		Text:         "hello",
		Thinking:     "hmm",
		ModelName:    "claude-sonnet",
		InputTokens:  100,
		OutputTokens: 50,
		Timestamp:    time.Now(),
	}
	if turn.Role != "assistant" {
		t.Errorf("expected Role=assistant, got %q", turn.Role)
	}
	if turn.InputTokens != 100 {
		t.Errorf("expected InputTokens=100, got %d", turn.InputTokens)
	}
}
