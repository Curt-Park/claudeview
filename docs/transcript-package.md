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
| `types.go`  | `ParsedTranscript`, `Turn`, `ToolCallEntry`, `SessionInfo`, `ProjectInfo`, `SessionAggregates`, `Usage` — intermediate parsing types |
| `parser.go` | `ParseFile(path) (*ParsedTranscript, error)` — reads JSONL line by line; extracts turns, tool calls, cost, token usage |
| `scanner.go`| `ScanProjects(claudeDir) ([]ProjectInfo, error)` — walks `~/.claude/projects/`; `ScanSubagents(dir) ([]SessionInfo, error)` |

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

- `ParseFile(path)` — full parse; used for session metadata extraction
- `Parse(r io.Reader)` — parse from any reader
- `ParseAggregatesIncremental(path, agg)` — offset-based re-read; avoids re-parsing from the beginning on each refresh tick
- `ScanProjects(claudeDir)` — enumerate all projects+sessions; used by liveDataProvider
- `ScanSubagents(dir)` — enumerate subagent transcripts for a session
- `CountSubagents(dir)` — count subagent transcripts without full enumeration

## Related

- [[model-package]] — Session, Agent, ToolCall types populated by this package
- [[architecture]] — transcript package role in the data flow
