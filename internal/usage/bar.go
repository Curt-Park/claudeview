package usage

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorNormal   = lipgloss.Color("82")
	colorWarning  = lipgloss.Color("214")
	colorCritical = lipgloss.Color("196")
	colorEmpty    = lipgloss.Color("240") // slightly lighter than bg so ░ cells are visible
	colorDim      = lipgloss.Color("243")
	colorBg       = lipgloss.Color("238") // matches StyleCrumbs background
)

// RenderBar renders a 2-line (or 3-line if SevenDayOpus is set) progress bar panel.
// Returns "" if data is nil.
func RenderBar(data *Data, stale bool, width int) string {
	if data == nil {
		return ""
	}

	const labelW = 4
	const rightW = 28
	barW := width - labelW - 3 - rightW
	if barW < 8 {
		barW = 8
	}
	if barW > 80 {
		barW = 80
	}

	var rows []string
	if w := data.FiveHour; w != nil {
		rows = append(rows, renderProgressRow("5h", w, stale, labelW, barW, width))
	}
	if w := data.SevenDay; w != nil {
		rows = append(rows, renderProgressRow("7d", w, stale, labelW, barW, width))
	}
	if w := data.SevenDayOpus; w != nil {
		rows = append(rows, renderProgressRow("opus", w, stale, labelW, barW, width))
	}

	if len(rows) == 0 {
		return ""
	}
	return strings.Join(rows, "\n")
}

func renderProgressRow(label string, w *Window, stale bool, labelW, barW, width int) string {
	pct := w.Utilization
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	filled := int(pct / 100.0 * float64(barW))
	empty := barW - filled

	var fg lipgloss.Color
	if stale {
		fg = colorDim
	} else if pct > 95 {
		fg = colorCritical
	} else if pct > 80 {
		fg = colorWarning
	} else {
		fg = colorNormal
	}
	barStyle := lipgloss.NewStyle().Foreground(fg)
	if !stale && pct > 80 {
		barStyle = barStyle.Bold(true)
	}
	emptyStyle := lipgloss.NewStyle().Foreground(colorEmpty)

	bar := barStyle.Render(strings.Repeat("█", filled)) + emptyStyle.Render(strings.Repeat("░", empty))
	pctStr := fmt.Sprintf("%3.0f%%", pct)

	var resetStr string
	if w.ResetsAt != nil {
		resetStr = "   reset in " + formatCountdown(*w.ResetsAt)
	}

	labelPad := strings.Repeat(" ", labelW-lipgloss.Width(label))
	dimStyle := lipgloss.NewStyle().Foreground(colorDim).Background(colorBg)
	bgStyle := lipgloss.NewStyle().Background(colorBg)
	content := barStyle.Render(label) + labelPad + " [" + bar + "] " + barStyle.Render(pctStr) + dimStyle.Render(resetStr)
	return bgStyle.Width(width).Render(content)
}

func formatCountdown(t time.Time) string {
	d := time.Until(t)
	if d <= 0 {
		return "soon"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh %dm", h, m)
	}
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	if h == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, h)
}
