---
title: "View Package (internal/view)"
type: component
tags: [views, internals]
---

# View Package — `internal/view`

Resource-specific table renderers. One file per resource type plus shared helpers.

## Files

| File              | Purpose                                                              |
|-------------------|----------------------------------------------------------------------|
| `resource_view.go`| `ResourceView[T]` — generic table view; `RowBuilder[T]`; `Sync()`   |
| `projects.go`     | columns + `projectRow` for `ResourceView[*model.Project]`            |
| `sessions.go`     | columns + `sessionRow`; flat mode; subtitle line (model/cost/status) |
| `agents.go`       | columns + `agentRow`; flat mode; tree prefix for subagents           |
| `plugins.go`      | columns + `pluginRow`; scope, enabled/disabled, skill/cmd/hook counts|
| `memories.go`     | columns + `memoryRow` for `ResourceView[*model.Memory]`              |
| `helpers.go`      | Shared formatting utilities (`truncateHash`, `ShortID`)              |

## Generic ResourceView[T]

All 5 resource views use the same generic type:

```go
type RowBuilder[T any] func(items []T, index int, flatMode bool) ui.Row

type ResourceView[T any] struct {
    Table    ui.TableView
    FlatMode bool
    Items    []T
}

func (v *ResourceView[T]) Sync(items []T, w, h, sel, off int, filter string, flat bool) ui.TableView
```

## Flat Mode

When `flat=true`, extra parent-context columns are prepended:
- Sessions: `PROJECT` (20)
- Agents: `SESSION` (12)

## Column Widths (summary)

| Resource  | Base columns                                              |
|-----------|-----------------------------------------------------------|
| Projects  | NAME(flex,55%), SESSIONS(8), LAST ACTIVE(11)             |
| Sessions  | NAME(10), TOPIC(flex,35%), TURNS(6), AGENTS(6), TOKENS(flex,25%), LAST ACTIVE(11) |
| Agents    | NAME(flex,20%), TYPE(16), STATUS(10), LAST ACTIVITY(flex,35%) |
| Plugins   | NAME(flex,25%), VERSION(10), SCOPE(8), STATUS(10), SKILLS(7), COMMANDS(9), HOOKS(6), AGENTS(7), MCPS(5), INSTALLED(12) |
| Memories  | NAME(18), TITLE(flex,45%), SIZE(8), MODIFIED(11)         |

## Related

- [[ui-spec]] — column definitions this package implements
- [[ui-package]] — consumes `ResourceView[T]` via `AppModel.Table`
- [[model-package]] — data types used by row builders
- [[architecture]] — view package role in the rendering pipeline
