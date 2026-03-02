---
title: "Session Chat View Design"
date: 2026-03-02
status: approved
---

# Session Chat View Design

## Overview

Replace the current `sessions → agents` leaf navigation with a rich **session-chat** view
that renders a session's conversation as a mobile-chat-style timeline. Each turn
(user, Claude, subagent) is displayed as a styled bubble with metadata, tool calls,
thinking blocks, and token counts.

---

## Navigation Change

```
[Before]  projects → sessions → agents  [leaf]
[After]   projects → sessions → session-chat  [leaf]
```

- `Sessions → enter` → `session-chat`
- `session-chat → esc` → back to Sessions
- Breadcrumb: `sessions > session-chat`

---

## Visual Layout

### User bubble (right-aligned, ~70% width)

```
             ╭─ 09:13 · You ──────────────────────────╮
             │ Agent 뷰를 대폭 개선해볼 생각이다.        │
             ╰────────────────────────────────────────╯
```

Multi-line messages expand the bubble height naturally; newlines are preserved.

### Claude bubble (left-aligned, full width)

```
╭─ 🤖 Claude · sonnet · 09:14 ──────────────────────────────────╮
│ ···thinking: 사용자가 뷰 개선을 원하는 것 같다···              │  (dim)
│                                                               │
│ 네, 작업을 시작하겠습니다. 먼저 파일 구조를 확인할게요.          │
│                                                               │
│  ▸ Read   internal/ui/app.go        →  120 lines             │
│  ▸ Grep   "ResourceType"            →  8 files               │
│  ▸ Bash   make test                 →  OK ✓                  │
│  ▸ Task   Explore codebase          →  [subagent ↓]          │
│                                            ░ 1,234 tok       │
╰───────────────────────────────────────────────────────────────╯
```

### Subagent bubble (indented, narrower)

```
  └─ 🔍 Explorer · 09:15 ─────────────────────────────────╮
     │ 파일을 찾았습니다.                                    │
     │  ▸ Glob  "**/*.go"  →  42 files                    │
     │                              ░ 234 tok             │
     ╰────────────────────────────────────────────────────╯
```

---

## Bubble Specifications

### Color scheme

| Sender        | Border color    | Position        | Icon |
|---------------|-----------------|-----------------|------|
| You (user)    | `"33"` (blue)   | Right, 70% wide | —    |
| Claude (main) | `"82"` (green)  | Left, full wide | 🤖   |
| Subagent      | `"135"` (purple)| Left, indented  | 🔍/📋/💻 |

### Tool call lines (inside Claude/subagent bubble)

| Outcome | Format |
|---------|--------|
| Success | `  ▸ Read  path/to/file  →  42 lines  ✓` |
| Error   | `  ▸ Bash  command       →  error  ✗`  (red foreground) |
| Subagent| `  ▸ Task  description   →  [subagent ↓]` |

### Thinking block

Rendered as the first line(s) inside the Claude bubble, dim (color `"243"`):
```
│ ···thinking: {first 80 chars of thinking text}···              │
```

### Token line

Right-justified inside the bubble footer:
```
│                                            ░ 1,234 tok       │
```

### Subagent icons

| Agent type      | Icon |
|-----------------|------|
| Explore         | 🔍   |
| Plan            | 📋   |
| Bash            | 💻   |
| general-purpose | ⚙️   |
| (unknown)       | 🤖   |

---

## Width Handling

```
userBubbleWidth  = min(int(termWidth * 0.70), termWidth - 4)
claudeBubbleWidth = termWidth - 2   // border takes 2 cols
subagentWidth    = termWidth - 6    // 4-col indent + border
```

For very narrow terminals (< 60 cols), user bubble falls back to full width.

---

## Data Flow

### On `drillDown()` from Sessions

```go
t, err := transcript.ParseFile(session.FilePath)
// store in AppModel.SelectedTranscript
m.drillInto(model.ResourceSessionChat)
```

Parsing is synchronous on drill-down (acceptable for MVP).

### Subagent matching

- Call `DataProvider.GetAgents(sessionID)` to get the agent list (already available).
- Task tool calls in the main transcript are matched to subagents by index order
  (1st Task → 1st subagent, etc.).
- Each subagent's turns are parsed from `agent.FilePath` on demand.

---

## Scrolling

Reuses existing `ContentOffset` mechanism (same as `memory-detail`, `plugin-item-detail`).

- `j/k`, `g/G`, `ctrl+d/u`, `pgdn/pgup`
- `isContentView()` extended to include `ResourceSessionChat`
- `contentMaxOffset()` extended to compute line count for session-chat

---

## Files to Change

| File | Change |
|------|--------|
| `internal/model/resource.go` | Add `ResourceSessionChat = "session-chat"` |
| `internal/ui/app.go` | `SelectedTranscript` field; `drillDown()` for sessions → session-chat; `isContentView()` + `contentMaxOffset()` + `View()` extended; `navigateBack()` for session-chat |
| `internal/ui/detail_render.go` | Add `RenderSessionChat()` |
| `internal/ui/styles.go` | Add bubble border styles |
| `internal/ui/menu.go` | Nav hints for `session-chat` |

---

## Out of Scope (Future)

- Full subagent turn interleaving by timestamp (currently matched by index order)
- Multi-session continuity ("plan → writing" session chaining)
- `bubbles/viewport` for smoother scrolling
- Glamour markdown rendering inside bubbles
- Toggle to expand/collapse thinking or tool results
