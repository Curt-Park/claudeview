# Design: Split Cache-Read Tokens from TOKEN_IN Display

## Problem

`TOKEN_IN` currently shows `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` summed across all API turns. In long sessions this is misleading: a 105-turn session showed 7.6M "input tokens" where 7.4M were cache reads — context that Claude re-used cheaply each turn, not unique information sent by the user.

## Goal

Show `IN` (genuinely new tokens) and `CACHE` (re-used context) separately so token consumption is not overstated.

## Display Format

Column header: `MODEL:IN+CACHE/OUT`

Example value: `haiku:243k+7.4M/26k`

- `IN` = `input_tokens + cache_creation_input_tokens` (new tokens sent this turn)
- `CACHE` = `cache_read_input_tokens` (context re-used from cache)
- `OUT` = `output_tokens`

## Data Model Changes

### `internal/transcript/types.go`
- Add `NewInputTokens() int` → `InputTokens + CacheCreationInputTokens`
- Keep `TotalInputTokens()` (now only used if a full sum is ever needed)

### `internal/model/turn.go`
- Add `CacheReadTokens int`

### `internal/model/session.go` (`TokenCount`)
- Add `CacheReadTokens int`

## Data Flow Changes

### `internal/transcript/parser.go`
Three call sites switch from `TotalInputTokens()` to `NewInputTokens()`, with a parallel `CacheReadInputTokens` accumulation added:
- `flushPendingTurn`: accumulate `u.CacheReadTokens += turn.Usage.CacheReadInputTokens`
- `mergeTurns`: accumulate cache reads on the pending turn's Usage
- live session aggregation: same pattern

### `internal/provider/live.go`
- Turn conversion: `InputTokens: t.Usage.NewInputTokens()`, add `CacheReadTokens`
- Subagent merge loop: same pattern

## Display Changes

### `internal/model/format.go`
Add:
```go
func FormatTokenInOutCache(in, cache, out int) string {
    return FormatTokenCount(in) + "+" + FormatTokenCount(cache) + "/" + FormatTokenCount(out)
}
```

### Three call sites updated
- `model/session.go` `TokenString()`
- `ui/chat_item.go` `ModelTokenLabel()`
- `ui/app.go` `buildToolCallSubRow()`

### Column headers
`internal/view/chat.go` and `internal/view/sessions.go`:
```
"MODEL:TOKEN_IN/OUT"  →  "MODEL:IN+CACHE/OUT"
```
