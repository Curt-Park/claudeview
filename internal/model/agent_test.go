package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestAgentShortID(t *testing.T) {
	tests := []struct {
		name  string
		agent *model.Agent
		want  string
	}{
		{"standard agent ID with dash", &model.Agent{ID: "agent-abc1234"}, "abc1234"},
		{"empty ID returns main", &model.Agent{ID: ""}, "main"},
		{"ID with no dash short enough", &model.Agent{ID: "abcdefghij"}, "abcdefgh"},
		{"long ID after dash truncated to 7", &model.Agent{ID: "agent-abcdefghijk"}, "abcdefg"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.agent.ShortID(); got != tc.want {
				t.Errorf("ShortID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAgentDisplayName(t *testing.T) {
	tests := []struct {
		name  string
		agent *model.Agent
		want  string
	}{
		{"main", &model.Agent{Type: model.AgentTypeMain}, "Claude"},
		{"Explore", &model.Agent{Type: model.AgentTypeExplore}, "Explorer"},
		{"Plan", &model.Agent{Type: model.AgentTypePlan}, "Planner"},
		{"Bash", &model.Agent{Type: model.AgentTypeBash}, "Bash-runner"},
		{"general-purpose", &model.Agent{Type: model.AgentTypeGeneral}, "General"},
		{"unknown with ID", &model.Agent{Type: "unknown", ID: "agent-abc1234"}, "abc1234"},
		{"unknown empty ID", &model.Agent{Type: "unknown", ID: ""}, "Agent"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.agent.DisplayName(); got != tc.want {
				t.Errorf("DisplayName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAgentTreePrefix(t *testing.T) {
	tests := []struct {
		name   string
		agent  *model.Agent
		isLast bool
		want   string
	}{
		{"main agent not subagent", &model.Agent{IsSubagent: false}, false, "\u25ba "},
		{"subagent isLast true", &model.Agent{IsSubagent: true}, true, "  \u2514\u2500 "},
		{"subagent isLast false", &model.Agent{IsSubagent: true}, false, "  \u251c\u2500 "},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.agent.TreePrefix(tc.isLast); got != tc.want {
				t.Errorf("TreePrefix(%v) = %q, want %q", tc.isLast, got, tc.want)
			}
		})
	}
}
