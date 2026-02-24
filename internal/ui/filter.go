package ui

import "strings"

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

// Clear clears the filter but keeps it active.
func (f *FilterModel) Clear() {
	f.Input = ""
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

// Matches returns true if the given string matches the current filter.
func (f *FilterModel) Matches(s string) bool {
	if f.Input == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(f.Input))
}

// View renders the filter bar.
func (f FilterModel) View() string {
	if !f.Active {
		return ""
	}
	return StyleFilter.Width(f.Width).Render("/" + f.Input + "█")
}
