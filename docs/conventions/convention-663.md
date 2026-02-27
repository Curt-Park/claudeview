---
confidence: 0.8
created: "2026-02-26T19:46:35+10:00"
id: convention-663
modified: "2026-02-26T19:46:35+10:00"
references: []
relations:
  - type: relates_to
    target: convention-3108
    description: Pre-completion checklist and refactoring standards are both mandatory before finishing any task
    confidence: 0.8
  - type: implements
    target: decision-1695
    description: Pre-completion checklist is the concrete testing directive referenced from the minimal CLAUDE.md
    confidence: 0.8
source: manual
status: active
tags:
  - testing
  - ci
  - make
  - pre-completion
  - quality-gate
title: Pre-completion checklist — make fmt, lint, test
type: convention
---

# Pre-Completion Checklist

Run the following commands in order and confirm all pass before finishing any task:

```bash
make fmt    # Format all Go source files (gofmt + goimports)
make lint   # Static analysis via golangci-lint — enforce style and catch bugs early
make test   # Run all tests with race detector — unit, render integration, and BDD-style tests
```

**Purpose:** Prevent regressions, enforce consistent style, and ensure the codebase remains shippable after every change — no matter how small.


## Related
- [[convention-3108]]
- [[decision-1695]]
