package transcript

import "encoding/json"

// entry represents a single JSONL line in a transcript file.
type entry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	UUID      string          `json:"uuid"`
	SessionID string          `json:"sessionId"`
	GitBranch string          `json:"gitBranch"`
	Message   json.RawMessage `json:"message"`
}

// messageContent is a polymorphic content block inside a message.
type messageContent struct {
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

// assistantMessage is the message field when type=="assistant".
type assistantMessage struct {
	Role    string           `json:"role"`
	Content []messageContent `json:"content"`
	Model   string           `json:"model"`
	Usage   Usage            `json:"usage"`
}

// userMessage is the message field when type=="user".
// Content can be either a JSON array of messageContent blocks (old format)
// or a plain JSON string (new format).
type userMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// textContent returns all plain-text strings from the content field,
// handling both the array-of-blocks format and the plain-string format.
func (m *userMessage) textContent() string {
	if len(m.Content) == 0 {
		return ""
	}
	// Try plain string first
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return s
	}
	// Fall back to array of content blocks
	var blocks []messageContent
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

// toolResults returns tool_result blocks from the content array.
// Returns nil if content is a plain string.
func (m *userMessage) toolResults() []messageContent {
	if len(m.Content) == 0 {
		return nil
	}
	var blocks []messageContent
	if err := json.Unmarshal(m.Content, &blocks); err != nil {
		return nil
	}
	var results []messageContent
	for _, c := range blocks {
		if c.Type == "tool_result" {
			results = append(results, c)
		}
	}
	return results
}
