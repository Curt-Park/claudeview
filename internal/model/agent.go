package model

import (
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
	case AgentTypeExplore:
		return "Explorer"
	case AgentTypePlan:
		return "Planner"
	case AgentTypeBash:
		return "Bash-runner"
	case AgentTypeGeneral:
		return "General"
	default:
		if a.ID != "" {
			return a.ShortID()
		}
		return "Agent"
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
