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
| `types.go`  | Wire types for JSONL decoding: `entry` (single JSONL line, includes `Slug` field), `messageContent` (polymorphic content block), `Usage` (token counts: InputTokens, OutputTokens, CacheCreationInputTokens, CacheReadInputTokens), `assistantMessage`, `userMessage` (with `textContent()` and `toolResults()` helpers) |
| `parser.go` | `ParsedTranscript`, `Turn`, `ToolCall`, `SessionAggregates` (includes `Slug`), `TranscriptCache` — intermediate parsing types; `ParseFile(path)`, `Parse(r)`, `ParseAggregatesIncremental(path, agg)` (captures slug from first entry), `ParseFileIncremental(path, cache)` |
| `scanner.go`| `SessionInfo`, `ProjectInfo` — directory scan types; `ScanProjects(claudeDir)` (uses `parallel.Map` for concurrent directory scanning), `ScanSubagents(dir)`, `CountSubagents(dir)` |

## JSONL Format

Claude Code writes one JSON object per line. Each line is either:
- A conversation turn (`role: "user" | "assistant"`) with text, thinking, tool_use, tool_result
- A summary/metadata record with cost/token totals

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
- `ParseAggregatesIncremental(path, agg)` — offset-based re-read for session-level metrics; avoids re-parsing from the beginning on each refresh tick
- `ParseFileIncremental(path, cache)` — offset-based incremental turns parsing via `TranscriptCache`; used by `provider.Live.GetTurns` for the history view. `TranscriptCache` tracks committed turns, a pending assistant turn, and unmatched tool results across calls. `Turns()` returns a snapshot including the pending turn; `Offset()` exposes the read position
- `ScanProjects(claudeDir)` — enumerate all projects+sessions using `parallel.Map` (from [[parallel-package]]) for concurrent directory scanning; used by [[provider-package]]
- `ScanSubagents(dir)` — enumerate subagent transcripts for a session
- `CountSubagents(dir)` — count subagent transcripts without full enumeration

## Related

- [[model-package]] — Session, Agent, ToolCall types populated by this package
- [[architecture]] — transcript package role in the data flow
- [[stringutil-package]] — `ExtractXMLTag` used in `extractTopic`
- [[parallel-package]] — `ScanProjects` uses `parallel.Map` for concurrent scanning
- [[provider-package]] — primary consumer of this package's functions
