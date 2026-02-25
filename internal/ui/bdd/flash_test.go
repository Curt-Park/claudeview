package bdd

// TestFlash — flash message display on resource switch.

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestFlashShownOnCommandSwitch(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Projects(all)[2]"))

	// Switch resource via command
	sendKey(tm, ":")
	for _, ch := range "sessions" {
		sendKey(tm, string(ch))
	}
	tm.Type("\r")

	// Flash message "switched to sessions" should appear
	waitForOutput(t, tm, notContainsStr("Projects(all)"))
}
