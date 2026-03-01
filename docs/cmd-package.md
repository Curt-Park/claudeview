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
    app      ui.AppModel
    dp       ui.DataProvider
    projects []*model.Project
}
```

On `Init`, it fires `loadDataAsync()` which reads the filesystem and sends a `dataLoadedMsg` back into the update loop. This keeps the initial render fast while data loads in the background.

## DataProvider Implementations

Both implement `ui.DataProvider`:

- **`liveDataProvider`** — reads `~/.claude/` via `transcript.ScanProjects`, `transcript.ParseAggregatesIncremental`, `config.LoadInstalledPlugins`, and `config.LoadSettings`. Populates `model.Project`, `model.Session`, `model.Agent`, `model.Plugin`, `model.Memory`.
- **`demoDataProvider`** — delegates to `internal/demo` for synthetic data; used with `--demo` flag.

## CLI Flags

| Flag           | Effect                                                     |
|----------------|------------------------------------------------------------|
| `--demo`       | Use `demoDataProvider` instead of live filesystem data     |
| `--render-once`| Render a single frame and exit (for snapshot testing)      |

## Helper Functions

- `parseAgentsFromSession(session, transcript)` — builds `[]*model.Agent` from parsed transcript
- `populateToolCalls(agent, turns)` — fills agent's `ToolCalls` slice
- `detectAgentType(agentID, turns)` — infers `AgentType` from transcript content
- `mdTitle(text)` — extracts title from Markdown content

## Related

- [[architecture]] — how cmd wires packages together
- [[ui-package]] — `AppModel` and `DataProvider` interface consumed here
- [[transcript-package]] — filesystem scanning and parsing used by `liveDataProvider`
- [[config-package]] — settings and plugin loading used by `liveDataProvider`
- [[demo-package]] — synthetic data used by `demoDataProvider`
