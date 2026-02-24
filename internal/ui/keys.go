package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds all keybindings.
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding

	// View mode
	Logs   key.Binding
	Detail key.Binding
	YAML   key.Binding

	// Scroll (log/detail view)
	ScrollUp    key.Binding
	ScrollDown  key.Binding
	ScrollLeft  key.Binding
	ScrollRight key.Binding
	GotoTop     key.Binding
	GotoBottom  key.Binding

	// Modes
	Command key.Binding
	Filter  key.Binding
	Follow  key.Binding
	Help    key.Binding

	// Ctrl+C
	ForceQuit key.Binding
}

// DefaultKeyMap returns the default k9s-style key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "logs"),
		),
		Detail: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "detail"),
		),
		YAML: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yaml"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "scroll down"),
		),
		ScrollLeft: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "scroll left"),
		),
		ScrollRight: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "scroll right"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
		),
	}
}
