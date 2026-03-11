---
title: "lipgloss Background Layering Convention"
type: convention
tags: [ui, convention, lipgloss]
---

# lipgloss Background Layering Convention

When rendering a row where all characters should share a background color, every nested `lipgloss.Render()` call must explicitly include `Background(colorBg)`. The outer `bgStyle.Width(N).Render(content)` alone is insufficient.

## Why

Each inner `style.Render(s)` emits `\x1b[0m` (reset) at its end. This reset clears any background set by the outer wrapper for all subsequent characters until the next explicit background escape. Result: literal strings and padding between styled segments render on the terminal's default background.

## Rule

```go
// WRONG — literal parts lose background after inner Render() resets
barStyle := lipgloss.NewStyle().Foreground(fg)
content := barStyle.Render(label) + " [" + bar + "] " + barStyle.Render(pct)
return bgStyle.Width(width).Render(content)   // gaps have no bg

// CORRECT — every segment explicitly carries the background
barStyle   := lipgloss.NewStyle().Foreground(fg).Background(colorBg)
emptyStyle := lipgloss.NewStyle().Foreground(colorEmpty).Background(colorBg)
bgStyle    := lipgloss.NewStyle().Background(colorBg)
content := barStyle.Render(label) + bgStyle.Render(" [") + bar + bgStyle.Render("] ") + barStyle.Render(pct)
return bgStyle.Width(width).Render(content)   // full row covered
```

## Applies To

Any multi-segment lipgloss row that needs a uniform background (e.g., usage bar rows, styled table cells, crumb-style panels).

## Related

- [[ui-package]] — usage bar in `internal/usage/bar.go` uses this pattern
- [[usage-package]] — discovered during usage bar background fix
