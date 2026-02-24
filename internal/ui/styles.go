package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorYellow   = lipgloss.Color("214")
	colorBlue     = lipgloss.Color("33")
	colorGreen    = lipgloss.Color("82")
	colorPurple   = lipgloss.Color("135")
	colorOrange   = lipgloss.Color("208")
	colorGray     = lipgloss.Color("243")
	colorRed      = lipgloss.Color("196")
	colorWhite    = lipgloss.Color("255")
	colorDimGray  = lipgloss.Color("238")
	colorCyan     = lipgloss.Color("51")
	colorBgSel    = lipgloss.Color("237")
	colorBgHeader = lipgloss.Color("234")

	// Base styles
	StyleNormal = lipgloss.NewStyle()
	StyleBold   = lipgloss.NewStyle().Bold(true)
	StyleDim    = lipgloss.NewStyle().Foreground(colorGray)

	// Status styles
	StyleActive    = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	StyleThinking  = lipgloss.NewStyle().Foreground(colorYellow)
	StyleReading   = lipgloss.NewStyle().Foreground(colorBlue)
	StyleWriting   = lipgloss.NewStyle().Foreground(colorGreen)
	StyleSearching = lipgloss.NewStyle().Foreground(colorPurple)
	StyleExecuting = lipgloss.NewStyle().Foreground(colorOrange)
	StyleDone      = lipgloss.NewStyle().Foreground(colorGray)
	StyleEnded     = lipgloss.NewStyle().Foreground(colorGray)
	StyleError     = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	StyleRunning   = lipgloss.NewStyle().Foreground(colorGreen)

	// Layout styles
	StyleHeader = lipgloss.NewStyle().
			Background(colorBgHeader).
			Foreground(colorWhite).
			Padding(0, 1)

	StyleMenu = lipgloss.NewStyle().
			Background(colorDimGray).
			Foreground(colorWhite).
			Padding(0, 1)

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

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray)

	StyleTitle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	StyleKey = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	StyleKeyDesc = lipgloss.NewStyle().
			Foreground(colorGray)

	StyleCommand = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	StyleFilter = lipgloss.NewStyle().
			Foreground(colorYellow)

	StyleLogTool   = lipgloss.NewStyle().Foreground(colorBlue)
	StyleLogText   = lipgloss.NewStyle().Foreground(colorWhite)
	StyleLogThink  = lipgloss.NewStyle().Foreground(colorPurple)
	StyleLogResult = lipgloss.NewStyle().Foreground(colorGray)
	StyleLogTime   = lipgloss.NewStyle().Foreground(colorDimGray)
)

// StatusStyle returns the lipgloss style for a given status string.
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "active":
		return StyleActive
	case "thinking":
		return StyleThinking
	case "reading":
		return StyleReading
	case "writing":
		return StyleWriting
	case "searching":
		return StyleSearching
	case "executing":
		return StyleExecuting
	case "done", "ended", "completed":
		return StyleDone
	case "error", "failed":
		return StyleError
	case "running":
		return StyleRunning
	default:
		return StyleNormal
	}
}
