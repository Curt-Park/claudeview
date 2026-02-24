package ui

import (
	"strings"
)

// CrumbsModel holds the breadcrumb navigation state.
type CrumbsModel struct {
	Items []string
	Width int
}

// Push adds a new crumb to the trail.
func (c *CrumbsModel) Push(item string) {
	c.Items = append(c.Items, item)
}

// Pop removes the last crumb.
func (c *CrumbsModel) Pop() {
	if len(c.Items) > 0 {
		c.Items = c.Items[:len(c.Items)-1]
	}
}

// Reset clears and sets the crumbs.
func (c *CrumbsModel) Reset(items ...string) {
	c.Items = items
}

// View renders the breadcrumb bar.
func (c CrumbsModel) View() string {
	if len(c.Items) == 0 {
		return StyleCrumbs.Width(c.Width).Render("")
	}
	line := strings.Join(c.Items, " > ")
	return StyleCrumbs.Width(c.Width).Render(line)
}
