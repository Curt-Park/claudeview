package model

import (
	"fmt"
	"time"
)

// Status represents the current state of a session/agent.
type Status string

const (
	StatusActive    Status = "active"
	StatusThinking  Status = "thinking"
	StatusReading   Status = "reading"
	StatusExecuting Status = "executing"
	StatusDone      Status = "done"
	StatusEnded     Status = "ended"
	StatusError     Status = "error"
	StatusFailed    Status = "failed"
	StatusRunning   Status = "running"
	StatusPending   Status = "pending"
)

// Session represents a Claude Code session.
type Session struct {
	ID           string
	ProjectHash  string
	FilePath     string
	SubagentDir  string
	Model        string
	Status       Status
	Agents       []*Agent
	TotalCost    float64
	InputTokens  int
	OutputTokens int
	NumTurns     int
	DurationMS   int64
	StartTime    time.Time
	EndTime      time.Time
	ModTime      time.Time
}

// Age returns a human-friendly elapsed time string.
func (s *Session) Age() string {
	t := s.ModTime
	if !s.EndTime.IsZero() {
		t = s.EndTime
	}
	return FormatAge(time.Since(t))
}

// TotalTokens returns combined input+output token count.
func (s *Session) TotalTokens() int {
	return s.InputTokens + s.OutputTokens
}

// CostString returns formatted cost.
func (s *Session) CostString() string {
	if s.TotalCost == 0 {
		return "-"
	}
	return fmt.Sprintf("$%.4f", s.TotalCost)
}

// TokenString returns formatted token count.
func (s *Session) TokenString() string {
	t := s.TotalTokens()
	if t == 0 {
		return "-"
	}
	if t >= 1000 {
		return fmt.Sprintf("%.1fk", float64(t)/1000)
	}
	return fmt.Sprintf("%d", t)
}

// ShortID returns the first 8 chars of the session ID.
func (s *Session) ShortID() string {
	if len(s.ID) > 8 {
		return s.ID[:8]
	}
	return s.ID
}

// ToolCount returns total tool calls across all agents.
func (s *Session) ToolCount() int {
	count := 0
	for _, a := range s.Agents {
		count += len(a.ToolCalls)
	}
	return count
}

// FormatAge converts a duration into a human-friendly string (e.g. "5m", "2h", "3d").
func FormatAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
