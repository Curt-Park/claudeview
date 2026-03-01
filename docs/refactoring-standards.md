---
title: "Refactoring Standards — Kent Beck & Martin Fowler"
type: convention
tags: [refactoring, code-quality, clean-code]
---

# Refactoring Standards (Kent Beck & Martin Fowler)

Apply the following principles:

### 1. Separate Refactoring from Feature Work
Never mix refactoring with behavior changes in the same step. First make the code work correctly, then improve its structure.

> *"Refactoring is a disciplined technique for restructuring existing body of code, altering its internal structure without changing its external behavior."* — Martin Fowler

### 2. Take Small, Verifiable Steps
Each refactoring must be a single, safe transformation that keeps all tests green. If a refactoring breaks tests, revert immediately.

> *"The key to keeping code working while refactoring is to take small steps."* — Martin Fowler

### 3. Eliminate Duplication (DRY)
When the same logic appears in two or more places, extract it. Use Extract Function, Extract Variable, or Pull Up Method.

> *"Once, and only once."* — Kent Beck

### 4. Reveal Intent — Name for the Reader, Not the Machine
Rename variables, functions, and types until the code reads like a clear statement of what it does. If a comment is needed to explain a block, extract it into a named function.

> *"Any fool can write code that a computer can understand. Good programmers write code that humans can understand."* — Martin Fowler

### 5. Apply the Rule of Three
First time: just write it. Second time: note the duplication. Third time: refactor.

> *"Three strikes and you refactor."* — Martin Fowler

### 6. Eliminate Code Smells Before They Accumulate
Mandatory cleanup triggers:
- **Long Function**: extract sections that need comments into named functions
- **Large Class**: if a struct handles more than one responsibility, split it
- **Duplicate Code**: consolidate before adding new behavior on top
- **Dead Code**: delete unused variables, functions, and types immediately
- **Primitive Obsession**: replace bare `string`/`int`/`bool` flags with named types when they carry domain meaning

### 7. Preserve Observable Behavior — Tests Are the Safety Net
Never refactor untested code without first writing a characterization test.

> *"Make it work, make it right, make it fast."* — Kent Beck

## Related

- [[pre-completion-checklist]]
