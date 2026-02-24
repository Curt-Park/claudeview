package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestCommandActivateDeactivate(t *testing.T) {
	c := ui.CommandModel{}
	c.Activate()
	if !c.Active {
		t.Error("expected Active=true after Activate()")
	}
	if c.Input != "" || c.Error != "" || c.Suggestion != "" {
		t.Error("expected Input/Error/Suggestion cleared on Activate()")
	}

	c.Input = "foo"
	c.Deactivate()
	if c.Active {
		t.Error("expected Active=false after Deactivate()")
	}
	if c.Input != "" {
		t.Error("expected Input cleared on Deactivate()")
	}
}

func TestCommandAddCharAndBackspace(t *testing.T) {
	c := ui.CommandModel{}
	c.Activate()
	c.AddChar('s')
	c.AddChar('e')
	c.AddChar('s')
	if c.Input != "ses" {
		t.Errorf("Input = %q, want %q", c.Input, "ses")
	}
	c.Backspace()
	if c.Input != "se" {
		t.Errorf("Input after Backspace = %q, want %q", c.Input, "se")
	}
}

func TestCommandBackspaceEmpty(t *testing.T) {
	c := ui.CommandModel{}
	c.Activate()
	c.Backspace() // should be no-op
	if c.Input != "" {
		t.Errorf("Input = %q after Backspace on empty, want %q", c.Input, "")
	}
}

func TestCommandSuggestion(t *testing.T) {
	c := ui.CommandModel{}
	c.Activate()
	c.AddChar('s')
	c.AddChar('e')
	c.AddChar('s')
	if c.Suggestion == "" {
		t.Error("expected a suggestion for prefix 'ses'")
	}

	// Full name typed — no suggestion needed
	c2 := ui.CommandModel{}
	c2.Activate()
	for _, ch := range "sessions" {
		c2.AddChar(ch)
	}
	if c2.Suggestion != "" {
		t.Errorf("expected no suggestion for complete name, got %q", c2.Suggestion)
	}

	// Unrecognized prefix
	c3 := ui.CommandModel{}
	c3.Activate()
	c3.AddChar('z')
	if c3.Suggestion != "" {
		t.Errorf("expected no suggestion for 'z', got %q", c3.Suggestion)
	}
}

func TestCommandAccept(t *testing.T) {
	c := ui.CommandModel{}
	c.Activate()
	c.AddChar('s')
	c.AddChar('e')
	c.AddChar('s')
	got := c.Accept()
	if got != c.Input {
		t.Errorf("Accept() = %q, want %q", got, c.Input)
	}
}

func TestCommandSubmitValid(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  model.ResourceType
	}{
		{"sessions", model.ResourceSessions},
		{"s", model.ResourceSessions},
		{"agents", model.ResourceAgents},
		{"a", model.ResourceAgents},
		{"tools", model.ResourceTools},
		{"plugins", model.ResourcePlugins},
		{"mcp", model.ResourceMCP},
	} {
		c := ui.CommandModel{Input: tc.input}
		rt, ok := c.Submit()
		if !ok {
			t.Errorf("Submit(%q): expected ok=true", tc.input)
		}
		if rt != tc.want {
			t.Errorf("Submit(%q) = %q, want %q", tc.input, rt, tc.want)
		}
	}
}

func TestCommandSubmitInvalid(t *testing.T) {
	c := ui.CommandModel{Input: "nonexistent"}
	_, ok := c.Submit()
	if ok {
		t.Error("Submit(nonexistent): expected ok=false")
	}
	if c.Error == "" {
		t.Error("expected Error to be set on invalid submit")
	}
}
