package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
)

// ChatItem represents a single selectable item in the chat table.
// It wraps a Turn with metadata about whether it's from a subagent.
// ExtraTurns holds subsequent tool-only assistant turns grouped under this item.
type ChatItem struct {
	Turn         model.Turn
	ExtraTurns   []model.Turn // subsequent tool-only turns merged into this group
	IsSubagent   bool
	AgentType    model.AgentType
	SubagentIdx  int // index into subagentTurns; -1 for non-subagent items
	IsDivider    bool
	DividerLabel string
}

// AllToolCalls collects tool calls from the primary Turn and all ExtraTurns.
func (c ChatItem) AllToolCalls() []*model.ToolCall {
	all := make([]*model.ToolCall, 0, len(c.Turn.ToolCalls))
	all = append(all, c.Turn.ToolCalls...)
	for _, et := range c.ExtraTurns {
		all = append(all, et.ToolCalls...)
	}
	return all
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

// WhoLabel returns a short label for the message author.
func (c ChatItem) WhoLabel() string {
	if c.IsDivider {
		return ""
	}
	if c.IsSubagent {
		return agentDisplayName(c.AgentType)
	}
	switch c.Turn.Role {
	case "user":
		return "You"
	case "assistant":
		return "Claude"
	case "system":
		return "System"
	}
	return c.Turn.Role
}

// MessagePreview returns a single-line summary of the turn content.
// Combines text preview and first tool call info for maximum context.
func (c ChatItem) MessagePreview(max int) string {
	if c.IsDivider {
		return c.DividerLabel
	}
	var parts []string

	// Add text preview (first non-blank line, whitespace-collapsed).
	// Check Turn.Text first, then fall back to ExtraTurns for collapsed subagents
	// where the first turn may be tool-only.
	text := c.Turn.Text
	if text == "" {
		for _, et := range c.ExtraTurns {
			if et.Text != "" {
				text = et.Text
				break
			}
		}
	}
	if text != "" {
		line := cleanTextPreview(text)
		if line != "" {
			parts = append(parts, line)
		}
	}

	// Add first tool call with its input summary
	if allTC := c.AllToolCalls(); len(allTC) > 0 {
		tc := allTC[0]
		toolStr := "▸ " + tc.Name
		if summary := tc.InputSummary(); summary != "" {
			toolStr += " " + summary
		}
		parts = append(parts, toolStr)
	}

	if len(parts) == 0 {
		return "-"
	}
	result := strings.Join(parts, " | ")
	if len([]rune(result)) > max {
		result = string([]rune(result)[:max-3]) + "..."
	}
	return result
}

// cleanTextPreview extracts a meaningful single-line preview from turn text.
// Handles system-injected content (command tags, skill responses, local command
// output) and skips leading blank lines.
func cleanTextPreview(text string) string {
	// Handle <command-message> / <command-name> user turns (slash commands).
	// Prefer <command-name> (includes / prefix) when present.
	switch {
	case strings.HasPrefix(text, "<command-message>"):
		if name := extractXMLContent(text, "command-name"); name != "" {
			return name
		}
		if name := extractXMLContent(text, "command-message"); name != "" {
			return "/" + name
		}
	case strings.HasPrefix(text, "<command-name>"):
		if name := extractXMLContent(text, "command-name"); name != "" {
			return name
		}
	case strings.HasPrefix(text, "Base directory for this skill:"):
		return "(skill loaded)"
	case strings.HasPrefix(text, "<local-command-caveat>"),
		strings.HasPrefix(text, "<local-command-stdout>"),
		strings.HasPrefix(text, "<local-command-stderr>"):
		return "(command output)"
	}

	// Collapse newlines into spaces so the preview shows as much context as
	// possible in a single line (the caller truncates to the column width).
	line := strings.Join(strings.Fields(text), " ")
	return line
}

// extractXMLContent returns the trimmed text inside the first <tag>…</tag> in s.
func extractXMLContent(s, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(s, open)
	end := strings.Index(s, close)
	if start >= 0 && end > start+len(open) {
		return strings.TrimSpace(s[start+len(open) : end])
	}
	return ""
}

// ActionLabel returns the first tool name + "+N" count, or "-".
// For Agent/Task calls, appends the agent type. For Skill calls, appends the skill name.
func (c ChatItem) ActionLabel() string {
	if c.IsDivider {
		return ""
	}
	allTC := c.AllToolCalls()
	if len(allTC) == 0 {
		return "-"
	}
	tc := allTC[0]
	name := tc.Name
	switch tc.Name {
	case "Task", "Agent":
		name = "Agent"
		if agentType := extractStringField(tc, "subagent_type"); agentType != "" {
			name += ":" + agentType
		}
	case "Skill":
		if skillName := extractStringField(tc, "skill"); skillName != "" {
			name = "Skill:" + skillName
		}
	}
	if len(allTC) > 1 {
		return fmt.Sprintf("%s+%d", name, len(allTC)-1)
	}
	return name
}

// extractStringField reads a string field from a tool call's JSON input.
func extractStringField(tc *model.ToolCall, field string) string {
	if tc.Input == nil {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(tc.Input, &m); err != nil {
		return ""
	}
	if v, ok := m[field].(string); ok {
		return v
	}
	return ""
}

// ModelTokenLabel returns per-model token totals (e.g. "opus:1.5k sonnet:300") or "-".
func (c ChatItem) ModelTokenLabel() string {
	if c.IsDivider {
		return ""
	}
	byModel := make(map[string]int)
	addTurn := func(t model.Turn) {
		if t.ModelName == "" {
			return
		}
		byModel[t.ModelName] += t.InputTokens + t.OutputTokens
	}
	addTurn(c.Turn)
	for _, et := range c.ExtraTurns {
		addTurn(et)
	}
	if len(byModel) == 0 {
		return "-"
	}
	models := make([]string, 0, len(byModel))
	for m := range byModel {
		models = append(models, m)
	}
	sort.Strings(models)
	var parts []string
	for _, m := range models {
		parts = append(parts, model.ShortModelName(m)+":"+model.FormatTokenCount(byModel[m]))
	}
	return strings.Join(parts, " ")
}

// TimeLabel returns elapsed time since prev turn, or "HH:MM" for the first item.
func (c ChatItem) TimeLabel(prev *ChatItem) string {
	if c.IsDivider {
		return ""
	}
	if c.Turn.Timestamp.IsZero() {
		return "-"
	}
	if prev == nil || prev.Turn.Timestamp.IsZero() {
		return "-"
	}
	d := c.Turn.Timestamp.Sub(prev.Turn.Timestamp)
	if d < 0 {
		return "-"
	}
	return model.FormatAge(d)
}

// BuildChatItems flattens main turns and subagent turns into a single
// selectable list. Consecutive assistant turns where the 2nd+ has no text
// and no thinking are grouped into the preceding ChatItem's ExtraTurns.
// For each assistant turn's Task tool call, subagent turns are interleaved.
func BuildChatItems(turns []model.Turn, subagentTurns [][]model.Turn, subagentTypes []model.AgentType) []ChatItem {
	var items []ChatItem
	subIdx := 0

	// interleaveSubagents appends one collapsed ChatItem per subagent for each Agent/Task tool call.
	// The first assistant turn becomes the primary Turn; remaining assistant turns go into ExtraTurns.
	interleaveSubagents := func(toolCalls []*model.ToolCall) {
		for _, tc := range toolCalls {
			if (tc.Name == "Task" || tc.Name == "Agent") && subIdx < len(subagentTurns) {
				agentType := model.AgentTypeGeneral
				if subIdx < len(subagentTypes) {
					agentType = subagentTypes[subIdx]
				}
				// Collect assistant turns, preferring one with text or tool calls as primary.
				var allAssistant []model.Turn
				for _, st := range subagentTurns[subIdx] {
					if st.Role != "assistant" {
						continue
					}
					allAssistant = append(allAssistant, st)
				}
				var first *model.Turn
				var extra []model.Turn
				for i := range allAssistant {
					if first == nil {
						if allAssistant[i].Text != "" || len(allAssistant[i].ToolCalls) > 0 {
							first = &allAssistant[i]
						} else {
							extra = append(extra, allAssistant[i]) // content-less leading turns → ExtraTurns
						}
					} else {
						extra = append(extra, allAssistant[i])
					}
				}
				// Fallback: use the first assistant turn even if content-less.
				if first == nil && len(allAssistant) > 0 {
					first = &allAssistant[0]
					extra = allAssistant[1:]
				}
				if first != nil {
					items = append(items, ChatItem{
						Turn:        *first,
						ExtraTurns:  extra,
						IsSubagent:  true,
						AgentType:   agentType,
						SubagentIdx: subIdx,
					})
				}
				subIdx++
			}
		}
	}

	for _, turn := range turns {
		// Group no-text assistant turns (tool-only or thinking+tools) into the preceding assistant ChatItem.
		// Thinking is internal reasoning, not user-visible dialog, so it doesn't prevent grouping.
		if turn.Role == "assistant" && turn.Text == "" && len(items) > 0 {
			last := &items[len(items)-1]
			if last.Turn.Role == "assistant" && !last.IsSubagent {
				last.ExtraTurns = append(last.ExtraTurns, turn)
				interleaveSubagents(turn.ToolCalls)
				continue
			}
		}

		items = append(items, ChatItem{Turn: turn, SubagentIdx: -1})
		if turn.Role == "assistant" {
			interleaveSubagents(turn.ToolCalls)
		}
	}
	return items
}

// BuildMergedChatItems merges turns from multiple sessions into a single ChatItem list
// with divider rows at session boundaries.
// sessionIDs provides the short session ID for each session (used in divider labels).
// For a single session, it delegates to BuildChatItems.
func BuildMergedChatItems(
	sessionTurns [][]model.Turn,
	subTurnsBySession [][][]model.Turn,
	subTypesBySession [][]model.AgentType,
	sessionIDs []string,
) []ChatItem {
	if len(sessionTurns) <= 1 {
		var subTurns [][]model.Turn
		var subTypes []model.AgentType
		if len(subTurnsBySession) > 0 {
			subTurns = subTurnsBySession[0]
		}
		if len(subTypesBySession) > 0 {
			subTypes = subTypesBySession[0]
		}
		var turns []model.Turn
		if len(sessionTurns) > 0 {
			turns = sessionTurns[0]
		}
		return BuildChatItems(turns, subTurns, subTypes)
	}

	total := len(sessionTurns)
	var all []ChatItem
	for i, turns := range sessionTurns {
		var subTurns [][]model.Turn
		var subTypes []model.AgentType
		if i < len(subTurnsBySession) {
			subTurns = subTurnsBySession[i]
		}
		if i < len(subTypesBySession) {
			subTypes = subTypesBySession[i]
		}
		items := BuildChatItems(turns, subTurns, subTypes)
		if i > 0 {
			label := fmt.Sprintf("── session %d/%d", i+1, total)
			if i < len(sessionIDs) && sessionIDs[i] != "" {
				label += fmt.Sprintf(" (%s)", sessionIDs[i])
			}
			label += " ──"
			all = append(all, ChatItem{
				IsDivider:    true,
				DividerLabel: label,
				SubagentIdx:  -1,
			})
		}
		all = append(all, items...)
	}
	return all
}
