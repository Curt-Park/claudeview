package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ToolCall represents a single tool invocation and its result.
type ToolCall struct {
	ID        string
	SessionID string
	AgentID   string
	Name      string
	Input     json.RawMessage
	Result    json.RawMessage
	IsError   bool
	Timestamp time.Time
	Duration  time.Duration
}

// InputSummary returns a one-line summary of the tool input (not truncated;
// the caller is responsible for display-time truncation or wrapping).
func (tc *ToolCall) InputSummary() string {
	if tc.Input == nil {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(tc.Input, &m); err != nil {
		return singleLine(string(tc.Input))
	}

	switch tc.Name {
	case "Read":
		if fp, ok := m["file_path"].(string); ok {
			return singleLine(fp)
		}
	case "Write":
		if fp, ok := m["file_path"].(string); ok {
			return singleLine(fp)
		}
	case "Edit":
		if fp, ok := m["file_path"].(string); ok {
			return singleLine(fp)
		}
	case "Bash":
		if cmd, ok := m["command"].(string); ok {
			return singleLine(cmd)
		}
	case "Grep":
		pattern, _ := m["pattern"].(string)
		path, _ := m["path"].(string)
		if path == "" {
			path = "."
		}
		return singleLine(fmt.Sprintf("%q in %s", pattern, path))
	case "Glob":
		if p, ok := m["pattern"].(string); ok {
			return singleLine(p)
		}
	case "Task":
		if desc, ok := m["description"].(string); ok {
			return singleLine(desc)
		}
	case "WebFetch":
		if url, ok := m["url"].(string); ok {
			return singleLine(url)
		}
	}

	// Fallback: first string value
	for _, v := range m {
		if s, ok := v.(string); ok && s != "" {
			return singleLine(s)
		}
	}
	return singleLine(string(tc.Input))
}

// singleLine replaces newlines with spaces so callers get a flat string
// suitable for wrapping at arbitrary column widths.
func singleLine(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

// ResultSummary returns a one-line summary of the tool result.
func (tc *ToolCall) ResultSummary() string {
	if tc.IsError {
		return "error"
	}
	if tc.Result == nil {
		return "-"
	}
	// Try string result
	var s string
	if err := json.Unmarshal(tc.Result, &s); err == nil {
		lines := strings.Split(s, "\n")
		return truncate(fmt.Sprintf("%d lines", len(lines)), 20)
	}
	// Try array of content blocks
	var arr []map[string]any
	if err := json.Unmarshal(tc.Result, &arr); err == nil {
		return fmt.Sprintf("%d blocks", len(arr))
	}
	return truncate(string(tc.Result), 20)
}

// DurationString returns formatted duration.
func (tc *ToolCall) DurationString() string {
	if tc.Duration == 0 {
		return "-"
	}
	if tc.Duration < time.Second {
		return fmt.Sprintf("%.1fms", float64(tc.Duration.Milliseconds()))
	}
	return fmt.Sprintf("%.1fs", tc.Duration.Seconds())
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
