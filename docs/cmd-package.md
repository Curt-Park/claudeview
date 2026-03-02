---
title: "Command Package (cmd)"
type: component
tags: [cmd, internals]
---

# Command Package — `cmd`

Wires all internal packages into a runnable Bubble Tea application. Contains the Cobra CLI definition, the top-level Bubble Tea model, both `DataProvider` implementations, and async data loading.

## Files

| File       | Purpose                                                                         |
|------------|---------------------------------------------------------------------------------|
| `root.go`  | Cobra `rootCmd`; `rootModel`; `liveDataProvider`; `demoDataProvider`; async load |

## rootModel

`rootModel` is the top-level `tea.Model` submitted to `tea.NewProgram`. It wraps `ui.AppModel` and owns the data loading lifecycle:

```go
type rootModel struct {
    // Data slices
    app         ui.AppModel
    dp          ui.DataProvider
    projects    []*model.Project
    sessions    []*model.Session
    agents      []*model.Agent
    plugins     []*model.Plugin
    pluginItems []*model.PluginItem
    memories    []*model.Memory

    // Resource views (eagerly initialized)
    projectsView    *view.ResourceView[*model.Project]
    sessionsView    *view.ResourceView[*model.Session]
    agentsView      *view.ResourceView[*model.Agent]
    pluginsView     *view.ResourceView[*model.Plugin]
    pluginItemsView *view.ResourceView[*model.PluginItem]
    memoriesView    *view.ResourceView[*model.Memory]

    // Static info (set once at startup)
    userStr       string
    claudeVersion string

    // Async loading state
    loading      bool
    cursor       map[model.ResourceType]struct{ sel, off int }
    lastResource model.ResourceType
}
```

On `Init`, it fires `loadData()` synchronously, then async reloads via `loadDataAsync()` which sends a `dataLoadedMsg` back into the update loop. This keeps the initial render fast while data refreshes in the background.

`dataLoadedMsg` carries resource-specific payloads including `turns []model.Turn`, `subagentTurns [][]model.Turn`, and `subagentTypes []model.AgentType` for session-chat refresh. `loadDataAsync()` handles `ResourceSessionChat` by reading `app.SelectedSessionFilePath` and `app.SelectedSessionSubagentDir` (captured before the goroutine) and calling `dp.GetTurns()` and `transcript.ScanSubagents()`. On receipt, `app.SelectedTurns`, `app.SubagentTurns`, and `app.SubagentTypes` are updated directly (bypassing `syncView` table logic).

## DataProvider Implementations

Both implement `ui.DataProvider`:

- **`liveDataProvider`** — reads `~/.claude/` via `transcript.ScanProjects`, `transcript.ParseAggregatesIncremental`, `config.LoadInstalledPlugins`, and `config.LoadSettings`. Populates `model.Project`, `model.Session`, `model.Agent`, `model.Plugin`, `model.Memory`. `GetTurns(filePath)` calls `transcript.ParseFile` on the given JSONL path and maps the result to `[]model.Turn`.
- **`demoDataProvider`** — delegates to `internal/demo` for synthetic data; used with `--demo` flag. `GetTurns` returns nil (no demo turn data).

## CLI Flags

| Flag           | Effect                                                     |
|----------------|------------------------------------------------------------|
| `--demo`       | Use `demoDataProvider` instead of live filesystem data     |
| `--render-once`| Render a single frame and exit (for snapshot testing)      |

## Helper Functions

- `parseAgentsFromSession(s *model.Session)` — builds `[]*model.Agent` by parsing the session's transcript and subagent transcripts
- `populateToolCalls(agent *model.Agent, sessionID string, parsed *transcript.ParsedTranscript)` — fills agent's `ToolCalls` slice and sets `LastActivity`
- `detectAgentType(id string)` — infers `AgentType` from the agent ID string
- `mdTitle(path string)` — reads a Markdown file and returns the first `# Heading` text

## Related

- [[architecture]] — how cmd wires packages together
- [[ui-package]] — `AppModel` and `DataProvider` interface consumed here
- [[transcript-package]] — filesystem scanning and parsing used by `liveDataProvider`
- [[config-package]] — settings and plugin loading used by `liveDataProvider`
- [[demo-package]] — synthetic data used by `demoDataProvider`
