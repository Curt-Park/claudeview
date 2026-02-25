package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// ScrollDown scrolls down.
func (d *DetailView) ScrollDown() {
	maxOff := max(0, len(d.Lines)-d.visibleLines())
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
	d.Offset = max(0, len(d.Lines)-d.visibleLines())
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
		}
	}
	return false
}

// View renders the detail view.
func (d DetailView) View() string {
	var sb strings.Builder

	title := fmt.Sprintf("── Detail: %s ", d.Title)
	sb.WriteString(StyleTitle.Render(title))
	sb.WriteString("\n")

	visible := d.visibleLines()

	for i := d.Offset; i < len(d.Lines) && i < d.Offset+visible; i++ {
		line := d.Lines[i]
		if len(line) > d.Width {
			line = line[:d.Width-1] + "…"
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	rendered := min(len(d.Lines)-d.Offset, visible)
	for i := rendered; i < visible; i++ {
		sb.WriteString("\n")
	}

	status := fmt.Sprintf("%d/%d lines", min(d.Offset+visible, len(d.Lines)), len(d.Lines))
	sb.WriteString(StyleDim.Render(status))

	return sb.String()
}
