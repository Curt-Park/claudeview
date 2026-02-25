package bdd

// TestInfoContext — info panel context values change with navigation depth.

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestInfoPanelAtProjectsLevelShowsDashes(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	// At projects level, context fields should show "--"
	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("Project:")(bts) && containsStr("Session:")(bts)
	})
}

func TestInfoPanelMenuShownForSessions(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	// Sessions view should show enter/logs/detail menu items
	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("<enter>")(bts) && containsStr("<l>")(bts) && containsStr("<d>")(bts)
	})
}
