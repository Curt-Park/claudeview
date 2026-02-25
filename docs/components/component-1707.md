---
confidence: 0.8
created: "2026-02-25T23:20:08+10:00"
id: component-1707
modified: "2026-02-25T23:20:08+10:00"
references: []
relations:
  - type: implements
    target: component-1348
    description: BDD test functions (TEST-NAV-001 etc.) verify the behavior of each UI Spec section in executable form
    confidence: 0.8
  - type: depends_on
    target: component-2468
    description: BDD tests drive AppModel directly and assert on rendered output
    confidence: 0.8
source: manual
status: active
tags:
  - testing
  - bdd
  - ui
title: BDD Test Suite (internal/ui/bdd)
type: component
---

# BDD Test Suite — `internal/ui/bdd/`

9 teatest-based integration tests covering spec behaviors. These tests drive the full Bubble Tea app model end-to-end and verify rendered output.

## Test Files

| File | Test IDs / Coverage |
|------|---------------------|
| `navigation_test.go` | TEST-NAV-001 — keyboard navigation (j/k, enter drill-down, esc back) |
| `drilldown_test.go` | Drill-down from projects → sessions → agents → tools |
| `filter_test.go` | Filter mode (/ key, live filtering, cursor `/█`, esc to exit) |
| `flash_test.go` | Flash message display on resource switch |
| `viewmode_test.go` | d/l/y/esc view mode switching |
| `parent_columns_test.go` | Parent context columns shown in flat-mode resource views |
| `resize_test.go` | Window resize handling |
| `initial_test.go` | Initial state and startup resource |
| `info_context_test.go` | Info panel context (project/session/user/version display) |
| `helpers_test.go` | Shared test helpers (mock DataProvider, key send utilities) |

## Framework

Uses `github.com/charmbracelet/x/exp/teatest` — renders frames to a buffer and asserts on string content. Tests run with `make bdd` (`go test -race -count=1 ./internal/ui/bdd/...`).

## Pattern

Each test:
1. Constructs a mock `DataProvider` with synthetic data
2. Creates `AppModel` + `rootModel`
3. Sends key messages via `teatest`
4. Asserts rendered output contains expected strings

## Relationship to Spec

- Test IDs (TEST-NAV-001, etc.) map directly to sections in [[UI Specification (docs/ui-spec.md)]]
- Adding new spec behavior → add corresponding BDD test


## Related
- [[component-1348]]
- [[component-2468]]
