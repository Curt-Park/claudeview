# claudeview

Terminal dashboard for Claude Code. Browse sessions, agents, tool calls, tasks, plugins, and MCP servers — all from one place, in real time.

## Screenshots

### Sessions — all active and past sessions at a glance

```
Project:      /Users/alice/.claude/projects/my-app
Session:      --           <j/k> down/up    </> filter    <t> tasks
User:         alice        <g/G> bot/top    <l> logs      <p> plugins
Claude Code:  1.2.3        <ctrl+d/u> page  <d> detail    <m> mcps
claudeview:   0.5.0        <enter> drill    <esc> back
──────────────────────────── Sessions(all)[3] ────────────────────────────────
  NAME       MODEL              STATUS        AGENTS  TOOLS  TOKENS   COST     AGE
► abc123ef   claude-opus-4-6    active             4     47   145.2k   $0.42    5m
  def456ab   claude-sonnet-4-6  ended              2     12    23.1k   $0.08    2h
  ghi789cd   claude-haiku-4-5   ended              1      5     5.0k   $0.01    1d
projects > sessions
```

### Agents — main agent and subagents as a tree

```
Project:      /Users/alice/.claude/projects/my-app
Session:      abc123ef     <j/k> down/up    </> filter    <t> tasks
User:         alice        <g/G> bot/top    <l> logs      <p> plugins
Claude Code:  1.2.3        <ctrl+d/u> page  <d> detail    <m> mcps
claudeview:   0.5.0        <enter> drill    <esc> back
────────────────────────── Agents(all)[4] ────────────────────────────────────
  NAME          TYPE               STATUS      TOOLS  LAST ACTIVITY
► Claude        main               running        12  Edit src/app.py
  ├─ Explorer   general-purpose    running         5  Read src/config.py
  ├─ Planner    Plan               done ✓          3  Read CLAUDE.md
  └─ Tester     Bash               running         2  Bash: npm test
projects > sessions > abc123ef > agents
```

### Tools — full history of tool calls

```
Project:      /Users/alice/.claude/projects/my-app
Session:      abc123ef     <j/k> down/up    </> filter
User:         alice        <g/G> bot/top    <d> detail
Claude Code:  1.2.3        <ctrl+d/u> page  <esc> back
claudeview:   0.5.0
──────────────────────────── Tools(all)[5] ───────────────────────────────────
  TIME       TOOL    INPUT SUMMARY                RESULT           DURATION
  17:42:10   Read    src/app.py                   142 lines        0.3s
  17:42:12   Grep    "handleAuth" in src/         3 matches        0.1s
► 17:42:14   Bash    npm test                     exit 0           2.1s
  17:42:18   Edit    src/app.py                   success          0.2s
  17:42:20   Glob    **/*.py                      12 files         0.1s
projects > sessions > abc123ef > agents > Claude > tools
```

### Logs — live-follow agent transcripts

```
──── Logs: Explorer (agent-a42f831) ──────────────────────── [follow] 10/10 ──
  17:42:15  [tool]   Grep "TODO" in **/*.py
            →        3 matches found
  17:42:16  [tool]   Read src/utils.py
            →        89 lines
  17:42:17  [text]   "I found 3 TODO items in the codebase..."
  17:42:18  [tool]   Read src/config.py
            →        45 lines
  17:42:20  [tool]   Grep "handleAuth" in src/
            →        2 matches
```

### Tasks — task list with dependency tracking

```
Project:      /Users/alice/.claude/projects/my-app
Session:      abc123ef     <j/k> down/up    </> filter    <t> tasks
User:         alice        <g/G> bot/top    <d> detail    <p> plugins
Claude Code:  1.2.3        <ctrl+d/u> page  <esc> back    <m> mcps
claudeview:   0.5.0
────────────────────────── Tasks(all)[4] ─────────────────────────────────────
  ID   STATUS                    SUBJECT                           BLOCKED BY
   1   ✓ completed               Explore project context
   2   ✓ completed               Ask clarifying questions           1
► 3   ► in_progress              Propose approaches                 2
   4   ○ pending                 Present design                     3
projects > sessions > abc123ef > tasks
```

## Features

- **Vim-style navigation** — `j/k` to move, `enter` to drill down, `esc` to go back
- **Jump shortcuts** — `t` tasks, `p` plugins, `m` MCP servers from anywhere
- **Live transcript parsing** — reads `~/.claude/projects/` JSONL files directly, no hooks needed
- **Drill-down navigation** — projects → sessions → agents → tool calls
- **Expanded row view** — selected row expands to show full content; collapses on move
- **Log view** — scrollable transcript with follow mode (`f`)
- **Detail view** — full JSON input/output for any selected row (`d`)
- **YAML/JSON dump** — raw data view for any row (`y`)
- **`:command` mode** — switch resource view from anywhere (`:sessions`, `:agents`, …)
- **Demo mode** — `--demo` flag for trying without a live Claude session

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Curt-Park/claudeview/main/install.sh | bash
```

Or build from source:

```bash
git clone https://github.com/Curt-Park/claudeview.git
cd claudeview
make install
```

## Usage

```bash
claudeview                        # Start on projects view
claudeview --demo                 # Run with synthetic demo data
claudeview --resource sessions    # Start on sessions view
claudeview --project <hash>       # Filter to a specific project
```

## Key Bindings

### Global

| Key | Action |
|-----|--------|
| `t` | Jump to tasks |
| `p` | Jump to plugins |
| `m` | Jump to MCP servers |
| `/` | Enter filter mode |
| `Ctrl+C` | Quit |

### Table View

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `g` / `G` | Go to top / bottom |
| `Ctrl+d` / `Ctrl+u` | Page down / up |
| `Enter` | Drill down |
| `l` | Log view |
| `d` | Detail view |
| `y` | YAML/JSON dump view |
| `Esc` | Navigate back |

### Log View

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down / up |
| `g` / `G` | Top / bottom |
| `Ctrl+d` / `Ctrl+u` | Page down / up |
| `f` | Toggle follow mode |
| `/` | Filter log lines |
| `Esc` | Back to table |

### Detail / YAML View

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down / up |
| `g` / `G` | Top / bottom |
| `Ctrl+d` / `Ctrl+u` | Page down / up |
| `Esc` | Back to table |

### Filter Mode (`/`)

| Key | Action |
|-----|--------|
| typing | Live filter table rows |
| `Enter` | Confirm filter |
| `Esc` | Clear filter and exit |
| `Backspace` | Delete last character |

### Command Mode (`:`)

Type `:` then a resource name to switch views:

| Command | Shows |
|---------|-------|
| `:projects` | All projects |
| `:sessions` | All sessions (with PROJECT column) |
| `:agents` | All agents (with SESSION column) |
| `:tools` | All tool calls (with SESSION + AGENT columns) |
| `:tasks` | All tasks (with SESSION column) |
| `:plugins` | All installed plugins |
| `:mcp` | All MCP servers |

## Data Sources

claudeview reads directly from `~/.claude/`:

```
~/.claude/projects/<hash>/
├── <session-id>.jsonl          ← main agent transcript
└── <session-id>/subagents/
    └── agent-<id>.jsonl        ← subagent transcripts
~/.claude/tasks/<session>/      ← task JSON files
~/.claude/plugins/              ← plugin metadata
~/.claude/settings.json         ← MCP server config
```

## Development Setup

### Recommended: mise

```bash
curl https://mise.run | sh
echo 'eval "$(mise activate bash)"' >> ~/.bashrc
source ~/.bashrc

git clone https://github.com/Curt-Park/claudeview.git
cd claudeview
mise trust && mise install
make build
```

### Manual

Install Go 1.26+ and golangci-lint, then:

```bash
make build    # build binary → bin/claudeview
make test     # run tests
make bdd      # run BDD integration tests
make lint     # run golangci-lint
make demo     # build and run demo mode
```

## License

MIT
