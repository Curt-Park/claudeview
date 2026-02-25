# claudeview — Development Guidelines for Claude Code

## Go Toolchain Rules

Run these commands **in order** after any Go file is created or modified:

```bash
go fmt ./...          # format all Go source files
go fix ./...          # apply Go API migrations (interface{} → any, etc.)
go vet ./...          # static analysis
go build ./...        # verify compilation
go test -race ./...   # run tests with race detector
```

### When to run each tool

| Event | Required commands |
|-------|-------------------|
| After writing/editing any `.go` file | `go fmt ./...` → `go fix ./...` → `go vet ./...` → `go build ./...` |
| Before committing | All of the above + `go test -race ./...` |
| After adding/updating dependencies | `go mod tidy` → `go fmt ./...` → `go build ./...` |
| Before creating a PR | `make lint` (runs golangci-lint) |

### Quick reference

```bash
make build   # go build -ldflags "..." -o bin/claudeview .
make test    # go test -race -count=1 ./...
make bdd     # go test -race -count=1 ./internal/ui/bdd/...
make lint    # golangci-lint run ./...
make demo    # build + run --demo mode
```

## Code Style

- **Use `any` instead of `interface{}`** — enforced by `go fix`
- **No unused imports** — `go fmt` and `go vet` will catch these
- **Error handling** — always check errors; do not use `_` for errors at package boundaries
- **Module path**: `github.com/Curt-Park/claudeview`
- **Go version**: 1.26+ (set in `go.mod` and `.mise.toml`)

## Project Structure

```
main.go                          # entrypoint
cmd/root.go                      # Cobra CLI (--demo, --project, --resource)
internal/
  transcript/                    # JSONL parser, watcher, scanner
  config/                        # settings.json, plugins, tasks parsers
  model/                         # data models (Project, Session, Agent, …)
  ui/                            # Bubble Tea app + chrome components
  view/                          # resource-specific table views
  demo/                          # synthetic demo data generator
```

## UI Specification

- **Spec document**: `docs/ui-spec.md` — all UI changes must be kept in sync with this document
- **BDD tests**: `internal/ui/bdd/` — teatest-based integration tests covering all spec behaviors
- **Test IDs**: each spec section references a test function (e.g. `TEST-NAV-001`)
- **Run BDD tests**: `make bdd`

## Development Setup

Use [mise](https://mise.jdx.dev/) for automatic tool version management:

```bash
mise trust && mise install       # installs Go 1.26 + golangci-lint 1.64
```
