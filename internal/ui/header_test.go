package ui_test

import (
	"strings"
	"testing"

	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestInfoModelHeightWithUsageLine(t *testing.T) {
	info := ui.InfoModel{}
	base := info.Height(4, 3, 1)

	info.UsageLine = "line1\nline2" // 2-line usage bar
	withUsage := info.Height(4, 3, 1)

	if withUsage != base+3 { // 2 usage lines + 1 separator
		t.Errorf("expected height %d with 2-line usage + separator, got %d", base+3, withUsage)
	}
}

func TestInfoModelViewWithUsageLine(t *testing.T) {
	info := ui.InfoModel{
		UsageLine: "USAGE_BAR_LINE",
		Width:     80,
	}
	menu := ui.MenuModel{}
	out := info.ViewWithMenu(menu)
	lines := strings.Split(out, "\n")
	if lines[0] != "USAGE_BAR_LINE" {
		t.Errorf("expected usage line as first line, got %q", lines[0])
	}
	// Line 1 (index 1) should be the separator: a background-colored blank line.
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines (usage + separator + info), got %d", len(lines))
	}
}

func TestInfoModelViewNoUsageLine(t *testing.T) {
	info := ui.InfoModel{Width: 80}
	menu := ui.MenuModel{}
	out := info.ViewWithMenu(menu)
	if strings.HasPrefix(out, "\n") {
		t.Error("should not start with newline when UsageLine is empty")
	}
}
