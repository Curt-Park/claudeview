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

	if withUsage != base+2 {
		t.Errorf("expected height %d with 2-line usage, got %d", base+2, withUsage)
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
}

func TestInfoModelViewNoUsageLine(t *testing.T) {
	info := ui.InfoModel{Width: 80}
	menu := ui.MenuModel{}
	out := info.ViewWithMenu(menu)
	if strings.HasPrefix(out, "\n") {
		t.Error("should not start with newline when UsageLine is empty")
	}
}
