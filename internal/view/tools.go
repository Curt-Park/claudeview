package view

import (
	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

var toolColumnsBase = []ui.Column{
	{Title: "TIME", Width: 10},
	{Title: "TOOL", Width: 10},
	{Title: "INPUT SUMMARY", Width: 10, Flex: true, MaxPercent: 0.40},
	{Title: "RESULT", Width: 16},
	{Title: "DURATION", Width: 10},
}

var toolColumnsFlat = []ui.Column{
	{Title: "SESSION", Width: 10},
	{Title: "AGENT", Width: 10},
	{Title: "TIME", Width: 10},
	{Title: "TOOL", Width: 10},
	{Title: "INPUT SUMMARY", Width: 10, Flex: true, MaxPercent: 0.40},
	{Title: "RESULT", Width: 16},
	{Title: "DURATION", Width: 10},
}

// NewToolsView creates a tools view.
func NewToolsView(width, height int) *ResourceView[*model.ToolCall] {
	return NewResourceView(toolColumnsBase, toolColumnsFlat, toolRow, width, height)
}

func toolRow(items []*model.ToolCall, i int, flatMode bool) ui.Row {
	tc := items[i]
	timeStr := ""
	if !tc.Timestamp.IsZero() {
		timeStr = tc.Timestamp.Format("15:04:05")
	}
	resultStr := tc.ResultSummary()
	if tc.IsError {
		resultStr = ui.StyleError.Render("error")
	}
	var cells []string
	if flatMode {
		sessionID := tc.SessionID
		if len(sessionID) > 8 {
			sessionID = sessionID[:8]
		}
		agentID := tc.AgentID
		if agentID == "" {
			agentID = "main"
		} else if len(agentID) > 8 {
			agentID = agentID[:8]
		}
		cells = append(cells, sessionID, agentID)
	}
	cells = append(cells,
		timeStr,
		tc.Name,
		tc.InputSummary(),
		resultStr,
		tc.DurationString(),
	)
	return ui.Row{Cells: cells, Data: tc}
}
