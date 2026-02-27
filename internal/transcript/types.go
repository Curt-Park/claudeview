package transcript

import "encoding/json"

// Entry represents a single JSONL line in a transcript file.
type Entry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	UUID      string          `json:"uuid"`
	SessionID string          `json:"sessionId"`
	GitBranch string          `json:"gitBranch"`
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
// Content can be either a JSON array of MessageContent blocks (old format)
// or a plain JSON string (new format).
type UserMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// TextContent returns all plain-text strings from the content field,
// handling both the array-of-blocks format and the plain-string format.
func (m *UserMessage) TextContent() string {
	if len(m.Content) == 0 {
		return ""
	}
	// Try plain string first
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return s
	}
	// Fall back to array of content blocks
	var blocks []MessageContent
	if err := json.Unmarshal(m.Content, &blocks); err != nil {
		return ""
	}
	var text string
	for _, c := range blocks {
		if c.Type == "text" {
			text += c.Text
		}
	}
	return text
}

// ToolResults returns tool_result blocks from the content array.
// Returns nil if content is a plain string.
func (m *UserMessage) ToolResults() []MessageContent {
	if len(m.Content) == 0 {
		return nil
	}
	var blocks []MessageContent
	if err := json.Unmarshal(m.Content, &blocks); err != nil {
		return nil
	}
	var results []MessageContent
	for _, c := range blocks {
		if c.Type == "tool_result" {
			results = append(results, c)
		}
	}
	return results
}
