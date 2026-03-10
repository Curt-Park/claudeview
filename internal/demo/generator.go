package demo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Curt-Park/claudeview/internal/model"
)

// GenerateMemories creates synthetic demo memory files.
func GenerateMemories() []*model.Memory {
	now := time.Now()
	return []*model.Memory{
		{
			Name:    "MEMORY.md",
			Path:    "/demo/.claude/projects/demo-project-1/memory/MEMORY.md",
			Title:   "Project Memory",
			Size:    2048,
			ModTime: now.Add(-10 * time.Minute),
			Content: "# Project Memory\n\nThis project uses OAuth2 for authentication.\n\n## Stack\n- Python 3.11\n- FastAPI\n- PostgreSQL\n\n## Conventions\n- Use snake_case for all identifiers\n- Tests live in `tests/` mirroring `src/`\n- All endpoints require JWT bearer tokens\n",
		},
		{
			Name:    "patterns.md",
			Path:    "/demo/.claude/projects/demo-project-1/memory/patterns.md",
			Title:   "Code Patterns",
			Size:    512,
			ModTime: now.Add(-2 * time.Hour),
			Content: "# Code Patterns\n\n## Repository pattern\nAll DB access goes through `repo/` classes — never query directly from routes.\n\n## Error handling\nRaise `HTTPException` with structured `detail` dicts, not plain strings.\n",
		},
		{
			Name:    "debugging.md",
			Path:    "/demo/.claude/projects/demo-project-1/memory/debugging.md",
			Title:   "Debugging Notes",
			Size:    1024,
			ModTime: now.Add(-24 * time.Hour),
			Content: "# Debugging Notes\n\n## OAuth redirect loop\nWas caused by missing `SESSION_SECRET` env var in staging. Fixed 2025-11-14.\n\n## Slow test suite\nParallel fixture teardown races with DB reset. Use `--forked` flag for now.\n",
		},
	}
}

// GenerateTurns creates a realistic demo conversation history.
func GenerateTurns() []model.Turn {
	now := time.Now()
	taskInput, _ := json.Marshal(map[string]string{
		"description":   "Explore OAuth2 library options for Python",
		"subagent_type": "general-purpose",
	})
	calls := []*model.ToolCall{
		{Name: "Read", Input: []byte(`{"file_path": "src/auth.py"}`), Result: []byte(`"142 lines"`), Timestamp: now.Add(-4 * time.Minute), Duration: 80 * time.Millisecond},
		{Name: "Grep", Input: []byte(`{"pattern": "handleAuth", "path": "src/"}`), Result: []byte(`"3 matches in src/auth.py"`), Timestamp: now.Add(-3*time.Minute - 40*time.Second), Duration: 120 * time.Millisecond},
		{Name: "Task", Input: taskInput, Result: []byte(`"authlib and python-jose are both suitable; authlib has better async support"`), Timestamp: now.Add(-3 * time.Minute), Duration: 8 * time.Second},
		{Name: "Edit", Input: []byte(`{"file_path": "src/auth.py", "old_string": "import jwt", "new_string": "from authlib.integrations.starlette_client import OAuth"}`), Result: []byte(`"success"`), Timestamp: now.Add(-90 * time.Second), Duration: 60 * time.Millisecond},
		{Name: "Bash", Input: []byte(`{"command": "python -m pytest tests/test_auth.py -v"}`), Result: []byte(`"5 passed in 1.23s"`), Timestamp: now.Add(-45 * time.Second), Duration: 1400 * time.Millisecond},
	}
	return []model.Turn{
		{
			Role:      "user",
			Text:      "Refactor the authentication module to use OAuth2 with the authlib library.",
			Timestamp: now.Add(-5 * time.Minute),
		},
		{
			Role:         "assistant",
			Text:         "I'll start by reading the current auth implementation and exploring suitable OAuth2 libraries.",
			Thinking:     "The user wants OAuth2. I should check the existing auth code first, then evaluate library options via a subagent before making changes.",
			ToolCalls:    calls[:3],
			ModelName:    "claude-opus-4-6",
			InputTokens:  5200,
			OutputTokens: 820,
			Timestamp:    now.Add(-4 * time.Minute),
		},
		{
			Role:         "assistant",
			Text:         "The subagent recommends authlib for its async support. I'll update the import and refactor the token verification logic now.",
			ToolCalls:    calls[3:5],
			ModelName:    "claude-opus-4-6",
			InputTokens:  8400,
			OutputTokens: 1150,
			Timestamp:    now.Add(-90 * time.Second),
		},
		{
			Role:      "user",
			Text:      "Looks good! Can you also add a refresh token endpoint?",
			Timestamp: now.Add(-30 * time.Second),
		},
		{
			Role:         "assistant",
			Text:         "All 5 auth tests pass after the refactor. I'll add the refresh token endpoint next — I'll write the route handler and a corresponding test.",
			ModelName:    "claude-opus-4-6",
			InputTokens:  2100,
			OutputTokens: 380,
			Timestamp:    now.Add(-20 * time.Second),
		},
	}
}

// GeneratePluginItems creates synthetic plugin items for a named demo plugin.
func GeneratePluginItems(pluginName string) []*model.PluginItem {
	switch pluginName {
	case "superpowers":
		return []*model.PluginItem{
			{Name: "brainstorming", Category: "skill", Content: "# brainstorming\n\nExplore requirements and design before implementation. Identifies unknowns, suggests approaches, and surfaces trade-offs.\n\n## When to use\nBefore any feature or component that involves design choices.\n"},
			{Name: "systematic-debugging", Category: "skill", Content: "# systematic-debugging\n\nStructured bug diagnosis: reproduce → isolate → hypothesize → verify → fix.\n\n## Steps\n1. Confirm the failure is reproducible\n2. Narrow to the smallest failing case\n3. Form a falsifiable hypothesis\n4. Verify with a targeted test\n5. Apply the fix and re-run all tests\n"},
			{Name: "test-driven-development", Category: "skill", Content: "# test-driven-development\n\nRed → Green → Refactor cycle. Write the test first, then the minimum code to pass it.\n"},
			{Name: "commit", Category: "command", Content: "Stage all changes and create a conventional commit message summarizing the diff.\n"},
			{Name: "review-pr", Category: "command", Content: "Fetch the specified PR diff and produce a structured code review with actionable feedback.\n"},
			{Name: "PostToolUse", Category: "hook", Content: `{"hooks": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "echo tool-used"}]}]}` + "\n"},
		}
	case "Notion":
		return []*model.PluginItem{
			{Name: "create-page", Category: "skill", Content: "# create-page\n\nCreate a new Notion page under a specified parent using the Notion MCP server.\n"},
			{Name: "search", Category: "skill", Content: "# search\n\nSearch the connected Notion workspace by keyword and return matching pages.\n"},
			{Name: "notion", Category: "mcp", Content: `{
  "command": "npx",
  "args": ["-y", "@notionhq/notion-mcp-server"],
  "env": {
    "OPENAPI_MCP_HEADERS": "{\"Authorization\": \"Bearer ${NOTION_TOKEN}\"}"
  }
}` + "\n"},
		}
	case "code-review":
		return []*model.PluginItem{
			{Name: "code-review", Category: "skill", Content: "# code-review\n\nReview a pull request: fetch the diff, analyze code quality, identify bugs, and produce a structured review report.\n"},
			{Name: "receiving-code-review", Category: "skill", Content: "# receiving-code-review\n\nProcess incoming review feedback with rigor — verify claims before accepting suggestions.\n"},
			{Name: "requesting-code-review", Category: "skill", Content: "# requesting-code-review\n\nPrepare a change for review: run tests, confirm lint passes, write a clear PR description.\n"},
		}
	default:
		return []*model.PluginItem{}
	}
}

// GenerateProjects creates synthetic demo projects.
func GenerateProjects() []*model.Project {
	now := time.Now()

	sessions1 := []*model.Session{
		{
			ID:            "abc12345-demo-0001-0000-000000000001",
			ProjectHash:   "demo-project-1",
			FilePath:      "demo://session-abc12345",
			Topic:         "Refactor authentication module to use OAuth2",
			Branch:        "feat/auth-refactor",
			FileSize:      1363149,
			TokensByModel: map[string]model.TokenCount{"claude-opus-4-6": {InputTokens: 120000, OutputTokens: 25200}},
			AgentCount:    4,
			ToolCallCount: 18,
			NumTurns:      12,
			StartTime:     now.Add(-5 * time.Minute),
			ModTime:       now.Add(-30 * time.Second),
			Agents:        generateAgents("abc12345"),
		},
		{
			ID:            "def45678-demo-0002-0000-000000000002",
			ProjectHash:   "demo-project-1",
			FilePath:      "demo://session-def45678",
			Topic:         "Fix login redirect bug after password reset",
			Branch:        "fix/login-redirect",
			FileSize:      510054,
			TokensByModel: map[string]model.TokenCount{"claude-sonnet-4-6": {InputTokens: 20000, OutputTokens: 3100}},
			AgentCount:    4,
			ToolCallCount: 18,
			NumTurns:      5,
			StartTime:     now.Add(-2 * time.Hour),
			EndTime:       now.Add(-90 * time.Minute),
			ModTime:       now.Add(-90 * time.Minute),
			Agents:        generateAgents("def45678"),
		},
	}

	sessions2 := []*model.Session{
		{
			ID:            "ghi78901-demo-0003-0000-000000000003",
			ProjectHash:   "demo-project-2",
			FilePath:      "demo://session-ghi78901",
			Topic:         "Update test coverage for API endpoints",
			Branch:        "main",
			FileSize:      2662400,
			TokensByModel: map[string]model.TokenCount{"claude-haiku-4-5": {InputTokens: 5000, OutputTokens: 800}},
			AgentCount:    4,
			ToolCallCount: 18,
			NumTurns:      3,
			StartTime:     now.Add(-24 * time.Hour),
			EndTime:       now.Add(-23 * time.Hour),
			ModTime:       now.Add(-23 * time.Hour),
			Agents:        generateAgents("ghi78901"),
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

// GeneratePlugins creates synthetic demo plugins.
func GeneratePlugins() []*model.Plugin {
	return []*model.Plugin{
		{
			Name:         "superpowers",
			Version:      "4.3.1",
			Marketplace:  "claude-plugins-official",
			Scope:        "user",
			Enabled:      true,
			InstalledAt:  "2025-12-15",
			SkillCount:   15,
			CommandCount: 2,
			HookCount:    1,
			AgentCount:   0,
			MCPCount:     0,
		},
		{
			Name:         "Notion",
			Version:      "1.2.0",
			Marketplace:  "claude-plugins-official",
			Scope:        "user",
			Enabled:      true,
			InstalledAt:  "2025-11-20",
			SkillCount:   8,
			CommandCount: 1,
			AgentCount:   1,
			MCPCount:     1,
		},
		{
			Name:        "code-review",
			Version:     "2.0.1",
			Marketplace: "claude-plugins-official",
			Scope:       "project",
			Enabled:     false,
			InstalledAt: "2025-10-01",
			SkillCount:  3,
			AgentCount:  0,
			MCPCount:    0,
		},
	}
}
