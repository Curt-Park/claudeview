---
confidence: 0.8
created: "2026-02-25T23:19:45+10:00"
id: component-1712
modified: "2026-02-25T23:19:45+10:00"
references: []
relations:
  - type: relates_to
    target: component-1287
    description: 'High tag overlap: 10%'
    confidence: 0.7000000000000001
  - type: relates_to
    target: component-1669
    description: transcript 파싱 결과(SessionInfo, Turn 등)가 cmd/root.go에서 model 타입으로 변환됨
    confidence: 0.8
source: manual
status: active
tags:
  - internals
  - transcript
  - parser
title: Transcript Package (internal/transcript)
type: component
---

# Transcript Package — `internal/transcript`

Reads and watches Claude Code JSONL transcript files. Provides the data backbone for the live data provider.

## Files

| File | Purpose |
|------|---------|
| `types.go` | `ParsedTranscript`, `Turn`, `ToolCallEntry`, `SessionInfo`, `ProjectInfo` — intermediate parsing types |
| `parser.go` | `ParseFile(path) (*ParsedTranscript, error)` — reads a JSONL file line by line, extracts turns, tool calls, cost, token usage |
| `scanner.go` | `ScanProjects(claudeDir) ([]ProjectInfo, error)` — walks `~/.claude/projects/` to enumerate projects and their sessions; `ScanSubagents(dir) ([]SessionInfo, error)` |
| `watcher.go` | `Watcher` — fsnotify-based file watcher; emits `RefreshMsg` to Bubble Tea program when transcript files change |

## JSONL Format

Claude Code writes one JSON object per line. Each line can be:
- A conversation turn (role: "user" | "assistant") with text, thinking, tool_use, tool_result
- A summary/metadata record with cost/token usage

`ParseFile` accumulates turns and aggregates totals (TotalCost, NumTurns, DurationMS).

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

- `ParseFile(path)` — full parse; used for log view and session metadata extraction
- `ScanProjects(claudeDir)` — enumerate all projects+sessions; used by liveDataProvider
- `ScanSubagents(dir)` — enumerate subagent transcripts for a session
- `Watcher.Watch(paths, program)` — start background file watching


## Related
- [[component-1287]]
- [[component-1669]]
