# Usage Limit Bar Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a persistent 2-line usage panel at the very top of every view showing filled progress bars for the 5-hour and 7-day Claude Max windows with reset countdowns. Fetched from Anthropic's internal OAuth usage API, refreshed every 60 seconds. Panel is hidden when no credentials file exists.

**Architecture:** New `internal/usage/` package handles credential reading and HTTP fetch with 60s cache. `RenderBar(data, stale, width)` returns a 2-line string (one row per window) with `█░` progress bars, percentage, and "reset in Xh Xm" labels. Stored in `InfoModel.UsageLine`; `InfoModel.Height()` counts newlines in `UsageLine` to grow by the right number of lines. `rootModel` owns the `*usage.Client`, refreshes on a 60-tick interval, and sets `app.Info.UsageLine` in `syncView`.

**Tech Stack:** Go standard `net/http`, lipgloss for styling, Bubble Tea async cmd pattern.

**Rendered output (2 lines, terminal-width bars):**
```
5h [████░░░░░░░░░░░░░░░░░░░░░░░░░░]  8%   reset in 4h 9m
7d [████████████████████░░░░░░░░░░] 68%   reset in 1d 2h
```
Color thresholds: >95% red (bold), >80% yellow (bold), otherwise green.

---

## Key File Map

| File | Role |
|------|------|
| `internal/usage/credentials.go` | Read OAuth token from `~/.claude/.credentials.json` |
| `internal/usage/client.go` | HTTP client, `Window`/`Data` types, 60s in-memory cache |
| `internal/usage/bar.go` | `RenderBar(data, stale, width) string` — 2-line progress bar output |
| `internal/usage/credentials_test.go` | Token reading tests |
| `internal/usage/client_test.go` | httptest-based fetch + cache tests |
| `internal/usage/bar_test.go` | Bar rendering tests |
| `internal/ui/header.go` | Add `UsageLine string` to `InfoModel`; update `Height()` and `ViewWithMenu()` |
| `internal/ui/header_test.go` | Tests for usage line rendering in header |
| `cmd/root.go` | Add `usageClient`, `usageData`, async refresh, set `UsageLine` in `syncView` |

> **No new ResourceType, no new view, no 'u' key.** The bar is always present at the top.

---

### Task 1: `internal/usage/credentials.go`

**Files:**
- Create: `internal/usage/credentials.go`
- Create: `internal/usage/credentials_test.go`

**Step 1: Write the failing test**

```go
// internal/usage/credentials_test.go
package usage_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/Curt-Park/claudeview/internal/usage"
)

func TestReadToken(t *testing.T) {
    dir := t.TempDir()
    creds := map[string]any{
        "claudeAiOauth": map[string]any{
            "accessToken": "test-token-abc",
        },
    }
    data, _ := json.Marshal(creds)
    path := filepath.Join(dir, ".credentials.json")
    os.WriteFile(path, data, 0600)

    token, err := usage.ReadToken(path)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if token != "test-token-abc" {
        t.Errorf("got %q, want %q", token, "test-token-abc")
    }
}

func TestReadTokenMissingFile(t *testing.T) {
    _, err := usage.ReadToken("/nonexistent/.credentials.json")
    if err == nil {
        t.Fatal("expected error for missing file")
    }
}

func TestReadTokenEmptyToken(t *testing.T) {
    dir := t.TempDir()
    creds := map[string]any{"claudeAiOauth": map[string]any{"accessToken": ""}}
    data, _ := json.Marshal(creds)
    os.WriteFile(filepath.Join(dir, ".credentials.json"), data, 0600)

    _, err := usage.ReadToken(filepath.Join(dir, ".credentials.json"))
    if err == nil {
        t.Fatal("expected error for empty token")
    }
}
```

**Step 2: Run tests to confirm they fail**

```
go test ./internal/usage/... -run TestReadToken -v
```
Expected: compile error — package does not exist yet.

**Step 3: Implement**

```go
// internal/usage/credentials.go
package usage

import (
    "encoding/json"
    "fmt"
    "os"
)

type credentials struct {
    ClaudeAiOauth struct {
        AccessToken string `json:"accessToken"`
    } `json:"claudeAiOauth"`
}

// ReadToken reads the OAuth access token from the given credentials file path.
// Returns an error if the file is missing, malformed, or the token is empty.
func ReadToken(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return "", fmt.Errorf("reading credentials: %w", err)
    }
    var creds credentials
    if err := json.Unmarshal(data, &creds); err != nil {
        return "", fmt.Errorf("parsing credentials: %w", err)
    }
    if creds.ClaudeAiOauth.AccessToken == "" {
        return "", fmt.Errorf("no accessToken in credentials")
    }
    return creds.ClaudeAiOauth.AccessToken, nil
}
```

**Step 4: Run tests to confirm they pass**

```
go test ./internal/usage/... -run TestReadToken -v
```
Expected: PASS (3 tests).

**Step 5: Commit**

```bash
git add internal/usage/credentials.go internal/usage/credentials_test.go
git commit -m "feat(usage): add credentials reader for OAuth token"
```

---

### Task 2: `internal/usage/client.go`

**Files:**
- Create: `internal/usage/client.go`
- Create: `internal/usage/client_test.go`

**Step 1: Write the failing tests**

```go
// internal/usage/client_test.go
package usage_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/Curt-Park/claudeview/internal/usage"
)

func makeServer(t *testing.T, fiveHourPct, sevenDayPct float64) *httptest.Server {
    t.Helper()
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        payload := map[string]any{
            "five_hour": map[string]any{
                "utilization": fiveHourPct,
                "resets_at":   time.Now().Add(2 * time.Hour).Format(time.RFC3339Nano),
            },
            "seven_day": map[string]any{
                "utilization": sevenDayPct,
                "resets_at":   time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339Nano),
            },
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(payload)
    }))
}

func TestClientFetch(t *testing.T) {
    srv := makeServer(t, 42.5, 18.0)
    defer srv.Close()

    c := usage.NewClient("test-token", srv.URL)
    data, stale, err := c.Fetch(t.Context())
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if stale {
        t.Error("expected fresh result on first fetch")
    }
    if data.FiveHour == nil || data.FiveHour.Utilization != 42.5 {
        t.Errorf("unexpected FiveHour: %v", data.FiveHour)
    }
    if data.SevenDay == nil || data.SevenDay.Utilization != 18.0 {
        t.Errorf("unexpected SevenDay: %v", data.SevenDay)
    }
}

func TestClientFetchCaches(t *testing.T) {
    calls := 0
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        calls++
        payload := map[string]any{
            "five_hour": map[string]any{"utilization": 10.0, "resets_at": time.Now().Add(time.Hour).Format(time.RFC3339Nano)},
            "seven_day": map[string]any{"utilization": 5.0, "resets_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano)},
        }
        json.NewEncoder(w).Encode(payload)
    }))
    defer srv.Close()

    c := usage.NewClient("tok", srv.URL)
    c.Fetch(t.Context())
    c.Fetch(t.Context()) // should hit cache, not server

    if calls != 1 {
        t.Errorf("expected 1 HTTP call, got %d (caching not working)", calls)
    }
}

func TestClientFetchHTTPError(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        http.Error(w, "internal server error", http.StatusInternalServerError)
    }))
    defer srv.Close()

    c := usage.NewClient("tok", srv.URL)
    _, _, err := c.Fetch(t.Context())
    if err == nil {
        t.Fatal("expected error for HTTP 500")
    }
}

func TestClientFetchStaleOnError(t *testing.T) {
    // First call succeeds, second call fails → should return stale data.
    callCount := 0
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        callCount++
        if callCount == 1 {
            payload := map[string]any{
                "five_hour": map[string]any{"utilization": 50.0, "resets_at": time.Now().Add(time.Hour).Format(time.RFC3339Nano)},
                "seven_day": map[string]any{"utilization": 25.0, "resets_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano)},
            }
            json.NewEncoder(w).Encode(payload)
        } else {
            http.Error(w, "error", http.StatusInternalServerError)
        }
    }))
    defer srv.Close()

    c := usage.NewClient("tok", srv.URL)
    // Force TTL to 0 so second call isn't cached.
    c.SetTTL(0)
    c.Fetch(t.Context())
    data, stale, err := c.Fetch(t.Context())
    if err != nil {
        t.Fatalf("expected stale fallback, got error: %v", err)
    }
    if !stale {
        t.Error("expected stale=true on second failing fetch")
    }
    if data == nil || data.FiveHour == nil {
        t.Error("expected stale data to be non-nil")
    }
}
```

**Step 2: Run tests to confirm they fail**

```
go test ./internal/usage/... -run TestClient -v
```
Expected: compile error — `usage.NewClient` not defined.

**Step 3: Implement**

```go
// internal/usage/client.go
package usage

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

const (
    defaultBaseURL = "https://api.anthropic.com"
    usagePath      = "/api/oauth/usage"
    defaultTTL     = 60 * time.Second
    requestTimeout = 5 * time.Second
    betaHeader     = "oauth-2025-04-20"
)

// Window holds utilization data for one rate-limit window.
type Window struct {
    Utilization float64    // 0-100 percentage
    ResetsAt    *time.Time // nil if not parseable
}

// Data holds the full API response.
type Data struct {
    FiveHour     *Window
    SevenDay     *Window
    SevenDayOpus *Window // nil for non-Opus tiers
}

// Client fetches and caches usage data from the Anthropic OAuth usage API.
type Client struct {
    token   string
    baseURL string
    ttl     time.Duration

    mu       sync.Mutex
    cached   *Data
    cachedAt time.Time
    lastGood *Data // last successful response, for stale fallback
}

// NewClient creates a new usage client.
// baseURL is optional; leave empty for the real Anthropic API.
// Tests pass a local httptest server URL.
func NewClient(token, baseURL string) *Client {
    if baseURL == "" {
        baseURL = defaultBaseURL
    }
    return &Client{token: token, baseURL: baseURL, ttl: defaultTTL}
}

// SetTTL overrides the cache TTL (used in tests to bypass caching).
func (c *Client) SetTTL(d time.Duration) { c.ttl = d }

// Fetch returns usage data, using cache if fresh.
// stale=true means a previously successful response is returned due to a fetch error.
func (c *Client) Fetch(ctx context.Context) (*Data, bool, error) {
    c.mu.Lock()
    if c.cached != nil && time.Since(c.cachedAt) < c.ttl {
        cached := c.cached
        c.mu.Unlock()
        return cached, false, nil
    }
    c.mu.Unlock()

    ctx, cancel := context.WithTimeout(ctx, requestTimeout)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+usagePath, nil)
    if err != nil {
        return c.stale(fmt.Errorf("building request: %w", err))
    }
    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("anthropic-beta", betaHeader)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return c.stale(fmt.Errorf("fetching usage: %w", err))
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return c.stale(fmt.Errorf("usage API returned HTTP %d", resp.StatusCode))
    }

    var raw struct {
        FiveHour     *rawWindow `json:"five_hour"`
        SevenDay     *rawWindow `json:"seven_day"`
        SevenDayOpus *rawWindow `json:"seven_day_opus"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
        return c.stale(fmt.Errorf("decoding response: %w", err))
    }

    data := &Data{
        FiveHour:     parseWindow(raw.FiveHour),
        SevenDay:     parseWindow(raw.SevenDay),
        SevenDayOpus: parseWindow(raw.SevenDayOpus),
    }

    c.mu.Lock()
    c.cached = data
    c.cachedAt = time.Now()
    c.lastGood = data
    c.mu.Unlock()

    return data, false, nil
}

type rawWindow struct {
    Utilization float64 `json:"utilization"`
    ResetsAtRaw *string `json:"resets_at"`
}

func parseWindow(r *rawWindow) *Window {
    if r == nil {
        return nil
    }
    w := &Window{Utilization: r.Utilization}
    if r.ResetsAtRaw != nil {
        if t, err := time.Parse(time.RFC3339Nano, *r.ResetsAtRaw); err == nil {
            w.ResetsAt = &t
        }
    }
    return w
}

// stale returns the last successful response (stale=true) or an error if none.
func (c *Client) stale(err error) (*Data, bool, error) {
    c.mu.Lock()
    last := c.lastGood
    c.mu.Unlock()
    if last != nil {
        return last, true, nil
    }
    return nil, false, err
}
```

**Step 4: Run tests to confirm they pass**

```
go test ./internal/usage/... -run TestClient -v
```
Expected: PASS (4 tests).

**Step 5: Commit**

```bash
git add internal/usage/client.go internal/usage/client_test.go
git commit -m "feat(usage): add HTTP client with 60s caching and stale fallback"
```

---

### Task 3: `internal/usage/bar.go` — progress bar renderer

**Files:**
- Create: `internal/usage/bar.go`
- Create: `internal/usage/bar_test.go`

**Output format** (2 lines, scales with terminal width):
```
5h [████░░░░░░░░░░░░░░░░░░░░░░░░░░]  8%   reset in 4h 9m
7d [████████████████████░░░░░░░░░░] 68%   reset in 1d 2h
```

Color thresholds: >95% red bold, >80% yellow bold, otherwise green.
`width` parameter controls how wide the `█░` bar segment is.

**Step 1: Write the failing tests**

```go
// internal/usage/bar_test.go
package usage_test

import (
    "strings"
    "testing"
    "time"

    "github.com/Curt-Park/claudeview/internal/usage"
)

func TestRenderBar_Normal(t *testing.T) {
    resetsAt := time.Now().Add(4*time.Hour + 9*time.Minute)
    data := &usage.Data{
        FiveHour: &usage.Window{Utilization: 8.0, ResetsAt: &resetsAt},
        SevenDay: &usage.Window{Utilization: 68.0, ResetsAt: &resetsAt},
    }
    out := usage.RenderBar(data, false, 80)
    lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
    }
    if !strings.Contains(out, "5h") {
        t.Error("expected '5h' label")
    }
    if !strings.Contains(out, "7d") {
        t.Error("expected '7d' label")
    }
    if !strings.Contains(out, "reset in") {
        t.Error("expected 'reset in' countdown label")
    }
}

func TestRenderBar_NilData(t *testing.T) {
    out := usage.RenderBar(nil, false, 80)
    if out != "" {
        t.Errorf("expected empty string for nil data, got %q", out)
    }
}

func TestRenderBar_NoResetTime(t *testing.T) {
    data := &usage.Data{
        FiveHour: &usage.Window{Utilization: 50.0, ResetsAt: nil},
        SevenDay: &usage.Window{Utilization: 20.0, ResetsAt: nil},
    }
    out := usage.RenderBar(data, false, 80)
    if !strings.Contains(out, "5h") {
        t.Error("expected '5h' label even without reset time")
    }
    // No crash; reset in label omitted when ResetsAt is nil
    if strings.Contains(out, "reset in") {
        t.Error("should not show 'reset in' when ResetsAt is nil")
    }
}

func TestRenderBar_Stale(t *testing.T) {
    resetsAt := time.Now().Add(time.Hour)
    data := &usage.Data{
        FiveHour: &usage.Window{Utilization: 30.0, ResetsAt: &resetsAt},
        SevenDay: &usage.Window{Utilization: 10.0, ResetsAt: &resetsAt},
    }
    out := usage.RenderBar(data, true, 80)
    if !strings.Contains(out, "5h") {
        t.Error("expected bar to render even when stale")
    }
}

func TestRenderBar_ProgressFilled(t *testing.T) {
    resetsAt := time.Now().Add(time.Hour)
    data := &usage.Data{
        FiveHour: &usage.Window{Utilization: 100.0, ResetsAt: &resetsAt},
        SevenDay: &usage.Window{Utilization: 0.0, ResetsAt: &resetsAt},
    }
    out := usage.RenderBar(data, false, 80)
    // 100% should have no empty cells, 0% should have no filled cells
    lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
    if len(lines) != 2 {
        t.Fatalf("expected 2 lines, got %d", len(lines))
    }
}

func TestRenderBar_WithOpus(t *testing.T) {
    resetsAt := time.Now().Add(time.Hour)
    data := &usage.Data{
        FiveHour:     &usage.Window{Utilization: 10.0, ResetsAt: &resetsAt},
        SevenDay:     &usage.Window{Utilization: 20.0, ResetsAt: &resetsAt},
        SevenDayOpus: &usage.Window{Utilization: 5.0, ResetsAt: &resetsAt},
    }
    out := usage.RenderBar(data, false, 80)
    lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
    if len(lines) != 3 {
        t.Fatalf("expected 3 lines (5h + 7d + opus), got %d", len(lines))
    }
}
```

**Step 2: Run tests to confirm they fail**

```
go test ./internal/usage/... -run TestRenderBar -v
```
Expected: compile error — `usage.RenderBar` not defined.

**Step 3: Implement**

```go
// internal/usage/bar.go
package usage

import (
    "fmt"
    "strings"
    "time"

    "github.com/charmbracelet/lipgloss"
)

var (
    colorNormal   = lipgloss.Color("82")  // green
    colorWarning  = lipgloss.Color("214") // yellow
    colorCritical = lipgloss.Color("196") // red
    colorEmpty    = lipgloss.Color("238") // dark gray for empty bar cells
    colorDim      = lipgloss.Color("243") // gray for stale
)

// RenderBar renders a 2-line (or 3-line if SevenDayOpus is set) progress bar panel:
//
//	5h [████░░░░░░░░░░░░░░░░░░░░░░░░░░]  8%   reset in 4h 9m
//	7d [████████████████████░░░░░░░░░░] 68%   reset in 1d 2h
//
// width is the total terminal width; the bar segment scales to fill it.
// Returns "" if data is nil.
func RenderBar(data *Data, stale bool, width int) string {
    if data == nil {
        return ""
    }

    // Label width: "opus" is widest at 4 chars.
    const labelW = 4
    // Right side: "100%   reset in 99d 23h" ≈ 26 chars max
    const rightW = 28
    // Bar segment fills the rest: [label] [bar] [right]
    // 3 = space + "[" + "]"
    barW := width - labelW - 3 - rightW
    if barW < 8 {
        barW = 8
    }
    if barW > 80 {
        barW = 80
    }

    var rows []string
    if w := data.FiveHour; w != nil {
        rows = append(rows, renderProgressRow("5h", w, stale, labelW, barW))
    }
    if w := data.SevenDay; w != nil {
        rows = append(rows, renderProgressRow("7d", w, stale, labelW, barW))
    }
    if w := data.SevenDayOpus; w != nil {
        rows = append(rows, renderProgressRow("opus", w, stale, labelW, barW))
    }

    if len(rows) == 0 {
        return ""
    }
    return strings.Join(rows, "\n")
}

func renderProgressRow(label string, w *Window, stale bool, labelW, barW int) string {
    pct := w.Utilization
    if pct < 0 {
        pct = 0
    }
    if pct > 100 {
        pct = 100
    }

    filled := int(pct / 100.0 * float64(barW))
    empty := barW - filled

    // Choose color based on utilization and stale state.
    var fg lipgloss.Color
    if stale {
        fg = colorDim
    } else if pct > 95 {
        fg = colorCritical
    } else if pct > 80 {
        fg = colorWarning
    } else {
        fg = colorNormal
    }
    barStyle := lipgloss.NewStyle().Foreground(fg)
    if !stale && pct > 80 {
        barStyle = barStyle.Bold(true)
    }
    emptyStyle := lipgloss.NewStyle().Foreground(colorEmpty)

    bar := barStyle.Render(strings.Repeat("█", filled)) + emptyStyle.Render(strings.Repeat("░", empty))

    pctStr := fmt.Sprintf("%3.0f%%", pct)

    var resetStr string
    if w.ResetsAt != nil {
        resetStr = "   reset in " + formatCountdown(*w.ResetsAt)
    }

    labelPad := strings.Repeat(" ", labelW-len([]rune(label)))
    return barStyle.Render(label) + labelPad + " [" + bar + "] " + barStyle.Render(pctStr) + lipgloss.NewStyle().Foreground(colorDim).Render(resetStr)
}

// formatCountdown formats a future time as a human-readable countdown.
func formatCountdown(t time.Time) string {
    d := time.Until(t)
    if d <= 0 {
        return "soon"
    }
    if d < time.Minute {
        return fmt.Sprintf("%ds", int(d.Seconds()))
    }
    if d < time.Hour {
        return fmt.Sprintf("%dm", int(d.Minutes()))
    }
    if d < 24*time.Hour {
        h := int(d.Hours())
        m := int(d.Minutes()) % 60
        if m == 0 {
            return fmt.Sprintf("%dh", h)
        }
        return fmt.Sprintf("%dh %dm", h, m)
    }
    days := int(d.Hours()) / 24
    h := int(d.Hours()) % 24
    if h == 0 {
        return fmt.Sprintf("%dd", days)
    }
    return fmt.Sprintf("%dd %dh", days, h)
}
```

**Step 4: Run tests to confirm they pass**

```
go test ./internal/usage/... -run TestRenderBar -v
```
Expected: PASS (6 tests).

**Step 5: Commit**

```bash
git add internal/usage/bar.go internal/usage/bar_test.go
git commit -m "feat(usage): add progress bar renderer with reset countdown"
```

---

### Task 4: Add `UsageLine` to `InfoModel` in `internal/ui/header.go`

**Files:**
- Modify: `internal/ui/header.go`
- Modify: `internal/ui/header_test.go` (or create if it doesn't exist — check first)

**Changes:**

1. Add `UsageLine string` field to `InfoModel`.

2. Update `Height()` to count actual lines in `UsageLine` (2 for 5h+7d, 3 if Opus window also present):
```go
func (info InfoModel) Height(navCount, actionCount, utilCount int) int {
    base := max(5, 1+max(navCount, max(actionCount, utilCount)))
    if info.UsageLine != "" {
        return base + strings.Count(info.UsageLine, "\n") + 1
    }
    return base
}
```

3. In `ViewWithMenu()`, prepend the usage line as the very first line when set:
```go
func (info InfoModel) ViewWithMenu(menu MenuModel) string {
    // ... existing code ...

    lines := []string{projectLine}
    // ... existing loop building lines ...

    result := strings.Join(lines, "\n")
    if info.UsageLine != "" {
        return info.UsageLine + "\n" + result
    }
    return result
}
```

**Step 1: Check if header_test.go exists**

```
ls internal/ui/header_test.go 2>/dev/null || echo "missing"
```

**Step 2: Write tests** (add to existing file or create new)

```go
func TestInfoModelHeightWithUsageLine(t *testing.T) {
    info := ui.InfoModel{}
    base := info.Height(4, 3, 1)

    info.UsageLine = "[5h: 8% 4h 9m] [7d: 68% 1d 2h]"
    withUsage := info.Height(4, 3, 1)

    if withUsage != base+1 {
        t.Errorf("expected height %d with usage line, got %d", base+1, withUsage)
    }
}

func TestInfoModelViewWithUsageLine(t *testing.T) {
    info := ui.InfoModel{
        UsageLine: "[5h: 8%] [7d: 68%]",
        Width:     80,
    }
    menu := ui.MenuModel{}
    out := info.ViewWithMenu(menu)
    lines := strings.Split(out, "\n")
    if lines[0] != "[5h: 8%] [7d: 68%]" {
        t.Errorf("expected usage line as first line, got %q", lines[0])
    }
}

func TestInfoModelViewNoUsageLine(t *testing.T) {
    info := ui.InfoModel{Width: 80}
    menu := ui.MenuModel{}
    out := info.ViewWithMenu(menu)
    // First line should be the Project line, not an empty usage line
    if strings.HasPrefix(out, "\n") {
        t.Error("should not start with newline when UsageLine is empty")
    }
}
```

**Step 3: Run tests to confirm they fail**

```
go test ./internal/ui/... -run TestInfoModel -v
```
Expected: FAIL — `UsageLine` field not yet added.

**Step 4: Make the changes** to `internal/ui/header.go` as described above.

**Step 5: Run tests to confirm they pass**

```
go test ./internal/ui/... -run TestInfoModel -v
```
Expected: PASS

**Step 6: Run full test suite to catch regressions**

```
make test
```
Expected: PASS

**Step 7: Commit**

```bash
git add internal/ui/header.go internal/ui/header_test.go
git commit -m "feat(usage): add UsageLine field to InfoModel for usage bar display"
```

---

### Task 5: Wire usage client into `cmd/root.go`

**Files:**
- Modify: `cmd/root.go`

**Step 1: Add fields to `rootModel`**

```go
type rootModel struct {
    // ... existing fields ...

    // Usage bar
    usageClient *usage.Client
    usageData   *usage.Data
    usageStale  bool
    usageTick   int // increments each tick; refresh at multiples of 60
}
```

**Step 2: Initialize in `newRootModel()`**

After the existing initialization block, add:

```go
// Initialize usage client if credentials are available.
credPath := filepath.Join(config.ClaudeDir(), ".credentials.json")
if token, err := usage.ReadToken(credPath); err == nil {
    rm.usageClient = usage.NewClient(token, "")
    // Trigger an immediate fetch so the bar shows on startup.
    if data, stale, err := rm.usageClient.Fetch(context.Background()); err == nil {
        rm.usageData = data
        rm.usageStale = stale
    }
}
```

Add imports:
```go
"context"
"path/filepath"
"github.com/Curt-Park/claudeview/internal/usage"
```

**Step 3: Add `usageLoadedMsg` type**

```go
type usageLoadedMsg struct {
    data  *usage.Data
    stale bool
}
```

**Step 4: Add `loadUsageAsync()` method**

```go
func (rm *rootModel) loadUsageAsync() tea.Cmd {
    if rm.usageClient == nil {
        return nil
    }
    client := rm.usageClient
    return func() tea.Msg {
        data, stale, err := client.Fetch(context.Background())
        if err != nil {
            return usageLoadedMsg{stale: true}
        }
        return usageLoadedMsg{data: data, stale: stale}
    }
}
```

**Step 5: Handle refresh in `Update()`**

In the `ui.TickMsg` case:
```go
case ui.TickMsg:
    rm.syncView()
    rm.usageTick++
    if !rm.loading {
        rm.loading = true
        extraCmd = rm.loadDataAsync()
    }
    // Refresh usage every 60 ticks (≈60 seconds).
    if rm.usageTick%60 == 0 {
        extraCmd = tea.Batch(extraCmd, rm.loadUsageAsync())
    }
```

Add a case for `usageLoadedMsg` in the switch:
```go
case usageLoadedMsg:
    rm.usageData = msg.data
    rm.usageStale = msg.stale
    rm.syncView()
```

**Step 6: Set `UsageLine` in `syncView()`**

At the start of `syncView()`, add before `rm.updateInfo()`:
```go
// Render usage bar (empty string if no data).
rm.app.Info.UsageLine = usage.RenderBar(rm.usageData, rm.usageStale, w)
```
(`w` is already computed at the top of `syncView()` as the content width.)

**Step 7: Add demo usage data**

In `internal/demo/generator.go`, add:

```go
// GenerateUsage returns synthetic usage data for demo mode.
func GenerateUsage() *usage.Data {
    resetsAt5h := time.Now().Add(4*time.Hour + 9*time.Minute)
    resetsAt7d := time.Now().Add(24*time.Hour + 2*time.Hour)
    return &usage.Data{
        FiveHour: &usage.Window{Utilization: 8.0, ResetsAt: &resetsAt5h},
        SevenDay: &usage.Window{Utilization: 68.0, ResetsAt: &resetsAt7d},
    }
}
```

Import `"github.com/Curt-Park/claudeview/internal/usage"` in `generator.go`.

In `newRootModel()` in `cmd/root.go`, when demo mode is active:
```go
if demoMode {
    rm.usageData = demo.GenerateUsage()
    rm.app.Info.UsageLine = usage.RenderBar(rm.usageData, false, rm.app.Width)
}
```

> **Note:** Check if there's a `demoMode` variable accessible in `newRootModel` or if it needs to be passed in. Currently `demoMode` is a package-level var in `cmd/root.go`, so it's accessible.

**Step 8: Build and verify**

```
go build ./...
```
Expected: success.

**Step 9: Run full test suite**

```
make test
```
Expected: PASS

**Step 10: Manual smoke test**

```
# With real credentials:
./claudeview
# → first line should show [5h: X% Xh Xm] [7d: X% Xd Xh]

# Demo mode:
./claudeview --demo
# → first line should show [5h: 8% 4h 9m] [7d: 68% 1d 2h]
```

**Step 11: Commit**

```bash
git add cmd/root.go internal/demo/generator.go
git commit -m "feat(usage): wire usage client into rootModel with async refresh and usage bar"
```

---

### Task 6: Pre-completion checklist

**Step 1: Format**
```
make fmt
```
Fix any formatting issues, then re-stage changed files.

**Step 2: Lint**
```
make lint
```
Fix any lint issues.

**Step 3: Full test suite with race detector**
```
make test
```
All tests must pass.

**Step 4: Commit any fixes**
```bash
git add -p
git commit -m "fix(usage): address fmt/lint issues"
```

---

## Summary

| File | Change |
|------|--------|
| `internal/usage/credentials.go` | New — reads OAuth token |
| `internal/usage/credentials_test.go` | New — 3 tests |
| `internal/usage/client.go` | New — HTTP client with 60s cache + stale fallback |
| `internal/usage/client_test.go` | New — 4 tests using httptest |
| `internal/usage/bar.go` | New — compact bracket renderer + countdown formatter |
| `internal/usage/bar_test.go` | New — 5 tests |
| `internal/ui/header.go` | Add `UsageLine string` to `InfoModel`, update `Height()` and `ViewWithMenu()` |
| `internal/ui/header_test.go` | Add 3 tests for usage line behavior |
| `cmd/root.go` | Add client init, async refresh, `usageLoadedMsg`, `syncView` wiring |
| `internal/demo/generator.go` | Add `GenerateUsage()` with demo values |
