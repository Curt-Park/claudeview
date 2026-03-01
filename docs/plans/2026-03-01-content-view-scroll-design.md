---
title: "Content View Scrolling + Sub-View Key Blocking"
type: decision
tags: [ui, navigation, keybindings, views]
---

# Content View Scrolling + Sub-View Key Blocking

## Problem

Two issues with content-only views (`plugin-item-detail`, `memory-detail`):

1. **Movement keys don't scroll**: j/k/G/g/ctrl+d/u silently mutate the
   background table selection with no visible effect.
2. **p/m jump keys disrupt navigation**: Pressing `p` or `m` while inside
   a plugin or memory sub-view performs a flat jump that discards hierarchical context.

## Approach: ContentOffset + isSubView guard

### 1. `AppModel.ContentOffset int`

A dedicated scroll offset for content views (views rendering flat text rather than
a table). Follows the same pattern as `Table.Offset`.

**Scroll key mapping:**

| Key | Action |
|-----|--------|
| `j` | offset++ |
| `k` | offset-- (floor 0) |
| `ctrl+d` | offset += contentHeight/2 |
| `ctrl+u` | offset -= contentHeight/2 (floor 0) |
| `G` | offset = sentinel (View() caps to actual max) |
| `g` | offset = 0 |

**Cap strategy**: `View()` always applies `min(offset, max(0, len(lines)-height))`,
so `G` can safely set a sentinel value (e.g. `1<<30`).

**Reset**: `drillInto()`, `navigateBack()`, `jumpTo()` all reset ContentOffset to 0.

### 2. Content view detection

```go
func isContentView(rt model.ResourceType) bool {
    return rt == model.ResourcePluginItemDetail || rt == model.ResourceMemoryDetail
}
```

Used in `updateTable()` default branch: route key messages to `updateContentScroll()`
instead of `m.Table.Update()`.

Used in `View()`: apply ContentOffset slicing before the height-cap.

### 3. Sub-view key blocking

```go
func isSubView(rt model.ResourceType) bool {
    return rt == model.ResourcePluginDetail ||
           rt == model.ResourcePluginItemDetail ||
           rt == model.ResourceMemoryDetail
}
```

`p` key: skip `jumpTo(ResourcePlugins)` when `isSubView(m.Resource)`.
`m` key: skip `jumpTo(ResourceMemory)` when `isSubView(m.Resource)`.

## Files Changed

| File | Change |
|------|--------|
| `internal/ui/app.go` | ContentOffset field, helpers, updateContentScroll(), View() slicing, p/m guard, resets |
| `internal/ui/app_test.go` | Scroll tests, p/m blocking tests |
| `internal/ui/menu.go` | No change (j/k hints already present; now they actually work) |

## Non-Goals

- Viewport component (charmbracelet/bubbles/viewport) — unnecessary complexity
- Preserving scroll position across navigation — always reset to top
- Horizontal scrolling
