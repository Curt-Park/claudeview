package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailView displays detailed information about a selected item.
type DetailView struct {
	Title  string
	Lines  []string
	Offset int
	Width  int
	Height int
}

// NewDetailView creates a new detail view.
func NewDetailView(title string, width, height int) DetailView {
	return DetailView{
		Title:  title,
		Width:  width,
		Height: height,
	}
}

// SetContent sets the content lines.
func (d *DetailView) SetContent(lines []string) {
	d.Lines = lines
	d.Offset = 0
}

// SetContentString sets content from a single string.
func (d *DetailView) SetContentString(s string) {
	d.Lines = strings.Split(s, "\n")
	d.Offset = 0
}

// ScrollUp scrolls up.
func (d *DetailView) ScrollUp() {
	if d.Offset > 0 {
		d.Offset--
	}
}

// visibleLines returns the number of content lines that fit in the view.
func (d *DetailView) visibleLines() int {
	v := d.Height - 2 // subtract title and status bar
	if v <= 0 {
		return 10
	}
	return v
}

// getDisplayLines returns the content lines wrapped to the current width.
func (d DetailView) getDisplayLines() []string {
	if d.Width <= 0 {
		return d.Lines
	}
	var out []string
	for _, line := range d.Lines {
		out = append(out, wrapText(line, d.Width)...)
	}
	return out
}

// ScrollDown scrolls down.
func (d *DetailView) ScrollDown() {
	dl := d.getDisplayLines()
	maxOff := max(0, len(dl)-d.visibleLines())
	if d.Offset < maxOff {
		d.Offset++
	}
}

// GotoTop scrolls to top.
func (d *DetailView) GotoTop() {
	d.Offset = 0
}

// GotoBottom scrolls to bottom.
func (d *DetailView) GotoBottom() {
	dl := d.getDisplayLines()
	d.Offset = max(0, len(dl)-d.visibleLines())
}

// PageUp scrolls up by half a page.
func (d *DetailView) PageUp() {
	half := max(1, d.visibleLines()/2)
	d.Offset = max(0, d.Offset-half)
}

// PageDown scrolls down by half a page.
func (d *DetailView) PageDown() {
	half := max(1, d.visibleLines()/2)
	dl := d.getDisplayLines()
	maxOff := max(0, len(dl)-d.visibleLines())
	d.Offset = min(maxOff, d.Offset+half)
}

// Update handles key input.
func (d *DetailView) Update(msg tea.Msg) bool {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k", "up":
			d.ScrollUp()
			return true
		case "j", "down":
			d.ScrollDown()
			return true
		case "g":
			d.GotoTop()
			return true
		case "G":
			d.GotoBottom()
			return true
		case "ctrl+u", "pgup":
			d.PageUp()
			return true
		case "ctrl+d", "pgdown":
			d.PageDown()
			return true
		}
	}
	return false
}

// wrapText splits s into lines of at most width visible characters.
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
			end = 1
		}
		lines = append(lines, string(runes[:end]))
		runes = runes[end:]
	}
	return lines
}

// View renders the detail view.
func (d DetailView) View() string {
	var sb strings.Builder

	title := fmt.Sprintf("── Detail: %s ", d.Title)
	sb.WriteString(StyleTitle.Render(title))
	sb.WriteString("\n")

	displayLines := d.getDisplayLines()
	visible := d.visibleLines()

	for i := d.Offset; i < len(displayLines) && i < d.Offset+visible; i++ {
		sb.WriteString(displayLines[i])
		sb.WriteString("\n")
	}

	rendered := min(len(displayLines)-d.Offset, visible)
	for i := rendered; i < visible; i++ {
		sb.WriteString("\n")
	}

	status := fmt.Sprintf("%d/%d lines", min(d.Offset+visible, len(displayLines)), len(displayLines))
	sb.WriteString(StyleDim.Render(status))

	return sb.String()
}
