---
title: "Transcript Package (internal/transcript)"
type: component
tags: [transcript, parser, internals]
---

# Transcript Package — `internal/transcript`

Reads Claude Code JSONL transcript files. Provides the data backbone for the live data provider.

## Files

| File        | Purpose                                                                           |
|-------------|-----------------------------------------------------------------------------------|
| `types.go`  | Wire types for JSONL decoding: `entry` (single JSONL line, includes `Slug` and `RequestID` fields), `messageContent` (polymorphic content block), `Usage` (token counts: InputTokens, OutputTokens, CacheCreationInputTokens, CacheReadInputTokens), `assistantMessage`, `userMessage` (with `textContent()` and `toolResults()` helpers) |
| `parser.go` | `ParsedTranscript`, `Turn` (includes `RequestID string` for streaming dedup), `ToolCall`, `SessionAggregates` (includes `Slug` and 4 unexported streaming-dedup fields), `TranscriptCache` — intermediate parsing types; `ParseFile(path)`, `Parse(r)`, `ParseAggregatesIncremental(path, agg)` (captures slug from first entry), `ParseFileIncremental(path, cache)` |
| `scanner.go`| `SessionInfo`, `ProjectInfo` — directory scan types; `ScanProjects(claudeDir)` (uses `parallel.Map` for concurrent directory scanning), `ScanSubagents(dir)`, `CountSubagents(dir)` |

## JSONL Format

Claude Code writes one JSON object per line. Each line is either:
- A conversation turn (`role: "user" | "assistant"`) with text, thinking, tool_use, tool_result
- A summary/metadata record with cost/token totals

Each `entry` carries an optional `requestId` field. When Claude Code writes streaming responses, it may emit multiple assistant entries for the same API request, each with the same `requestId`. The parser deduplicates these: only the final entry's data is kept, preventing double-counted tokens and duplicate tool calls.

`ParseFile` accumulates turns and aggregates TotalCost, NumTurns, DurationMS.

## Directory Layout Expected

```
~/.claude/
  projects/
    <hash>/           ← ProjectInfo.Hash, decoded path → ProjectInfo.Path
      <session-id>.jsonl
      <session-id>/   ← SubagentDir
        <subagent-id>.jsonl
```

## Key Functions

- `ParseFile(path)` — full parse; used for agent metadata extraction
- `Parse(r io.Reader)` — parse from any reader
- `mergeAssistantTurn(pending, next)` — merges consecutive assistant entries into the pending turn; when both share the same non-empty `RequestID`, replaces the pending turn wholesale (streaming dedup) instead of accumulating
- `flushPendingTurn(result, turn, ...)` — matches tool results into the pending turn, accumulates metrics, and appends it; when the last committed turn shares the same non-empty `RequestID`, replaces it and undoes/re-does token accounting (streaming dedup for interleaved entries)
- `ParseAggregatesIncremental(path, agg)` — offset-based re-read for session-level metrics; avoids re-parsing from the beginning on each refresh tick. When the same `requestId` is seen again, undoes the previous accumulation before re-accumulating (streaming dedup). `SessionAggregates` carries 4 unexported streaming-dedup fields: `lastRequestID`, `lastRequestModel`, `lastRequestUsage`, `lastToolCallDelta`
- `ParseFileIncremental(path, cache)` — offset-based incremental turns parsing via `TranscriptCache`; used by `provider.Live.GetTurns` for the history view. `TranscriptCache` tracks committed turns, a pending assistant turn, and unmatched tool results across calls. At flush time, if the last committed turn shares the same non-empty `RequestID`, it is replaced instead of appended (streaming dedup). `Turns()` returns a snapshot including the pending turn; `Offset()` exposes the read position
- `ScanProjects(claudeDir)` — enumerate all projects+sessions using `parallel.Map` (from [[parallel-package]]) for concurrent directory scanning; used by [[provider-package]]
- `ScanSubagents(dir)` — enumerate subagent transcripts for a session
- `CountSubagents(dir)` — count subagent transcripts without full enumeration

## Related

- [[model-package]] — Session, Agent, ToolCall types populated by this package
- [[architecture]] — transcript package role in the data flow
- [[stringutil-package]] — `ExtractXMLTag` used in `extractTopic`
- [[parallel-package]] — `ScanProjects` uses `parallel.Map` for concurrent scanning
- [[provider-package]] — primary consumer of this package's functions
- [[streaming-dedup-convention]] — convention for handling streaming deduplication across all three code paths
