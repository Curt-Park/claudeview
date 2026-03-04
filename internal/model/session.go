package model

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TokenCount holds per-model token usage.
type TokenCount struct {
	InputTokens  int
	OutputTokens int
}

// Session represents a Claude Code session.
type Session struct {
	ID            string
	ProjectHash   string
	FilePath      string
	SubagentDir   string
	Branch        string
	Slug          string
	FileSize      int64
	Topic         string
	TokensByModel map[string]TokenCount
	AgentCount    int
	ToolCallCount int
	Agents        []*Agent
	NumTurns      int
	StartTime     time.Time
	EndTime       time.Time
	ModTime       time.Time

	// GroupSessions holds all sessions in the slug group (oldest-first).
	// nil for solo sessions (no slug or single-member group).
	GroupSessions []*Session
}

// IsGroupRepresentative returns true if this session represents a collapsed slug group.
func (s *Session) IsGroupRepresentative() bool {
	return len(s.GroupSessions) > 1
}

// GroupNameCell returns the display name for the NAME column.
// For groups: "d2559feb..360eb907"; for solo sessions: "d2559feb".
func (s *Session) GroupNameCell() string {
	if !s.IsGroupRepresentative() {
		return s.ShortID()
	}
	first := s.GroupSessions[0].ShortID()
	last := s.GroupSessions[len(s.GroupSessions)-1].ShortID()
	return first + ".." + last
}

// LastActive returns a human-friendly elapsed time string based on ModTime.
func (s *Session) LastActive() string {
	return FormatAge(time.Since(s.ModTime))
}

// TokenString returns a compact per-model token string (e.g. "opus:125k sonnet:50k").
func (s *Session) TokenString() string {
	if len(s.TokensByModel) == 0 {
		return "-"
	}
	models := make([]string, 0, len(s.TokensByModel))
	for m := range s.TokensByModel {
		models = append(models, m)
	}
	sort.Strings(models)

	var parts []string
	for _, m := range models {
		tc := s.TokensByModel[m]
		total := tc.InputTokens + tc.OutputTokens
		parts = append(parts, fmt.Sprintf("%s:%s", ShortModelName(m), FormatTokenCount(total)))
	}
	return strings.Join(parts, " ")
}

// TopicShort returns a normalized, truncated topic string.
// Newlines are replaced with spaces (matching claude -r style) so the full
// content is visible on a single line.
func (s *Session) TopicShort(maxLen int) string {
	if s.Topic == "" {
		return "-"
	}
	topic := strings.ReplaceAll(s.Topic, "\n", " ")
	runes := []rune(topic)
	if len(runes) > maxLen {
		return string(runes[:maxLen-1]) + "…"
	}
	return topic
}

// MetaLine returns a compact metadata string: "branch · size".
func (s *Session) MetaLine() string {
	size := FormatSize(s.FileSize)
	if s.Branch == "" {
		return size
	}
	return s.Branch + " · " + size
}

// ShortID returns the first 8 chars of the session ID.
func (s *Session) ShortID() string {
	if len(s.ID) > 8 {
		return s.ID[:8]
	}
	return s.ID
}
