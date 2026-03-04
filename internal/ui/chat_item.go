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
	Turn       model.Turn
	ExtraTurns []model.Turn // subsequent tool-only turns merged into this group
	IsSubagent bool
	AgentType  model.AgentType
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

// WhoLabel returns a short label for the message author.
func (c ChatItem) WhoLabel() string {
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
	var parts []string

	// Add text preview (first line, whitespace-collapsed)
	if c.Turn.Text != "" {
		line := strings.SplitN(c.Turn.Text, "\n", 2)[0]
		line = strings.TrimSpace(line)
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

// ActionLabel returns the first tool name + "+N" count, or "-".
// For Agent/Task calls, appends the agent type. For Skill calls, appends the skill name.
func (c ChatItem) ActionLabel() string {
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
	if c.Turn.Timestamp.IsZero() {
		return "-"
	}
	if prev == nil || prev.Turn.Timestamp.IsZero() {
		return "-"
	}
	return model.FormatAge(c.Turn.Timestamp.Sub(prev.Turn.Timestamp))
}

// BuildChatItems flattens main turns and subagent turns into a single
// selectable list. Consecutive assistant turns where the 2nd+ has no text
// and no thinking are grouped into the preceding ChatItem's ExtraTurns.
// For each assistant turn's Task tool call, subagent turns are interleaved.
func BuildChatItems(turns []model.Turn, subagentTurns [][]model.Turn, subagentTypes []model.AgentType) []ChatItem {
	var items []ChatItem
	subIdx := 0

	// interleaveSubagents appends subagent ChatItems for each Task tool call in the turn.
	interleaveSubagents := func(toolCalls []*model.ToolCall) {
		for _, tc := range toolCalls {
			if tc.Name == "Task" && subIdx < len(subagentTurns) {
				agentType := model.AgentTypeGeneral
				if subIdx < len(subagentTypes) {
					agentType = subagentTypes[subIdx]
				}
				for _, st := range subagentTurns[subIdx] {
					if st.Role == "assistant" {
						items = append(items, ChatItem{
							Turn:       st,
							IsSubagent: true,
							AgentType:  agentType,
						})
					}
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

		items = append(items, ChatItem{Turn: turn})
		if turn.Role == "assistant" {
			interleaveSubagents(turn.ToolCalls)
		}
	}
	return items
}
