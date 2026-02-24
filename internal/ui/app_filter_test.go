package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestFilterWiredToTable — regression for Bug 5.
func TestFilterWiredToTable(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(5))

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(ui.AppModel)

	for _, ch := range "foo" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}

	if m.Table.Filter != "foo" {
		t.Errorf("Table.Filter=%q, want %q", m.Table.Filter, "foo")
	}
	if m.Log.Filter != "foo" {
		t.Errorf("Log.Filter=%q, want %q", m.Log.Filter, "foo")
	}
}

func TestFilterClearedOnEsc(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(5))

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(ui.AppModel)
	for _, ch := range "bar" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}

	m = specialKey(m, tea.KeyEsc)

	if m.Table.Filter != "" {
		t.Errorf("Table.Filter=%q after Esc, want empty", m.Table.Filter)
	}
}

func TestFilterBackspace(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(ui.AppModel)
	for _, ch := range "abc" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(ui.AppModel)

	if m.Table.Filter != "ab" {
		t.Errorf("Table.Filter=%q after backspace, want %q", m.Table.Filter, "ab")
	}
}
