---
title: "Model Package (internal/model)"
type: component
tags: [models, internals]
---

# Model Package — `internal/model`

Core data models used across transcript parsing, config loading, UI rendering, and view layers. Plain Go structs — no ORM or DB layer.

## Files & Types

| File          | Types / Purpose                                                         |
|---------------|-------------------------------------------------------------------------|
| `project.go`  | `Project` — Hash, Path, LastSeen, Sessions `[]*Session`                 |
| `session.go`  | `Session` — ID, ProjectHash, FilePath, SubagentDir, Branch, Slug, FileSize, Topic, TokensByModel (`map[string]TokenCount`), AgentCount, ToolCallCount, Agents, NumTurns, StartTime, EndTime, ModTime; `TokenCount` struct (InputTokens, OutputTokens) |
| `slug_group.go` | `GroupSessionsBySlug(sessions)` — groups sessions sharing a slug (sorted by latest ModTime desc, within-group by ModTime asc); `SessionTreePrefix(sessions, idx)` — returns tree prefix (`""`, `"├ "`, `"└ "`) for contiguous slug groups |
| `agent.go`    | `Agent` — ID, SessionID, Type (`AgentType`), Status, IsSubagent, ToolCalls, LastActivity, FilePath, StartTime, Depth |
| `turn.go`     | `Turn` — Role, Text, Thinking, ToolCalls, ModelName, InputTokens, OutputTokens, Timestamp |
| `tool_call.go`| `ToolCall` — ID, SessionID, AgentID, Name, Input/Result (json.RawMessage), IsError, Timestamp; `InputSummary()` |
| `plugin.go`   | `Plugin` — Name, Version, Scope, Marketplace, Enabled, InstalledAt, CacheDir, SkillCount, CommandCount, HookCount, AgentCount, MCPCount; `CountSkills/Commands/Hooks/Agents/MCPs(cacheDir)` + `List*` variants; `PluginItem` — Name, Category, CacheDir; `ListPluginItems(cacheDir)`, `ReadPluginItemContent(item)`, `HookScript` — Path, Content; `ReadHookCommandScripts(item)` — reads script files referenced by hook commands (expands `${CLAUDE_PLUGIN_ROOT}`); `normalizeJSON(raw)` |
| `memory.go`   | `Memory` — Name, Title, Path, Size, ModTime; `SizeStr()`, `LastModified()` |
| `resource.go` | `ResourceType` constants; `ResolveResource(s)`; `AllResourceNames()`    |
| `status.go`   | `Status` string type and constants                                      |
| `format.go`   | `FormatAge(d)` — human-friendly duration; `FormatTokenCount(n)` — "1.5k", "1.5M"; `FormatSize(b)` — human-friendly byte size |

## ResourceType Constants

```go
ResourceProjects         = "projects"
ResourceSessions         = "sessions"
ResourceAgents           = "agents"
ResourcePlugins          = "plugins"
ResourceMemory           = "memories"
ResourcePluginDetail     = "plugin-detail"
ResourcePluginItemDetail = "plugin-item-detail"
ResourceMemoryDetail     = "memory-detail"
ResourceHistory          = "history"
ResourceHistoryDetail    = "history-detail"
```

## Status Constants

`StatusActive`, `StatusThinking`, `StatusReading`, `StatusExecuting`, `StatusDone`, `StatusEnded`, `StatusError`, `StatusFailed`, `StatusRunning`, `StatusPending`, `StatusCompleted`

## AgentType

`AgentType` string — values derived from transcript data (e.g. `"main"`, `"general-purpose"`, `"bash"`, etc.)

## Related

- [[architecture]] — how models flow through the application
- [[view-package]] — consumes model types for table rendering
- [[ui-package]] — uses `ResourceType` constants and selection state
- [[transcript-package]] — populates Session, Agent, ToolCall from JSONL
- [[config-package]] — populates Plugin from installed_plugins.json
