package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Curt-Park/claudeview/internal/stringutil"
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

// matchToolResults fills Result, IsError, and Duration for each ToolCall
// whose ID appears in the provided maps.
func matchToolResults(calls []ToolCall, results map[string]json.RawMessage, errors map[string]bool, timestamps map[string]time.Time) {
	for i := range calls {
		tc := &calls[i]
		if res, ok := results[tc.ID]; ok {
			tc.Result = res
			tc.IsError = errors[tc.ID]
			if resultTS, ok := timestamps[tc.ID]; ok && !tc.Timestamp.IsZero() {
				tc.Duration = resultTS.Sub(tc.Timestamp)
			}
		}
	}
}

// flushPendingTurn matches tool results into the pending turn, accumulates metrics, and appends it.
// toolTimestamps maps tool_use_id -> timestamp of the user turn that delivered the result;
// used to compute per-tool-call Duration.
func flushPendingTurn(result *ParsedTranscript, turn *Turn, toolResults map[string]json.RawMessage, toolErrors map[string]bool, toolTimestamps map[string]time.Time) {
	matchToolResults(turn.ToolCalls, toolResults, toolErrors, toolTimestamps)
	// Accumulate per-model token usage
	u := result.TokensByModel[turn.Model]
	u.InputTokens += turn.Usage.NewInputTokens()
	u.CacheReadInputTokens += turn.Usage.CacheReadInputTokens
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

// formatCompactText builds display text for a compact_boundary system entry.
func formatCompactText(e entry) string {
	text := "Conversation compacted"
	if e.CompactMetadata != nil && e.CompactMetadata.PreTokens > 0 {
		text += fmt.Sprintf(" (%dk tokens)", e.CompactMetadata.PreTokens/1000)
	}
	return text
}

// collectToolResults populates the result/error/timestamp maps from a user message's tool results.
func collectToolResults(msg userMessage, ts time.Time, results map[string]json.RawMessage, errors map[string]bool, timestamps map[string]time.Time) {
	for _, c := range msg.toolResults() {
		results[c.ToolUseID] = c.Content
		errors[c.ToolUseID] = c.IsError
		timestamps[c.ToolUseID] = ts
	}
}

// buildAssistantTurn constructs an assistant Turn from a parsed message.
func buildAssistantTurn(msg assistantMessage, ts time.Time) Turn {
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
	return turn
}

// mergeAssistantTurn merges next into pending (consecutive assistant entries).
func mergeAssistantTurn(pending *Turn, next Turn) {
	if next.Text != "" {
		if pending.Text != "" {
			pending.Text += "\n"
		}
		pending.Text += next.Text
	}
	if next.Thinking != "" {
		if pending.Thinking != "" {
			pending.Thinking += "\n"
		}
		pending.Thinking += next.Thinking
	}
	pending.ToolCalls = append(pending.ToolCalls, next.ToolCalls...)
	pending.Usage.InputTokens += next.Usage.InputTokens
	pending.Usage.CacheCreationInputTokens += next.Usage.CacheCreationInputTokens
	pending.Usage.CacheReadInputTokens += next.Usage.CacheReadInputTokens
	pending.Usage.OutputTokens += next.Usage.OutputTokens
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
	toolTimestamps := map[string]time.Time{}

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
			collectToolResults(msg, ts, toolResults, toolErrors, toolTimestamps)
			// Flush pending assistant turn with matched results
			if pendingAssistantTurn != nil {
				flushPendingTurn(result, pendingAssistantTurn, toolResults, toolErrors, toolTimestamps)
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
			turn := buildAssistantTurn(msg, ts)
			if pendingAssistantTurn != nil {
				// Merge consecutive assistant entries (e.g. text-only followed by tool-only)
				mergeAssistantTurn(pendingAssistantTurn, turn)
			} else {
				pendingAssistantTurn = &turn
			}

		case "system":
			if entry.Subtype == "compact_boundary" {
				if pendingAssistantTurn != nil {
					flushPendingTurn(result, pendingAssistantTurn, toolResults, toolErrors, toolTimestamps)
					pendingAssistantTurn = nil
				}
				result.Turns = append(result.Turns, Turn{
					Role: "system", Text: formatCompactText(entry), Timestamp: ts,
				})
			}
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
		flushPendingTurn(result, pendingAssistantTurn, toolResults, toolErrors, toolTimestamps)
	}

	return result, scanner.Err()
}

// TranscriptCache holds parser state for incremental turns parsing.
type TranscriptCache struct {
	committed      []Turn                     // fully flushed turns (tool results matched)
	pending        *Turn                      // current assistant turn awaiting tool results
	toolResults    map[string]json.RawMessage // collected but unmatched tool results
	toolErrors     map[string]bool            // error flags for unmatched tool results
	toolTimestamps map[string]time.Time       // timestamp of user turn delivering each tool result
	offset         int64                      // file position after last-read line
	topic          string                     // first real user message text
}

// Offset returns the current file read position.
func (c *TranscriptCache) Offset() int64 { return c.offset }

// Turns returns all turns including a snapshot of the pending assistant turn.
func (c *TranscriptCache) Turns() []Turn {
	if c.pending == nil {
		return c.committed
	}
	snapshot := *c.pending
	snapshot.ToolCalls = make([]ToolCall, len(c.pending.ToolCalls))
	copy(snapshot.ToolCalls, c.pending.ToolCalls)
	matchToolResults(snapshot.ToolCalls, c.toolResults, c.toolErrors, c.toolTimestamps)
	return append(c.committed, snapshot)
}

// ParseFileIncremental reads a JSONL file from the stored offset,
// processes new lines into turns, and returns the updated cache.
// If cache is nil, a new TranscriptCache is created (reading from offset 0).
func ParseFileIncremental(path string, cache *TranscriptCache) (*TranscriptCache, error) {
	if cache == nil {
		cache = &TranscriptCache{
			toolResults:    make(map[string]json.RawMessage),
			toolErrors:     make(map[string]bool),
			toolTimestamps: make(map[string]time.Time),
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return cache, err
	}
	defer func() { _ = f.Close() }()

	if cache.offset > 0 {
		if _, err := f.Seek(cache.offset, io.SeekStart); err != nil {
			return cache, err
		}
	}

	scanner := newJSONLScanner(f)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var e entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339Nano, e.Timestamp)

		switch e.Type {
		case "user":
			var msg userMessage
			if err := json.Unmarshal(e.Message, &msg); err != nil {
				continue
			}
			// Collect tool results to match with pending tool calls
			collectToolResults(msg, ts, cache.toolResults, cache.toolErrors, cache.toolTimestamps)
			// Flush pending assistant turn with matched results
			if cache.pending != nil {
				matchToolResults(cache.pending.ToolCalls, cache.toolResults, cache.toolErrors, cache.toolTimestamps)
				cache.committed = append(cache.committed, *cache.pending)
				cache.pending = nil
			}
			// Add user text turns
			turn := Turn{Role: "user", Timestamp: ts}
			turn.Text = msg.textContent()
			if turn.Text != "" {
				if cache.topic == "" && !strings.HasPrefix(turn.Text, "Base directory for this skill:") {
					cache.topic = extractTopic(turn.Text)
				}
				cache.committed = append(cache.committed, turn)
			}

		case "assistant":
			var msg assistantMessage
			if err := json.Unmarshal(e.Message, &msg); err != nil {
				continue
			}
			turn := buildAssistantTurn(msg, ts)
			if cache.pending != nil {
				// Merge consecutive assistant entries
				mergeAssistantTurn(cache.pending, turn)
			} else {
				cache.pending = &turn
			}

		case "system":
			if e.Subtype == "compact_boundary" {
				if cache.pending != nil {
					matchToolResults(cache.pending.ToolCalls, cache.toolResults, cache.toolErrors, cache.toolTimestamps)
					cache.committed = append(cache.committed, *cache.pending)
					cache.pending = nil
				}
				cache.committed = append(cache.committed, Turn{
					Role: "system", Text: formatCompactText(e), Timestamp: ts,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return cache, err
	}

	// Update offset to end of file
	if pos, err := f.Seek(0, io.SeekEnd); err == nil {
		cache.offset = pos
	}

	return cache, nil
}

// SessionAggregates holds cached session-level metrics for incremental parsing.
type SessionAggregates struct {
	Topic          string
	Branch         string
	Slug           string
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

		// Capture slug from first entry that has it
		if agg.Slug == "" && entry.Slug != "" {
			agg.Slug = entry.Slug
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
			u.InputTokens += msg.Usage.NewInputTokens()
			u.CacheReadInputTokens += msg.Usage.CacheReadInputTokens
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
		strings.HasPrefix(text, "<local-command-stderr>"),
		strings.HasPrefix(text, "[Request interrupted by user"):
		// These are injected by Claude Code and carry no useful topic text.
		return ""

	case strings.HasPrefix(text, "<command-name>"):
		// e.g. <command-name>/plugin</command-name>…
		if name := stringutil.ExtractXMLTag(text, "command-name"); name != "" {
			return name
		}

	case strings.HasPrefix(text, "<command-message>"):
		// Prefer <command-name> (includes / prefix) when also present in the same message.
		if name := stringutil.ExtractXMLTag(text, "command-name"); name != "" {
			return name
		}
		if name := stringutil.ExtractXMLTag(text, "command-message"); name != "" {
			return name
		}
	}
	return text
}
