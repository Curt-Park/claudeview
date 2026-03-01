---
title: "claudeview UI Specification"
type: concept
tags: [spec, ui, navigation, keybindings, layout]
---

# claudeview UI Specification

## Overview

claudeview is a k9s-style terminal dashboard for Claude Code sessions. The UI consists of a fixed chrome frame (info panel + title bar + breadcrumbs + status bar) surrounding a scrollable content area.

---

## Screen Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ INFO PANEL (5 rows min)                                                      │
│  col0: info   │  col1: nav cmds   │  col2: util cmds  │  col3: shortcuts    │
├─────────────────────────────────────────────────────────────────────────────┤
│ TITLE BAR (1 row): ──── Projects(all)[3] ────                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  CONTENT AREA  (dynamic height)                                             │
│  [ Table | Plugin Detail | Memory Detail ]                                  │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ BREADCRUMBS (1 row): projects > sessions > agents                           │
├─────────────────────────────────────────────────────────────────────────────┤
│ STATUS BAR (1 row): [flash | filter]                                        │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Chrome rows**: 5+ (info) + 1 (title) + 1 (crumbs) + 1 (status) = **8+ rows**
**Content height**: `terminal_height - chrome` (min 5), computed dynamically

---

## Info Panel

The info panel occupies 5+ rows and renders five columns:

```
Project:      <full-width path — left-truncated if long>
Session:      <value>   <j/k> down/up    </> filter    <p> plugins   <ctrl+c> quit
User:         <value>   <G/g> bottom/top               <m> memories
Claude Code:  <value>   <ctrl+d/u> page down/up
claudeview:   <value>   <enter> (context)
              ...       <esc> (context)
```

- **Col 0**: `labelW=14` + value; total ~32 chars
- **Col 1**: nav commands (j/k, G/g, ctrl+d/u, enter/esc — context-sensitive)
- **Col 2**: util commands (`/` filter)
- **Col 3**: p/m jump shortcuts (context-sensitive)
- **Col 4**: `ctrl+c quit` (first row only)

Panel height: `max(5, 1 + max(navCount, utilCount))`

### Context-Dependent Values

| Field        | Projects level | Sessions level   | Agents+ level          |
|--------------|----------------|------------------|------------------------|
| Project      | `--`           | selected project | selected project       |
| Session      | `--`           | `--`             | selected session (8ch) |
| User         | OS username    | OS username      | OS username            |
| Claude Code  | CLI version    | CLI version      | CLI version            |
| claudeview   | app version    | app version      | app version            |

### Jump Shortcuts (Col 3)

- `<p>` plugins — visible when not in plugins/memories/detail view
- `<m>` memories — visible only when a project is selected AND not in plugins/memories/detail view

Both hints hidden when active resource is `plugins`, `memories`, `plugin-detail`, or `memory-detail`.

---

## Resource Table Columns

### 1. Projects

| Column      | Width          | Description                    |
|-------------|----------------|--------------------------------|
| NAME        | flex (max 55%) | project directory hash         |
| SESSIONS    | 8              | total session count            |
| LAST ACTIVE | 11             | human-friendly age (e.g. `3d`) |

**Navigation**: Enter → Sessions (filtered to this project)

### 2. Sessions

| Column      | Width          | Description                                  |
|-------------|----------------|----------------------------------------------|
| PROJECT     | 20             | parent project hash (flat access only)       |
| NAME        | 10             | session short ID (first 8 chars)             |
| TOPIC       | flex (max 35%) | first line of session topic / summary        |
| TURNS       | 6              | conversation turn count                      |
| AGENTS      | 6              | agent count                                  |
| TOKENS      | flex (max 25%) | token usage string (input + output)          |
| LAST ACTIVE | 11             | time since last modification                 |

Each row optionally shows a **subtitle line** (dimmed) with model, cost, and status metadata, indented under TOPIC.

**Note**: PROJECT column only shown in flat access (via `p`/`m` jump, or no project selected).

**Navigation**: Enter → Agents (filtered to this session)

### 3. Agents

| Column        | Width          | Description                                |
|---------------|----------------|--------------------------------------------|
| SESSION       | 12             | parent session short ID (flat access only) |
| NAME          | flex (max 20%) | tree-prefixed display name                 |
| TYPE          | 16             | agent type string                          |
| STATUS        | 10             | colored status                             |
| LAST ACTIVITY | flex (max 35%) | last tool name + input summary             |

**Note**: SESSION column only shown in flat access. Agents is a **leaf** — `enter` has no effect.

### 4. Plugins

| Column    | Width          | Description                           |
|-----------|----------------|---------------------------------------|
| NAME      | flex (max 25%) | plugin name                           |
| VERSION   | 10             | semver string                         |
| SCOPE     | 8              | `user` / `project`                    |
| STATUS    | 10             | `enabled` / `disabled` (colored)      |
| SKILLS    | 7              | skill count                           |
| COMMANDS  | 9              | command count                         |
| HOOKS     | 6              | hook count                            |
| AGENTS    | 7              | agent definition count                |
| MCPS      | 5              | MCP server definition count           |
| INSTALLED | 12             | installation date (YYYY-MM-DD)        |

**Navigation**: Enter → Plugin Detail

### 5. Memories

| Column   | Width          | Description            |
|----------|----------------|------------------------|
| NAME     | 18             | memory file name       |
| TITLE    | flex (max 45%) | memory document title  |
| SIZE     | 8              | file size string       |
| MODIFIED | 11             | last modification date |

**Note**: `<m>` jump requires a project to be selected.

**Navigation**: Enter → Memory Detail

---

## Content Modes

### 1. Table (default)
- `j/k` / arrows: move selection; `g/G`: top/bottom; `ctrl+d/u` / `pgdn/pgup`: half-page
- `enter`: drill down; `/`: filter; `esc`: clear filter or navigate back
- Rows updated within 5 seconds highlighted ("hot" state)

### 2. Plugin Detail
- Activated by `enter` on a Plugins row (resource → `plugin-detail`)
- Shows plugin header + sections: Skills, Commands, Hooks, Agents, MCPs
- `esc`: return to Plugins table

### 3. Memory Detail
- Activated by `enter` on a Memories row (resource → `memory-detail`)
- Shows raw file content of the selected memory file
- `esc`: return to Memories table

---

## Keybindings Reference

### Global
| Key      | Action                                      |
|----------|---------------------------------------------|
| `ctrl+c` | quit immediately                            |
| `/`      | enter filter mode                           |
| `p`      | jump to plugins (always)                    |
| `m`      | jump to memories (requires project context) |

### Table Mode
| Key               | Action                                                            |
|-------------------|-------------------------------------------------------------------|
| `j` / `↓`         | move selection down                                               |
| `k` / `↑`         | move selection up                                                 |
| `g`               | go to top                                                         |
| `G`               | go to bottom                                                      |
| `ctrl+d` / `pgdn` | page down (half page)                                             |
| `ctrl+u` / `pgup` | page up (half page)                                               |
| `enter`           | drill down (projects→sessions; sessions→agents; plugins/memories→detail) |
| `esc`             | clear filter (if active); otherwise navigate back                 |

### Filter Mode (`/`)
| Key        | Action                            |
|------------|-----------------------------------|
| typing     | live filter rows                  |
| `enter`    | confirm filter (stay in table)    |
| `esc`      | clear filter and exit             |
| `backspace`| delete last character             |

---

## Navigation Hierarchy

```
projects
  └─→ sessions (filtered by project)
        └─→ agents (filtered by session)  [leaf]

[p] plugins  ──→  plugin-detail
[m] memories ──→  memory-detail  (project context required)
```

**Jump** (`p`/`m`): saves current state. `esc` restores it (resource, project, session, filter).

**Filter stack**: parent filter saved on drill-down, restored on `esc` back.

### Breadcrumb Examples
```
projects > sessions > agents
plugins > plugin-detail
memories > memory-detail
```

---

## Status Bar

1. **Filter mode active**: shows `/my-filter` live input
2. **Flash message**: info (yellow) or error (red), auto-expires

---

## Test Coverage

Tests in `internal/ui/` (`package ui_test`):

| File                    | Coverage                                          |
|-------------------------|---------------------------------------------------|
| `app_test.go`           | AppModel integration — key flows, state transitions |
| `render_test.go`        | Full render output / golden snapshots             |
| `detail_render_test.go` | Plugin detail and memory detail rendering         |
| `filter_test.go`        | Filter component unit tests                       |
| `crumbs_test.go`        | Breadcrumb component unit tests                   |
| `menu_test.go`          | Menu / nav hint unit tests                        |
| `testhelpers_test.go`   | Shared helpers (mock DataProvider, key senders)   |

---

## Related

- [[ui-package]] implements this spec (AppModel, keybindings, layout)
- [[view-package]] implements column definitions and resource tables
- [[pre-completion-checklist]] applies to all UI changes
