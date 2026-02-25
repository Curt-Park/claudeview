---
confidence: 0.8
created: "2026-02-25T23:20:00+10:00"
id: component-1348
modified: "2026-02-25T23:20:00+10:00"
references: []
relations:
  - type: affects
    target: component-2468
    description: UI Spec이 키 바인딩, 모드 전환, 레이아웃 등 UI 구현 방향을 결정
    confidence: 0.8
  - type: affects
    target: component-1792
    description: UI Spec이 각 리소스 뷰의 컬럼 정의, 표시 형식, 정렬 순서를 규정
    confidence: 0.8
source: manual
status: active
tags:
  - spec
  - ui
  - navigation
title: UI Specification (docs/ui-spec.md)
type: component
---

# UI Specification — `docs/ui-spec.md`

Comprehensive specification document that defines all UI behaviors, navigation flows, and view layouts for claudeview. All UI changes must be kept in sync with this document.

## What It Covers

- Screen layout (header, table, menu/info panel, crumbs)
- Navigation model (drill-down hierarchy, flat mode, mode switching)
- Key bindings and their effects per mode
- Resource views: columns, sort order, display formats per resource type
- Log view behavior (transcript rendering, scrolling)
- Detail view behavior (key/value pane)
- Filter mode (/ key, live filtering, cursor behavior)
- Command mode (: key, resource switching with autocomplete)
- Flash messages (ephemeral status notifications)
- Info panel layout (4-column format: nav hints, context, user info, versions)
- Breadcrumb display rules

## Relationship to Code

- Each spec section maps to one or more BDD test functions (e.g., `TEST-NAV-001`)
- BDD tests in `internal/ui/bdd/` are the executable verification of this spec
- View package columns/formats should match spec column definitions
- Navigation behavior in `internal/ui/app.go` implements spec navigation rules

## See Also

- [[BDD Test Suite (internal/ui/bdd)]] — executable spec verification
- [[UI Package (internal/ui)]] — implementation


## Related
- [[component-2468]]
- [[component-1792]]
