---
title: "Usage Package (internal/usage)"
type: component
tags: [usage, internals, api]
---

# Usage Package — `internal/usage`

Monitors Claude Max subscription utilization via Anthropic's internal OAuth API. Provides token reading, a caching HTTP client, and a progress bar renderer. Displayed as a 0–3 row panel above the info panel when credentials are available.

## Files

| File                    | Purpose                                                                 |
|-------------------------|-------------------------------------------------------------------------|
| `credentials.go`        | Reads OAuth access token from `~/.claude/.credentials.json`             |
| `client.go`             | HTTP client with 60s TTL in-memory cache and stale-fallback on error    |
| `bar.go`                | Renders `█░` progress bar rows with dark background (`colorBg "238"`)  |
| `credentials_test.go`   | Tests for token reading (happy path, missing file, malformed JSON)      |
| `client_test.go`        | Tests for cache TTL, stale fallback, error handling                     |
| `bar_test.go`           | Tests for bar rendering (line count, 100%/0% fill assertions)           |

## API

Undocumented Anthropic internal endpoint:
```
GET https://api.anthropic.com/api/oauth/usage
Headers: Authorization: Bearer <token>
         anthropic-beta: oauth-2025-04-20
```

Response fields mapped to `Data`:
```go
type Window struct {
    Utilization float64  // 0–100
    ResetsAt    *time.Time
}
type Data struct {
    FiveHour     *Window
    SevenDay     *Window
    SevenDayOpus *Window  // non-nil only for Opus-tier subscriptions
}
```

## Credentials

OAuth token read from `~/.claude/.credentials.json`:
```json
{ "claudeAiOauth": { "accessToken": "..." } }
```
`ReadToken(path string) (string, error)` — exported, path-parameterized for testing.

## Client

```go
func NewClient(token, baseURL string) *Client
func (c *Client) SetTTL(d time.Duration)          // mutex-protected; for tests
func (c *Client) Fetch(ctx context.Context) (*Data, bool, error)
// Returns: data, stale (true if serving cached data after error), error
```

- Cache TTL: 60 seconds (default)
- Stale fallback: on HTTP error, returns last good data with `stale=true`
- Thread-safe: mutex protects `ttl`, `cache`, `cachedAt`, `lastGood`

## Bar Renderer

```go
func RenderBar(data *Data, stale bool, width int) string
// Returns "" if data is nil
```

- Renders one row per non-nil Window (5h, 7d, opus)
- Color thresholds: `>95%` critical red, `>80%` warning orange, otherwise green; stale = dim
- Each row wrapped with `bgStyle.Width(width).Render(...)` (dark bg `"238"`, matches `StyleCrumbs`)
- See [[lipgloss-bg-convention]] — every sub-style must explicitly set `Background(colorBg)`

## Integration in rootModel

`cmd/root.go` wires usage monitoring:
- `newRootModel()` reads credentials, creates `Client`, fires initial `Fetch`
- Demo mode: calls `demo.GenerateUsage()` to skip HTTP
- `TickMsg` handler: increments `usageTick`, fires `loadUsageAsync()` every 60 ticks
- `syncView()`: assigns `usage.RenderBar(rm.usageData, rm.usageStale, w)` to `app.Info.UsageLine`

## Related

- [[cmd-package]] — wires usage client into rootModel
- [[ui-package]] — `InfoModel.UsageLine` consumed by `ViewWithMenu`
- [[ui-spec]] — usage bar appears above info panel in screen layout
- [[demo-package]] — `GenerateUsage()` provides synthetic usage for --demo
- [[architecture]] — usage package in package table
- [[lipgloss-bg-convention]] — background layering rule applied in `bar.go`
