package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestCrumbsPushPop(t *testing.T) {
	c := ui.CrumbsModel{}
	c.Push("projects")
	c.Push("my-app")
	c.Push("sessions")

	if len(c.Items) != 3 {
		t.Fatalf("len = %d, want 3", len(c.Items))
	}
	if c.Items[2] != "sessions" {
		t.Errorf("Items[2] = %q, want %q", c.Items[2], "sessions")
	}

	c.Pop()
	if len(c.Items) != 2 {
		t.Fatalf("len after Pop = %d, want 2", len(c.Items))
	}
	if c.Items[1] != "my-app" {
		t.Errorf("Items[1] = %q, want %q", c.Items[1], "my-app")
	}
}

func TestCrumbsPopEmpty(t *testing.T) {
	c := ui.CrumbsModel{}
	c.Pop() // should not panic
	if len(c.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.Items))
	}
}

func TestCrumbsReset(t *testing.T) {
	c := ui.CrumbsModel{Items: []string{"old", "items"}}
	c.Reset("projects", "my-app")
	if len(c.Items) != 2 {
		t.Fatalf("len after Reset = %d, want 2", len(c.Items))
	}
	if c.Items[0] != "projects" || c.Items[1] != "my-app" {
		t.Errorf("Items = %v, want [projects my-app]", c.Items)
	}
}
