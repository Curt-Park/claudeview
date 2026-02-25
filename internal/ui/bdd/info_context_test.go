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

func TestInfoPanelShortcutZeroAlwaysPresent(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceSessions, dp, sessionRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	// <0> all shortcut should always be shown in info panel right column
	waitForOutput(t, tm, func(bts []byte) bool {
		return containsStr("<0>")(bts) && containsStr("all")(bts)
	})
}
