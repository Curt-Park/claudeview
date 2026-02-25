# claudeview UI Specification

## Overview

claudeview is a k9s-style terminal dashboard for Claude Code sessions. The UI consists of a fixed chrome frame (info panel + title bar + breadcrumbs + status bar) surrounding a scrollable content area.

---

## Screen Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ INFO PANEL (7 rows)                                                          │
│  left column (info)     │  center column (keybindings)  │  right (shortcuts) │
├─────────────────────────────────────────────────────────────────────────────┤
│ TITLE BAR (1 row): ──── Projects(all)[3] ────                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  CONTENT AREA  (Height - 10 rows)                                           │
│  [ Table | Log | Detail | YAML | Help ]                                     │
│                                                                              │
├─────────────────────────────────────────────────────────────────────────────┤
│ BREADCRUMBS (1 row): projects > sessions > agents                           │
├─────────────────────────────────────────────────────────────────────────────┤
│ STATUS BAR (1 row): [flash | command | filter]                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Chrome rows**: 7 (info) + 1 (title) + 1 (crumbs) + 1 (status) = **10 rows**
**Content height**: `terminal_height - 10` (min 5)

---

## Info Panel (3 Columns)

The info panel occupies 7 rows and renders three columns side-by-side:

```
Project:      <value>          <enter> view      <0> all
Session:      <value>          <l> logs          <1> project-alpha
User:         <value>          <d> detail        <2> project-beta
Claude Code:  <value>          </> filter        <3> project-gamma
claudeview:   <value>          <:> cmd
CPU:          <value>%         <?> help
MEM:          <value> MiB      <q> quit
```

### Column Widths
- **Left (info)**: 46 visible chars (14 label + 32 value)
- **Center (keybindings)**: items from current `MenuItem` list (TableMenuItems by default)
- **Right (shortcuts)**: `0-9` parent shortcuts, one per row (max 7 shown, matching info rows)

### Context-Dependent Values

Values in the left column follow context rules based on navigation depth:

| Field        | Projects level | Sessions level       | Agents+ level          |
|--------------|----------------|----------------------|------------------------|
| Project      | `--`           | selected project     | selected project       |
| Session      | `--`           | `--`                 | selected session (8ch) |
| User         | `--`           | OS username          | OS username            |
| Claude Code  | `--`           | Claude CLI version   | Claude CLI version     |
| claudeview   | app version    | app version          | app version            |
| CPU          | `0%`           | session process CPU  | session process CPU    |
| MEM          | `0 MiB`        | session process MEM  | session process MEM    |

**Rule**: At the `projects` top level, Project/Session/User/Claude Code all show `--`. Values are progressively filled as the user drills down into the hierarchy.

### Parent Shortcuts (Column 3)

```
<0>  all          — clear filter, show all items
<1>  project-alpha — (active) filter by this parent
<2>  project-beta  — filter by this parent
...
<9>  last-parent   — filter by 9th parent
```

- **`0`** always means "all" (clear filter)
- **`1-9`** select from `ParentShortcuts` list
- Active parents are sorted first, then by item count descending
- Only shown when current resource has a meaningful parent context (sessions→project, agents→session, tools→session+agent, tasks→session)
- At most 9 shortcuts displayed (rows 1-7 in panel, overflow not shown)

---

## Resource Table Columns

### 1. Projects

| Column   | Width     | Description                     |
|----------|-----------|---------------------------------|
| NAME     | flex      | project hash (truncated)        |
| SESSIONS | 8         | total session count             |
| ACTIVE   | 6         | active session count            |
| LAST SEEN| 10        | human-friendly age (e.g. `3d`)  |

**Navigation**: Enter → Sessions (filtered to this project)

### 2. Sessions

| Column  | Width | Description                                    |
|---------|-------|------------------------------------------------|
| PROJECT | 20    | parent project name (flat access only)         |
| NAME    | 10    | session short ID (first 8 chars)               |
| MODEL   | 16    | Claude model name (flex)                       |
| STATUS  | 12    | colored status string                          |
| AGENTS  | 6     | agent count                                    |
| TOOLS   | 6     | total tool calls                               |
| TOKENS  | 8     | total tokens (k-suffixed)                      |
| COST    | 8     | `$X.XXXX` or `-`                               |
| AGE     | 6     | time since last modification                   |

**Note**: PROJECT column is only shown during flat access (`:sessions` command). During drill-down from a project, PROJECT column is hidden.

**Navigation**: Enter → Agents (filtered to this session)

### 3. Agents

| Column        | Width | Description                                |
|---------------|-------|--------------------------------------------|
| SESSION       | 12    | parent session short ID (flat access only) |
| NAME          | 26    | tree-prefixed display name (flex)          |
| TYPE          | 12    | agent type string                          |
| STATUS        | 14    | colored status                             |
| TOOLS         | 6     | tool call count                            |
| LAST ACTIVITY | 30    | last tool name + input summary             |

**Navigation**: Enter → Tools (filtered to this agent)

### 4. Tools (Tool Calls)

| Column       | Width | Description                                   |
|--------------|-------|-----------------------------------------------|
| SESSION      | 10    | parent session short ID (flat access only)    |
| AGENT        | 10    | parent agent short ID (flat access only)      |
| TIME         | 10    | timestamp `HH:MM:SS`                          |
| TOOL         | 10    | tool name                                     |
| INPUT SUMMARY| 30    | first N chars of input JSON (flex)            |
| RESULT       | 16    | result summary or `error` (red)               |
| DURATION     | 10    | milliseconds or `--`                          |

**Note**: SESSION and AGENT columns only shown during flat access.

### 5. Tasks

| Column     | Width | Description                               |
|------------|-------|-------------------------------------------|
| SESSION    | 12    | parent session short ID (flat access only)|
| ID         | 4     | task ID string                            |
| STATUS     | 12    | icon + status string (colored)            |
| SUBJECT    | 40    | task subject (flex)                       |
| BLOCKED BY | 14    | comma-separated blocker IDs               |

### 6. Plugins

| Column      | Width | Description                     |
|-------------|-------|---------------------------------|
| NAME        | 20    | plugin name (flex)              |
| VERSION     | 10    | semver string                   |
| ENABLED     | 8     | `yes` / `no`                    |
| SKILLS      | 6     | skill count                     |
| COMMANDS    | 8     | command count                   |
| HOOKS       | 6     | hook count                      |
| MARKETPLACE | 12    | marketplace name                |

### 7. MCP Servers

| Column    | Width | Description                       |
|-----------|-------|-----------------------------------|
| NAME      | 20    | server name (flex)                |
| TRANSPORT | 10    | `stdio` / `http` / `sse`          |
| STATUS    | 10    | colored status                    |
| COMMAND   | 30    | command + args                    |

---

## View Modes

The content area renders one of five modes:

### 1. Table (default)
- Scrollable table with column headers
- `j/k` or arrow keys: move selection
- `g/G`: go to top/bottom
- `ctrl+u/d` or `pgup/pgdn`: page up/down
- `enter`: drill down
- Filter applied: rows filtered by substring match (case-insensitive, all cells)

### 2. Log (`l` key)
- Scrollable transcript log view for sessions/agents
- Shows turns with role, timestamp, thinking, text, tool calls
- `h/j/k/l`: scroll (horizontal/vertical)
- `g/G`: top/bottom
- `f`: toggle follow mode (auto-scroll to newest)
- `/`: search within log
- `esc`: return to Table

### 3. Detail (`d` key)
- Resource-specific detail panel
- Scrollable multi-line text
- `h/j/k/l`: scroll
- `g/G`: top/bottom
- `esc`: return to Table

### 4. YAML (`y` key)
- JSON dump of selected row's data object
- Same navigation as Detail
- `esc`: return to Table

### 5. Help (`?` key) — Full Screen
- Full-screen help overlay replacing content area
- Shows complete keybinding reference
- Shows info panel context rules
- Scrollable with `j/k`
- `esc`: return to previous mode (Table)

---

## Keybindings Reference

### Global (all modes)
| Key      | Action                            |
|----------|-----------------------------------|
| `ctrl+c` | quit immediately                  |
| `:`      | enter command mode                |
| `/`      | enter filter mode                 |
| `?`      | show full-screen help             |

### Table Mode
| Key            | Action                                  |
|----------------|-----------------------------------------|
| `j` / `↓`      | move selection down                     |
| `k` / `↑`      | move selection up                       |
| `g`            | go to top                               |
| `G`            | go to bottom                            |
| `ctrl+d` / `pgdn` | page down (half page)               |
| `ctrl+u` / `pgup` | page up (half page)                 |
| `enter`        | drill down (projects→sessions→agents→tools) |
| `l`            | log view                                |
| `d`            | detail view                             |
| `y`            | YAML/JSON dump view                     |
| `0`            | clear parent filter (show all)          |
| `1`-`9`        | filter by Nth parent shortcut           |
| `esc` / `q`    | navigate back (or back to projects)     |

### Log Mode
| Key       | Action              |
|-----------|---------------------|
| `j` / `↓` | scroll down         |
| `k` / `↑` | scroll up           |
| `h` / `←` | scroll left         |
| `l` / `→` | scroll right        |
| `g`       | go to top           |
| `G`       | go to bottom        |
| `f`       | toggle follow mode  |
| `/`       | search within log   |
| `esc`     | return to table     |

### Detail / YAML Mode
| Key       | Action              |
|-----------|---------------------|
| `j` / `↓` | scroll down         |
| `k` / `↑` | scroll up           |
| `h` / `←` | scroll left         |
| `l` / `→` | scroll right        |
| `g`       | go to top           |
| `G`       | go to bottom        |
| `esc`     | return to table     |

### Command Mode (`:`)
| Key        | Action                            |
|------------|-----------------------------------|
| typing     | input resource name               |
| `tab`      | accept autocomplete suggestion    |
| `enter`    | execute command (switch resource) |
| `esc`      | cancel                            |

Available resources: `projects`, `sessions`, `agents`, `tools`, `tasks`, `plugins`, `mcp`

### Filter Mode (`/`)
| Key        | Action                            |
|------------|-----------------------------------|
| typing     | live filter table rows            |
| `enter`    | confirm filter (stay in table)    |
| `esc`      | clear filter and exit filter mode |
| `backspace`| delete last character             |

---

## Navigation Hierarchy

```
projects
  └─→ sessions (filtered by project)
        └─→ agents (filtered by session)
              └─→ tools (filtered by agent)
```

**Drill-down** (`enter`): moves deeper, pushing breadcrumb
**Navigate back** (`esc`/`q`): pops breadcrumb, returns to parent
**Flat access** (`:command`): always shows full unfiltered data with parent columns

### Breadcrumb Examples
```
projects
projects > sessions
projects > sessions > agents
projects > sessions > agents > tools
```

When using `:command` to switch resources, breadcrumbs reset to just the new resource name.

---

## Command Mode Resources

`:` followed by any resource name (with tab-autocomplete) switches the view:

| Command     | Shows                              |
|-------------|------------------------------------|
| `:projects` | all projects                       |
| `:sessions` | all sessions (with PROJECT column) |
| `:agents`   | all agents (with SESSION column)   |
| `:tools`    | all tools (with SESSION+AGENT cols)|
| `:tasks`    | all tasks (with SESSION column)    |
| `:plugins`  | all plugins                        |
| `:mcp`      | all MCP servers                    |

---

## Filter (`/`) Behavior

- Filters are applied as case-insensitive substring match across **all cells** in a row
- Filter is shown in the title bar: `Sessions(my-filter)[3]`
- `0` key clears the filter when in table mode (equivalent to parent-filter "all")
- Filter resets when switching resource or navigating back

---

## Status Bar

The single status bar row shows (in priority order):
1. **Command mode**: `:sessions` with ghost autocomplete hint
2. **Filter mode**: `/my-filter` input
3. **Flash message**: info (yellow) or error (red), with auto-expiry

---

## Test IDs

Each behavior maps to a BDD test in `internal/ui/bdd/`:

| Behavior                           | Test File              | Test ID Prefix       |
|------------------------------------|------------------------|----------------------|
| Initial render (projects table)    | `initial_test.go`      | `TestInitial`        |
| Table navigation (j/k/g/G)         | `navigation_test.go`   | `TestNav`            |
| Drill-down + breadcrumbs           | `drilldown_test.go`    | `TestDrilldown`      |
| `:command` resource switch         | `command_test.go`      | `TestCommand`        |
| `/` filter                         | `filter_test.go`       | `TestFilter`         |
| d/l/y/esc view mode switch         | `viewmode_test.go`     | `TestViewMode`       |
| `?` full-screen help               | `help_test.go`         | `TestHelp`           |
| 0-9 number shortcuts               | `shortcuts_test.go`    | `TestShortcut`       |
| Info panel context values          | `info_context_test.go` | `TestInfoContext`    |
| Parent columns (flat access)       | `parent_columns_test.go` | `TestParentCols`   |
| Window resize                      | `resize_test.go`       | `TestResize`         |
| Flash message display/expiry       | `flash_test.go`        | `TestFlash`          |
| Detail/YAML view content           | `detail_yaml_test.go`  | `TestDetailYAML`     |

**Golden file regeneration**:
```bash
UPDATE_SNAPSHOTS=1 go test ./internal/ui/bdd/...
```
