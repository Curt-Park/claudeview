---
title: "Provider Package (internal/provider)"
type: component
tags: [provider, internals]
---

# Provider Package — `internal/provider`

Live `ui.DataProvider` implementation. Extracted from `cmd/root.go`'s `liveDataProvider` struct to give it a dedicated home alongside its helpers.

## Files

| File      | Purpose                                                            |
|-----------|--------------------------------------------------------------------|
| `live.go` | `Live` struct, `NewLive()`, all `DataProvider` methods and helpers |

## API

```go
func NewLive(claudeDir string) ui.DataProvider
```

Returns a `*Live` that reads `~/.claude/` (or the given `claudeDir`) to satisfy the [[ui-package]] `DataProvider` interface.

## Methods

| Method | What it does |
|--------|-------------|
| `GetProjects()` | Scans `~/.claude/projects/`, builds `Project`+`Session` models; uses `parallel.Map` for concurrent `sessionFromInfo` calls |
| `GetSessions(projectHash)` | Filters by project, parallel `sessionFromInfo`, applies `model.GroupSessionsBySlug` |
| `GetAgents(sessionID)` | Calls `parseAgentsFromSession` (package-level helper) |
| `GetPlugins(projectHash)` | Reads `installed_plugins.json` via [[config-package]] |
| `GetPluginItems(plugin)` | Delegates to `model.ListPluginItems` |
| `GetMemories(projectHash)` | Reads `memory/` dir; uses `stringutil.MdTitle` for heading extraction |
| `GetTurns(filePath)` | Incremental parse via `transcript.ParseFileIncremental` with `turnsCache` |

## Package-Level Helpers

- **`sessionFromInfo(si)`** — incremental aggregate parse with `aggCache` (mutex-protected); merges subagent token counts via `parallel.Map`
- **`populateToolCalls(agent, sessionID, parsed)`** — fills `agent.ToolCalls` from a `ParsedTranscript`; sets `LastActivity`
- **`parseAgentsFromSession(s)`** — builds `[]*model.Agent` including subagents; uses `model.ExtractAgentTypesFromCalls` to assign `AgentType` by position; `parallel.Map` for concurrent subagent transcript parsing

## Caches

Both fields are protected by `mu sync.Mutex` for goroutine-safe concurrent access:

| Field | Type | Purpose |
|-------|------|---------|
| `aggCache` | `map[string]*transcript.SessionAggregates` | Avoids re-parsing full files for session metrics |
| `turnsCache` | `map[string]*transcript.TranscriptCache` | Offset-based incremental turn parsing |

## Related

- [[architecture]] — DataProvider implementations diagram
- [[cmd-package]] — `run()` calls `provider.NewLive(config.ClaudeDir())`
- [[transcript-package]] — primary data source; `ScanProjects`, `ParseAggregatesIncremental`, `ParseFileIncremental`, `ParseFile`
- [[ui-package]] — `DataProvider` interface this package implements
- [[model-package]] — all returned types (`Project`, `Session`, `Agent`, `Plugin`, `Memory`, `Turn`)
- [[parallel-package]] — used for concurrent I/O in `sessionFromInfo`, `parseAgentsFromSession`, `GetProjects`, `GetSessions`
- [[stringutil-package]] — `MdTitle` for memory file headings
- [[config-package]] — `LoadInstalledPlugins`, `EnabledPlugins`, `ProjectEnabledPlugins`
