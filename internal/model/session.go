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
		parts = append(parts, fmt.Sprintf("%s:%s", shortModelName(m), formatTokens(total)))
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

// shortModelName extracts a short identifier from a model name.
func shortModelName(model string) string {
	lower := strings.ToLower(model)
	switch {
	case strings.Contains(lower, "opus"):
		return "opus"
	case strings.Contains(lower, "sonnet"):
		return "sonnet"
	case strings.Contains(lower, "haiku"):
		return "haiku"
	default:
		parts := strings.Split(model, "-")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return model
	}
}

// formatTokens formats a token count as "125k", "1.5M", etc.
func formatTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1000:
		return fmt.Sprintf("%dk", n/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
