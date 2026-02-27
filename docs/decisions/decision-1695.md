---
confidence: 0.8
created: "2026-02-26T19:38:07+10:00"
id: decision-1695
modified: "2026-02-26T19:38:07+10:00"
references: []
relations: []
source: manual
status: active
tags:
  - claude-md
  - documentation
  - system-prompt
  - dx
title: Keep CLAUDE.md minimal — reference docs/ for detailed guidelines
type: decision
---

# Keep CLAUDE.md Minimal — Reference docs/ for Detailed Guidelines

## Context
CLAUDE.md is injected as a system prompt on every Claude Code session for the claudeview project. As the file grew to include detailed refactoring principles, toolchain tables, and project structure, it increased token overhead on every request — even when that detail was not relevant to the task at hand. Most of the content is reference material consulted rarely, not directives needed every session.

## Decision
Keep CLAUDE.md as a lightweight entry point containing only high-level directives and pointers to `docs/` files. Detailed testing and refactoring guidelines live in a dedicated file under `docs/` (e.g., `docs/conventions/development-guide.md`). CLAUDE.md instructs Claude to read the relevant docs/ file before finishing a task, rather than embedding the full content inline.

## Alternatives Considered
- **All content in CLAUDE.md**: Rejected. Every session pays the token cost regardless of task relevance. The file becomes harder to maintain as a living document.
- **No documentation at all**: Rejected. Without structure, conventions drift and quality degrades across sessions.

## Consequences
### Positive
- System prompt stays lean — only actionable per-session directives are loaded
- Detailed guidelines remain versioned and searchable in the repo under `docs/`
- Easier to update guidelines without touching the system-prompt-critical CLAUDE.md

### Negative
- Claude must explicitly read the docs/ file when needed (one extra tool call)
- Risk of Claude skipping the docs/ read if the CLAUDE.md pointer is unclear
