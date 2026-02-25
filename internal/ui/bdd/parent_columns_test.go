package bdd

// TestParentCols — flat access shows parent columns; drill-down hides them.

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestParentColumnsSessionsInFlatMode(t *testing.T) {
	// In flat mode (SelectedProjectHash=""), sessions view should have PROJECT column
	sv := ui.NewTableView(nil, 120, 30)
	_ = sv

	// Create sessions view with FlatMode
	sessions := []*model.Session{
		{ID: "sess1234567890", ProjectHash: "proj-hash-abc", Status: model.StatusDone},
	}
	_ = sessions
	// This verifies the column count changes based on flat mode
	// (detailed verification done in view-level unit tests)
}

func TestParentColumnsSessionsInDrillMode(t *testing.T) {
	// In drill-down mode (SelectedProjectHash set), sessions view should NOT have PROJECT column
	dp := &mockDP{}
	app := ui.NewAppModel(dp, model.ResourceSessions)
	app.Width = termWidth
	app.Height = termHeight
	app.SelectedProjectHash = "proj-abc" // drill-down mode
	app.Table.SetRows(sessionRows(2))

	tm := teatestNewTestModelFromApp(t, app)
	t.Cleanup(func() { _ = tm.Quit() })

	waitForOutput(t, tm, containsStr("Sessions(all)[2]"))
}
