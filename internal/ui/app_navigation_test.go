package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestJKNavigation — regression for Bug 1: cursor resets on every key press.
func TestJKNavigation(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(5))

	m = key(m, "j")
	if m.Table.Selected != 1 {
		t.Errorf("after j: Selected=%d, want 1", m.Table.Selected)
	}
	m = key(m, "j")
	m = key(m, "j")
	if m.Table.Selected != 3 {
		t.Errorf("after 3×j: Selected=%d, want 3", m.Table.Selected)
	}

	m = key(m, "k")
	if m.Table.Selected != 2 {
		t.Errorf("after k: Selected=%d, want 2", m.Table.Selected)
	}
}

func TestJKDoesNotGoOutOfBounds(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))

	m = key(m, "k")
	if m.Table.Selected != 0 {
		t.Errorf("k at top: Selected=%d, want 0", m.Table.Selected)
	}

	m = key(m, "j")
	m = key(m, "j")
	m = key(m, "j")
	m = key(m, "j")
	if m.Table.Selected != 2 {
		t.Errorf("j past bottom: Selected=%d, want 2", m.Table.Selected)
	}
}

func TestGotoTopBottom(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(10))

	for range 5 {
		m = key(m, "j")
	}
	if m.Table.Selected != 5 {
		t.Fatalf("setup: Selected=%d, want 5", m.Table.Selected)
	}

	m = key(m, "G")
	if m.Table.Selected != 9 {
		t.Errorf("G: Selected=%d, want 9", m.Table.Selected)
	}

	m = key(m, "g")
	if m.Table.Selected != 0 {
		t.Errorf("g: Selected=%d, want 0", m.Table.Selected)
	}
}

func TestPageUpDown(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(20))
	m.Table.Height = 10

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = next.(ui.AppModel)
	if m.Table.Selected < 4 {
		t.Errorf("ctrl+d: Selected=%d, want >=4", m.Table.Selected)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = next.(ui.AppModel)
	if m.Table.Selected > 1 {
		t.Errorf("ctrl+u: Selected=%d, want <=1", m.Table.Selected)
	}
}
