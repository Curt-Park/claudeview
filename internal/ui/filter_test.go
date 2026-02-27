package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestFilterActivateDeactivate(t *testing.T) {
	f := ui.FilterModel{}
	f.Activate()
	if !f.Active {
		t.Error("expected Active=true after Activate()")
	}
	if f.Input != "" {
		t.Error("expected Input cleared on Activate()")
	}
	f.Input = "foo"
	f.Deactivate()
	if f.Active {
		t.Error("expected Active=false after Deactivate()")
	}
}

func TestFilterAddCharAndBackspace(t *testing.T) {
	f := ui.FilterModel{}
	f.Activate()
	f.AddChar('f')
	f.AddChar('o')
	f.AddChar('o')
	if f.Input != "foo" {
		t.Errorf("Input = %q, want %q", f.Input, "foo")
	}
	f.Backspace()
	if f.Input != "fo" {
		t.Errorf("Input after Backspace = %q, want %q", f.Input, "fo")
	}
	// Backspace on empty — no-op
	f2 := ui.FilterModel{}
	f2.Backspace()
	if f2.Input != "" {
		t.Error("Backspace on empty should be no-op")
	}
}
