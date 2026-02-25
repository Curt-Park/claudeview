---
confidence: 0.8
created: "2026-02-25T23:19:35+10:00"
id: component-1669
modified: "2026-02-25T23:19:35+10:00"
references: []
relations:
  - type: relates_to
    target: component-2095
    description: 'High tag overlap: 10%'
    confidence: 0.7000000000000001
  - type: relates_to
    target: component-2425
    description: 'High tag overlap: 10%'
    confidence: 0.7000000000000001
source: manual
status: active
tags:
  - internals
  - models
title: Model Package (internal/model)
type: component
---

# Model Package — `internal/model`

Core data models used across transcript parsing, config loading, UI rendering, and view layers.

## Files & Types

| File | Types |
|------|-------|
| `project.go` | `Project` — Hash, Path, LastSeen, Sessions []*Session |
| `session.go` | `Session` — ID, FilePath, SubagentDir, ProjectHash, Status, TotalCost, NumTurns, DurationMS, InputTokens, OutputTokens, Model, ModTime, Agents |
| `agent.go` | `Agent` — ID, SessionID, Type (AgentType), Status, FilePath, IsSubagent, StartTime, ToolCalls, LastActivity |
| `tool_call.go` | `ToolCall` — ID, SessionID, AgentID, Name, Input/Result (json.RawMessage), IsError, Timestamp; `InputSummary()` method |
| `task.go` | `Task` — ID, SessionID, Subject, Description, Status, Owner, BlockedBy, Blocks, ActiveForm |
| `plugin.go` | `Plugin` — Name, Version, Marketplace, Enabled, InstalledAt, CacheDir, SkillCount, CommandCount, HookCount; `CountSkills/CountCommands/CountHooks(dir)` |
| `mcp_server.go` | `MCPServer` — Name, Transport, Command, Args, URL, Status |
| `resource.go` | `ResourceType` string constants (ResourceProjects, ResourceSessions, ResourceAgents, ResourceTools, ResourceTasks, ResourcePlugins, ResourceMCP); `ResolveResource(s)` |

## Status Type

`Status` string: `StatusActive`, `StatusEnded`, `StatusDone`, `StatusRunning`, `StatusError`

## AgentType

`AgentType` string: `AgentTypeMain`, `AgentTypeGeneral`, `AgentTypeExplore`, `AgentTypePlan`, `AgentTypeBash`

## Notes

- All models are plain Go structs with no ORM or DB layer
- Models are populated by `transcript` and `config` packages, consumed by `view` and `ui`


## Related
- [[component-2095]]
- [[component-2425]]
