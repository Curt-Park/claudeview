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
	s := &model.Session{
		TokensByModel: map[string]model.TokenCount{
			"claude-sonnet-4-6": {InputTokens: 50000, OutputTokens: 12500},
		},
	}
	got := s.TokenString()
	if got != "sonnet:62k" {
		t.Errorf("TokenString() = %q, want %q", got, "sonnet:62k")
	}
}

func TestSessionTokenStringEmpty(t *testing.T) {
	s := &model.Session{}
	if got := s.TokenString(); got != "-" {
		t.Errorf("TokenString() empty = %q, want %q", got, "-")
	}
}

func TestSessionTokenStringMultiModel(t *testing.T) {
	s := &model.Session{
		TokensByModel: map[string]model.TokenCount{
			"claude-opus-4-6":   {InputTokens: 1000000, OutputTokens: 500000},
			"claude-sonnet-4-6": {InputTokens: 50000, OutputTokens: 12500},
		},
	}
	got := s.TokenString()
	// Alphabetically: opus before sonnet
	if got != "opus:1.5M sonnet:62k" {
		t.Errorf("TokenString() = %q, want %q", got, "opus:1.5M sonnet:62k")
	}
}

func TestSessionLastActive(t *testing.T) {
	s := &model.Session{ModTime: time.Now().Add(-5 * time.Minute)}
	got := s.LastActive()
	if got == "" {
		t.Error("LastActive() returned empty string")
	}
	if got != "5m" {
		t.Errorf("LastActive() = %q, expected ~5m", got)
	}
}

func TestSessionTopicShort(t *testing.T) {
	tests := []struct {
		name   string
		topic  string
		maxLen int
		want   string
	}{
		{"empty topic", "", 20, "-"},
		{"short topic", "Hello world", 20, "Hello world"},
		{"multiline topic", "First line\nSecond line", 20, "First line"},
		{"truncated topic", "This is a very long topic text", 15, "This is a very…"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &model.Session{Topic: tc.topic}
			if got := s.TopicShort(tc.maxLen); got != tc.want {
				t.Errorf("TopicShort(%d) = %q, want %q", tc.maxLen, got, tc.want)
			}
		})
	}
}
