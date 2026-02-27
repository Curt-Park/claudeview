# Plugin & Memory Detail View Design

**Date:** 2026-02-27
**Branch:** feat/ui-ux-improvements

## Problem

In the plugins and memories views, pressing enter does nothing. Users cannot inspect the contents of a plugin (its skills, commands, hooks, agents, MCPs) or read the full text of a memory file.

## Goal

When the user selects an item in the plugins or memories view and presses enter, show a detail view that fills the full content area and supports j/k scrolling. Pressing esc returns to the parent view.

## Approach: New ResourceType per Detail View

Add two new `ResourceType` values and reuse the existing `drillDown` / `drillInto` / `navigateBack` navigation pattern unchanged.

### New ResourceTypes

```go
ResourcePluginDetail = "plugin-detail"
ResourceMemoryDetail = "memory-detail"
```

### Navigation Flow

```
plugins  →(enter)→ plugin-detail  →(esc)→ plugins
memories →(enter)→ memory-detail  →(esc)→ memories
```

Both detail views are terminal nodes (no further enter action).

## Data Model

### AppModel additions

```go
SelectedPlugin *model.Plugin
SelectedMemory *model.Memory
```

Set in `drillDown()` before calling `drillInto()`.

### New model functions (plugin.go)

```go
func ListSkills(cacheDir string) []string    // subdirectory names under skills/
func ListCommands(cacheDir string) []string  // .md filenames under commands/
func ListHooks(cacheDir string) []string     // hook event names from hooks.json or files
func ListAgents(cacheDir string) []string    // .md filenames under agents/
func ListMCPs(cacheDir string) []string      // mcpServer keys from .mcp.json or plugin.json
```

## Plugin Detail Content

Rendered as plain text sections. Sections with zero items are omitted.

```
Skills:    brainstorming, debugging, writing-plans
Commands:  commit, review-pr
Hooks:     PreToolUse, PostToolUse
Agents:    code-reviewer
MCPs:      my-mcp-server
```

Each section: bold/colored header, then a comma-separated list (or one item per line if many).

## Memory Detail Content

Read the file at `model.Memory.Path` and render the raw text. Lines are fed into the table as single-cell rows so existing j/k/G/g/ctrl+d/u scroll works without new scroll infrastructure.

## Key Handling

| Key | plugins | plugin-detail | memories | memory-detail |
|-----|---------|---------------|----------|---------------|
| enter | drill to detail | — | drill to detail | — |
| esc | back | back to plugins | back | back to memories |
| j/k/G/g/ctrl+d/u | scroll table | scroll detail | scroll table | scroll detail |

## Rendering

`AppModel.View()` gains a switch on `m.Resource` before calling `m.Table.View()`:

```go
switch m.Resource {
case model.ResourcePluginDetail:
    contentStr = view.RenderPluginDetail(m.SelectedPlugin, w, h)
case model.ResourceMemoryDetail:
    contentStr = view.RenderMemoryDetail(m.SelectedMemory, w, h)
default:
    contentStr = m.Table.View()
}
```

Both render functions return a string of exactly `h` lines.

## Menu Hints

```
plugins:       <enter> detail   <esc> back
plugin-detail: <esc>   back     (no enter)
memories:      <enter> detail   <esc> back
memory-detail: <esc>   back     (no enter)
```

## Files Changed

| File | Change |
|------|--------|
| `internal/model/resource.go` | Add `ResourcePluginDetail`, `ResourceMemoryDetail` |
| `internal/model/plugin.go` | Add `ListSkills`, `ListCommands`, `ListHooks`, `ListAgents`, `ListMCPs` |
| `internal/ui/app.go` | Add `SelectedPlugin`/`SelectedMemory`, extend `drillDown`, `navigateBack`, `View` |
| `internal/ui/menu.go` | Add detail cases to `TableNavItems` |
| `internal/view/plugins.go` | Add `RenderPluginDetail` |
| `internal/view/memories.go` | Add `RenderMemoryDetail` |

## Testing

- `internal/model/plugin_test.go` — unit tests for each `List*` function
- `internal/ui/app_test.go` — enter on plugin row → `ResourcePluginDetail`; esc → `ResourcePlugins`; enter on memory row → `ResourceMemoryDetail`; esc → `ResourceMemories`
