package transcript

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
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
	Turns          []Turn
	Topic          string
	TokensByModel  map[string]Usage
	TotalToolCalls int
	TotalCost      float64
	DurationMS     int64
	NumTurns       int
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

// flushPendingTurn matches tool results into the pending turn, accumulates metrics, and appends it.
func flushPendingTurn(result *ParsedTranscript, turn *Turn, toolResults map[string]json.RawMessage, toolErrors map[string]bool) {
	for i := range turn.ToolCalls {
		tc := &turn.ToolCalls[i]
		if res, ok := toolResults[tc.ID]; ok {
			tc.Result = res
			tc.IsError = toolErrors[tc.ID]
		}
	}
	// Accumulate per-model token usage
	u := result.TokensByModel[turn.Model]
	u.InputTokens += turn.Usage.InputTokens
	u.OutputTokens += turn.Usage.OutputTokens
	result.TokensByModel[turn.Model] = u
	// Count tool calls
	result.TotalToolCalls += len(turn.ToolCalls)
	result.Turns = append(result.Turns, *turn)
}

// Parse reads from an io.Reader and parses JSONL transcript entries.
func Parse(r io.Reader) (*ParsedTranscript, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB buffer

	result := &ParsedTranscript{
		TokensByModel: make(map[string]Usage),
	}
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
			for _, c := range msg.ToolResults() {
				toolResults[c.ToolUseID] = c.Content
				toolErrors[c.ToolUseID] = c.IsError
			}
			// Flush pending assistant turn with matched results
			if pendingAssistantTurn != nil {
				flushPendingTurn(result, pendingAssistantTurn, toolResults, toolErrors)
				pendingAssistantTurn = nil
			}
			// Add user text turns
			turn := Turn{Role: "user", Timestamp: ts}
			turn.Text = msg.TextContent()
			if turn.Text != "" {
				// Set topic from first real user message, matching claude -r display (skip skill prefix lines)
				if result.Topic == "" && !strings.HasPrefix(turn.Text, "Base directory for this skill:") {
					result.Topic = turn.Text
				}
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
			// Two formats in the wild:
			//   Old: system fields inside a "message" object, duration_ms snake_case.
			//   New: system fields at top level, durationMs camelCase (no cost/turns).
			if len(entry.Message) > 0 {
				// Old format
				var msg struct {
					Subtype      string  `json:"subtype"`
					DurationMS   int64   `json:"duration_ms"`
					NumTurns     int     `json:"num_turns"`
					TotalCostUSD float64 `json:"total_cost_usd"`
				}
				if err := json.Unmarshal(entry.Message, &msg); err == nil && msg.Subtype == "turn_duration" {
					result.TotalCost = msg.TotalCostUSD
					result.DurationMS += msg.DurationMS
					result.NumTurns = msg.NumTurns
				}
			} else {
				// New format: fields at top level
				var msg struct {
					Subtype    string `json:"subtype"`
					DurationMS int64  `json:"durationMs"`
				}
				if err := json.Unmarshal(line, &msg); err == nil && msg.Subtype == "turn_duration" {
					result.DurationMS += msg.DurationMS
				}
			}
		}
	}

	// Flush any remaining pending turn
	if pendingAssistantTurn != nil {
		flushPendingTurn(result, pendingAssistantTurn, toolResults, toolErrors)
	}

	return result, scanner.Err()
}

// SessionAggregates holds cached session-level metrics for incremental parsing.
type SessionAggregates struct {
	Topic          string
	Branch         string
	TokensByModel  map[string]Usage
	TotalToolCalls int
	DurationMS     int64
	NumTurns       int
	Offset         int64 // next read start position
}

// ParseAggregatesIncremental reads a JSONL file from the stored offset,
// accumulates metrics into agg, and returns the updated aggregates.
// If agg is nil, a new SessionAggregates is created (reading from offset 0).
func ParseAggregatesIncremental(path string, agg *SessionAggregates) (*SessionAggregates, error) {
	if agg == nil {
		agg = &SessionAggregates{
			TokensByModel: make(map[string]Usage),
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return agg, err
	}
	defer func() { _ = f.Close() }()

	if agg.Offset > 0 {
		if _, err := f.Seek(agg.Offset, io.SeekStart); err != nil {
			return agg, err
		}
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		// Capture git branch from first entry that has it
		if agg.Branch == "" && entry.GitBranch != "" {
			agg.Branch = entry.GitBranch
		}

		switch entry.Type {
		case "user":
			var msg UserMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				break
			}
			// Set topic from first real user message, matching claude -r display (skip skill prefix lines)
			if agg.Topic == "" {
				if text := msg.TextContent(); text != "" {
					if !strings.HasPrefix(text, "Base directory for this skill:") {
						agg.Topic = text
					}
				}
			}

		case "assistant":
			var msg AssistantMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				break
			}
			u := agg.TokensByModel[msg.Model]
			u.InputTokens += msg.Usage.InputTokens
			u.OutputTokens += msg.Usage.OutputTokens
			agg.TokensByModel[msg.Model] = u
			for _, c := range msg.Content {
				if c.Type == "tool_use" {
					agg.TotalToolCalls++
				}
			}

		case "system":
			if len(entry.Message) > 0 {
				var msg struct {
					Subtype    string `json:"subtype"`
					DurationMS int64  `json:"duration_ms"`
					NumTurns   int    `json:"num_turns"`
				}
				if err := json.Unmarshal(entry.Message, &msg); err == nil && msg.Subtype == "turn_duration" {
					agg.DurationMS += msg.DurationMS
					agg.NumTurns = msg.NumTurns
				}
			} else {
				var msg struct {
					Subtype    string `json:"subtype"`
					DurationMS int64  `json:"durationMs"`
				}
				if err := json.Unmarshal(line, &msg); err == nil && msg.Subtype == "turn_duration" {
					agg.DurationMS += msg.DurationMS
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return agg, err
	}

	// Update offset to end of file
	if pos, err := f.Seek(0, io.SeekEnd); err == nil {
		agg.Offset = pos
	}

	return agg, nil
}
