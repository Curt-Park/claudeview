---
confidence: 0.8
created: "2026-02-25T23:19:27+10:00"
id: component-1792
modified: "2026-02-25T23:19:27+10:00"
references: []
relations:
  - type: relates_to
    target: component-2250
    description: 'High tag overlap: 10%'
    confidence: 0.7000000000000001
  - type: depends_on
    target: component-2468
    description: view 패키지가 ui.TableView 타입을 직접 임포트해서 사용
    confidence: 0.8
  - type: depends_on
    target: component-1669
    description: view 패키지가 model.Project/Session/Agent 등 타입을 직접 참조해서 렌더링
    confidence: 0.8
source: manual
status: active
tags:
  - internals
  - views
title: View Package (internal/view)
type: component
---

# View Package — `internal/view`

Resource-specific table view and detail rendering. One view file per resource type, plus helpers and a detail renderer.

## Files

| File | Purpose |
|------|---------|
| `resource_view.go` | `ResourceView[T]` — generic table view; `RowBuilder[T]` type; `Sync()` method |
| `projects.go` | columns + `projectRow` for `ResourceView[*model.Project]` |
| `sessions.go` | columns + `sessionRow` for `ResourceView[*model.Session]`; supports flat mode |
| `agents.go` | columns + `agentRow` for `ResourceView[*model.Agent]`; flat mode |
| `tools.go` | columns + `toolRow` for `ResourceView[*model.ToolCall]`; flat mode |
| `tasks.go` | columns + `taskRow` for `ResourceView[*model.Task]`; flat mode |
| `plugins.go` | columns + `pluginRow` for `ResourceView[*model.Plugin]` — name, version, status (enabled/disabled), skill/cmd/hook counts, install date |
| `mcp.go` | columns + `mcpRow` for `ResourceView[*model.MCPServer]` — name, plugin, transport, tool count, command |
| `detail.go` | `SessionDetailLines`, `AgentDetailLines`, `TaskDetailLines`, `PluginDetailLines`, `MCPDetailLines` |
| `helpers.go` | Shared formatting utilities (hash truncation) |

## Generic ResourceView[T]

All 7 resource views use the generic `ResourceView[T any]` type:

```go
type RowBuilder[T any] func(items []T, index int, flatMode bool) ui.Row

type ResourceView[T any] struct {
    Table    ui.TableView
    FlatMode bool
    Items    []T
    // ...
}

func (v *ResourceView[T]) Sync(items []T, w, h, sel, off int, filter string, flat bool) ui.TableView
```

## Flat Mode

Views that can appear at multiple levels of the drill-down hierarchy pass `flat bool` to `Sync()`. When `true`, extra parent-context columns (e.g., PROJECT, SESSION, AGENT) are prepended.

## View Lifecycle (managed by rootModel)

```
rootModel.syncView()
  → calls view.Sync(data, w, h, sel, off, filter, flat)
  → Sync rebuilds TableView with correct columns and rows
  → assigns returned TableView to app.Table
```

Views are eagerly initialized in `newRootModel()` — no lazy nil checks needed. `Sync()` preserves cursor and filter state across `RefreshMsg` and `WindowSizeMsg`.

## Detail Rendering

`detail.go` exports per-resource `*DetailLines(model) []string` functions, called by `rootModel.populateDetail()` to populate the `DetailView` pane.


## Related
- [[component-2250]]
- [[component-2468]]
- [[component-1669]]
