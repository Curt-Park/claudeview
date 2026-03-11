package model

import (
	"encoding/json"
	"strings"
	"time"
)

// AgentType classifies the type of agent.
type AgentType string

const (
	AgentTypeMain    AgentType = "main"
	AgentTypeExplore AgentType = "Explore"
	AgentTypePlan    AgentType = "Plan"
	AgentTypeBash    AgentType = "Bash"
	AgentTypeGeneral AgentType = "general-purpose"
)

// DisplayLabel returns the human-readable display label for the agent type.
func (t AgentType) DisplayLabel() string {
	switch t {
	case AgentTypeExplore:
		return "Explorer"
	case AgentTypePlan:
		return "Planner"
	case AgentTypeBash:
		return "Bash"
	case AgentTypeGeneral:
		return "Agent"
	default:
		s := string(t)
		if i := strings.LastIndex(s, ":"); i >= 0 {
			s = s[i+1:]
		}
		if s == "" {
			return "Agent"
		}
		return s
	}
}

// Icon returns the icon character for the agent type.
func (t AgentType) Icon() string {
	switch t {
	case AgentTypeExplore:
		return "🔍"
	case AgentTypePlan:
		return "📋"
	case AgentTypeBash:
		return "💻"
	default:
		return "⚙️"
	}
}

// Agent represents a Claude Code agent (main or subagent).
type Agent struct {
	ID           string
	SessionID    string
	Type         AgentType
	Status       Status
	ToolCalls    []*ToolCall
	LastActivity string
	FilePath     string
	StartTime    time.Time
	IsSubagent   bool
	Depth        int // tree depth for display
}

// ShortID returns a display-friendly ID.
func (a *Agent) ShortID() string {
	if a.ID == "" {
		return "main"
	}
	// agent-a42f831 -> a42f831
	parts := strings.SplitN(a.ID, "-", 2)
	if len(parts) == 2 {
		id := parts[1]
		if len(id) > 7 {
			return id[:7]
		}
		return id
	}
	if len(a.ID) > 8 {
		return a.ID[:8]
	}
	return a.ID
}

// DisplayName returns a human-friendly name for the agent.
func (a *Agent) DisplayName() string {
	switch a.Type {
	case AgentTypeMain:
		return "Claude"
	case AgentTypeExplore, AgentTypePlan, AgentTypeBash, AgentTypeGeneral:
		return a.Type.DisplayLabel()
	default:
		if a.ID != "" {
			return a.ShortID()
		}
		return a.Type.DisplayLabel()
	}
}

// TreePrefix returns the tree-drawing prefix for display.
func (a *Agent) TreePrefix(isLast bool) string {
	if !a.IsSubagent {
		return "► "
	}
	if isLast {
		return "  └─ "
	}
	return "  ├─ "
}

// AgentTypeFromInput reads the subagent_type field from a tool call's JSON input.
func AgentTypeFromInput(input json.RawMessage) AgentType {
	if input == nil {
		return AgentTypeGeneral
	}
	var m map[string]any
	if err := json.Unmarshal(input, &m); err != nil {
		return AgentTypeGeneral
	}
	v, ok := m["subagent_type"].(string)
	if !ok || v == "" {
		return AgentTypeGeneral
	}
	return AgentType(v)
}
