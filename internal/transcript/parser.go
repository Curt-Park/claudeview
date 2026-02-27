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

// newJSONLScanner creates a bufio.Scanner with a 10MB buffer for JSONL parsing.
func newJSONLScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 1<<20), 10<<20)
	return s
}

// parseSystemDuration extracts duration, turn count, and cost from a system entry.
// It handles both the old format (fields in Message) and new format (fields at top level).
func parseSystemDuration(message json.RawMessage, rawLine []byte) (durationMS int64, numTurns int, costUSD float64) {
	if len(message) > 0 {
		var msg struct {
			Subtype      string  `json:"subtype"`
			DurationMS   int64   `json:"duration_ms"`
			NumTurns     int     `json:"num_turns"`
			TotalCostUSD float64 `json:"total_cost_usd"`
		}
		if err := json.Unmarshal(message, &msg); err == nil && msg.Subtype == "turn_duration" {
			return msg.DurationMS, msg.NumTurns, msg.TotalCostUSD
		}
	} else {
		var msg struct {
			Subtype    string `json:"subtype"`
			DurationMS int64  `json:"durationMs"`
		}
		if err := json.Unmarshal(rawLine, &msg); err == nil && msg.Subtype == "turn_duration" {
			return msg.DurationMS, 0, 0
		}
	}
	return 0, 0, 0
}

// Parse reads from an io.Reader and parses JSONL transcript entries.
func Parse(r io.Reader) (*ParsedTranscript, error) {
	scanner := newJSONLScanner(r)

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

		var entry entry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, entry.Timestamp)

		switch entry.Type {
		case "user":
			var msg userMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				continue
			}
			// Collect tool results to match with pending tool calls
			for _, c := range msg.toolResults() {
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
			turn.Text = msg.textContent()
			if turn.Text != "" {
				// Set topic from first real user message, matching claude -r display (skip skill prefix lines)
				if result.Topic == "" && !strings.HasPrefix(turn.Text, "Base directory for this skill:") {
					result.Topic = extractTopic(turn.Text)
				}
				result.Turns = append(result.Turns, turn)
			}

		case "assistant":
			var msg assistantMessage
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
			// Two formats: old (fields in Message, snake_case) and new (top-level, camelCase).
			dur, turns, cost := parseSystemDuration(entry.Message, line)
			result.DurationMS += dur
			if turns > 0 {
				result.NumTurns = turns
			}
			if cost > 0 {
				result.TotalCost = cost
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

	scanner := newJSONLScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry entry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		// Capture git branch from first entry that has it
		if agg.Branch == "" && entry.GitBranch != "" {
			agg.Branch = entry.GitBranch
		}

		switch entry.Type {
		case "user":
			var msg userMessage
			if err := json.Unmarshal(entry.Message, &msg); err != nil {
				break
			}
			// Set topic from first real user message, matching claude -r display (skip skill prefix lines)
			if agg.Topic == "" {
				if text := msg.textContent(); text != "" {
					if !strings.HasPrefix(text, "Base directory for this skill:") {
						agg.Topic = extractTopic(text)
					}
				}
			}

		case "assistant":
			var msg assistantMessage
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
			agg.NumTurns++

		case "system":
			dur, turns, _ := parseSystemDuration(entry.Message, line)
			agg.DurationMS += dur
			if turns > 0 {
				agg.NumTurns = turns
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

// extractTopic normalizes raw user message text for use as a session topic.
//
// Claude Code wraps slash-command invocations in XML tags. Known formats:
//
//   - <local-command-caveat>…</local-command-caveat>           (caveat only, no command yet)
//   - <command-name>/skill</command-name>…                     (command in next turn)
//   - <command-message>skill</command-message><command-name>/skill</command-name>…
//   - <local-command-stdout>…                                  (tool output, not useful)
//
// Returns "" when no human-readable content is found so the caller skips the
// message and uses the next real user turn as the topic.
func extractTopic(text string) string {
	switch {
	case strings.HasPrefix(text, "<local-command-caveat>"),
		strings.HasPrefix(text, "<local-command-stdout>"),
		strings.HasPrefix(text, "<local-command-stderr>"):
		// These are injected by Claude Code and carry no useful topic text.
		return ""

	case strings.HasPrefix(text, "<command-name>"):
		// e.g. <command-name>/plugin</command-name>…
		if name := extractXMLTag(text, "command-name"); name != "" {
			return name
		}

	case strings.HasPrefix(text, "<command-message>"):
		// Prefer <command-name> (includes / prefix) when also present in the same message.
		if name := extractXMLTag(text, "command-name"); name != "" {
			return name
		}
		if name := extractXMLTag(text, "command-message"); name != "" {
			return name
		}
	}
	return text
}

// extractXMLTag returns the trimmed content of the first occurrence of
// <tag>…</tag> in s, or "" if not found.
func extractXMLTag(s, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(s, open)
	end := strings.Index(s, close)
	if start >= 0 && end > start+len(open) {
		return strings.TrimSpace(s[start+len(open) : end])
	}
	return ""
}
