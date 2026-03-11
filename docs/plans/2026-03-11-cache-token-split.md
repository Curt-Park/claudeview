# Cache Token Split Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Separate `cache_read_input_tokens` from TOKEN_IN so the display shows genuinely new input tokens and cached context separately, preventing inflated token counts in long sessions.

**Architecture:** Add `CacheReadTokens` field through the data stack (transcript → model → provider), add `FormatTokenInOutCache` formatter, update three display call sites and two column headers. `FormatTokenInOutCache` omits `+cache` when cache is zero so existing display is unchanged for non-cached sessions.

**Tech Stack:** Go, `internal/transcript`, `internal/model`, `internal/provider`, `internal/ui`

---

### Task 1: Add `NewInputTokens()` to `transcript.Usage` and update sample data

**Files:**
- Modify: `internal/transcript/types.go`
- Modify: `internal/transcript/testdata/sample_transcript.jsonl`
- Modify: `internal/transcript/parser_test.go`

**Step 1: Add cache tokens to the third assistant turn in sample_transcript.jsonl**

Find the third `"type":"assistant"` entry. Change its `usage` field from:
```json
"usage": {"input_tokens": 300, "output_tokens": 40}
```
to:
```json
"usage": {"input_tokens": 300, "cache_creation_input_tokens": 500, "cache_read_input_tokens": 1000, "output_tokens": 40}
```

**Step 2: Write failing test for `NewInputTokens()` in parser_test.go**

Add to `TestParseFile` after the existing `InputTokens` assertion:
```go
if usage.CacheReadTokens != 1000 {
    t.Errorf("expected CacheReadTokens=1000, got %d", usage.CacheReadTokens)
}
```

Also update the `InputTokens` assertion comment to clarify it excludes cache reads:
```go
// InputTokens should be sum of input_tokens + cache_creation only (not cache_read)
// turn1: 100, turn2: 200, turn3: 300+500=800 → total 1100
if usage.InputTokens != 1100 {
    t.Errorf("expected InputTokens=1100, got %d", usage.InputTokens)
}
```

**Step 3: Run test to verify it fails**

```bash
go test ./internal/transcript/... -run TestParseFile -v
```
Expected: FAIL — `CacheReadTokens` field doesn't exist yet, `InputTokens` assertion is wrong.

**Step 4: Add `NewInputTokens()` to `transcript/types.go`**

```go
// NewInputTokens returns input_tokens + cache_creation_input_tokens,
// excluding cache_read_input_tokens (which represent re-used context, not new input).
func (u Usage) NewInputTokens() int {
    return u.InputTokens + u.CacheCreationInputTokens
}
```

Keep `TotalInputTokens()` as-is (it may be used externally or for future cost calculations).

**Step 5: Run test again**

```bash
go test ./internal/transcript/... -run TestParseFile -v
```
Expected: Still FAIL — `CacheReadTokens` field doesn't exist in the model yet (Task 2).

**Step 6: Commit what compiles so far**

```bash
git add internal/transcript/types.go internal/transcript/testdata/sample_transcript.jsonl
git commit -m "feat: add NewInputTokens() excluding cache reads to transcript.Usage"
```

---

### Task 2: Add `CacheReadTokens` to `model.Turn` and `model.TokenCount`

**Files:**
- Modify: `internal/model/turn.go`
- Modify: `internal/model/session.go`

**Step 1: Add field to `model.Turn`**

In `internal/model/turn.go`, add after `InputTokens`:
```go
CacheReadTokens int
```

**Step 2: Add field to `model.TokenCount`**

In `internal/model/session.go`, add after `InputTokens`:
```go
CacheReadTokens int
```

**Step 3: Run all tests**

```bash
go test ./...
```
Expected: PASS (new fields are zero-valued, no existing behaviour changes yet).

**Step 4: Commit**

```bash
git add internal/model/turn.go internal/model/session.go
git commit -m "feat: add CacheReadTokens field to model.Turn and model.TokenCount"
```

---

### Task 3: Update `transcript/parser.go` to use `NewInputTokens()` and accumulate cache reads

**Files:**
- Modify: `internal/transcript/parser.go`
- Modify: `internal/transcript/parser_test.go`

**Step 1: Write failing test assertion for `CacheReadTokens` in `TokensByModel`**

The `TestParseFile` addition from Task 1 (Step 2) is already there — it will fail because `Usage.CacheReadTokens` doesn't exist yet on the `transcript.Usage` type returned by `TokensByModel`.

`transcript.Usage` already has `CacheReadInputTokens` — we use this field directly in `TokensByModel` accumulation (it's the same `Usage` struct stored in `ParsedTranscript.TokensByModel`).

Update the test assertion to use the correct field name:
```go
if usage.CacheReadInputTokens != 1000 {
    t.Errorf("expected CacheReadInputTokens=1000, got %d", usage.CacheReadInputTokens)
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/transcript/... -run TestParseFile -v
```
Expected: FAIL — `InputTokens` is still 2100 (using `TotalInputTokens()`), `CacheReadInputTokens` is 0.

**Step 3: Update the three call sites in `parser.go`**

**Call site 1** — `flushPendingTurn` (line ~80):
```go
// before:
u.InputTokens += turn.Usage.TotalInputTokens()
u.OutputTokens += turn.Usage.OutputTokens

// after:
u.InputTokens += turn.Usage.NewInputTokens()
u.CacheReadInputTokens += turn.Usage.CacheReadInputTokens
u.OutputTokens += turn.Usage.OutputTokens
```

**Call site 2** — `mergeTurns` (line ~179):
```go
// before:
pending.Usage.InputTokens += next.Usage.TotalInputTokens()
pending.Usage.OutputTokens += next.Usage.OutputTokens

// after:
pending.Usage.InputTokens += next.Usage.NewInputTokens()
pending.Usage.CacheReadInputTokens += next.Usage.CacheReadInputTokens
pending.Usage.OutputTokens += next.Usage.OutputTokens
```

**Call site 3** — `ParseAggregatesIncremental` live aggregation (line ~483):
```go
// before:
u.InputTokens += msg.Usage.TotalInputTokens()
u.OutputTokens += msg.Usage.OutputTokens

// after:
u.InputTokens += msg.Usage.NewInputTokens()
u.CacheReadInputTokens += msg.Usage.CacheReadInputTokens
u.OutputTokens += msg.Usage.OutputTokens
```

**Step 4: Run tests**

```bash
go test ./internal/transcript/... -v
```
Expected: PASS — `InputTokens=1100`, `CacheReadInputTokens=1000`.

**Step 5: Run all tests**

```bash
go test ./...
```
Expected: PASS.

**Step 6: Commit**

```bash
git add internal/transcript/parser.go internal/transcript/parser_test.go
git commit -m "feat: split cache reads from new input tokens in parser accumulation"
```

---

### Task 4: Update `provider/live.go` to pass `CacheReadTokens` through

**Files:**
- Modify: `internal/provider/live.go`

**Step 1: Update turn conversion (line ~210)**

```go
// before:
turn := model.Turn{
    Role:         t.Role,
    Text:         t.Text,
    Thinking:     t.Thinking,
    ModelName:    t.Model,
    InputTokens:  t.Usage.TotalInputTokens(),
    OutputTokens: t.Usage.OutputTokens,
    Timestamp:    t.Timestamp,
}

// after:
turn := model.Turn{
    Role:            t.Role,
    Text:            t.Text,
    Thinking:        t.Thinking,
    ModelName:       t.Model,
    InputTokens:     t.Usage.NewInputTokens(),
    CacheReadTokens: t.Usage.CacheReadInputTokens,
    OutputTokens:    t.Usage.OutputTokens,
    Timestamp:       t.Timestamp,
}
```

**Step 2: Update session `TokensByModel` population (line ~268)**

```go
// before:
s.TokensByModel[m] = model.TokenCount{InputTokens: u.InputTokens, OutputTokens: u.OutputTokens}

// after:
s.TokensByModel[m] = model.TokenCount{
    InputTokens:     u.InputTokens,
    CacheReadTokens: u.CacheReadInputTokens,
    OutputTokens:    u.OutputTokens,
}
```

**Step 3: Update subagent merge loop (line ~296)**

```go
// before:
cur.InputTokens += u.InputTokens
cur.OutputTokens += u.OutputTokens

// after:
cur.InputTokens += u.InputTokens
cur.CacheReadTokens += u.CacheReadInputTokens
cur.OutputTokens += u.OutputTokens
```

**Step 4: Run all tests**

```bash
go test ./...
```
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/provider/live.go
git commit -m "feat: propagate CacheReadTokens through provider/live.go"
```

---

### Task 5: Add `FormatTokenInOutCache` and update display layer

**Files:**
- Modify: `internal/model/format.go`
- Modify: `internal/model/session.go`
- Modify: `internal/ui/chat_item.go`
- Modify: `internal/ui/app.go`
- Modify: `internal/view/chat.go`
- Modify: `internal/view/sessions.go`

**Step 1: Write failing test for `FormatTokenInOutCache` in `internal/model/format_test.go`**

Check if `format_test.go` exists first. If not, create it:

```go
package model_test

import (
    "testing"
    "github.com/Curt-Park/claudeview/internal/model"
)

func TestFormatTokenInOutCache(t *testing.T) {
    tests := []struct {
        in, cache, out int
        want           string
    }{
        {50000, 7400000, 26000, "50k+7.4M/26k"},
        {243000, 0, 26000, "243k/26k"},   // no cache: same as old format
        {0, 0, 0, "0/0"},
    }
    for _, tt := range tests {
        got := model.FormatTokenInOutCache(tt.in, tt.cache, tt.out)
        if got != tt.want {
            t.Errorf("FormatTokenInOutCache(%d,%d,%d) = %q, want %q",
                tt.in, tt.cache, tt.out, got, tt.want)
        }
    }
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/model/... -run TestFormatTokenInOutCache -v
```
Expected: FAIL — function doesn't exist yet.

**Step 3: Add `FormatTokenInOutCache` to `internal/model/format.go`**

```go
// FormatTokenInOutCache formats input/cache-read/output token counts.
// When cache is zero, omits the +cache section (e.g. "50k/26k").
// When cache is non-zero, shows "50k+7.4M/26k".
func FormatTokenInOutCache(in, cache, out int) string {
    if cache == 0 {
        return FormatTokenCount(in) + "/" + FormatTokenCount(out)
    }
    return FormatTokenCount(in) + "+" + FormatTokenCount(cache) + "/" + FormatTokenCount(out)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/model/... -run TestFormatTokenInOutCache -v
```
Expected: PASS.

**Step 5: Update `TokenString()` in `internal/model/session.go`**

```go
// before:
parts = append(parts, fmt.Sprintf("%s:%s", ShortModelName(m), FormatTokenInOut(tc.InputTokens, tc.OutputTokens)))

// after:
parts = append(parts, fmt.Sprintf("%s:%s", ShortModelName(m), FormatTokenInOutCache(tc.InputTokens, tc.CacheReadTokens, tc.OutputTokens)))
```

**Step 6: Update `TestSessionTokenString` in `internal/model/session_test.go`**

The test with zero cache stays the same — `FormatTokenInOutCache(50000, 0, 12500)` returns `"50k/12k"`, so no change needed for existing tests. Add a new test case:

```go
func TestSessionTokenStringWithCache(t *testing.T) {
    s := &model.Session{
        TokensByModel: map[string]model.TokenCount{
            "claude-haiku-4-5": {InputTokens: 243000, CacheReadTokens: 7400000, OutputTokens: 26000},
        },
    }
    got := s.TokenString()
    if got != "haiku:243k+7.4M/26k" {
        t.Errorf("TokenString() = %q, want %q", got, "haiku:243k+7.4M/26k")
    }
}
```

**Step 7: Update `ModelTokenLabel()` in `internal/ui/chat_item.go`**

The `byModel` map stores `in` and `out`. Add `cache`:

```go
// before:
type tokenPair struct{ in, out int }
byModel := make(map[string]tokenPair)
addTurn := func(t model.Turn) {
    if t.ModelName == "" {
        return
    }
    p := byModel[t.ModelName]
    p.in += t.InputTokens
    p.out += t.OutputTokens
    byModel[t.ModelName] = p
}
// ...
parts = append(parts, model.ShortModelName(m)+":"+model.FormatTokenInOut(p.in, p.out))

// after:
type tokenPair struct{ in, cache, out int }
byModel := make(map[string]tokenPair)
addTurn := func(t model.Turn) {
    if t.ModelName == "" {
        return
    }
    p := byModel[t.ModelName]
    p.in += t.InputTokens
    p.cache += t.CacheReadTokens
    p.out += t.OutputTokens
    byModel[t.ModelName] = p
}
// ...
parts = append(parts, model.ShortModelName(m)+":"+model.FormatTokenInOutCache(p.in, p.cache, p.out))
```

**Step 8: Update `buildToolCallSubRow()` in `internal/ui/app.go`**

```go
// before:
in, out := tr.ParentTurn.InputTokens, tr.ParentTurn.OutputTokens
if in > 0 || out > 0 {
    modelTok = m + ":" + model.FormatTokenInOut(in, out)
}

// after:
in, cache, out := tr.ParentTurn.InputTokens, tr.ParentTurn.CacheReadTokens, tr.ParentTurn.OutputTokens
if in > 0 || cache > 0 || out > 0 {
    modelTok = m + ":" + model.FormatTokenInOutCache(in, cache, out)
}
```

**Step 9: Update column headers**

In `internal/view/chat.go`:
```go
// before:
{Title: "MODEL:TOKEN_IN/OUT", ...}
// after:
{Title: "MODEL:IN+CACHE/OUT", ...}
```

In `internal/view/sessions.go`:
```go
// before:
{Title: "MODEL:TOKEN_IN/OUT", ...}
// after:
{Title: "MODEL:IN+CACHE/OUT", ...}
```

**Step 10: Run all tests**

```bash
go test ./...
```
Expected: PASS.

**Step 11: Run pre-completion checks**

```bash
make fmt && make lint && make test
```
Expected: all pass, no lint errors.

**Step 12: Commit**

```bash
git add internal/model/format.go internal/model/session.go internal/ui/chat_item.go internal/ui/app.go internal/view/chat.go internal/view/sessions.go internal/model/session_test.go
git commit -m "feat: display cache-read tokens separately as IN+CACHE/OUT"
```
