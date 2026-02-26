package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Column defines a column in the table.
type Column struct {
	Title      string
	Width      int     // fixed width, or minimum width when Flex=true
	Flex       bool    // if true, grows to fill remaining space
	MaxPercent float64 // if Flex and > 0, caps expansion at this fraction of terminal width (0–1)
}

// Row is a single row in the table (slice of cell strings + raw data).
type Row struct {
	Cells []string
	Data  any // original data object
}

// TableView is a generic scrollable table component.
type TableView struct {
	Columns      []Column
	Rows         []Row
	Selected     int
	Offset       int // scroll offset
	ExpandOffset int // lines of expanded selected row scrolled off the top
	Width        int
	Height       int
	Filter       string
}

// NewTableView creates a new table view.
func NewTableView(cols []Column, width, height int) TableView {
	return TableView{
		Columns: cols,
		Width:   width,
		Height:  height,
	}
}

// SetRows sets the table rows and clamps Selected to the filtered set.
func (t *TableView) SetRows(rows []Row) {
	t.Rows = rows
	t.ExpandOffset = 0
	n := t.FilteredCount()
	if n == 0 {
		t.Selected = 0
	} else if t.Selected >= n {
		t.Selected = n - 1
	}
}

// SelectedRow returns the currently selected row (from the filtered set).
func (t *TableView) SelectedRow() *Row {
	rows := t.filteredRows()
	if len(rows) == 0 || t.Selected >= len(rows) {
		return nil
	}
	return &rows[t.Selected]
}

// MoveUp moves the selection up within the filtered set.
// If the selected row's expanded content is partially scrolled (ExpandOffset > 0),
// it scrolls back up within the expansion before moving to the previous row.
func (t *TableView) MoveUp() {
	if t.ExpandOffset > 0 {
		t.ExpandOffset--
		return
	}
	if t.Selected > 0 {
		t.Selected--
	}
	t.ensureVisible()
}

// MoveDown moves the selection down within the filtered set.
// If the selected row's expanded content overflows the viewport, it scrolls
// through the hidden lines one by one before advancing to the next row.
func (t *TableView) MoveDown() {
	rows := t.filteredRows()
	if t.Selected < len(rows) {
		cols := t.calculateWidths()
		allLines := strings.Split(t.renderExpandedRow(rows[t.Selected], cols), "\n")
		if t.ExpandOffset < max(0, len(allLines)-t.dataRows()) {
			t.ExpandOffset++
			return
		}
	}
	t.ExpandOffset = 0
	if t.Selected < t.FilteredCount()-1 {
		t.Selected++
	}
	t.ensureVisible()
}

// GotoTop moves to the first row of the filtered set.
func (t *TableView) GotoTop() {
	t.Selected = 0
	t.Offset = 0
	t.ExpandOffset = 0
}

// GotoBottom moves to the last row of the filtered set.
func (t *TableView) GotoBottom() {
	n := t.FilteredCount()
	if n == 0 {
		return
	}
	t.Selected = n - 1
	t.ExpandOffset = 0
	t.ensureVisible()
}

// dataRows returns the number of visible data rows (total height minus the header line).
func (t *TableView) dataRows() int {
	return max(t.Height-1, 1)
}

// PageUp moves up by half a page.
func (t *TableView) PageUp() {
	half := max(1, t.dataRows()/2)
	t.Selected = max(0, t.Selected-half)
	t.ExpandOffset = 0
	t.ensureVisible()
}

// PageDown moves down by half a page within the filtered set.
func (t *TableView) PageDown() {
	half := max(1, t.dataRows()/2)
	t.Selected = min(t.FilteredCount()-1, t.Selected+half)
	t.ExpandOffset = 0
	t.ensureVisible()
}

func (t *TableView) ensureVisible() {
	dr := t.dataRows()
	if t.Selected < t.Offset {
		t.Offset = t.Selected
	}
	if t.Selected >= t.Offset+dr {
		t.Offset = t.Selected - dr + 1
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

	visible := t.dataRows()

	// Clamp offset/selected against filtered set
	offset := t.Offset
	selected := t.Selected
	if len(rows) == 0 {
		offset = 0
		selected = -1
	} else {
		if offset >= len(rows) {
			offset = len(rows) - 1
		}
		if selected >= len(rows) {
			selected = len(rows) - 1
		}
	}

	// Pre-render the expanded (selected) row so we know its line count.
	// Nudge offset up to maximise space for the expansion, then apply
	// ExpandOffset to slide the visible window within the expanded content.
	var expandedLines []string
	if selected >= offset && selected < len(rows) {
		allExpLines := strings.Split(t.renderExpandedRow(rows[selected], cols), "\n")
		linesBeforeSel := selected - offset
		if linesBeforeSel+len(allExpLines) > visible {
			offset = min(selected, max(0, selected-(visible-len(allExpLines))))
		}
		start := min(t.ExpandOffset, len(allExpLines))
		expandedLines = allExpLines[start:]
	}

	linesUsed := 0
	for i := offset; i < len(rows) && linesUsed < visible; i++ {
		row := rows[i]
		if i == selected {
			remaining := visible - linesUsed
			lines := expandedLines
			if len(lines) > remaining {
				lines = lines[:remaining]
			}
			for _, l := range lines {
				sb.WriteString(l)
				sb.WriteString("\n")
				linesUsed++
			}
		} else {
			sb.WriteString(t.renderRow(row, cols, false))
			sb.WriteString("\n")
			linesUsed++
		}
	}

	// Fill empty lines
	for linesUsed < visible {
		sb.WriteString(strings.Repeat(" ", t.Width))
		sb.WriteString("\n")
		linesUsed++
	}

	return sb.String()
}

func (t TableView) calculateWidths() []int {
	widths := make([]int, len(t.Columns))
	spacing := len(t.Columns) - 1

	// Pass 1: assign fixed columns and sum their widths.
	fixedTotal := 0
	for i, col := range t.Columns {
		if !col.Flex {
			widths[i] = col.Width
			fixedTotal += col.Width
		}
	}
	avail := max(0, t.Width-fixedTotal-spacing)

	// Pass 2: collect flex columns and compute their desired widths.
	// Fixed columns always take priority; flex columns share whatever remains.
	// Flex column minimum widths (col.Width) are treated as soft hints — they
	// will NOT cause the table to overflow the terminal.
	type fInfo struct {
		idx     int
		desired int
		pct     float64 // MaxPercent, or 0 if uncapped
	}
	var fCols []fInfo
	for i, col := range t.Columns {
		if !col.Flex {
			continue
		}
		desired := avail
		pct := col.MaxPercent
		if pct > 0 {
			if capped := int(float64(t.Width) * pct); capped < desired {
				desired = capped
			}
		}
		// Do NOT enforce col.Width here: that would cause overflow on narrow terminals.
		fCols = append(fCols, fInfo{i, desired, pct})
	}
	if len(fCols) == 0 {
		return widths
	}

	// If the sum of desired widths fits within avail, use them directly.
	totalDesired := 0
	for _, f := range fCols {
		totalDesired += f.desired
	}
	if totalDesired <= avail {
		for _, f := range fCols {
			widths[f.idx] = f.desired
		}
		return widths
	}

	// Otherwise scale down proportionally using MaxPercent as weights.
	totalPct := 0.0
	for _, f := range fCols {
		totalPct += f.pct
	}
	if totalPct == 0 {
		// No MaxPercent set: distribute equally.
		share := avail / len(fCols)
		for _, f := range fCols {
			widths[f.idx] = max(0, share)
		}
		return widths
	}
	remaining := avail
	for j, f := range fCols {
		var w int
		if j == len(fCols)-1 {
			w = remaining
		} else {
			w = int(float64(avail) * f.pct / totalPct)
			remaining -= w
		}
		widths[f.idx] = max(0, w)
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
		if selected {
			// Strip ANSI codes so the selection style renders uniformly across the whole row.
			// Pre-rendered cells contain reset sequences (\x1b[0m) that would cut off the
			// selection background mid-row.
			padded = ansi.Strip(padded)
		}
		parts = append(parts, padded)
	}
	line := strings.Join(parts, " ")
	if selected {
		return StyleSelected.Width(t.Width).Render(line)
	}
	return line
}

// renderExpandedRow renders the selected row as multiple lines, wrapping each
// cell's full content within its column width so nothing is truncated.
func (t TableView) renderExpandedRow(row Row, widths []int) string {
	colLines := make([][]string, len(t.Columns))
	maxLines := 1
	for i := range t.Columns {
		cell := ""
		if i < len(row.Cells) {
			cell = ansi.Strip(row.Cells[i])
		}
		wrapped := wrapText(cell, widths[i])
		colLines[i] = wrapped
		if len(wrapped) > maxLines {
			maxLines = len(wrapped)
		}
	}

	var sb strings.Builder
	for lineIdx := 0; lineIdx < maxLines; lineIdx++ {
		var parts []string
		for i := range t.Columns {
			cell := ""
			if lineIdx < len(colLines[i]) {
				cell = colLines[i][lineIdx]
			}
			parts = append(parts, padRight(cell, widths[i]))
		}
		line := strings.Join(parts, " ")
		if lineIdx > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(StyleSelected.Width(t.Width).Render(line))
	}
	return sb.String()
}

// wrapText splits s into lines of at most width visible characters.
// It works correctly with multibyte and wide (CJK) characters.
func wrapText(s string, width int) []string {
	if width <= 0 || s == "" {
		return []string{""}
	}
	runes := []rune(s)
	var lines []string
	for len(runes) > 0 {
		if lipgloss.Width(string(runes)) <= width {
			lines = append(lines, string(runes))
			break
		}
		// Greedily consume runes up to width visible chars.
		lineW := 0
		end := 0
		for _, r := range runes {
			rw := lipgloss.Width(string(r))
			if lineW+rw > width {
				break
			}
			lineW += rw
			end++
		}
		if end == 0 {
			end = 1 // force-include at least one rune to avoid infinite loop
		}
		lines = append(lines, string(runes[:end]))
		runes = runes[end:]
	}
	return lines
}

func padRight(s string, n int) string {
	visible := lipgloss.Width(s)
	if visible > n {
		if n > 1 {
			truncated := ansi.Truncate(s, n-1, "…")
			tw := lipgloss.Width(truncated)
			return truncated + strings.Repeat(" ", max(n-tw, 0))
		}
		return ansi.Truncate(s, n, "")
	}
	return s + strings.Repeat(" ", n-visible)
}

// FilteredCount returns the number of rows after applying the current filter.
func (t TableView) FilteredCount() int {
	return len(t.filteredRows())
}
