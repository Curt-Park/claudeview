---
title: "Demo Package (internal/demo)"
type: component
tags: [demo, internals]
---

# Demo Package — `internal/demo`

Generates synthetic data for `--demo` mode. Allows claudeview to be run and demonstrated without a live `~/.claude/` directory.

## Files

| File           | Purpose                                                         |
|----------------|-----------------------------------------------------------------|
| `generator.go` | All generator functions; hardcoded synthetic data               |
| `provider.go`  | `Provider` struct; `NewProvider() ui.DataProvider`              |

## Exported Functions

- `NewProvider() ui.DataProvider` — constructs a `Provider` backed by the generators; used by `cmd/root.go` with `--demo`
- `GenerateProjects() []*model.Project` — returns a fixed set of synthetic projects with sessions and agents
- `GeneratePlugins() []*model.Plugin` — returns synthetic plugin entries
- `GenerateMemories() []*model.Memory` — returns synthetic memory entries
- `GenerateUsage() *usage.Data` — returns synthetic usage data for `--demo` mode (5h: 8%, resets in ~4h9m; 7d: 68%, resets in ~1d2h). Called by `cmd/root.go` to set initial `usageData`, bypassing the HTTP fetch.

## Usage

`cmd/root.go` calls `demo.NewProvider()` when `--demo` is passed. `Provider` implements `ui.DataProvider` by delegating each method to the `Generate*` functions.

## No Tests

This package has no test files. Its correctness is validated visually via `make demo`.

## Related

- [[cmd-package]] — calls `demo.NewProvider()` to wire the demo `DataProvider`
- [[model-package]] — returns types from `internal/model`
- [[architecture]] — demo package role in the data flow
- [[usage-package]] — `GenerateUsage()` provides synthetic `*usage.Data` for demo mode
