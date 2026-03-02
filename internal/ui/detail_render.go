package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Curt-Park/claudeview/internal/model"
)

// RenderSessionChat renders a session's conversation as a chat timeline.
func RenderSessionChat(turns []model.Turn, subagentTurns [][]model.Turn, subagentTypes []model.AgentType, width int) string {
	if len(turns) == 0 {
		return ""
	}

	userBubbleWidth := int(float64(width) * 0.70)
	if userBubbleWidth < 20 {
		userBubbleWidth = width - 4
	}
	if userBubbleWidth > width-4 {
		userBubbleWidth = width - 4
	}

	var sb strings.Builder
	subIdx := 0
	for _, turn := range turns {
		switch turn.Role {
		case "user":
			sb.WriteString(renderUserBubble(turn, userBubbleWidth, width))
			sb.WriteString("\n")
		case "assistant":
			sb.WriteString(renderClaudeBubble(turn, width))
			sb.WriteString("\n")
			// Append subagent bubble for each Task tool call in this turn
			for _, tc := range turn.ToolCalls {
				if tc.Name == "Task" && subIdx < len(subagentTurns) {
					agentType := model.AgentTypeGeneral
					if subIdx < len(subagentTypes) {
						agentType = subagentTypes[subIdx]
					}
					sb.WriteString(renderSubagentBubbles(subagentTurns[subIdx], agentType, width))
					sb.WriteString("\n")
					subIdx++
				}
			}
		}
	}
	return sb.String()
}

func renderUserBubble(turn model.Turn, bubbleWidth, fullWidth int) string {
	ts := StyleChatTimestamp.Render(turn.Timestamp.Format("15:04"))
	header := ts + " · " + StyleChatHeader.Render("You")
	content := header + "\n" + turn.Text
	bubble := StyleUserBubble.Width(bubbleWidth).Render(content)
	return lipgloss.PlaceHorizontal(fullWidth, lipgloss.Right, bubble)
}

func renderClaudeBubble(turn model.Turn, width int) string {
	claudeWidth := width - 2 // border takes 2 cols

	ts := StyleChatTimestamp.Render(turn.Timestamp.Format("15:04"))
	modelShort := shortModel(turn.ModelName)
	header := "🤖 " + StyleChatHeader.Render("Claude") + " · " + StyleDim.Render(modelShort) + " · " + ts

	var parts []string
	parts = append(parts, header)

	if turn.Thinking != "" {
		thinking := turn.Thinking
		if len([]rune(thinking)) > 80 {
			thinking = string([]rune(thinking)[:77]) + "..."
		}
		parts = append(parts, StyleChatThinking.Render("···thinking: "+thinking+"···"))
	}

	if turn.Text != "" {
		parts = append(parts, "")
		parts = append(parts, turn.Text)
	}

	if len(turn.ToolCalls) > 0 {
		parts = append(parts, "")
		for _, tc := range turn.ToolCalls {
			parts = append(parts, renderToolLine(tc))
		}
	}

	totalTok := turn.InputTokens + turn.OutputTokens
	if totalTok > 0 {
		tokStr := StyleChatTokens.Render(fmt.Sprintf("░ %s tok", model.FormatTokenCount(totalTok)))
		parts = append(parts, "")
		parts = append(parts, lipgloss.PlaceHorizontal(claudeWidth-2, lipgloss.Right, tokStr))
	}

	content := strings.Join(parts, "\n")
	return StyleClaudeBubble.Width(claudeWidth).Render(content)
}

func renderToolLine(tc *model.ToolCall) string {
	name := StyleChatToolName.Render("▸ " + tc.Name)
	summary := tc.InputSummary()
	if len([]rune(summary)) > 40 {
		summary = string([]rune(summary)[:37]) + "..."
	}

	var outcome string
	if tc.IsError {
		outcome = StyleChatToolErr.Render("✗ error")
	} else {
		result := tc.ResultSummary()
		outcome = StyleChatToolOK.Render("→ "+result) + " " + StyleChatToolOK.Render("✓")
	}
	return "  " + name + "  " + StyleDim.Render(summary) + "  " + outcome
}

func renderSubagentBubbles(turns []model.Turn, agentType model.AgentType, width int) string {
	if len(turns) == 0 {
		return ""
	}
	indent := "  "
	subWidth := width - 6 // indent + border

	icon := subagentIcon(agentType)
	label := icon + " " + StyleChatHeader.Render(agentDisplayName(agentType))

	var parts []string
	for _, turn := range turns {
		if turn.Role != "assistant" {
			continue
		}
		ts := StyleChatTimestamp.Render(turn.Timestamp.Format("15:04"))
		header := label + " · " + ts

		var lines []string
		lines = append(lines, header)
		if turn.Text != "" {
			lines = append(lines, "")
			lines = append(lines, turn.Text)
		}
		for _, tc := range turn.ToolCalls {
			lines = append(lines, renderToolLine(tc))
		}
		totalTok := turn.InputTokens + turn.OutputTokens
		if totalTok > 0 {
			tokStr := StyleChatTokens.Render(fmt.Sprintf("░ %s tok", model.FormatTokenCount(totalTok)))
			lines = append(lines, "")
			lines = append(lines, lipgloss.PlaceHorizontal(subWidth-2, lipgloss.Right, tokStr))
		}

		bubble := StyleSubagentBubble.Width(subWidth).Render(strings.Join(lines, "\n"))
		for _, line := range strings.Split(bubble, "\n") {
			parts = append(parts, indent+line)
		}
		parts = append(parts, "")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n")
}

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
	default:
		return "Agent"
	}
}

// RenderPluginItemDetail renders the content of a selected plugin item.
func RenderPluginItemDetail(item *model.PluginItem) string {
	if item == nil {
		return ""
	}
	header := StyleTitle.Render(item.Name) + "  " + StyleDim.Render(item.Category)
	content := model.ReadPluginItemContent(item)
	result := header + "\n\n" + content
	if item.Category == "hook" {
		scripts := model.ReadHookCommandScripts(item)
		if len(scripts) > 0 {
			result += "\n\ncommand scripts below:\n"
			for _, s := range scripts {
				result += "\n" + StyleDim.Render("--- "+s.Path+" ---") + "\n" + s.Content
			}
		}
	}
	return result
}

// RenderMemoryDetail reads and returns the raw content of a memory file.
func RenderMemoryDetail(m *model.Memory) string {
	if m == nil {
		return ""
	}
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Sprintf("error reading %s: %v", m.Path, err)
	}
	return string(data)
}
