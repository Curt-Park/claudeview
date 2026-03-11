package model

import "time"

// Turn represents a single conversation turn (user or assistant).
type Turn struct {
	Role         string // "user" or "assistant"
	Text         string
	Thinking     string
	ToolCalls    []*ToolCall
	ModelName    string
	InputTokens      int
	CacheReadTokens  int
	OutputTokens     int
	Timestamp    time.Time
}
