---
title: "Documentation Hierarchy — CLAUDE.md as Entry Point, docs/ as Reference"
type: convention
tags: [documentation, project-structure, claude-md]
---

# Documentation Hierarchy — CLAUDE.md as Entry Point, docs/ as Reference

## Convention

CLAUDE.md contains only what must be active on every session:
- Mandatory pre-completion commands (make fmt, lint, test)
- Pointers to relevant docs/ files for detailed guidelines

All deep reference material lives under `docs/` (flat structure, one file per topic):
- `docs/ui-spec.md` — UI behavior specification
- `docs/pre-completion-checklist.md` — testing and formatting standards
- `docs/refactoring-standards.md` — refactoring principles
- `docs/architecture.md` — project architecture overview
- Component, model, and convention docs for each package

## Rationale

Keeping CLAUDE.md minimal reduces system prompt token overhead. The docs/ directory is versioned in the repo and readable on demand — it does not need to be loaded on every session.

## Application

When CLAUDE.md instructs Claude to consult a docs/ file before finishing a task, Claude must read that file and apply its content. The pointer must be explicit enough that it cannot be skipped.

## Related

- [[claude-md-minimal]]
