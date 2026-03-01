---
title: "Test Suite"
type: component
tags: [testing, ui, internals]
---

# Test Suite

Tests span four packages. `internal/ui` has the largest test surface (integration + render), while `internal/config`, `internal/model`, and `internal/transcript` each have unit tests for their own logic. The `internal/view` and `internal/demo` packages have no test files.

## Test Files — `internal/ui` (`package ui_test`)

| File                    | Coverage                                                    |
|-------------------------|-------------------------------------------------------------|
| `app_test.go`           | AppModel integration — key flows, navigation, state transitions |
| `render_test.go`        | Full render output assertions / golden snapshots            |
| `detail_render_test.go` | `RenderPluginItemDetail` and `RenderMemoryDetail` output    |
| `filter_test.go`        | `FilterModel` unit tests                                    |
| `crumbs_test.go`        | `CrumbsModel` unit tests                                    |
| `menu_test.go`          | `MenuModel` and nav hint unit tests                         |
| `testhelpers_test.go`   | Shared helpers: `mockDP`, key senders, row builders         |

## Other Test Packages

| Package                | Files                                        | Count |
|------------------------|----------------------------------------------|-------|
| `internal/config`      | `settings_test.go`, `plugins_test.go`        | ~15   |
| `internal/model`       | `agent_test.go`, `session_test.go`, `project_test.go`, `tool_call_test.go`, `plugin_test.go`, `resource_test.go` | ~33 |
| `internal/transcript`  | `scanner_test.go`, `parser_test.go`          | ~11   |

## Pattern

Each integration test:
1. Constructs a `mockDP` (implements `DataProvider` with typed stubs)
2. Creates `AppModel` via `NewAppModel(&mockDP{}, resource)`
3. Sends `tea.KeyMsg` via `AppModel.Update()` — pure function, no program needed
4. Asserts on `AppModel.View()` output or internal state fields

## Run

```bash
make test   # go test -race -count=1 ./...
```

## Related

- [[ui-package]] — `AppModel` under test
- [[ui-spec]] — spec behaviors verified by these tests
- [[pre-completion-checklist]] — requires `make test` before completing any task
