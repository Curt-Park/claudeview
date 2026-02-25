package ui

import (
	"strings"

	"github.com/Curt-Park/claudeview/internal/model"
)

// CommandModel handles the `:` command mode with autocomplete.
type CommandModel struct {
	Active     bool
	Input      string
	Suggestion string
	Error      string
	Width      int
}

// Activate starts command mode.
func (c *CommandModel) Activate() {
	c.Active = true
	c.Input = ""
	c.Error = ""
	c.Suggestion = ""
}

// Deactivate stops command mode.
func (c *CommandModel) Deactivate() {
	c.Active = false
	c.Input = ""
	c.Error = ""
	c.Suggestion = ""
}

// AddChar adds a character to the input.
func (c *CommandModel) AddChar(ch rune) {
	c.Input += string(ch)
	c.updateSuggestion()
}

// Backspace removes the last character.
func (c *CommandModel) Backspace() {
	if len(c.Input) > 0 {
		c.Input = c.Input[:len(c.Input)-1]
		c.updateSuggestion()
	}
}

// Accept tries to accept the current suggestion.
func (c *CommandModel) Accept() string {
	if c.Suggestion != "" {
		c.Input = c.Suggestion
	}
	return c.Input
}

// Submit parses the input and returns the resolved resource type (or "").
func (c *CommandModel) Submit() (model.ResourceType, bool) {
	input := strings.TrimSpace(c.Input)
	rt, ok := model.ResolveResource(input)
	if !ok {
		c.Error = "unknown resource: " + input
	}
	return rt, ok
}

func (c *CommandModel) updateSuggestion() {
	if c.Input == "" {
		c.Suggestion = ""
		return
	}
	for _, name := range model.AllResourceNames() {
		if strings.HasPrefix(name, c.Input) && name != c.Input {
			c.Suggestion = name
			return
		}
	}
	c.Suggestion = ""
}

// View renders the command bar.
func (c CommandModel) View() string {
	if !c.Active {
		return ""
	}
	if c.Error != "" {
		return StyleFlashError.Width(c.Width).Render(":" + c.Error)
	}
	display := ":" + c.Input
	if c.Suggestion != "" {
		// Show ghost suggestion
		ghost := StyleDim.Render(c.Suggestion[len(c.Input):])
		display += ghost
	}
	return StyleCommand.Width(c.Width).Render(display)
}
