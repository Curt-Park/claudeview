---
confidence: 0.8
created: "2026-02-25T23:19:17+10:00"
id: component-2468
modified: "2026-02-25T23:19:17+10:00"
references: []
relations:
  - type: relates_to
    target: component-2866
    description: 'High tag overlap: 10%'
    confidence: 0.7000000000000001
  - type: depends_on
    target: component-1669
    description: AppModel uses model.ResourceType, model.ResourceProjects, and other constants
    confidence: 0.8
source: manual
status: active
tags:
  - internals
  - ui
  - bubble-tea
title: UI Package (internal/ui)
type: component
---

# UI Package — `internal/ui`

The largest package in claudeview (~25 .go files). Implements the Bubble Tea application model and all reusable chrome components.

## Files

| File | Purpose |
|------|---------|
| `app.go` | `AppModel` — root Bubble Tea model; handles key events, layout, mode switching |
| `table_view.go` | `TableView` — scrollable table with filter, selection, expandable rows |
| `log_view.go` | `LogView` — scrollable JSONL transcript log renderer |
| `detail_view.go` | `DetailView` — key/value detail pane |
| `header.go` | Top bar with app name, version, resource type |
| `menu.go` | Bottom info panel showing navigation hints (4-column layout) |
| `crumbs.go` | Breadcrumb trail (project → session → agent) |
| `flash.go` | Ephemeral status/error message overlay |
| `filter.go` | `/`-triggered filter input bar |
| `styles.go` | Lip Gloss style definitions shared across components |

## AppModel

`AppModel` is the central Bubble Tea model. It holds:
- `Resource` — current `model.ResourceType`
- `Table` — the active `TableView` (swapped by `rootModel.syncView`)
- `Log`, `Detail` — secondary view panes
- `Info` — `InfoData` struct (project, session, user, versions)
- `Width`, `Height` — terminal dimensions
- `SelectedProjectHash`, `SelectedSessionID`, `SelectedAgentID` — drill-down context
- Mode state: normal / filter / command / log / detail

## Key Messages

| Message | Meaning |
|---------|---------|
| `RefreshMsg` | Reload data from DataProvider |
| `DetailRequestMsg` | Populate DetailView for selected row |
| `LogRequestMsg` | Populate LogView for selected row |
| `YAMLRequestMsg` | Populate DetailView with JSON dump |

## DataProvider Interface

```go
type DataProvider interface {
    GetProjects() []*model.Project
    GetSessions(projectHash string) []*model.Session
    GetAgents(sessionID string) []*model.Agent
    GetTools(agentID string) []*model.ToolCall
    GetTasks(sessionID string) []*model.Task
    GetPlugins() []*model.Plugin
    GetMCPServers() []*model.MCPServer
}
```

## Navigation Keys

- `j`/`k` — move up/down in table
- `enter` — drill down
- `l` — log view
- `d` — detail view
- `y` — YAML/JSON dump view
- `t`/`p`/`m` — jump to tasks / plugins / MCP servers
- `/` — filter mode
- `esc` — back / exit mode
- `ctrl+c` — quit

## BDD Tests

9 teatest-based integration tests in `internal/ui/bdd/` covering spec behaviors (TEST-NAV-001 etc.).
See [[BDD Test Suite (internal/ui/bdd)]].


## Related
- [[component-2866]]
- [[component-1669]]
