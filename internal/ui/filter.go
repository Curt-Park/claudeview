package ui

// FilterModel handles the `/` filter mode.
type FilterModel struct {
	Active bool
	Input  string
	Width  int
}

// Activate starts filter mode.
func (f *FilterModel) Activate() {
	f.Active = true
	f.Input = ""
}

// Deactivate stops filter mode.
func (f *FilterModel) Deactivate() {
	f.Active = false
}

// AddChar adds a character to the filter input.
func (f *FilterModel) AddChar(ch rune) {
	f.Input += string(ch)
}

// Backspace removes the last character.
func (f *FilterModel) Backspace() {
	if len(f.Input) > 0 {
		f.Input = f.Input[:len(f.Input)-1]
	}
}

// Height returns the number of terminal lines rendered by View.
func (f FilterModel) Height() int { return 1 }

// View renders the filter bar.
func (f FilterModel) View() string {
	if !f.Active {
		return ""
	}
	return StyleFilter.Width(f.Width).Render("/" + f.Input + "█")
}
