package usage_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/usage"
)

func TestRenderBar_Normal(t *testing.T) {
	resetsAt := time.Now().Add(4*time.Hour + 9*time.Minute)
	data := &usage.Data{
		FiveHour: &usage.Window{Utilization: 8.0, ResetsAt: &resetsAt},
		SevenDay: &usage.Window{Utilization: 68.0, ResetsAt: &resetsAt},
	}
	out := usage.RenderBar(data, false, 80)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if !strings.Contains(out, "5h") {
		t.Error("expected '5h' label")
	}
	if !strings.Contains(out, "7d") {
		t.Error("expected '7d' label")
	}
	if !strings.Contains(out, "reset in") {
		t.Error("expected 'reset in' countdown label")
	}
}

func TestRenderBar_NilData(t *testing.T) {
	out := usage.RenderBar(nil, false, 80)
	if out != "" {
		t.Errorf("expected empty string for nil data, got %q", out)
	}
}

func TestRenderBar_NoResetTime(t *testing.T) {
	data := &usage.Data{
		FiveHour: &usage.Window{Utilization: 50.0, ResetsAt: nil},
		SevenDay: &usage.Window{Utilization: 20.0, ResetsAt: nil},
	}
	out := usage.RenderBar(data, false, 80)
	if !strings.Contains(out, "5h") {
		t.Error("expected '5h' label even without reset time")
	}
	if strings.Contains(out, "reset in") {
		t.Error("should not show 'reset in' when ResetsAt is nil")
	}
}

func TestRenderBar_Stale(t *testing.T) {
	resetsAt := time.Now().Add(time.Hour)
	data := &usage.Data{
		FiveHour: &usage.Window{Utilization: 30.0, ResetsAt: &resetsAt},
		SevenDay: &usage.Window{Utilization: 10.0, ResetsAt: &resetsAt},
	}
	out := usage.RenderBar(data, true, 80)
	if !strings.Contains(out, "5h") {
		t.Error("expected bar to render even when stale")
	}
}

func TestRenderBar_ProgressFilled(t *testing.T) {
	resetsAt := time.Now().Add(time.Hour)
	data := &usage.Data{
		FiveHour: &usage.Window{Utilization: 100.0, ResetsAt: &resetsAt},
		SevenDay: &usage.Window{Utilization: 0.0, ResetsAt: &resetsAt},
	}
	out := usage.RenderBar(data, false, 80)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestRenderBar_WithOpus(t *testing.T) {
	resetsAt := time.Now().Add(time.Hour)
	data := &usage.Data{
		FiveHour:     &usage.Window{Utilization: 10.0, ResetsAt: &resetsAt},
		SevenDay:     &usage.Window{Utilization: 20.0, ResetsAt: &resetsAt},
		SevenDayOpus: &usage.Window{Utilization: 5.0, ResetsAt: &resetsAt},
	}
	out := usage.RenderBar(data, false, 80)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (5h + 7d + opus), got %d", len(lines))
	}
}
