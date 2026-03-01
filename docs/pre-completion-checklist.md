---
title: "Pre-Completion Checklist — make fmt, lint, test"
type: convention
tags: [testing, ci, make, pre-completion, quality-gate]
---

# Pre-Completion Checklist

Run the following commands in order and confirm all pass before finishing any task:

```bash
make fmt    # Format all Go source files (go fmt ./...)
make lint   # Static analysis via golangci-lint — enforce style and catch bugs early
make test   # Run all tests with race detector — unit and render integration tests
```

**Purpose:** Prevent regressions, enforce consistent style, and ensure the codebase remains shippable after every change — no matter how small.

## Related

- [[refactoring-standards]]
- [[test-suite]]
- [[ui-spec]]
