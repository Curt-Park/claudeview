package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// LogLine is a single line in the log view.
type LogLine struct {
	Text  string
	Style string // "tool", "text", "think", "result", "time", "normal"
}

// LogView displays scrollable transcript logs.
type LogView struct {
	Title  string
	Lines  []LogLine
	Offset int
	ColOff int // horizontal scroll
	Follow bool
	Width  int
	Height int
	Filter string
}

// NewLogView creates a new log view.
func NewLogView(title string, width, height int) LogView {
	return LogView{
		Title:  title,
		Width:  width,
		Height: height,
	}
}

// SetLines replaces the log lines.
func (l *LogView) SetLines(lines []LogLine) {
	l.Lines = lines
	if l.Follow {
		l.ScrollToBottom()
	}
}

// AppendLines adds lines to the log.
func (l *LogView) AppendLine(line LogLine) {
	l.Lines = append(l.Lines, line)
	if l.Follow {
		l.ScrollToBottom()
	}
}

// ScrollToBottom scrolls to the last line.
func (l *LogView) ScrollToBottom() {
	if len(l.Lines) > l.Height {
		l.Offset = len(l.Lines) - l.Height
	}
}

// ScrollUp scrolls up by one line.
func (l *LogView) ScrollUp() {
	if l.Offset > 0 {
		l.Offset--
	}
}

// ScrollDown scrolls down by one line.
func (l *LogView) ScrollDown() {
	maxOff := max(0, len(l.Lines)-l.Height)
	if l.Offset < maxOff {
		l.Offset++
	}
}

// ScrollLeft scrolls left.
func (l *LogView) ScrollLeft() {
	if l.ColOff > 0 {
		l.ColOff -= 4
	}
}

// ScrollRight scrolls right.
func (l *LogView) ScrollRight() {
	l.ColOff += 4
}

// GotoTop scrolls to the top.
func (l *LogView) GotoTop() {
	l.Offset = 0
}

// GotoBottom scrolls to the bottom.
func (l *LogView) GotoBottom() {
	l.ScrollToBottom()
}

// ToggleFollow toggles follow mode.
func (l *LogView) ToggleFollow() {
	l.Follow = !l.Follow
	if l.Follow {
		l.ScrollToBottom()
	}
}

// Update handles key events for the log view.
func (l *LogView) Update(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k", "up":
			l.ScrollUp()
			return true
		case "j", "down":
			l.ScrollDown()
			return true
		case "h", "left":
			l.ScrollLeft()
			return true
		case "l", "right":
			l.ScrollRight()
			return true
		case "g":
			l.GotoTop()
			return true
		case "G":
			l.GotoBottom()
			return true
		case "f":
			l.ToggleFollow()
			return true
		}
	}
	return false
}

// filteredLines returns lines matching the current filter.
func (l *LogView) filteredLines() []LogLine {
	if l.Filter == "" {
		return l.Lines
	}
	var out []LogLine
	for _, line := range l.Lines {
		if strings.Contains(strings.ToLower(line.Text), strings.ToLower(l.Filter)) {
			out = append(out, line)
		}
	}
	return out
}

// View renders the log view.
func (l LogView) View() string {
	lines := l.filteredLines()
	var sb strings.Builder

	// Title bar
	title := fmt.Sprintf("── Logs: %s ", l.Title)
	if l.Follow {
		title += StyleActive.Render("[follow] ")
	}
	sb.WriteString(StyleTitle.Render(title))
	sb.WriteString("\n")

	// Content
	visible := l.Height - 2 // -2 for title and status
	if visible <= 0 {
		visible = 10
	}

	for i := l.Offset; i < len(lines) && i < l.Offset+visible; i++ {
		line := lines[i]
		rendered := l.renderLine(line)
		// Horizontal scroll
		if l.ColOff > 0 && len(rendered) > l.ColOff {
			rendered = rendered[l.ColOff:]
		}
		// Truncate to width
		if len(rendered) > l.Width {
			rendered = rendered[:l.Width]
		}
		sb.WriteString(rendered)
		sb.WriteString("\n")
	}

	// Fill
	rendered := min(len(lines)-l.Offset, visible)
	for i := rendered; i < visible; i++ {
		sb.WriteString("\n")
	}

	// Status line
	total := len(lines)
	status := fmt.Sprintf("%d/%d lines", min(l.Offset+visible, total), total)
	if l.Filter != "" {
		status += fmt.Sprintf("  filter: %s", l.Filter)
	}
	sb.WriteString(StyleDim.Render(status))

	return sb.String()
}

func (l LogView) renderLine(line LogLine) string {
	switch line.Style {
	case "tool":
		return StyleLogTool.Render(line.Text)
	case "text":
		return StyleLogText.Render(line.Text)
	case "think":
		return StyleLogThink.Render(line.Text)
	case "result":
		return StyleLogResult.Render(line.Text)
	case "time":
		return StyleLogTime.Render(line.Text)
	default:
		return line.Text
	}
}
