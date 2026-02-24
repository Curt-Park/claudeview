package transcript

import "encoding/json"

// Entry represents a single JSONL line in a transcript file.
type Entry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	UUID      string          `json:"uuid"`
	SessionID string          `json:"sessionId"`
	Message   json.RawMessage `json:"message"`
}

// MessageContent is a polymorphic content block inside a message.
type MessageContent struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
	// tool_result fields
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	// thinking
	Thinking string `json:"thinking,omitempty"`
}

// Usage tracks token consumption.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// AssistantMessage is the message field when type=="assistant".
type AssistantMessage struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
	Model   string           `json:"model"`
	Usage   Usage            `json:"usage"`
}

// UserMessage is the message field when type=="user".
type UserMessage struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

// SystemMessage is the message field when type=="system".
type SystemMessage struct {
	Subtype       string  `json:"subtype"`
	DurationMS    int64   `json:"duration_ms"`
	DurationAPIMS int64   `json:"duration_api_ms"`
	NumTurns      int     `json:"num_turns"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
}

// ProgressMessage is the message field when type=="progress".
type ProgressMessage struct {
	Subtype   string          `json:"subtype"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
}
