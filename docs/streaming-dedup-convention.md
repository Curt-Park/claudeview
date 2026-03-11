---
title: "Claude Code JSONL Streaming Deduplication"
type: convention
tags: [transcript, parser, internals, convention]
---

# Claude Code JSONL Streaming Deduplication

## JSONL Streaming Format

Claude Code's JSONL format writes multiple assistant entries per API call — streaming events — all sharing the same `requestId` field on the top-level JSONL entry, interleaved with user/progress entries. Each streaming event for the same API call is a **complete snapshot** (not a delta), so the last event has the most accurate token counts and tool calls.

## The Bug This Fixed

The parser was treating each streaming event as a separate turn, overcounting tokens by ~1.8x in long sessions.

## Convention

When an assistant entry's `requestId` is non-empty and matches a previously seen entry's `requestId`, **REPLACE** the previous entry's data instead of accumulating/merging.

## Three Code Paths That Enforce This

1. **`mergeAssistantTurn`** — consecutive entries: when both share the same non-empty `RequestID`, replaces the pending turn wholesale instead of accumulating.

2. **`flushPendingTurn` / `ParseFileIncremental` flush** — interleaved entries: when the last committed turn shares the same non-empty `RequestID`, replaces it and undoes/re-does token accounting.

3. **`ParseAggregatesIncremental`** — aggregate metrics: when the same `requestId` is seen again, undoes the previous accumulation before re-accumulating.

## SessionAggregates Tracking Fields

`SessionAggregates` carries 4 unexported fields to track the last assistant entry across incremental calls so dedup can undo/redo across call boundaries:

- `lastRequestID`
- `lastRequestModel`
- `lastRequestUsage`
- `lastToolCallDelta`

## Related

- [[transcript-package]] — the package where all three code paths are implemented
- [[architecture]] — overall data flow context
