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
| `detail_render.go`    | `RenderPluginItemDetail`, `RenderMemoryDetail`, `RenderSessionChat` — string renderers |
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
- `SelectedTurns []model.Turn` — main agent turns for session-chat view
- `SubagentTurns [][]model.Turn` — per-subagent turn slices (parallel to Task tool calls)
- `SubagentTypes []model.AgentType` — agent type for each subagent turn slice
- `ChatFollow bool` — follow mode flag; when true, session-chat view auto-scrolls to bottom (tail -f)
- `SelectedSessionFilePath string` — JSONL file path of selected session (for async refresh)
- `SelectedSessionSubagentDir string` — subagent directory for selected session (for async refresh)
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
    GetTurns(filePath string) []model.Turn
}
```

## Navigation Keys

| Key      | Action                                      |
|----------|---------------------------------------------|
| `j/k`    | move up/down in table; in session-chat: scroll down/up (k disables follow mode) |
| `g/G`    | top/bottom; in session-chat: G re-enables follow mode |
| `ctrl+d/u` | page down/up; in session-chat: ctrl+u disables follow mode |
| `enter`  | drill down                                  |
| `p`      | jump to plugins                             |
| `m`      | jump to memories (requires project context) |
| `/`      | filter mode                                 |
| `esc`    | clear filter / navigate back                |
| `ctrl+c` | quit                                        |

### Follow Mode (session-chat only)

`ChatFollow = true` is set on drill-down into session-chat (auto-scroll to latest content, like `tail -f`). `View()` overrides `ContentOffset` with `maxOffset` when follow mode is active. `updateContentScroll()` syncs `ContentOffset` from the virtual follow position before applying a relative delta, then toggles `ChatFollow`:
- Enabled by: `G`, reaching the bottom via `j` or `ctrl+d`
- Disabled by: `k`, `g`, `ctrl+u`

## Key Messages

| Message            | Trigger                     |
|--------------------|-----------------------------|
| `TickMsg`          | 1-second timer tick         |
| `RefreshMsg`       | data reload signal          |
| `HighlightClearMsg`| key highlight expiry (150ms)|

## detail_render.go

Three top-level renderers:

- **`RenderSessionChat(turns, subagentTurns, subagentTypes, width)`** — renders a session's conversation as a scrollable chat timeline. User turns appear as right-aligned rounded-border bubbles (blue); Claude turns as left-aligned green bubbles with thinking, tool call lines, and token counts; subagent turns as indented purple bubbles. Subagent bubbles are inserted after the Claude turn that contains the matching `Task` tool call.
- **`RenderPluginItemDetail(item, width)`** — renders a plugin item's content with header and optional hook script blocks.
- **`RenderMemoryDetail(m, width)`** — reads and wraps a memory file's raw Markdown content.

## styles.go — Chat Bubble Styles

In addition to base, status, and layout styles, `styles.go` defines:

| Style               | Description                                      |
|---------------------|--------------------------------------------------|
| `StyleUserBubble`   | Rounded border, blue foreground, right-aligned   |
| `StyleClaudeBubble` | Rounded border, green foreground, left-aligned   |
| `StyleSubagentBubble` | Rounded border, purple foreground, indented    |
| `StyleChatThinking` | Gray — thinking block label                      |
| `StyleChatToolOK`   | Green — successful tool call outcome             |
| `StyleChatToolErr`  | Red — failed tool call outcome                   |
| `StyleChatToolName` | Blue — tool name prefix (`▸ ToolName`)           |
| `StyleChatTokens`   | Gray — token count footer (`░ N tok`)            |
| `StyleChatTimestamp`| Dim gray — `HH:MM` timestamp                    |
| `StyleChatHeader`   | Cyan bold — speaker name ("Claude", "You", etc.) |

## Related

- [[ui-spec]] — behavior specification this package implements
- [[view-package]] — provides `ResourceView[T]` consumed by AppModel
- [[model-package]] — `ResourceType` constants and data types
- [[test-suite]] — AppModel integration tests
