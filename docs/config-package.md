---
title: "Config Package (internal/config)"
type: component
tags: [config, internals]
---

# Config Package — `internal/config`

Reads Claude Code configuration files from `~/.claude/`. Provides parsers for settings and plugins.

## Files

| File          | Purpose                                                                      |
|---------------|------------------------------------------------------------------------------|
| `settings.go` | `LoadSettings(claudeDir)` — parses `settings.json`; `ClaudeDir()` helper    |
| `plugins.go`  | `LoadInstalledPlugins(claudeDir)` — parses `installed_plugins.json` (v1 + v2 formats); `EnabledPlugins(claudeDir)`; `ProjectEnabledPlugins(projectRoot)`; `PluginCacheDir(...)` |
| `json.go`     | Shared JSON decoding helpers                                                 |

## Key Types

- `Settings` — top-level settings.json structure (Model, EnabledMCPJSONs, MCPServers, Hooks, Permissions)
- `MCPServer` — single MCP server config (Command, Args, Env, Type, URL)
- `InstalledPlugin` — Name, Version, Marketplace, Scope, ProjectPath, InstalledAt, CacheDir

## Usage Pattern

All functions accept `claudeDir string` (from `ClaudeDir()`) and return parsed structs plus an error. Errors are non-fatal in the data provider — empty slices returned on failure.

## ClaudeDir()

Returns `~/.claude` (expands `$HOME`). Root for all config and transcript file discovery.

## Related

- [[model-package]] — `Plugin` type populated from config data
- [[architecture]] — config package role in the data flow
