package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Column defines a column in the table.
type Column struct {
	Title string
	Width int
	Flex  bool // if true, grows to fill remaining space
}

// Row is a single row in the table (slice of cell strings + raw data).
type Row struct {
	Cells []string
	Data  any // original data object
}

// TableView is a generic scrollable table component.
type TableView struct {
	Columns  []Column
	Rows     []Row
	Selected int
	Offset   int // scroll offset
	Width    int
	Height   int
	Filter   string
}

// NewTableView creates a new table view.
func NewTableView(cols []Column, width, height int) TableView {
	return TableView{
		Columns: cols,
		Width:   width,
		Height:  height,
	}
}

// SetRows sets the table rows.
func (t *TableView) SetRows(rows []Row) {
	t.Rows = rows
	if t.Selected >= len(rows) {
		t.Selected = max(0, len(rows)-1)
	}
}

// SelectedRow returns the currently selected row data.
func (t *TableView) SelectedRow() *Row {
	if len(t.Rows) == 0 || t.Selected >= len(t.Rows) {
		return nil
	}
	return &t.Rows[t.Selected]
}

// MoveUp moves the selection up.
func (t *TableView) MoveUp() {
	if t.Selected > 0 {
		t.Selected--
	}
	t.ensureVisible()
}

// MoveDown moves the selection down.
func (t *TableView) MoveDown() {
	if t.Selected < len(t.Rows)-1 {
		t.Selected++
	}
	t.ensureVisible()
}

// GotoTop moves to the first row.
func (t *TableView) GotoTop() {
	t.Selected = 0
	t.Offset = 0
}

// GotoBottom moves to the last row.
func (t *TableView) GotoBottom() {
	if len(t.Rows) == 0 {
		return
	}
	t.Selected = len(t.Rows) - 1
	t.ensureVisible()
}

// PageUp moves up by half a page.
func (t *TableView) PageUp() {
	half := max(1, t.Height/2)
	t.Selected = max(0, t.Selected-half)
	t.ensureVisible()
}

// PageDown moves down by half a page.
func (t *TableView) PageDown() {
	half := max(1, t.Height/2)
	t.Selected = min(len(t.Rows)-1, t.Selected+half)
	t.ensureVisible()
}

func (t *TableView) ensureVisible() {
	if t.Selected < t.Offset {
		t.Offset = t.Selected
	}
	if t.Selected >= t.Offset+t.Height {
		t.Offset = t.Selected - t.Height + 1
	}
}

// Update handles key events.
func (t *TableView) Update(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k", "up":
			t.MoveUp()
			return true, nil
		case "j", "down":
			t.MoveDown()
			return true, nil
		case "g":
			t.GotoTop()
			return true, nil
		case "G":
			t.GotoBottom()
			return true, nil
		case "ctrl+u", "pgup":
			t.PageUp()
			return true, nil
		case "ctrl+d", "pgdown":
			t.PageDown()
			return true, nil
		}
	}
	return false, nil
}

// filteredRows returns rows matching the current filter (case-insensitive).
func (t TableView) filteredRows() []Row {
	if t.Filter == "" {
		return t.Rows
	}
	needle := strings.ToLower(t.Filter)
	var out []Row
	for _, row := range t.Rows {
		for _, cell := range row.Cells {
			if strings.Contains(strings.ToLower(cell), needle) {
				out = append(out, row)
				break
			}
		}
	}
	return out
}

// View renders the table.
func (t TableView) View() string {
	if t.Width == 0 {
		return ""
	}

	rows := t.filteredRows()

	// Calculate column widths
	cols := t.calculateWidths()

	// Header
	var sb strings.Builder
	sb.WriteString(t.renderHeader(cols))
	sb.WriteString("\n")

	// Rows
	visible := t.Height
	if visible <= 0 {
		visible = 20
	}

	// Clamp offset/selected against filtered set
	offset := t.Offset
	selected := t.Selected
	if offset >= len(rows) {
		offset = max(0, len(rows)-1)
	}

	for i := offset; i < len(rows) && i < offset+visible; i++ {
		row := rows[i]
		line := t.renderRow(row, cols, i == selected)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Fill empty rows
	rendered := min(len(rows)-offset, visible)
	for i := rendered; i < visible; i++ {
		sb.WriteString(strings.Repeat(" ", t.Width))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (t TableView) calculateWidths() []int {
	widths := make([]int, len(t.Columns))
	total := 0
	flexIdx := -1

	for i, col := range t.Columns {
		if col.Flex {
			flexIdx = i
		} else {
			widths[i] = col.Width
			total += col.Width + 1 // +1 for spacing
		}
	}

	if flexIdx >= 0 {
		widths[flexIdx] = max(10, t.Width-total-1)
	}

	return widths
}

func (t TableView) renderHeader(widths []int) string {
	var parts []string
	for i, col := range t.Columns {
		cell := padRight(col.Title, widths[i])
		parts = append(parts, StyleColumnHeader.Render(cell))
	}
	return strings.Join(parts, " ")
}

func (t TableView) renderRow(row Row, widths []int, selected bool) string {
	var parts []string
	for i := range t.Columns {
		cell := ""
		if i < len(row.Cells) {
			cell = row.Cells[i]
		}
		padded := padRight(cell, widths[i])
		parts = append(parts, padded)
	}
	line := strings.Join(parts, " ")
	if selected {
		return StyleSelected.Width(t.Width).Render(line)
	}
	return line
}

func padRight(s string, n int) string {
	visible := lipgloss.Width(s)
	if visible >= n {
		if n > 1 {
			return ansi.Truncate(s, n-1, "…")
		}
		return ansi.Truncate(s, n, "")
	}
	return s + strings.Repeat(" ", n-visible)
}

// RowCount returns the number of rows.
func (t *TableView) RowCount() int {
	return len(t.Rows)
}

// StatusLine returns a status string like "1/10".
func (t *TableView) StatusLine() string {
	if len(t.Rows) == 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", t.Selected+1, len(t.Rows))
}
