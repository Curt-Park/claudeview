package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestDrillDownSetsProjectHash — regression for Bug 4.
func TestDrillDownSetsProjectHash(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{
		projectRow("proj-abc123"),
		projectRow("proj-def456"),
	})

	m = key(m, "j")
	m = specialKey(m, tea.KeyEnter)

	if m.SelectedProjectHash != "proj-def456" {
		t.Errorf("SelectedProjectHash=%q, want %q", m.SelectedProjectHash, "proj-def456")
	}
	if m.Resource != model.ResourceSessions {
		t.Errorf("Resource=%q after drill-down, want sessions", m.Resource)
	}
}

func TestDrillDownSetsSessionID(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{
		sessionRow("aaaaaaaa-1111-2222-3333-444444444444"),
		sessionRow("bbbbbbbb-1111-2222-3333-444444444444"),
	})

	m = key(m, "j")
	m = specialKey(m, tea.KeyEnter)

	want := "bbbbbbbb-1111-2222-3333-444444444444"
	if m.SelectedSessionID != want {
		t.Errorf("SelectedSessionID=%q, want %q", m.SelectedSessionID, want)
	}
	if m.Resource != model.ResourceAgents {
		t.Errorf("Resource=%q, want agents", m.Resource)
	}
}

func TestDrillDownSetsAgentID(t *testing.T) {
	m := makeModel(model.ResourceAgents, []ui.Row{
		{Cells: []string{"agent-xyz", "main", "done", "3", "read file"}, Data: &model.Agent{ID: "agent-xyz"}},
	})

	m = specialKey(m, tea.KeyEnter)

	if m.SelectedAgentID != "agent-xyz" {
		t.Errorf("SelectedAgentID=%q, want %q", m.SelectedAgentID, "agent-xyz")
	}
	if m.Resource != model.ResourceTools {
		t.Errorf("Resource=%q, want tools", m.Resource)
	}
}

func TestDrillDownOnEmptyTableIsNoop(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{})
	before := m.Resource

	m = specialKey(m, tea.KeyEnter)

	if m.Resource != before {
		t.Errorf("drill-down on empty table changed resource to %q", m.Resource)
	}
}

func TestNavigateBackClearsContext(t *testing.T) {
	m := makeModel(model.ResourceTools, []ui.Row{})
	m.SelectedAgentID = "agent-xyz"
	m.Crumbs.Items = []string{"projects", "sessions", "agents", "tools"}

	m = specialKey(m, tea.KeyEsc)

	if m.Resource != model.ResourceAgents {
		t.Errorf("Resource=%q, want agents", m.Resource)
	}
	if m.SelectedAgentID != "" {
		t.Errorf("SelectedAgentID=%q, want empty after back", m.SelectedAgentID)
	}
}

func TestNavigateBackFromSessionsClearsProjectHash(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{})
	m.SelectedProjectHash = "proj-abc"
	m.Crumbs.Items = []string{"projects", "sessions"}

	m = specialKey(m, tea.KeyEsc)

	if m.Resource != model.ResourceProjects {
		t.Errorf("Resource=%q, want projects", m.Resource)
	}
	if m.SelectedProjectHash != "" {
		t.Errorf("SelectedProjectHash=%q, want empty after back", m.SelectedProjectHash)
	}
}
