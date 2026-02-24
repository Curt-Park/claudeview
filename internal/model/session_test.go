package model_test

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestSessionShortID(t *testing.T) {
	s := &model.Session{ID: "abc12345-def6-7890-ghij-klmnopqrstuv"}
	if got := s.ShortID(); got != "abc12345" {
		t.Errorf("ShortID() = %q, want %q", got, "abc12345")
	}
}

func TestSessionTokenString(t *testing.T) {
	s := &model.Session{InputTokens: 50000, OutputTokens: 12500}
	got := s.TokenString()
	if got != "62.5k" {
		t.Errorf("TokenString() = %q, want %q", got, "62.5k")
	}
}

func TestSessionCostString(t *testing.T) {
	s := &model.Session{TotalCost: 0.4234}
	got := s.CostString()
	if got != "$0.4234" {
		t.Errorf("CostString() = %q, want %q", got, "$0.4234")
	}
}

func TestSessionAge(t *testing.T) {
	s := &model.Session{ModTime: time.Now().Add(-5 * time.Minute)}
	age := s.Age()
	if age == "" {
		t.Error("Age() returned empty string")
	}
	// Should be "5m" approximately
	if age != "5m" {
		t.Logf("Age() = %q (expected ~5m)", age)
	}
}

func TestSessionToolCount(t *testing.T) {
	s := &model.Session{
		Agents: []*model.Agent{
			{ToolCalls: []*model.ToolCall{{}, {}, {}}},
			{ToolCalls: []*model.ToolCall{{}}},
		},
	}
	if got := s.ToolCount(); got != 4 {
		t.Errorf("ToolCount() = %d, want 4", got)
	}
}
