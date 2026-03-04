---
title: "claudeview Architecture Overview"
type: concept
tags: [architecture, overview]
---

# claudeview Architecture Overview

claudeview is a Go TUI application for monitoring Claude Code sessions. It reads JSONL transcript files written by Claude Code and presents them in an interactive Bubble Tea dashboard.

## Entry Points

- `main.go` — sets `AppVersion` from build-time ldflags, calls `cmd.Execute()`
- `cmd/root.go` — Cobra CLI; `--demo` flag; wires `DataProvider` into `AppModel`
- `cmd/update.go` — `--update` flag; self-update from GitHub releases

## Top-Level Architecture

```
main.go
  └─ cmd/root.go (Cobra)
       ├─ ui.AppModel     — Bubble Tea model; owns keyboard, layout, mode, navigation state
       ├─ ui.DataProvider — interface; two implementations:
       │    ├─ liveDataProvider  — reads ~/.claude/ via transcript + config packages
       │    └─ demoDataProvider  — returns synthetic data (internal/demo)
       └─ view.ResourceView[T] — generic table renderer; one constructor per resource type
```

## Internal Packages

| Package              | Role                                                            |
|----------------------|-----------------------------------------------------------------|
| `internal/transcript`| JSONL parser, directory scanner                                 |
| `internal/config`    | settings.json, installed_plugins.json parsers                   |
| `internal/model`     | Data models: Project, Session, Agent, ToolCall, Plugin, Memory  |
| `internal/ui`        | Bubble Tea AppModel + chrome components                         |
| `internal/view`      | Generic `ResourceView[T]` + 7 resource constructors             |
| `internal/demo`      | Synthetic demo data generator                                   |

## DataProvider Interface

```go
type DataProvider interface {
    GetProjects() []*model.Project
    GetSessions(projectHash string) []*model.Session
    GetAgents(sessionID string) []*model.Agent
    GetPlugins(projectHash string) []*model.Plugin
    GetMemories(projectHash string) []*model.Memory
    GetTurns(filePath string) []model.Turn
}
```

## Resource Hierarchy

```
projects → sessions → history → history-detail  [leaf, content view]
plugins  → plugin-detail → plugin-item-detail  [leaf]
memories → memory-detail  (requires project context)
```

## Key Design Decisions

- `AppModel` owns all UI state and navigation; `rootModel` in `cmd/root.go` wraps it with async data loading and `DataProvider` wiring
- `DataProvider` interface returns typed slices (`[]*model.Project`, etc.)
- `ResourceView[T]` (generic) unifies all resource views; `Sync()` preserves cursor/scroll/filter
- Views eagerly initialized; `Sync()` replaces the old `Set*()` + lazy init pattern
- "Hot" row highlight: rows modified within 5 seconds are highlighted
- Session subtitle row shows model/cost/status metadata below the main row

## Module / Build

- Module: `github.com/Curt-Park/claudeview`
- `make build` / `make demo` / `make test` / `make lint` / `make fmt`

## Related

- [[cmd-package]] — Cobra CLI, rootModel, DataProvider wiring
- [[ui-package]] — Bubble Tea AppModel and chrome components
- [[view-package]] — ResourceView[T] and resource constructors
- [[model-package]] — data model types
- [[transcript-package]] — JSONL parsing and project scanning
- [[config-package]] — Claude settings and plugin config parsing
- [[demo-package]] — synthetic demo data generator
- [[test-suite]] — test coverage
- [[ui-spec]] — UI behavior specification
