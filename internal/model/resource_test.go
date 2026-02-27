package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestResolveResource(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      model.ResourceType
		wantFound bool
	}{
		{"alias p", "p", model.ResourceProjects, true},
		{"alias s", "s", model.ResourceSessions, true},
		{"alias a", "a", model.ResourceAgents, true},
		{"alias tk", "tk", model.ResourceTasks, true},
		{"alias pl", "pl", model.ResourcePlugins, true},
		{"alias m", "m", model.ResourceMCP, true},
		{"full name projects", "projects", model.ResourceProjects, true},
		{"full name sessions", "sessions", model.ResourceSessions, true},
		{"full name agents", "agents", model.ResourceAgents, true},
		{"full name tasks", "tasks", model.ResourceTasks, true},
		{"full name plugins", "plugins", model.ResourcePlugins, true},
		{"full name mcp", "mcp", model.ResourceMCP, true},
		{"invalid name", "foobar", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := model.ResolveResource(tc.input)
			if ok != tc.wantFound {
				t.Errorf("ResolveResource(%q) found = %v, want %v", tc.input, ok, tc.wantFound)
			}
			if got != tc.want {
				t.Errorf("ResolveResource(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestAllResourceNames(t *testing.T) {
	names := model.AllResourceNames()
	if len(names) == 0 {
		t.Fatal("AllResourceNames() returned empty slice")
	}

	expected := []string{"projects", "sessions", "agents", "tasks", "plugins", "mcp"}
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	for _, exp := range expected {
		if !nameSet[exp] {
			t.Errorf("AllResourceNames() missing expected name %q", exp)
		}
	}
}
