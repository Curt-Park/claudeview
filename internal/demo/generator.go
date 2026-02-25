package demo

import (
	"fmt"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

// GenerateProjects creates synthetic demo projects.
func GenerateProjects() []*model.Project {
	now := time.Now()

	sessions1 := []*model.Session{
		{
			ID:           "abc12345-demo-0001-0000-000000000001",
			ProjectHash:  "demo-project-1",
			Model:        "claude-opus-4-6",
			Status:       model.StatusActive,
			TotalCost:    0.4234,
			InputTokens:  120000,
			OutputTokens: 25200,
			NumTurns:     12,
			StartTime:    now.Add(-5 * time.Minute),
			ModTime:      now.Add(-30 * time.Second),
			Agents:       generateAgents("abc12345"),
		},
		{
			ID:           "def45678-demo-0002-0000-000000000002",
			ProjectHash:  "demo-project-1",
			Model:        "claude-sonnet-4-6",
			Status:       model.StatusEnded,
			TotalCost:    0.0823,
			InputTokens:  20000,
			OutputTokens: 3100,
			NumTurns:     5,
			StartTime:    now.Add(-2 * time.Hour),
			EndTime:      now.Add(-90 * time.Minute),
			ModTime:      now.Add(-90 * time.Minute),
			Agents:       generateAgents("def45678"),
		},
	}

	sessions2 := []*model.Session{
		{
			ID:           "ghi78901-demo-0003-0000-000000000003",
			ProjectHash:  "demo-project-2",
			Model:        "claude-haiku-4-5",
			Status:       model.StatusEnded,
			TotalCost:    0.0123,
			InputTokens:  5000,
			OutputTokens: 800,
			NumTurns:     3,
			StartTime:    now.Add(-24 * time.Hour),
			EndTime:      now.Add(-23 * time.Hour),
			ModTime:      now.Add(-23 * time.Hour),
			Agents:       generateAgents("ghi78901"),
		},
	}

	return []*model.Project{
		{
			Hash:     "-Users-mac-Repositories-my-awesome-app",
			Path:     "/Users/mac/.claude/projects/-Users-mac-Repositories-my-awesome-app",
			Sessions: sessions1,
			LastSeen: now.Add(-30 * time.Second),
		},
		{
			Hash:     "-Users-mac-Repositories-another-project",
			Path:     "/Users/mac/.claude/projects/-Users-mac-Repositories-another-project",
			Sessions: sessions2,
			LastSeen: now.Add(-23 * time.Hour),
		},
	}
}

func generateAgents(sessionID string) []*model.Agent {
	now := time.Now()
	tools1 := generateToolCalls(sessionID, "main", 12)
	tools2 := generateToolCalls(sessionID, "sub1", 5)
	tools3 := generateToolCalls(sessionID, "sub2", 3)
	tools4 := generateToolCalls(sessionID, "sub3", 2)

	return []*model.Agent{
		{
			ID:           "",
			SessionID:    sessionID,
			Type:         model.AgentTypeMain,
			Status:       model.StatusThinking,
			ToolCalls:    tools1,
			LastActivity: "Edit src/app.py",
			StartTime:    now.Add(-5 * time.Minute),
			IsSubagent:   false,
		},
		{
			ID:           fmt.Sprintf("agent-%s-sub1", sessionID[:4]),
			SessionID:    sessionID,
			Type:         model.AgentTypeExplore,
			Status:       model.StatusReading,
			ToolCalls:    tools2,
			LastActivity: "Read src/config.py",
			StartTime:    now.Add(-4 * time.Minute),
			IsSubagent:   true,
		},
		{
			ID:           fmt.Sprintf("agent-%s-sub2", sessionID[:4]),
			SessionID:    sessionID,
			Type:         model.AgentTypePlan,
			Status:       model.StatusDone,
			ToolCalls:    tools3,
			LastActivity: "Read CLAUDE.md",
			StartTime:    now.Add(-3 * time.Minute),
			IsSubagent:   true,
		},
		{
			ID:           fmt.Sprintf("agent-%s-sub3", sessionID[:4]),
			SessionID:    sessionID,
			Type:         model.AgentTypeBash,
			Status:       model.StatusExecuting,
			ToolCalls:    tools4,
			LastActivity: "Bash: npm test",
			StartTime:    now.Add(-2 * time.Minute),
			IsSubagent:   true,
		},
	}
}

func generateToolCalls(sessionID, agentID string, count int) []*model.ToolCall {
	tools := []struct {
		name   string
		input  string
		result string
	}{
		{"Read", `{"file_path": "src/app.py"}`, `"142 lines"`},
		{"Grep", `{"pattern": "handleAuth", "path": "src/"}`, `"3 matches"`},
		{"Bash", `{"command": "npm test"}`, `"exit 0"`},
		{"Edit", `{"file_path": "src/app.py"}`, `"success"`},
		{"Glob", `{"pattern": "**/*.py"}`, `"12 files"`},
		{"Write", `{"file_path": "src/new.py"}`, `"success"`},
		{"Read", `{"file_path": "CLAUDE.md"}`, `"45 lines"`},
		{"Task", `{"description": "Explore the codebase"}`, `"done"`},
	}

	now := time.Now()
	var calls []*model.ToolCall
	for i := 0; i < count && i < len(tools); i++ {
		tc := tools[i%len(tools)]
		calls = append(calls, &model.ToolCall{
			ID:        fmt.Sprintf("%s-tool-%d", agentID, i),
			SessionID: sessionID,
			AgentID:   agentID,
			Name:      tc.name,
			Input:     []byte(tc.input),
			Result:    []byte(tc.result),
			Timestamp: now.Add(time.Duration(-count+i) * 30 * time.Second),
			Duration:  time.Duration(100+i*50) * time.Millisecond,
		})
	}
	return calls
}

// GenerateTasks creates synthetic demo tasks.
func GenerateTasks(sessionID string) []*model.Task {
	return []*model.Task{
		{ID: "1", SessionID: sessionID, Subject: "Explore project context", Status: model.StatusCompleted},
		{ID: "2", SessionID: sessionID, Subject: "Ask clarifying questions", Status: model.StatusCompleted, BlockedBy: []string{"1"}},
		{ID: "3", SessionID: sessionID, Subject: "Propose approaches", Status: model.StatusInProgress, BlockedBy: []string{"2"}},
		{ID: "4", SessionID: sessionID, Subject: "Present design", Status: model.StatusPending, BlockedBy: []string{"3"}},
	}
}

// GeneratePlugins creates synthetic demo plugins.
func GeneratePlugins() []*model.Plugin {
	return []*model.Plugin{
		{
			Name:         "superpowers",
			Version:      "4.3.1",
			Marketplace:  "claude-plugins-official",
			Enabled:      true,
			InstalledAt:  "2025-12-15",
			SkillCount:   15,
			CommandCount: 2,
			HookCount:    1,
		},
		{
			Name:         "Notion",
			Version:      "1.2.0",
			Marketplace:  "claude-plugins-official",
			Enabled:      true,
			InstalledAt:  "2025-11-20",
			SkillCount:   8,
			CommandCount: 1,
		},
		{
			Name:        "code-review",
			Version:     "2.0.1",
			Marketplace: "claude-plugins-official",
			Enabled:     false,
			InstalledAt: "2025-10-01",
			SkillCount:  3,
		},
	}
}

// GenerateMCPServers creates synthetic demo MCP servers.
func GenerateMCPServers() []*model.MCPServer {
	return []*model.MCPServer{
		{
			Name:      "filesystem",
			Plugin:    "superpowers",
			Transport: "stdio",
			Command:   "npx",
			Args:      []string{"@anthropic/mcp-fs"},
			ToolCount: 8,
			Status:    model.StatusRunning,
		},
		{
			Name:      "github",
			Plugin:    "Notion",
			Transport: "stdio",
			Command:   "notion-mcp-server",
			ToolCount: 12,
			Status:    model.StatusRunning,
		},
		{
			Name:      "sqlite",
			Plugin:    "superpowers",
			Transport: "stdio",
			Command:   "mcp-server-sqlite",
			Args:      []string{"--db-path", "~/.claude/data.db"},
			ToolCount: 6,
			Status:    model.StatusDone,
		},
	}
}
