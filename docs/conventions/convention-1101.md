---
confidence: 0.8
created: "2026-02-26T19:38:29+10:00"
id: convention-1101
modified: "2026-02-26T19:38:29+10:00"
references: []
relations:
  - type: implements
    target: decision-1695
    description: This convention is the concrete application of the decision to keep CLAUDE.md minimal
    confidence: 0.8
source: manual
status: active
tags:
  - claude-md
  - documentation
  - docs
  - project-structure
title: Documentation hierarchy — CLAUDE.md as entry point, docs/ as reference
type: convention
---

# Documentation Hierarchy — CLAUDE.md as Entry Point, docs/ as Reference

## Convention
CLAUDE.md contains only what must be active on every session:
- Mandatory pre-completion commands (make fmt, lint, test, bdd)
- A pointer to the relevant docs/ file for detailed guidelines

All deep reference material lives under `docs/`:
- `docs/ui-spec.md` — UI behavior specification
- `docs/conventions/` — coding and refactoring standards
- `docs/decisions/` — architectural decision records (autology-managed)

## Rationale
Keeping CLAUDE.md minimal reduces system prompt token overhead. The docs/ directory is versioned in the repo and readable on demand — it does not need to be loaded on every session.

## Application
When CLAUDE.md instructs Claude to consult a docs/ file before finishing a task, Claude must read that file via the Read tool and apply its content. The pointer must be explicit enough that it cannot be skipped.

## References
- `docs/ui-spec.md`
- `docs/conventions/`
- Related decision: decision-1695


## Related
- [[decision-1695]]
