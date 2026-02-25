---
confidence: 0.8
created: "2026-02-25T23:19:52+10:00"
id: component-1365
modified: "2026-02-25T23:19:52+10:00"
references: []
relations:
  - type: relates_to
    target: component-1669
    description: config 파싱 결과(InstalledPlugin, TaskEntry 등)가 cmd/root.go에서 model 타입으로 변환됨
    confidence: 0.8
source: manual
status: active
tags:
  - internals
  - config
title: Config Package (internal/config)
type: component
---

# Config Package — `internal/config`

Reads Claude Code configuration files from `~/.claude/`. Provides parsers for settings, plugins, and tasks.

## Files

| File | Purpose |
|------|---------|
| `settings.go` | `LoadSettings(claudeDir)` — parses `settings.json`; returns struct with `MCPServers` map; `ClaudeDir()` helper returns `~/.claude` |
| `plugins.go` | `LoadInstalledPlugins(claudeDir)` — parses `installed_plugins.json`; `EnabledPlugins(claudeDir)` — returns map[name]bool from settings; `PluginCacheDir(...)` — constructs cache directory path |
| `tasks.go` | `LoadTasks(claudeDir, sessionID)` — reads `tasks/<sessionID>.json`; returns `[]TaskEntry` |

## Key Types (internal)

- `SettingsFile` — top-level settings.json structure (MCPServers, plugins config)
- `MCPServerConfig` — Type, Command, Args, URL fields
- `InstalledPlugin` — Name, Version, Marketplace, Enabled, InstalledAt
- `TaskEntry` — ID, Subject, Description, Status, Owner, BlockedBy, Blocks, ActiveForm

## Usage Pattern

All functions accept `claudeDir string` (result of `ClaudeDir()`) and return parsed structs plus an error. Errors are non-fatal in the data provider — empty slices are returned on failure.

## ClaudeDir()

Returns `~/.claude` (expands `$HOME`). Used as the root for all config and transcript file discovery.


## Related
- [[component-1669]]
