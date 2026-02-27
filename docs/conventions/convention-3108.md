---
confidence: 0.8
created: "2026-02-26T19:46:55+10:00"
id: convention-3108
modified: "2026-02-26T19:46:55+10:00"
references: []
relations:
  - type: implements
    target: decision-1695
    description: Refactoring standards is the concrete refactoring directive referenced from the minimal CLAUDE.md
    confidence: 0.8
source: manual
status: active
tags:
  - refactoring
  - kent-beck
  - martin-fowler
  - code-quality
  - clean-code
title: Refactoring standards — Kent Beck & Martin Fowler
type: convention
---

# Refactoring Standards (Kent Beck & Martin Fowler)

Apply the following principles:

### 1. Separate Refactoring from Feature Work
Never mix refactoring with behavior changes in the same step. First make the code work correctly, then improve its structure. If you find yourself changing both structure and behavior at once, stop and split the work.

> *"Refactoring is a disciplined technique for restructuring existing body of code, altering its internal structure without changing its external behavior."* — Martin Fowler

### 2. Take Small, Verifiable Steps
Each refactoring must be a single, safe transformation that keeps all tests green. If a refactoring breaks tests, revert immediately and try a smaller step. Large rewrites are not refactoring — they are rewriting.

> *"The key to keeping code working while refactoring is to take small steps."* — Martin Fowler

### 3. Eliminate Duplication (DRY)
When the same logic appears in two or more places, extract it. Duplication makes future changes risky because every copy must be updated consistently. Use Extract Function, Extract Variable, or Pull Up Method to consolidate.

> *"Once, and only once."* — Kent Beck (XP rule for expressing every concept in code exactly once)

### 4. Reveal Intent — Name for the Reader, Not the Machine
Rename variables, functions, and types until the code reads like a clear statement of what it does, not how it does it. If a comment is needed to explain a block of code, that block is a candidate for Extract Function with a descriptive name.

> *"Any fool can write code that a computer can understand. Good programmers write code that humans can understand."* — Martin Fowler

### 5. Apply the Rule of Three
The first time you write something, just write it. The second time you write something similar, note the duplication. The third time, refactor. This avoids premature abstraction while preventing harmful repetition.

> *"Three strikes and you refactor."* — Martin Fowler

### 6. Eliminate Code Smells Before They Accumulate
Treat the following as mandatory cleanup triggers — not optional polish:
- **Long Function**: if a function needs a comment to explain a section, extract that section into a named function.
- **Large Class**: if a struct handles more than one responsibility, split it.
- **Duplicate Code**: consolidate before adding new behavior on top.
- **Dead Code**: delete unused variables, functions, and types immediately — do not leave them commented out.
- **Primitive Obsession**: replace bare `string`, `int`, or `bool` flags with named types or small structs when they carry domain meaning.

### 7. Preserve Observable Behavior — Tests Are the Safety Net
Refactoring is only safe when tests exist and pass before and after. Never refactor untested code without first writing a characterization test. The test suite is the contract that proves structure changes did not alter behavior.

> *"Make it work, make it right, make it fast."* — Kent Beck


## Related
- [[decision-1695]]
