package transcript

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"time"
)

// ToolCall represents a matched tool_use + tool_result pair.
type ToolCall struct {
	ID        string
	Name      string
	Input     json.RawMessage
	Result    json.RawMessage
	IsError   bool
	Timestamp time.Time
	Duration  time.Duration
}

// Turn is a parsed conversation turn with tool calls extracted.
type Turn struct {
	Role      string
	Text      string
	Thinking  string
	ToolCalls []ToolCall
	Model     string
	Usage     Usage
	Timestamp time.Time
}

// ParsedTranscript is the result of parsing a JSONL file.
type ParsedTranscript struct {
	Turns      []Turn
	TotalCost  float64
	DurationMS int64
	NumTurns   int
}

// ParseFile reads and parses a JSONL transcript file.
func ParseFile(path string) (*ParsedTranscript, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return Parse(f)
}

// Parse reads from an io.Reader and parses JSONL transcript entries.
func Parse(r io.Reader) (*ParsedTranscript, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB buffer

	result := &ParsedTranscript{}
	// toolResults maps tool_use_id -> tool_result content
	toolResults := map[string]json.RawMessage{}
	toolErrors := map[string]bool{}

	var pendingAssistantTurn *Turn

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)

		switch entry.Type {
		case "user":
			var msg UserMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				continue
			}
			// Collect tool results to match with pending tool calls
			for _, c := range msg.Content {
				if c.Type == "tool_result" {
					toolResults[c.ToolUseID] = c.Content
					toolErrors[c.ToolUseID] = c.IsError
				}
			}
			// Flush pending assistant turn with matched results
			if pendingAssistantTurn != nil {
				for i := range pendingAssistantTurn.ToolCalls {
					tc := &pendingAssistantTurn.ToolCalls[i]
					if res, ok := toolResults[tc.ID]; ok {
						tc.Result = res
						tc.IsError = toolErrors[tc.ID]
					}
				}
				result.Turns = append(result.Turns, *pendingAssistantTurn)
				pendingAssistantTurn = nil
			}
			// Add user text turns
			turn := Turn{Role: "user", Timestamp: ts}
			for _, c := range msg.Content {
				if c.Type == "text" {
					turn.Text += c.Text
				}
			}
			if turn.Text != "" {
				result.Turns = append(result.Turns, turn)
			}

		case "assistant":
			var msg AssistantMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				continue
			}
			turn := Turn{
				Role:      "assistant",
				Model:     msg.Model,
				Usage:     msg.Usage,
				Timestamp: ts,
			}
			for _, c := range msg.Content {
				switch c.Type {
				case "text":
					turn.Text += c.Text
				case "thinking":
					turn.Thinking += c.Thinking
				case "tool_use":
					turn.ToolCalls = append(turn.ToolCalls, ToolCall{
						ID:        c.ID,
						Name:      c.Name,
						Input:     c.Input,
						Timestamp: ts,
					})
				}
			}
			pendingAssistantTurn = &turn

		case "system":
			var msg SystemMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				continue
			}
			if msg.Subtype == "turn_duration" {
				result.TotalCost = msg.TotalCostUSD
				result.DurationMS += msg.DurationMS
				result.NumTurns = msg.NumTurns
			}
		}
	}

	// Flush any remaining pending turn
	if pendingAssistantTurn != nil {
		for i := range pendingAssistantTurn.ToolCalls {
			tc := &pendingAssistantTurn.ToolCalls[i]
			if res, ok := toolResults[tc.ID]; ok {
				tc.Result = res
				tc.IsError = toolErrors[tc.ID]
			}
		}
		result.Turns = append(result.Turns, *pendingAssistantTurn)
	}

	return result, scanner.Err()
}
