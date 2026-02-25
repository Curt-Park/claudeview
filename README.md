# claudeview

Terminal dashboard for Claude Code. Browse sessions, agents, tool calls, tasks, plugins, and MCP servers — all from one place, in real time.

## Features

- **Zero-setup observability** — reads `~/.claude/projects/` JSONL files directly; no hooks, no config, no agents to run
- **Full session hierarchy** — projects, sessions, agents, tool calls, tasks, plugins, and MCP servers in a single view
- **Live updates** — transcripts stream in as Claude writes them; no polling delay
- **Agent tree** — main agent and subagents rendered as a navigable tree with type, status, and last activity
- **In-place log view** — read the raw transcript of any session or agent without leaving the dashboard
- **Detail and YAML views** — inspect any row's full data as structured text or raw JSON

For the full UI specification, keybindings, and column definitions see [`docs/ui-spec.md`](docs/ui-spec.md).

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
claudeview         # Start on projects view
claudeview --demo  # Run with synthetic demo data
```

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
