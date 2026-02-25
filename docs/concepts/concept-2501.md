---
confidence: 0.8
created: "2026-02-25T23:19:04+10:00"
id: concept-2501
modified: "2026-02-25T23:19:04+10:00"
references: []
relations:
  - type: relates_to
    target: concept-2723
    description: 'High tag overlap: 10%'
    confidence: 0.7000000000000001
  - type: relates_to
    target: component-2468
    description: 아키텍처 개요가 UI 패키지를 핵심 컴포넌트로 문서화
    confidence: 0.8
  - type: relates_to
    target: component-1792
    description: 아키텍처 개요가 View 패키지를 핵심 컴포넌트로 문서화
    confidence: 0.8
  - type: relates_to
    target: component-1712
    description: 아키텍처 개요가 Transcript 패키지를 핵심 컴포넌트로 문서화
    confidence: 0.8
  - type: relates_to
    target: component-1669
    description: 아키텍처 개요가 Model 패키지를 핵심 컴포넌트로 문서화
    confidence: 0.8
  - type: relates_to
    target: component-1365
    description: 아키텍처 개요가 Config 패키지를 핵심 컴포넌트로 문서화
    confidence: 0.8
  - type: relates_to
    target: component-1707
    description: 아키텍처 개요가 BDD 테스트 스위트를 품질 보증 레이어로 문서화
    confidence: 0.8
  - type: relates_to
    target: component-1348
    description: 아키텍처 개요가 UI Spec을 동작 정의 문서로 참조
    confidence: 0.8
source: manual
status: active
tags:
  - architecture
  - overview
title: claudeview — Project Architecture Overview
type: concept
---

# claudeview — Project Architecture Overview

claudeview is a Go TUI (terminal UI) application for monitoring and exploring Claude Code sessions. It reads JSONL transcript files written by Claude Code and presents them in an interactive Bubble Tea dashboard.

## Entry Points

- `main.go` — sets `AppVersion` from build-time ldflags, calls `cmd.Execute()`
- `cmd/root.go` — Cobra CLI; defines `--demo`, `--project`, `--resource`, `--render-once` flags; wires together `DataProvider`, `AppModel`, and `rootModel`

## Top-Level Architecture

```
main.go
  └─ cmd/root.go (Cobra)
       ├─ rootModel       — wraps AppModel; owns all resource data & view instances
       ├─ ui.AppModel     — Bubble Tea model; handles keyboard, layout, mode switching
       ├─ ui.DataProvider — interface; two implementations:
       │    ├─ liveDataProvider  — reads ~/.claude/ via transcript + config packages
       │    └─ demoDataProvider  — returns synthetic data from internal/demo
       └─ view.ResourceView[T] — generic resource table renderer; one constructor per resource type
```

## Internal Packages

| Package | Role |
|---------|------|
| `internal/transcript` | JSONL parser, directory scanner, file watcher |
| `internal/config` | settings.json, installed_plugins.json, tasks/*.json parsers |
| `internal/model` | Data models: Project, Session, Agent, ToolCall, Task, Plugin, MCPServer, Resource |
| `internal/ui` | Bubble Tea AppModel + chrome components (header, menu, crumbs, flash, filter, table/log/detail views) |
| `internal/view` | Generic `ResourceView[T]` + 7 resource constructors + `detail.go` |
| `internal/demo` | Synthetic demo data generator |

## Resource Hierarchy (drill-down navigation)

```
projects → sessions → agents → tool calls
                   → tasks
plugins (flat)
mcp servers (flat)
```

## Key Design Decisions

- `rootModel` owns all data loading and view lifecycle; `AppModel` owns only UI state
- `DataProvider` interface returns concrete slice types (`[]*model.Project`, etc.) — no `any` casts
- `ResourceView[T]` (generic) unifies all 7 resource views; `Sync()` replaces per-view `Set*()` + lazy init
- Views are eagerly initialized in `newRootModel()`; cursor/scroll/filter state preserved across refreshes
- Session "active" status is inferred from JSONL mod time < 5 minutes ago

## Module / Build

- Module path: `github.com/Curt-Park/claudeview`
- Go version: 1.26+
- Binary: `bin/claudeview` (~5.6 MB)
- `make build` / `make demo` / `make test` / `make bdd` / `make lint`


## Related
- [[concept-2723]]
- [[component-2468]]
- [[component-1792]]
- [[component-1712]]
- [[component-1669]]
- [[component-1365]]
- [[component-1707]]
- [[component-1348]]
