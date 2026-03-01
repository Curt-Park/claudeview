---
title: "UI Package (internal/ui)"
type: component
tags: [ui, internals, bubble-tea]
---

# UI Package — `internal/ui`

Implements the Bubble Tea application model and all reusable chrome components.

## Files

| File                  | Purpose                                                        |
|-----------------------|----------------------------------------------------------------|
| `app.go`              | `AppModel` — root Bubble Tea model; key events, layout, mode   |
| `table_view.go`       | `TableView` — scrollable table with filter, selection          |
| `detail_render.go`    | `RenderPluginItemDetail`, `RenderMemoryDetail` — string renderers |
| `header.go`           | Info panel (5-column layout: info, nav, util, shortcuts, quit) |
| `menu.go`             | `MenuModel` — nav/util item lists and key highlight state      |
| `crumbs.go`           | `CrumbsModel` — breadcrumb trail                               |
| `flash.go`            | `FlashModel` — ephemeral status/error message                  |
| `filter.go`           | `FilterModel` — `/`-triggered filter input bar                 |
| `styles.go`           | Lip Gloss style definitions shared across components           |

## AppModel

`AppModel` is the single Bubble Tea model. It holds:

- `Resource` — current `model.ResourceType`
- `Table` — active `TableView`
- `Info`, `Menu`, `Crumbs`, `Flash`, `Filter` — chrome components
- `Width`, `Height` — terminal dimensions
- `SelectedProjectHash`, `SelectedSessionID`, `SelectedAgentID` — drill-down context
- `SelectedPlugin`, `SelectedPluginItem`, `SelectedMemory` — detail view context
- `inFilter bool` — filter input mode flag
- `filterStack []string` — saved parent filters across drill-downs
- `jumpFrom *jumpFromState` — saved state for esc-to-restore after p/m jump

## DataProvider Interface

```go
type DataProvider interface {
    GetProjects() []*model.Project
    GetSessions(projectHash string) []*model.Session
    GetAgents(sessionID string) []*model.Agent
    GetPlugins(projectHash string) []*model.Plugin
    GetMemories(projectHash string) []*model.Memory
}
```

## Navigation Keys

| Key      | Action                                      |
|----------|---------------------------------------------|
| `j/k`    | move up/down in table                       |
| `g/G`    | top/bottom                                  |
| `ctrl+d/u` | page down/up                              |
| `enter`  | drill down                                  |
| `p`      | jump to plugins                             |
| `m`      | jump to memories (requires project context) |
| `/`      | filter mode                                 |
| `esc`    | clear filter / navigate back                |
| `ctrl+c` | quit                                        |

## Key Messages

| Message            | Trigger                     |
|--------------------|-----------------------------|
| `TickMsg`          | 1-second timer tick         |
| `RefreshMsg`       | data reload signal          |
| `HighlightClearMsg`| key highlight expiry (150ms)|

## Related

- [[ui-spec]] — behavior specification this package implements
- [[view-package]] — provides `ResourceView[T]` consumed by AppModel
- [[model-package]] — `ResourceType` constants and data types
- [[test-suite]] — AppModel integration tests
