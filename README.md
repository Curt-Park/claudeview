# claudeview

Terminal dashboard for Claude Code. Browse sessions, agents, tool calls, tasks, plugins, and MCP servers — all from one place, in real time.

## Demo

```bash
┌─ claudeview │ Project: my-app │ Model: opus-4-6 │ MCP: 3 ────────────────────┐
│ <enter> view  <l> logs  <d> detail  </> filter  <:> cmd  <?> help  <q> quit  │
├──────────────────────────────────────────────────────────────────────────────┤
│ NAME       MODEL        STATUS   AGENTS  TOOLS  TOKENS   COST     AGE        │
│ ► abc123   opus-4-6     active      4      47   145.2k   $0.42    5m         │
│   def456   sonnet-4-6   ended       2      12    23.1k   $0.08    2h         │
│   ghi789   haiku-4-5    ended       1       5     5.0k   $0.01    1d         │
├──────────────────────────────────────────────────────────────────────────────┤
│ projects > my-app > sessions                                                 │
│ 3 sessions found                                                             │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Features

- **Vim-style navigation** — `:resource` command to switch views, `j/k` to move up/down
- **Live transcript parsing** — reads `~/.claude/projects/` JSONL files directly, no hooks needed
- **Drill-down navigation** — projects → sessions → agents → tools
- **Log view** — scrollable transcript with follow mode (`f`)
- **Detail view** — full JSON/input/output for any tool call
- **Demo mode** — `--demo` flag for trying without a live Claude session
- **Tier 1 resources**: projects, sessions, agents, tools, tasks, plugins, MCP servers

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
claudeview                        # Start on sessions view
claudeview --demo                 # Run with synthetic demo data
claudeview --resource projects    # Start on projects view
claudeview --project my-app       # Filter to a specific project
```

## Key Bindings

### Global

| Key | Action |
|-----|--------|
| `:` | Command mode — switch resource (`:sessions`, `:agents`, `:tools`, `:tasks`, `:plugins`, `:mcp`) |
| `/` | Filter mode |
| `?` | Help |
| `q` / `Esc` | Back / exit filter |
| `Ctrl+C` | Quit |

### Table View

| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `Enter` | Drill down |
| `l` | Log view |
| `d` | Detail view |
| `Esc` | Go back |

### Log / Detail View

| Key | Action |
|-----|--------|
| `h/j/k/l` | Scroll |
| `g` / `G` | Top / bottom |
| `f` | Toggle follow mode |
| `Esc` | Back to table |

## Resources

| Command | Alias | Description |
|---------|-------|-------------|
| `:projects` | `:p` | Project directories |
| `:sessions` | `:s` | Sessions (model, tokens, cost, age) |
| `:agents` | `:a` | Agent tree (main + subagents) |
| `:tools` | `:t` | Tool call history |
| `:tasks` | `:tk` | Task list (status, dependencies) |
| `:plugins` | `:pl` | Installed plugins |
| `:mcp` | `:m` | MCP servers |

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
make lint     # run golangci-lint
make demo     # build and run demo mode
```

## License

MIT
