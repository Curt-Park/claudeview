---
title: "Command Package (cmd)"
type: component
tags: [cmd, internals]
---

# Command Package — `cmd`

Wires all internal packages into a runnable Bubble Tea application. Contains the Cobra CLI definition, the top-level Bubble Tea model, and async data loading. `DataProvider` implementations live in [[provider-package]] and [[demo-package]].

## Files

| File              | Purpose                                                                          |
|-------------------|----------------------------------------------------------------------------------|
| `root.go`         | Cobra `rootCmd`; `rootModel`; async data loading; wires `provider.NewLive` and `demo.NewProvider` |
| `update.go`       | `--update` self-update: fetch latest GitHub release, download, atomic replace    |
| `update_test.go`  | 4 tests for update logic using `httptest.NewServer` (no network)                 |

## rootModel

`rootModel` is the top-level `tea.Model` submitted to `tea.NewProgram`. It wraps `ui.AppModel` and owns the data loading lifecycle:

```go
type rootModel struct {
    // Data slices
    app         ui.AppModel
    dp          ui.DataProvider
    projects    []*model.Project
    sessions    []*model.Session
    plugins     []*model.Plugin
    pluginItems []*model.PluginItem
    memories    []*model.Memory

    // Resource views (eagerly initialized)
    projectsView    *view.ResourceView[*model.Project]
    sessionsView    *view.ResourceView[*model.Session]
    pluginsView     *view.ResourceView[*model.Plugin]
    pluginItemsView *view.ResourceView[*model.PluginItem]
    memoriesView    *view.ResourceView[*model.Memory]
    chatView        *view.ResourceView[ui.ChatItem]

    // Cached chat items for the chat table
    chatItems []ui.ChatItem

    // Static info (set once at startup)
    userStr       string
    claudeVersion string

    // Async loading state
    loading      bool
    cursor       map[model.ResourceType]struct{ sel, off int }
    lastResource model.ResourceType

    // Key-based cursor for history view (survives expansion-induced row shifts)
    historyCursorKey  string // ChatItemKey of selected ChatItem (or parent when on sub-row)
    historyToolCallID string // ToolCall.ID of selected sub-row; "" = cursor on parent
}
```

On `Init`, it fires `loadData()` synchronously, then async reloads via `loadDataAsync()` which sends a `dataLoadedMsg` back into the update loop. This keeps the initial render fast while data refreshes in the background.

`dataLoadedMsg` carries resource-specific payloads including `turns []model.Turn`, `subagentTurns [][]model.Turn`, `subagentTypes []model.AgentType`, and slug group fields (`slugGroupSessions`, `slugGroupTurns`, `slugGroupSubTurns`, `slugGroupSubTypes`) for history view refresh. `loadDataAsync()` handles `ResourceHistory`/`ResourceHistoryDetail`: it calls `refreshSlugGroup()` to re-scan sessions and detect newly created (or removed) sessions under the same slug. When the refreshed group has multiple sessions, it loads turns/subagents for each; otherwise it reads single-session data via `app.SelectedSessionFilePath` and `app.SelectedSessionSubagentDir`. On receipt, `SlugSessions` is updated if `slugGroupSessions` is non-nil, then either `app.SetSlugGroupData()` or the single-session fields are set, `RebuildChatItems()` refreshes the flattened chat item list, and `syncView` updates the table. `GetSessions` applies `model.GroupSessionsBySlug` before returning, sorting sessions into slug-grouped order with tree prefixes.

`syncView()` also handles expansion state: when `ExpandedItems` is non-empty, it resolves the cursor index from `historyCursorKey` (by scanning `chatItems` for a matching `ChatItemKey`) before calling `Sync`, then calls `app.ApplyExpansion()` to insert `ToolCallRow` sub-rows. If `historyToolCallID` is set, it scans `FilteredRows()` to restore the sub-row cursor position. `SyncViewMsg` (sent by `toggleExpansion`) is intercepted in `rootModel.Update` to immediately call `syncView` without a full data reload.

## DataProvider Implementations

Both implement `ui.DataProvider` and live in their own packages:

- **`provider.Live`** (`internal/provider`) — reads `~/.claude/`; see [[provider-package]] for details
- **`demo.Provider`** (`internal/demo`) — synthetic data for `--demo`; see [[demo-package]] for details

`run()` in `root.go` selects between them: `provider.NewLive(config.ClaudeDir())` or `demo.NewProvider()`.

## CLI Flags

| Flag           | Effect                                                  |
|----------------|---------------------------------------------------------|
| `--demo`       | Use `demo.Provider` instead of live filesystem data     |
| `--update`     | Self-update to the latest GitHub release                |

## Helper Functions

- `refreshSlugGroup(dp, projectHash, sessionID, currentSlug)` — re-scans sessions to detect new/removed sessions in a slug group during history view refresh

`loadDataAsync()` uses `parallel.Map` (from [[parallel-package]]) for concurrent slug-group and subagent turn loading. Agent/subagent type extraction is handled by `model.ExtractSubagentTypes` and `model.ExtractAgentTypesFromCalls` (see [[model-package]]).

## Related

- [[architecture]] — how cmd wires packages together
- [[ui-package]] — `AppModel` and `DataProvider` interface consumed here
- [[provider-package]] — live `DataProvider` wired in `run()`
- [[demo-package]] — demo `DataProvider` wired in `run()`
- [[parallel-package]] — used in `loadDataAsync` for concurrent turn loading
- [[transcript-package]] — `ScanSubagents` used directly in `loadDataAsync`
- [[config-package]] — `ClaudeDir()` used in `run()`
