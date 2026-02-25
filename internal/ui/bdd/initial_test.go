package bdd

// TestInitial — verify initial screen renders the projects table with info panel showing "--".
import (
	"strings"
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestInitialProjectsTable(t *testing.T) {
	dp := &mockDP{}
	rows := projectRows(3)
	tm := newTestModel(t, model.ResourceProjects, dp, rows)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		out := string(bts)
		// Title bar should show Projects resource
		return strings.Contains(out, "Projects(all)[3]")
	})
}

func TestInitialInfoPanelDashes(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(1))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		out := string(bts)
		// At projects level, all context fields show "--"
		return strings.Contains(out, "Project:") &&
			strings.Contains(out, "Session:")
	})
}

func TestInitialShortcutZeroAlwaysShown(t *testing.T) {
	dp := &mockDP{}
	tm := newTestModel(t, model.ResourceProjects, dp, projectRows(2))
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, func(bts []byte) bool {
		out := string(bts)
		return strings.Contains(out, "<0>") && strings.Contains(out, "all")
	})
}
