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
// If Subtitle is non-empty it is rendered as a second dimmed line below the row.
// SubtitleIndent specifies how many spaces to prepend to the subtitle (e.g. to
// align it under a specific column).
type Row struct {
	Cells          []string
	Subtitle       string // optional second line shown in dimmed style
	SubtitleIndent int    // leading spaces before the subtitle text
	Data           any    // original data object
}

// rowLineCount returns the number of display lines this row occupies.
func rowLineCount(row Row) int {
	if row.Subtitle != "" {
		return 2
	}
	return 1
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

// SetRows sets the table rows and clamps Selected to the filtered set.
func (t *TableView) SetRows(rows []Row) {
	t.Rows = rows
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
func (t *TableView) MoveUp() {
	if t.Selected > 0 {
		t.Selected--
	}
	t.ensureVisible()
}

// MoveDown moves the selection down within the filtered set.
func (t *TableView) MoveDown() {
	if t.Selected < t.FilteredCount()-1 {
		t.Selected++
	}
	t.ensureVisible()
}

// GotoTop moves to the first row of the filtered set.
func (t *TableView) GotoTop() {
	t.Selected = 0
	t.Offset = 0
}

// GotoBottom moves to the last row of the filtered set.
func (t *TableView) GotoBottom() {
	n := t.FilteredCount()
	if n == 0 {
		return
	}
	t.Selected = n - 1
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
	t.ensureVisible()
}

// PageDown moves down by half a page within the filtered set.
func (t *TableView) PageDown() {
	half := max(1, t.dataRows()/2)
	t.Selected = min(t.FilteredCount()-1, t.Selected+half)
	t.ensureVisible()
}

func (t *TableView) ensureVisible() {
	rows := t.filteredRows()
	dr := t.dataRows()
	if t.Selected < t.Offset {
		t.Offset = t.Selected
		return
	}
	// Count display lines from offset to selected (inclusive).
	lines := 0
	for i := t.Offset; i <= t.Selected && i < len(rows); i++ {
		lines += rowLineCount(rows[i])
	}
	// Advance offset until selected row fits within the viewport.
	for lines > dr && t.Offset < t.Selected {
		lines -= rowLineCount(rows[t.Offset])
		t.Offset++
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

	linesUsed := 0
	for i := offset; i < len(rows) && linesUsed < visible; i++ {
		row := rows[i]
		sb.WriteString(t.renderRow(row, cols, i == selected))
		sb.WriteString("\n")
		linesUsed++
		if row.Subtitle != "" && linesUsed < visible {
			sb.WriteString(t.renderSubtitleLine(row))
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

// renderSubtitleLine renders the subtitle string as a full-width dimmed line,
// indented by row.SubtitleIndent spaces to align under a specific column.
func (t TableView) renderSubtitleLine(row Row) string {
	prefix := strings.Repeat(" ", max(0, row.SubtitleIndent))
	text := prefix + row.Subtitle
	padded := padRight(text, t.Width)
	return StyleRowSubtitle.Render(padded)
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
