package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Curt-Park/claudeview/internal/model"
)

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

// renderAgentCallLine renders an Agent/Task tool call: icon + agent name header,
// followed by the full result message.
// subItem is the corresponding sub-agent ChatItem; its last turn text is used as
// a fallback when tc.Result is not yet available (e.g. agent still in progress).
func renderAgentCallLine(tc *model.ToolCall, subItem *ChatItem, maxWidth int) string {
	agentType := model.AgentTypeFromInput(tc.Input)
	icon := subagentIcon(agentType)
	name := agentDisplayName(agentType)
	headerParts := []string{StyleChatToolName.Render(icon + " " + name)}
	if subItem != nil {
		if m := model.ShortModelName(subItem.Turn.ModelName); m != "" {
			headerParts = append(headerParts, StyleDim.Render(m))
		}
		if !subItem.Turn.Timestamp.IsZero() {
			headerParts = append(headerParts, StyleChatTimestamp.Render(subItem.Turn.Timestamp.Format("15:04")))
		}
		totalTok := subItem.Turn.InputTokens + subItem.Turn.OutputTokens
		for _, et := range subItem.ExtraTurns {
			totalTok += et.InputTokens + et.OutputTokens
		}
		if totalTok > 0 {
			headerParts = append(headerParts, StyleChatTokens.Render(model.FormatTokenCount(totalTok)+" tok"))
		}
	}
	header := "  " + strings.Join(headerParts, "  ")
	result := expandResult(tc)
	if result == "" && subItem != nil {
		result = lastAssistantText(*subItem)
	}
	if result == "" {
		return header
	}
	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")
	for _, l := range strings.Split(result, "\n") {
		lines = append(lines, ansi.Wrap(StyleDim.Render(l), maxWidth, ""))
	}
	return strings.Join(lines, "\n")
}

// lastAssistantText returns the last non-empty text from a sub-agent ChatItem,
// checking ExtraTurns in reverse order before falling back to the primary Turn.
func lastAssistantText(item ChatItem) string {
	for i := len(item.ExtraTurns) - 1; i >= 0; i-- {
		if item.ExtraTurns[i].Text != "" {
			return item.ExtraTurns[i].Text
		}
	}
	return item.Turn.Text
}

// RenderChatItemDetail renders the detail view for a selected ChatItem.
// For subagent items, it renders the primary turn plus all ExtraTurns (the full transcript).
// For regular items (user, assistant, divider), it renders a single item.
func RenderChatItemDetail(items []ChatItem, selectedIdx, width int) string {
	if selectedIdx < 0 || selectedIdx >= len(items) {
		return ""
	}
	sel := items[selectedIdx]

	// Subagent: render primary turn + all extra turns as full transcript.
	if sel.IsSubagent {
		var allLines []string
		allLines = append(allLines, renderChatItem(ChatItem{
			Turn: sel.Turn, IsSubagent: sel.IsSubagent, AgentType: sel.AgentType, SubagentIdx: sel.SubagentIdx,
		}, nil, width)...)
		for _, et := range sel.ExtraTurns {
			allLines = append(allLines, "") // blank line between turns
			allLines = append(allLines, renderChatItem(ChatItem{
				Turn: et, IsSubagent: sel.IsSubagent, AgentType: sel.AgentType, SubagentIdx: sel.SubagentIdx,
			}, nil, width)...)
		}
		return strings.Join(allLines, "\n")
	}

	// For regular items, collect the sub-agent ChatItems that immediately follow
	// so Agent/Task call lines can show content even when tc.Result is not yet set.
	var subItems []ChatItem
	for i := selectedIdx + 1; i < len(items); i++ {
		if items[i].IsSubagent {
			subItems = append(subItems, items[i])
		} else if !items[i].IsDivider {
			break
		}
	}
	return strings.Join(renderChatItem(sel, subItems, width), "\n")
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

// renderChatItem renders a single ChatItem as flat text lines.
// subItems provides the sub-agent ChatItems that follow this item (positional match
// for Agent/Task calls) so their content can be shown when tc.Result is not set.
func renderChatItem(item ChatItem, subItems []ChatItem, width int) []string {
	turn := item.Turn

	// Build header: WHO · model · time · tokens
	var headerParts []string
	if item.IsSubagent {
		icon := subagentIcon(item.AgentType)
		headerParts = append(headerParts, StyleChatHeader.Render(icon+" "+agentDisplayName(item.AgentType)))
	} else if turn.Role == "user" {
		headerParts = append(headerParts, StyleChatHeader.Render("You"))
	} else {
		headerParts = append(headerParts, StyleChatHeader.Render("Claude"))
	}
	if m := model.ShortModelName(turn.ModelName); m != "" {
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

	agentIdx := 0
	renderTC := func(tc *model.ToolCall, currentTurn model.Turn) string {
		if tc.Name == "Agent" || tc.Name == "Task" {
			var si *ChatItem
			if agentIdx < len(subItems) {
				si = &subItems[agentIdx]
			}
			agentIdx++
			return renderAgentCallLine(tc, si, width)
		}
		return renderExpandedToolCall(tc, currentTurn, width)
	}

	for _, tc := range turn.ToolCalls {
		parts = append(parts, "")
		parts = append(parts, renderTC(tc, turn))
	}

	// ExtraTurns: separator + thinking + text + tool calls for each grouped turn
	for _, et := range item.ExtraTurns {
		parts = append(parts, "")
		if et.Thinking != "" {
			parts = append(parts, StyleChatThinking.Render("── thinking ──"))
			parts = append(parts, ansi.Wrap(et.Thinking, width, ""))
		}
		if et.Text != "" {
			parts = append(parts, ansi.Wrap(et.Text, width, ""))
		}
		for _, tc := range et.ToolCalls {
			parts = append(parts, "")
			parts = append(parts, renderTC(tc, et))
		}
	}

	return parts
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
	totalTok := turn.InputTokens + turn.OutputTokens
	if totalTok > 0 {
		headerParts = append(headerParts, StyleChatTokens.Render(model.FormatTokenCount(totalTok)+" tok"))
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
	if tc.Result == nil {
		return ""
	}
	// Try string result
	var s string
	if err := json.Unmarshal(tc.Result, &s); err == nil {
		return s
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
			return strings.Join(texts, "\n")
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
	if m.Content != "" {
		return ansi.Wrap(m.Content, width, "")
	}
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Sprintf("error reading %s: %v", m.Path, err)
	}
	return ansi.Wrap(string(data), width, "")
}
