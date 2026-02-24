package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestBreadcrumbPushOnDrillDown — regression for Bug 6.
func TestBreadcrumbPushOnDrillDown(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{projectRow("p1")})

	if len(m.Crumbs.Items) != 1 {
		t.Fatalf("initial crumbs=%v, want 1 item", m.Crumbs.Items)
	}

	m = specialKey(m, tea.KeyEnter)
	if len(m.Crumbs.Items) != 2 {
		t.Errorf("after drill-down crumbs=%v, want 2 items", m.Crumbs.Items)
	}
	if m.Crumbs.Items[1] != string(model.ResourceSessions) {
		t.Errorf("crumbs[1]=%q, want %q", m.Crumbs.Items[1], model.ResourceSessions)
	}

	m.Table.SetRows([]ui.Row{sessionRow("aaaaaaaa-0000-0000-0000-000000000000")})
	m = specialKey(m, tea.KeyEnter)
	if len(m.Crumbs.Items) != 3 {
		t.Errorf("after 2nd drill-down crumbs=%v, want 3 items", m.Crumbs.Items)
	}
}

func TestBreadcrumbPopOnBack(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{projectRow("p1")})
	m = specialKey(m, tea.KeyEnter)

	m = specialKey(m, tea.KeyEsc)
	if len(m.Crumbs.Items) != 1 {
		t.Errorf("after Esc crumbs=%v, want 1 item", m.Crumbs.Items)
	}
	if m.Resource != model.ResourceProjects {
		t.Errorf("Resource=%q after back, want projects", m.Resource)
	}
}

func TestCommandSwitchResetsCrumbs(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{projectRow("p1")})
	m = specialKey(m, tea.KeyEnter)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	m = next.(ui.AppModel)
	for _, ch := range "sessions" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}
	m = specialKey(m, tea.KeyEnter)

	if len(m.Crumbs.Items) != 1 {
		t.Errorf("after :sessions crumbs=%v, want 1 item", m.Crumbs.Items)
	}
}
