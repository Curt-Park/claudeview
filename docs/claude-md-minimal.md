---
title: "Keep CLAUDE.md Minimal — Reference docs/ for Detailed Guidelines"
type: decision
tags: [claude-md, documentation, dx, system-prompt]
---

# Keep CLAUDE.md Minimal — Reference docs/ for Detailed Guidelines

## Context

CLAUDE.md is injected as a system prompt on every Claude Code session for the claudeview project. As the file grew to include detailed refactoring principles, toolchain tables, and project structure, it increased token overhead on every request — even when that detail was not relevant to the task at hand. Most of the content is reference material consulted rarely, not directives needed every session.

## Decision

Keep CLAUDE.md as a lightweight entry point containing only high-level directives and pointers to `docs/` files. Detailed testing and refactoring guidelines live in dedicated files under `docs/`. CLAUDE.md instructs Claude to read the relevant docs/ file before finishing a task, rather than embedding the full content inline.

## Alternatives Considered

- **All content in CLAUDE.md**: Rejected. Every session pays the token cost regardless of task relevance.
- **No documentation at all**: Rejected. Without structure, conventions drift and quality degrades.

## Consequences

### Positive
- System prompt stays lean — only actionable per-session directives loaded
- Detailed guidelines versioned and searchable under `docs/`
- Easier to update guidelines without touching the system-prompt-critical CLAUDE.md

### Negative
- Claude must explicitly read the docs/ file (one extra tool call)
- Risk of skipping the docs/ read if the CLAUDE.md pointer is unclear

## Related

- [[docs-hierarchy]]
- [[pre-completion-checklist]]
