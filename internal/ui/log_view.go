package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
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

// visibleLines returns the number of content lines that fit in the view.
func (l *LogView) visibleLines() int {
	v := l.Height - 2 // subtract title and status bar
	if v <= 0 {
		return 10
	}
	return v
}

// ScrollToBottom scrolls to the last line.
func (l *LogView) ScrollToBottom() {
	visible := l.visibleLines()
	if len(l.Lines) > visible {
		l.Offset = len(l.Lines) - visible
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
	maxOff := max(0, len(l.Lines)-l.visibleLines())
	if l.Offset < maxOff {
		l.Offset++
	}
}

// PageUp scrolls up by half a page.
func (l *LogView) PageUp() {
	half := max(1, l.visibleLines()/2)
	l.Offset = max(0, l.Offset-half)
}

// PageDown scrolls down by half a page.
func (l *LogView) PageDown() {
	half := max(1, l.visibleLines()/2)
	maxOff := max(0, len(l.Lines)-l.visibleLines())
	l.Offset = min(maxOff, l.Offset+half)
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
		case "g":
			l.GotoTop()
			return true
		case "G":
			l.GotoBottom()
			return true
		case "f":
			l.ToggleFollow()
			return true
		case "ctrl+u", "pgup":
			l.PageUp()
			return true
		case "ctrl+d", "pgdown":
			l.PageDown()
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
	visible := l.visibleLines()

	for i := l.Offset; i < len(lines) && i < l.Offset+visible; i++ {
		line := lines[i]
		rendered := l.renderLine(line)
		rendered = ansi.Truncate(rendered, l.Width, "")
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
