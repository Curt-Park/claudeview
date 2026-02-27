package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/Curt-Park/claudeview/internal/model"
)

var (
	// Colors
	colorYellow  = lipgloss.Color("214")
	colorBlue    = lipgloss.Color("33")
	colorGreen   = lipgloss.Color("82")
	colorPurple  = lipgloss.Color("135")
	colorOrange  = lipgloss.Color("208")
	colorGray    = lipgloss.Color("243")
	colorRed     = lipgloss.Color("196")
	colorWhite   = lipgloss.Color("255")
	colorDimGray = lipgloss.Color("238")
	colorCyan    = lipgloss.Color("51")
	colorBgSel   = lipgloss.Color("237")

	// Base styles
	StyleNormal = lipgloss.NewStyle()
	StyleDim    = lipgloss.NewStyle().Foreground(colorGray)

	// Status styles
	StyleActive    = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	StyleThinking  = lipgloss.NewStyle().Foreground(colorYellow)
	StyleReading   = lipgloss.NewStyle().Foreground(colorBlue)
	StyleWriting   = lipgloss.NewStyle().Foreground(colorGreen)
	StyleSearching = lipgloss.NewStyle().Foreground(colorPurple)
	StyleExecuting = lipgloss.NewStyle().Foreground(colorOrange)
	StyleDone      = lipgloss.NewStyle().Foreground(colorGray)
	StyleError     = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	StyleRunning   = lipgloss.NewStyle().Foreground(colorGreen)

	// Layout styles
	StyleCrumbs = lipgloss.NewStyle().
			Background(colorDimGray).
			Foreground(colorCyan).
			Padding(0, 1)

	StyleFlash = lipgloss.NewStyle().
			Foreground(colorYellow).
			Padding(0, 1)

	StyleFlashError = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true).
			Padding(0, 1)

	StyleSelected = lipgloss.NewStyle().
			Background(colorBgSel).
			Foreground(colorWhite).
			Bold(true)

	StyleColumnHeader = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	StyleTitle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	StyleKey = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	StyleKeyDesc = lipgloss.NewStyle().
			Foreground(colorGray)

	StyleKeyHighlight = lipgloss.NewStyle().
				Foreground(colorWhite).
				Bold(true)

	StyleHotRow = lipgloss.NewStyle().Foreground(colorYellow)

	StyleFilter = lipgloss.NewStyle().
			Foreground(colorYellow)

	StyleLogTool   = lipgloss.NewStyle().Foreground(colorBlue)
	StyleLogText   = lipgloss.NewStyle().Foreground(colorWhite)
	StyleLogThink  = lipgloss.NewStyle().Foreground(colorPurple)
	StyleLogResult = lipgloss.NewStyle().Foreground(colorGray)
	StyleLogTime   = lipgloss.NewStyle().Foreground(colorDimGray)

	// StyleRowSubtitle is used for the optional second line of a table row.
	StyleRowSubtitle         = lipgloss.NewStyle().Foreground(colorGray)
	StyleRowSubtitleSelected = lipgloss.NewStyle().Background(colorBgSel).Foreground(colorGray)
)

// StatusStyle returns the lipgloss style for a given status.
func StatusStyle(status model.Status) lipgloss.Style {
	switch status {
	case model.StatusActive:
		return StyleActive
	case model.StatusThinking:
		return StyleThinking
	case model.StatusReading:
		return StyleReading
	case model.StatusExecuting:
		return StyleExecuting
	case model.StatusDone, model.StatusEnded, model.StatusCompleted:
		return StyleDone
	case model.StatusError, model.StatusFailed:
		return StyleError
	case model.StatusRunning:
		return StyleRunning
	default:
		return StyleNormal
	}
}
