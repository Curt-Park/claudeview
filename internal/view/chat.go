package view

import (
	"github.com/Curt-Park/claudeview/internal/ui"
)

var chatColumns = []ui.Column{
	{Title: "NAME", Width: 18},
	{Title: "MESSAGE", Width: 30, Flex: true, MaxPercent: 0.50},
	{Title: "ACTION", Width: 16},
	{Title: "MODEL:TOKEN", Width: 20, Flex: true, MaxPercent: 0.20},
	{Title: "DURATION", Width: 14},
}

func chatRow(items []ui.ChatItem, i int, _ bool) ui.Row {
	c := items[i]
	if c.IsDivider {
		return ui.Row{
			Cells: []string{
				"",
				c.DividerLabel,
				"",
				"",
				"",
			},
			Data: c,
			Skip: true,
		}
	}
	var prev *ui.ChatItem
	if i > 0 {
		prev = &items[i-1]
	}
	return ui.Row{
		Cells: []string{
			c.WhoLabel(),
			c.MessagePreview(120),
			c.ActionLabel(),
			c.ModelTokenLabel(),
			c.TimeLabel(prev),
		},
		Data: c,
	}
}

// NewChatView creates a ResourceView for the session chat table.
func NewChatView(width, height int) *ResourceView[ui.ChatItem] {
	return NewResourceView(chatColumns, nil, chatRow, width, height)
}
