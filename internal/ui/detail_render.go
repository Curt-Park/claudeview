package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Curt-Park/claudeview/internal/model"
)

// renderTurnBoundary renders a lightweight API-call boundary marker shown between
// ExtraTurns in both Claude and sub-agent detail views:
//
//	── sonnet  06:14  300 tok ──
func renderTurnBoundary(t model.Turn, width int) string {
	var parts []string
	if m := model.ShortModelName(t.ModelName); m != "" {
		parts = append(parts, StyleDim.Render(m))
	}
	if !t.Timestamp.IsZero() {
		parts = append(parts, StyleChatTimestamp.Render(t.Timestamp.Format("15:04")))
	}
	if t.InputTokens > 0 || t.CacheReadTokens > 0 || t.OutputTokens > 0 {
		parts = append(parts, StyleChatTokens.Render(model.FormatTokenInOutCache(t.InputTokens, t.CacheReadTokens, t.OutputTokens)+" tok"))
	}
	inner := strings.Join(parts, "  ")
	if inner == "" {
		return StyleChatThinking.Render("────────────")
	}
	return StyleChatThinking.Render("── " + inner + " ──")
}

// RenderChatItemDetail renders the detail view for a selected ChatItem.
// Shows the message text (thinking + text) and inline one-line tool call
// summaries across all turns.
func RenderChatItemDetail(items []ChatItem, selectedIdx, width int) string {
	if selectedIdx < 0 || selectedIdx >= len(items) {
		return ""
	}
	sel := items[selectedIdx]
	var lines []string

	// Header: WHO · model · time · tokens
	lines = append(lines, renderChatItemHeader(sel))

	// Render thinking, text, and tool call summaries for a single turn.
	renderTurn := func(t model.Turn) {
		if t.Thinking != "" {
			lines = append(lines, "")
			lines = append(lines, StyleChatThinking.Render("── thinking ──"))
			lines = append(lines, ansi.Wrap(t.Thinking, width, ""))
		}
		if t.Text != "" {
			lines = append(lines, "")
			lines = append(lines, ansi.Wrap(t.Text, width, ""))
		}
		for _, tc := range t.ToolCalls {
			lines = append(lines, renderToolCallOneliner(tc, width))
		}
	}

	renderTurn(sel.Turn)
	for _, et := range sel.ExtraTurns {
		if et.Thinking == "" && et.Text == "" && len(et.ToolCalls) == 0 {
			continue
		}
		lines = append(lines, "")
		lines = append(lines, renderTurnBoundary(et, width))
		renderTurn(et)
	}

	return strings.Join(lines, "\n")
}

// renderToolCallOneliner renders a tool call as a single compact summary line:
//
//	▸ NAME  input-summary  ✓/✗  duration
func renderToolCallOneliner(tc *model.ToolCall, maxWidth int) string {
	var statusStr string
	if tc.IsError {
		statusStr = StyleChatToolErr.Render("✗")
	} else {
		statusStr = StyleChatToolOK.Render("✓")
	}

	parts := []string{StyleChatToolName.Render("▸ " + tc.Name)}
	if s := tc.InputSummary(); s != "" {
		parts = append(parts, StyleDim.Render(s))
	}
	parts = append(parts, statusStr)
	if tc.Duration > 0 {
		parts = append(parts, StyleChatTimestamp.Render(tc.DurationString()))
	}
	line := "  " + strings.Join(parts, "  ")
	return ansi.Wrap(line, maxWidth, "")
}

// renderChatItemHeader builds the "WHO · model · time · tokens" header line.
func renderChatItemHeader(item ChatItem) string {
	turn := item.Turn
	var parts []string
	if item.IsSubagent {
		parts = append(parts, StyleChatHeader.Render(item.AgentType.Icon()+" "+item.AgentType.DisplayLabel()))
	} else if turn.Role == "user" {
		parts = append(parts, StyleChatHeader.Render("You"))
	} else {
		parts = append(parts, StyleChatHeader.Render("Claude"))
	}
	if m := model.ShortModelName(turn.ModelName); m != "" {
		parts = append(parts, StyleDim.Render(m))
	}
	if !turn.Timestamp.IsZero() {
		parts = append(parts, StyleChatTimestamp.Render(turn.Timestamp.Format("15:04")))
	}
	totalIn := turn.InputTokens
	totalCache := turn.CacheReadTokens
	totalOut := turn.OutputTokens
	for _, et := range item.ExtraTurns {
		totalIn += et.InputTokens
		totalCache += et.CacheReadTokens
		totalOut += et.OutputTokens
	}
	if totalIn > 0 || totalCache > 0 || totalOut > 0 {
		parts = append(parts, StyleChatTokens.Render(model.FormatTokenInOutCache(totalIn, totalCache, totalOut)+" tok"))
	}
	return strings.Join(parts, "  ")
}

// ChatItemKey returns a unique fingerprint for a ChatItem, used to re-resolve
// the selected item after async rebuilds.
func ChatItemKey(item ChatItem) string {
	key := item.Turn.Timestamp.String() + "|" + item.Turn.Role + "|" + fmt.Sprintf("%d", item.SubagentIdx)
	if tcs := item.AllToolCalls(); len(tcs) > 0 {
		key += "|" + tcs[0].Name
	}
	// Include text prefix to disambiguate items with identical timestamp/role
	// (e.g. consecutive user turns from local commands at the same second).
	if t := item.Turn.Text; len(t) > 32 {
		key += "|" + t[:32]
	} else if t != "" {
		key += "|" + t
	}
	return key
}

// renderExpandedToolCall renders a tool call in two-line Option-1 style:
//
//	▸ NAME  model  duration
//	    input summary  ✓/✗
//	    result...
func renderExpandedToolCall(tc *model.ToolCall, turn model.Turn, maxWidth int) string {
	// Line 1: name + model + duration + tokens
	headerParts := []string{StyleChatToolName.Render("▸ " + tc.Name)}
	if m := model.ShortModelName(turn.ModelName); m != "" {
		headerParts = append(headerParts, StyleDim.Render(m))
	}
	if tc.Duration > 0 {
		headerParts = append(headerParts, StyleChatTimestamp.Render(tc.DurationString()))
	}
	if turn.InputTokens > 0 || turn.CacheReadTokens > 0 || turn.OutputTokens > 0 {
		headerParts = append(headerParts, StyleChatTokens.Render(model.FormatTokenInOutCache(turn.InputTokens, turn.CacheReadTokens, turn.OutputTokens)+" tok"))
	}
	headerLine := "  " + strings.Join(headerParts, "  ")

	// Line 2 (indented): input summary + status
	var statusStr string
	if tc.IsError {
		statusStr = StyleChatToolErr.Render("✗ error")
	} else {
		statusStr = StyleChatToolOK.Render("✓")
	}
	inputFull := tc.InputSummary()
	inputLine := "    "
	if inputFull != "" {
		inputLine += ansi.Wrap(StyleDim.Render(inputFull)+"  "+statusStr, maxWidth-4, "")
	} else {
		inputLine += statusStr
	}

	var lines []string
	lines = append(lines, ansi.Wrap(headerLine, maxWidth, ""))
	lines = append(lines, inputLine)

	// Result lines (indented)
	if resultStr := expandResult(tc); resultStr != "" {
		indent := "    "
		contentWidth := maxWidth - len(indent)
		if contentWidth < 20 {
			contentWidth = 20
		}
		for _, rl := range strings.Split(resultStr, "\n") {
			lines = append(lines, indent+ansi.Wrap(StyleDim.Render(rl), contentWidth, ""))
		}
	}

	return strings.Join(lines, "\n")
}

// expandResult extracts the full result text from a tool call.
func expandResult(tc *model.ToolCall) string {
	return tc.ResultText()
}

// RenderToolCallDetail renders the full detail view for a single tool call.
func RenderToolCallDetail(tr *ToolCallRow, width int) string {
	if tr == nil {
		return ""
	}
	return renderExpandedToolCall(tr.ToolCall, tr.ParentTurn, width)
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
	if m.Content != "" {
		return ansi.Wrap(m.Content, width, "")
	}
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Sprintf("error reading %s: %v", m.Path, err)
	}
	return ansi.Wrap(string(data), width, "")
}
