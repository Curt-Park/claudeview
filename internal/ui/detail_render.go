package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Curt-Park/claudeview/internal/model"
)

func shortModel(m string) string {
	lower := strings.ToLower(m)
	switch {
	case strings.Contains(lower, "opus"):
		return "opus"
	case strings.Contains(lower, "sonnet"):
		return "sonnet"
	case strings.Contains(lower, "haiku"):
		return "haiku"
	default:
		return m
	}
}

func subagentIcon(t model.AgentType) string {
	switch t {
	case model.AgentTypeExplore:
		return "🔍"
	case model.AgentTypePlan:
		return "📋"
	case model.AgentTypeBash:
		return "💻"
	default:
		return "⚙️"
	}
}

func agentDisplayName(t model.AgentType) string {
	switch t {
	case model.AgentTypeExplore:
		return "Explorer"
	case model.AgentTypePlan:
		return "Planner"
	case model.AgentTypeBash:
		return "Bash"
	case model.AgentTypeGeneral:
		return "General"
	default:
		return "Agent"
	}
}

// RenderChatItemDetail renders a single ChatItem as flat text (no bubble borders),
// similar to RenderPluginItemDetail: a styled header line followed by full-width content.
// For grouped items (with ExtraTurns), all tool calls are rendered after the primary turn.
func RenderChatItemDetail(item ChatItem, width int) string {
	turn := item.Turn

	// Build header: WHO · model · time · tokens (aggregated)
	var headerParts []string
	if item.IsSubagent {
		icon := subagentIcon(item.AgentType)
		headerParts = append(headerParts, StyleChatHeader.Render(icon+" "+agentDisplayName(item.AgentType)))
	} else if turn.Role == "user" {
		headerParts = append(headerParts, StyleChatHeader.Render("You"))
	} else {
		headerParts = append(headerParts, StyleChatHeader.Render("Claude"))
	}
	if m := shortModel(turn.ModelName); m != "" {
		headerParts = append(headerParts, StyleDim.Render(m))
	}
	if !turn.Timestamp.IsZero() {
		headerParts = append(headerParts, StyleChatTimestamp.Render(turn.Timestamp.Format("15:04")))
	}
	totalTok := turn.InputTokens + turn.OutputTokens
	for _, et := range item.ExtraTurns {
		totalTok += et.InputTokens + et.OutputTokens
	}
	if totalTok > 0 {
		headerParts = append(headerParts, StyleChatTokens.Render(model.FormatTokenCount(totalTok)+" tok"))
	}

	var parts []string
	parts = append(parts, strings.Join(headerParts, "  "))

	// Primary turn: thinking → text → tool calls
	if turn.Thinking != "" {
		parts = append(parts, "")
		parts = append(parts, StyleChatThinking.Render("── thinking ──"))
		parts = append(parts, ansi.Wrap(turn.Thinking, width, ""))
	}

	if turn.Text != "" {
		parts = append(parts, "")
		parts = append(parts, ansi.Wrap(turn.Text, width, ""))
	}

	for _, tc := range turn.ToolCalls {
		parts = append(parts, "")
		parts = append(parts, renderExpandedToolCall(tc, width))
	}

	// ExtraTurns: separator + thinking + tool calls for each grouped turn
	for _, et := range item.ExtraTurns {
		parts = append(parts, "")
		parts = append(parts, StyleDim.Render("── continued ──"))
		if et.Thinking != "" {
			parts = append(parts, "")
			parts = append(parts, StyleChatThinking.Render("── thinking ──"))
			parts = append(parts, ansi.Wrap(et.Thinking, width, ""))
		}
		for _, tc := range et.ToolCalls {
			parts = append(parts, "")
			parts = append(parts, renderExpandedToolCall(tc, width))
		}
	}

	return strings.Join(parts, "\n")
}

// renderExpandedToolCall renders a tool call with full input and result.
func renderExpandedToolCall(tc *model.ToolCall, maxWidth int) string {
	name := StyleChatToolName.Render("▸ " + tc.Name)
	inputFull := tc.InputSummary()

	var statusStr string
	if tc.IsError {
		statusStr = StyleChatToolErr.Render("✗ error")
	} else {
		statusStr = StyleChatToolOK.Render("✓")
	}

	durationStr := ""
	if tc.Duration > 0 {
		durationStr = " " + StyleChatTimestamp.Render(tc.DurationString())
	}

	headerLine := "  " + name + "  " + StyleDim.Render(inputFull) + "  " + statusStr + durationStr

	var lines []string
	lines = append(lines, headerLine)

	// Show full result content
	resultStr := expandResult(tc)
	if resultStr != "" {
		for _, rl := range strings.Split(resultStr, "\n") {
			lines = append(lines, "    "+StyleDim.Render(rl))
		}
	}

	return strings.Join(lines, "\n")
}

// expandResult extracts the full result text from a tool call.
func expandResult(tc *model.ToolCall) string {
	if tc.Result == nil {
		return ""
	}
	// Try string result
	var s string
	if err := json.Unmarshal(tc.Result, &s); err == nil {
		// Cap at reasonable length for display
		lines := strings.Split(s, "\n")
		if len(lines) > 30 {
			lines = append(lines[:30], fmt.Sprintf("... (%d more lines)", len(lines)-30))
		}
		return strings.Join(lines, "\n")
	}
	// Try array of content blocks
	var arr []map[string]any
	if err := json.Unmarshal(tc.Result, &arr); err == nil {
		var texts []string
		for _, block := range arr {
			if text, ok := block["text"].(string); ok {
				texts = append(texts, text)
			}
		}
		if len(texts) > 0 {
			result := strings.Join(texts, "\n")
			lines := strings.Split(result, "\n")
			if len(lines) > 30 {
				lines = append(lines[:30], fmt.Sprintf("... (%d more lines)", len(lines)-30))
			}
			return strings.Join(lines, "\n")
		}
	}
	return ""
}

// RenderPluginItemDetail renders the content of a selected plugin item.
func RenderPluginItemDetail(item *model.PluginItem, width int) string {
	if item == nil {
		return ""
	}
	header := StyleTitle.Render(item.Name) + "  " + StyleDim.Render(item.Category)
	content := model.ReadPluginItemContent(item)
	result := header + "\n\n" + ansi.Wrap(content, width, "")
	if item.Category == "hook" {
		scripts := model.ReadHookCommandScripts(item)
		if len(scripts) > 0 {
			result += "\n\ncommand scripts below:\n"
			for _, s := range scripts {
				result += "\n" + StyleDim.Render("--- "+s.Path+" ---") + "\n" + ansi.Wrap(s.Content, width, "")
			}
		}
	}
	return result
}

// RenderMemoryDetail reads and returns the raw content of a memory file.
func RenderMemoryDetail(m *model.Memory, width int) string {
	if m == nil {
		return ""
	}
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Sprintf("error reading %s: %v", m.Path, err)
	}
	return ansi.Wrap(string(data), width, "")
}
