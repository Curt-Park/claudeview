---
title: "Demo Package (internal/demo)"
type: component
tags: [demo, internals]
---

# Demo Package — `internal/demo`

Generates synthetic data for `--demo` mode. Allows claudeview to be run and demonstrated without a live `~/.claude/` directory.

## Files

| File           | Purpose                                          |
|----------------|--------------------------------------------------|
| `generator.go` | All generator functions; hardcoded synthetic data |

## Exported Functions

- `GenerateProjects() []*model.Project` — returns a fixed set of synthetic projects with sessions and agents
- `GeneratePlugins() []*model.Plugin` — returns synthetic plugin entries
- `GenerateMemories() []*model.Memory` — returns synthetic memory entries

## Usage

`demoDataProvider` in `cmd/root.go` calls these functions to satisfy the `ui.DataProvider` interface when `--demo` is passed.

## No Tests

This package has no test files. Its correctness is validated visually via `make demo`.

## Related

- [[cmd-package]] — `demoDataProvider` consumes these generators
- [[model-package]] — returns types from `internal/model`
- [[architecture]] — demo package role in the data flow
