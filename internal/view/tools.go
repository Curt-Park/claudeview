package view

import (
	"fmt"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// ToolsView renders the tool calls list.
type ToolsView struct {
	ToolCalls []*model.ToolCall
	Table     ui.TableView
	FlatMode  bool // when true, SESSION+AGENT columns are shown (flat :command access)
}

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
func NewToolsView(width, height int) *ToolsView {
	return &ToolsView{
		Table: ui.NewTableView(toolColumnsBase, width, height),
	}
}

// SetToolCalls updates the tool calls list.
func (v *ToolsView) SetToolCalls(calls []*model.ToolCall) {
	v.ToolCalls = calls
	if v.FlatMode {
		v.Table.Columns = toolColumnsFlat
	} else {
		v.Table.Columns = toolColumnsBase
	}
	rows := make([]ui.Row, len(calls))
	for i, tc := range calls {
		timeStr := ""
		if !tc.Timestamp.IsZero() {
			timeStr = tc.Timestamp.Format("15:04:05")
		}
		resultStr := tc.ResultSummary()
		if tc.IsError {
			resultStr = ui.StyleError.Render("error")
		}
		var cells []string
		if v.FlatMode {
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
		rows[i] = ui.Row{Cells: cells, Data: tc}
	}
	v.Table.SetRows(rows)
}

// SelectedToolCall returns the currently selected tool call.
func (v *ToolsView) SelectedToolCall() *model.ToolCall {
	row := v.Table.SelectedRow()
	if row == nil {
		return nil
	}
	if tc, ok := row.Data.(*model.ToolCall); ok {
		return tc
	}
	return nil
}

// View renders the tools table.
func (v *ToolsView) View() string {
	return v.Table.View()
}

// DetailLines generates detail view lines for a tool call.
func ToolCallDetailLines(tc *model.ToolCall) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Tool:     %s", tc.Name))
	lines = append(lines, fmt.Sprintf("Agent:    %s", tc.AgentID))
	if !tc.Timestamp.IsZero() {
		lines = append(lines, fmt.Sprintf("Time:     %s (duration: %s)",
			tc.Timestamp.Format("15:04:05"), tc.DurationString()))
	}
	lines = append(lines, "")
	lines = append(lines, "Input:")
	if tc.Input != nil {
		lines = append(lines, "  "+string(tc.Input))
	}
	lines = append(lines, "")
	lines = append(lines, "Output:")
	if tc.IsError {
		lines = append(lines, "  [ERROR]")
	}
	if tc.Result != nil {
		lines = append(lines, "  "+string(tc.Result))
	}
	return lines
}
