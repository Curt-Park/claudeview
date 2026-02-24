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
}

var toolColumns = []ui.Column{
	{Title: "TIME", Width: 10},
	{Title: "TOOL", Width: 10},
	{Title: "INPUT SUMMARY", Width: 30, Flex: true},
	{Title: "RESULT", Width: 16},
	{Title: "DURATION", Width: 10},
}

// NewToolsView creates a tools view.
func NewToolsView(width, height int) *ToolsView {
	return &ToolsView{
		Table: ui.NewTableView(toolColumns, width, height),
	}
}

// SetToolCalls updates the tool calls list.
func (v *ToolsView) SetToolCalls(calls []*model.ToolCall) {
	v.ToolCalls = calls
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
		rows[i] = ui.Row{
			Cells: []string{
				timeStr,
				tc.Name,
				tc.InputSummary(),
				resultStr,
				tc.DurationString(),
			},
			Data: tc,
		}
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
