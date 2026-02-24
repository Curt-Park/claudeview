package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestDKeyEntersModeDetail(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "d")
	if m.ViewMode != ui.ModeDetail {
		t.Errorf("ViewMode=%v after d, want ModeDetail", m.ViewMode)
	}
}

func TestLKeyEntersModeLog(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "l")
	if m.ViewMode != ui.ModeLog {
		t.Errorf("ViewMode=%v after l, want ModeLog", m.ViewMode)
	}
}

func TestYKeyEntersModeYAML(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "y")
	if m.ViewMode != ui.ModeYAML {
		t.Errorf("ViewMode=%v after y, want ModeYAML", m.ViewMode)
	}
}

func TestEscReturnsModeTable(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(ui.AppModel) ui.AppModel
	}{
		{"detail", func(m ui.AppModel) ui.AppModel { return key(m, "d") }},
		{"log", func(m ui.AppModel) ui.AppModel { return key(m, "l") }},
		{"yaml", func(m ui.AppModel) ui.AppModel { return key(m, "y") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(makeModel(model.ResourceSessions, nRows(3)))
			m = specialKey(m, tea.KeyEsc)
			if m.ViewMode != ui.ModeTable {
				t.Errorf("ViewMode=%v after Esc, want ModeTable", m.ViewMode)
			}
		})
	}
}

// TestDetailViewHLNotVerticalScroll — regression for Bug 8.
func TestDetailViewHLNotVerticalScroll(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "d")

	m.Detail.Lines = make([]string, 50)
	for i := range m.Detail.Lines {
		m.Detail.Lines[i] = "line"
	}
	m.Detail.Offset = 5

	before := m.Detail.Offset
	m = key(m, "h")
	if m.Detail.Offset != before {
		t.Errorf("h in detail changed Offset %d→%d (should be no-op)", before, m.Detail.Offset)
	}
	m = key(m, "l")
	if m.Detail.Offset != before {
		t.Errorf("l in detail changed Offset %d→%d (should be no-op)", before, m.Detail.Offset)
	}
}
