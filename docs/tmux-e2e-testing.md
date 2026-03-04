---
title: "E2E Testing via tmux"
type: convention
tags: [testing, tmux, e2e, verification]
---

# E2E Testing via tmux

Visual verification of the TUI is done by driving claudeview inside a tmux session. This complements the automated [[test-suite]] and [[pre-completion-checklist]] with human-visible confirmation that layout, content, and navigation work end-to-end.

## Process

1. **Build** the binary:
   ```bash
   go build -o /tmp/claudeview .
   ```

2. **Launch** in a detached tmux session with a fixed geometry:
   ```bash
   tmux new-session -d -s cv -x 160 -y 45 '/tmp/claudeview'
   ```
   - Session name `cv` is the convention.
   - 160×45 ensures enough room for all columns and the info panel.

3. **Navigate** by sending keys:
   ```bash
   tmux send-keys -t cv Enter          # drill down
   tmux send-keys -t cv 'j'            # move down
   tmux send-keys -t cv '/' && tmux send-keys -t cv 'sparkling' && tmux send-keys -t cv Enter
   tmux send-keys -t cv Escape          # back
   tmux send-keys -t cv 'g'            # go to top
   tmux send-keys -t cv 'G'            # go to bottom
   tmux send-keys -t cv C-d            # page down (ctrl+d)
   ```

4. **Capture** the pane content for comparison:
   ```bash
   tmux capture-pane -t cv -p | head -50     # first 50 lines
   tmux capture-pane -t cv -p | sed -n '8,45p'  # content area only (skip chrome)
   ```

5. **Compare** table MESSAGE column with detail view content:
   - Select a row → `Enter` to expand → `tmux capture-pane` → `Escape` back.
   - Check that the preview accurately represents the detail content.

6. **Rebuild and relaunch** after code changes:
   ```bash
   go build -o /tmp/claudeview . && tmux send-keys -t cv C-c && sleep 0.5
   tmux send-keys -t cv '/tmp/claudeview' Enter
   ```
   If the tmux server has died, recreate the session from step 2.

## Tips

- Use `sleep 0.3` after `send-keys` before `capture-pane` to let the TUI redraw.
- Loop navigation keys for bulk scrolling: `for i in $(seq 1 30); do tmux send-keys -t cv 'j'; done`
- Filter to a specific session: `send-keys '/'` → type slug → `send-keys Enter` → `send-keys Enter`.
- Kill the app with `tmux send-keys -t cv C-c` (sends ctrl+c).

## When to Use

- After any UI change (column widths, preview logic, styles, layout)
- After changes to navigation (drill-down, back, follow mode)
- Before updating PR verification tables

## Related

- [[test-suite]] — automated unit and integration tests
- [[pre-completion-checklist]] — `make fmt`, `make lint`, `make test`
- [[ui-spec]] — the spec that visual checks verify
- [[ui-package]] — the code under visual test
