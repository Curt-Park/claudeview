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
| `sessions.go`     | columns + `sessionRow`; flat mode; subtitle line (branch/size); SLUG + SESSION_IDs columns with `GroupNameCell()` for slug groups |
| `plugins.go`      | columns + `pluginRow`; scope, enabled/disabled, skill/cmd/hook counts|
| `plugin_items.go` | columns + `pluginItemRow` for `ResourceView[*model.PluginItem]`; `NewPluginItemsView` |
| `memories.go`     | columns + `memoryRow` for `ResourceView[*model.Memory]`              |
| `chat.go`         | columns + `chatRow` for `ResourceView[ui.ChatItem]`; `NewChatView`; divider row handling for merged slug groups |
| `helpers.go`      | Shared formatting utilities (`truncateHash`, `ShortID`)              |

## Generic ResourceView[T]

All 6 resource views use the same generic type:

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

## Column Widths (summary)

| Resource  | Base columns                                              |
|-----------|-----------------------------------------------------------|
| Projects  | NAME(flex,55%), SESSIONS(8), LAST ACTIVE(11)             |
| Sessions  | SLUG(16), SESSION_IDs(19), TOPIC(flex,35%), TURNS(6), AGENTS(6), MODEL:TOKEN(flex,25%), LAST ACTIVE(11) |
| Plugins   | NAME(flex,25%), VERSION(10), SCOPE(8), STATUS(10), SKILLS(7), COMMANDS(9), HOOKS(6), AGENTS(7), MCPS(5), INSTALLED(12) |
| Memories  | NAME(18), TITLE(flex,45%), SIZE(8), MODIFIED(11)         |
| Chat      | NAME(10), MESSAGE(flex,50%), ACTION(16), MODEL:TOKEN(flex,20%), DURATION(14) |

## Related

- [[ui-spec]] — column definitions this package implements
- [[ui-package]] — consumes `ResourceView[T]` via `AppModel.Table`
- [[model-package]] — data types used by row builders
- [[architecture]] — view package role in the rendering pipeline
