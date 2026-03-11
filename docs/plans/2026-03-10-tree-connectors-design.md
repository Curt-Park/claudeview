# Tree Connectors for Sub-agent Rows

**Date:** 2026-03-10

## Goal

Show tree connectors (├─ / └─) in the NAME column of the chat history table to visualise the parent-child relationship between a Claude turn and its sub-agents.

**Target layout:**
```
NAME            MESSAGE            ACTION         MODEL:TOKEN  DURATION
Claude          planning next..    Agent:Plan     opus:1.5k    3s
└─ Planner      reading files..    Read+3         son:800      2s
You             looks good         -              -            5s
Claude          implementing..     Agent:Explore  son:2k       4s
├─ Explorer     searching..        Glob+2         son:600      1s
└─ Explorer     found matches..    Read+4         son:1.4k     3s
```

## Constraints

- One row per sub-agent (collapsed model — not one row per turn).
- `└─` / `├─` are scoped to each parent turn's batch of sub-agents:
  - A lone sub-agent gets `└─`.
  - Multiple sub-agents: all but the last get `├─`, the last gets `└─`.

## Design

### Option chosen: `TreeConnector` field on `ChatItem` (Option A)

Add one field to `ChatItem` in `internal/ui/chat_item.go`:

```go
TreeConnector string // "├─", "└─", or "" for non-sub-agent items
```

**`interleaveSubagents` closure** — pre-count Agent/Task calls in the current batch, then assign connectors by position:

```go
agentCallCount := 0
for _, tc := range toolCalls {
    if tc.Name == "Task" || tc.Name == "Agent" { agentCallCount++ }
}
batchPos := 0
// for each agent call:
connector := "├─"
if batchPos == agentCallCount-1 { connector = "└─" }
batchPos++
```

**`WhoLabel()`** — prepend connector for sub-agent items:

```go
if c.IsSubagent && c.TreeConnector != "" {
    return c.TreeConnector + " " + agentDisplayName(c.AgentType)
}
```

### View change (`internal/view/chat.go`)

NAME column width: `10` → `14` (fits `"└─ Explorer"` = 11 chars with margin).

### Tests (`internal/ui/chat_item_test.go`)

- Update existing tests that assert on `WhoLabel()` or `ChatItem` structure.
- Add: single Agent call → connector is `"└─"`; two Agent calls → `"├─"` then `"└─"`.

## Files to modify

| File | Change |
|------|--------|
| `internal/ui/chat_item.go` | Add `TreeConnector` field; update `interleaveSubagents`; update `WhoLabel()` |
| `internal/view/chat.go` | NAME column width 10 → 14 |
| `internal/ui/chat_item_test.go` | Update/add tests |
