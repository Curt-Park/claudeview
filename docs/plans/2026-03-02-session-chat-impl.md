# Session Chat View Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the `sessions → agents` leaf with a `sessions → session-chat` view that renders a Claude Code session as a mobile-chat-style timeline of bubbles (user on right, Claude on left, subagents indented).

**Architecture:** Add `model.Turn` to represent a conversation turn. Extend `DataProvider` with `GetTurns(filePath string) []model.Turn`. Pre-load turns in `AppModel.drillDown()` and render via `RenderSessionChat()` using the existing `ContentOffset` scroll mechanism.

**Tech Stack:** Go, lipgloss v1.1.0 (`RoundedBorder`, `PlaceHorizontal`, `Width`), internal transcript parser, Bubble Tea.

**Design doc:** `docs/plans/2026-03-02-session-chat-design.md`

---

## Before you start

Read the following files to get oriented:
- `internal/model/resource.go` — ResourceType constants
- `internal/ui/app.go` — AppModel, DataProvider interface, drillDown, navigateBack, isContentView, contentMaxOffset, View
- `internal/ui/detail_render.go` — existing content view render functions
- `internal/ui/styles.go` — color constants and existing styles
- `internal/ui/menu.go` — TableNavItems, TableUtilItems
- `internal/ui/testhelpers_test.go` — mockDP, helper functions
- `cmd/root.go` — liveDataProvider, demoDataProvider

---

### Task 1: Add `model.Turn` type

**Files:**
- Create: `internal/model/turn.go`

**Step 1: Write the failing test**

Add to `internal/model/agent_test.go` (or create `internal/model/turn_test.go`):
```go
func TestTurn_HasExpectedFields(t *testing.T) {
    turn := model.Turn{
        Role:         "assistant",
        Text:         "hello",
        Thinking:     "hmm",
        ModelName:    "claude-sonnet",
        InputTokens:  100,
        OutputTokens: 50,
    }
    if turn.Role != "assistant" {
        t.Errorf("expected Role=assistant, got %q", turn.Role)
    }
    if turn.InputTokens != 100 {
        t.Errorf("expected InputTokens=100, got %d", turn.InputTokens)
    }
}
```

**Step 2: Run test to verify it fails**

```bash
cd /home/curt/Repositories/claudeview && go test ./internal/model/... -run TestTurn_HasExpectedFields -v
```
Expected: `FAIL — undefined: model.Turn`

**Step 3: Create `internal/model/turn.go`**

```go
package model

import "time"

// Turn represents a single conversation turn (user or assistant).
type Turn struct {
    Role         string     // "user" or "assistant"
    Text         string
    Thinking     string
    ToolCalls    []*ToolCall
    ModelName    string
    InputTokens  int
    OutputTokens int
    Timestamp    time.Time
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/model/... -run TestTurn_HasExpectedFields -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/model/turn.go internal/model/turn_test.go
git commit -m "feat: add model.Turn type for conversation turns"
```

---

### Task 2: Add `ResourceSessionChat` to model

**Files:**
- Modify: `internal/model/resource.go`

**Step 1: Write the failing test**

Add to `internal/model/resource_test.go`:
```go
func TestResourceSessionChatExists(t *testing.T) {
    if model.ResourceSessionChat == "" {
        t.Error("ResourceSessionChat should be non-empty")
    }
    if string(model.ResourceSessionChat) != "session-chat" {
        t.Errorf("expected 'session-chat', got %q", model.ResourceSessionChat)
    }
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/model/... -run TestResourceSessionChatExists -v
```
Expected: `FAIL — undefined: model.ResourceSessionChat`

**Step 3: Add constant to `internal/model/resource.go`**

Add after `ResourceMemoryDetail`:
```go
ResourceSessionChat ResourceType = "session-chat"
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/model/... -run TestResourceSessionChatExists -v
```
Expected: `PASS`

**Step 5: Commit**

```bash
git add internal/model/resource.go internal/model/resource_test.go
git commit -m "feat: add ResourceSessionChat constant"
```

---

### Task 3: Extend `DataProvider` with `GetTurns` + update all implementors

**Files:**
- Modify: `internal/ui/app.go` (DataProvider interface)
- Modify: `internal/ui/testhelpers_test.go` (mockDP)
- Modify: `cmd/root.go` (liveDataProvider + demoDataProvider)

**Step 1: Add `GetTurns` to DataProvider in `internal/ui/app.go`**

In the `DataProvider` interface, add:
```go
GetTurns(filePath string) []model.Turn
```

**Step 2: Run `go build` to see all missing implementations**

```bash
go build ./...
```
Expected: compile errors for mockDP, liveDataProvider, demoDataProvider.

**Step 3: Add `GetTurns` to `mockDP` in `internal/ui/testhelpers_test.go`**

Add field and method:
```go
type mockDP struct {
    projects []*model.Project
    sessions []*model.Session
    agents   []*model.Agent
    plugins  []*model.Plugin
    memories []*model.Memory
    turns    []model.Turn // new
}

func (m *mockDP) GetTurns(_ string) []model.Turn { return m.turns }
```

**Step 4: Add `GetTurns` to `demoDataProvider` in `cmd/root.go`**

```go
func (d *demoDataProvider) GetTurns(_ string) []model.Turn { return nil }
```

**Step 5: Add `GetTurns` to `liveDataProvider` in `cmd/root.go`**

```go
func (l *liveDataProvider) GetTurns(filePath string) []model.Turn {
    parsed, err := transcript.ParseFile(filePath)
    if err != nil {
        return nil
    }
    turns := make([]model.Turn, 0, len(parsed.Turns))
    for _, t := range parsed.Turns {
        turn := model.Turn{
            Role:         t.Role,
            Text:         t.Text,
            Thinking:     t.Thinking,
            ModelName:    t.Model,
            InputTokens:  t.Usage.InputTokens,
            OutputTokens: t.Usage.OutputTokens,
            Timestamp:    t.Timestamp,
        }
        for _, tc := range t.ToolCalls {
            turn.ToolCalls = append(turn.ToolCalls, &model.ToolCall{
                ID:        tc.ID,
                Name:      tc.Name,
                Input:     tc.Input,
                Result:    tc.Result,
                IsError:   tc.IsError,
                Timestamp: tc.Timestamp,
                Duration:  tc.Duration,
            })
        }
        turns = append(turns, turn)
    }
    return turns
}
```

**Step 6: Verify it builds**

```bash
go build ./...
```
Expected: no errors.

**Step 7: Commit**

```bash
git add internal/ui/app.go internal/ui/testhelpers_test.go cmd/root.go
git commit -m "feat: add GetTurns to DataProvider + implement in all providers"
```

---

### Task 4: Wire session-chat into AppModel navigation

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/app_test.go`

**Step 1: Write failing navigation tests**

Add to `internal/ui/app_test.go`:

```go
func TestDrilldownSessionsToSessionChat(t *testing.T) {
    s := &model.Session{ID: "sess-abc123", FilePath: "/tmp/fake.jsonl"}
    app := newApp(model.ResourceSessions)
    app.Table.SetRows([]ui.Row{{
        Cells: []string{s.ShortID(), "topic", "2", "10", "1k", "1h"},
        Data:  s,
    }})

    app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

    if app.Resource != model.ResourceSessionChat {
        t.Errorf("expected resource=session-chat after Enter on sessions, got %s", app.Resource)
    }
}

func TestSessionChatEscReturnsToSessions(t *testing.T) {
    app := newApp(model.ResourceSessionChat)

    app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

    if app.Resource != model.ResourceSessions {
        t.Errorf("expected resource=sessions after Esc from session-chat, got %s", app.Resource)
    }
}

func TestSessionChatIsContentView(t *testing.T) {
    app := newApp(model.ResourceSessionChat)
    app.Width = 120
    app.Height = 40
    // j/k should adjust ContentOffset, not Table.Selected
    before := app.Table.Selected
    app = updateApp(app, keyMsg("j"))
    if app.Table.Selected != before {
        t.Error("session-chat j should not change Table.Selected")
    }
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestDrilldownSessionsToSessionChat|TestSessionChatEsc|TestSessionChatIsContentView" -v
```
Expected: FAIL.

**Step 3: Add `SelectedTurns`, `SubagentTurns`, `SubagentTypes` fields to AppModel**

In `internal/ui/app.go`, in the `AppModel` struct, add after `SelectedMemory`:
```go
// Session chat data (set on drill-down into session-chat)
SelectedTurns  []model.Turn
SubagentTurns  [][]model.Turn
SubagentTypes  []model.AgentType
```

**Step 4: Modify `drillDown()` — change sessions case**

Replace the `case model.ResourceSessions:` block:
```go
case model.ResourceSessions:
    if s, ok := row.Data.(*model.Session); ok {
        m.SelectedSessionID = s.ID
        m.SelectedTurns = m.DataProvider.GetTurns(s.FilePath)
        agents := m.DataProvider.GetAgents(s.ID)
        m.SubagentTurns = nil
        m.SubagentTypes = nil
        for _, a := range agents {
            if a.IsSubagent && a.FilePath != "" {
                m.SubagentTurns = append(m.SubagentTurns, m.DataProvider.GetTurns(a.FilePath))
                m.SubagentTypes = append(m.SubagentTypes, a.Type)
            }
        }
    }
    m.drillInto(model.ResourceSessionChat)
```

**Step 5: Modify `navigateBack()` — add session-chat case**

In the resource hierarchy switch, add:
```go
case model.ResourceSessionChat:
    m.SelectedSessionID = ""
    m.SelectedTurns = nil
    m.SubagentTurns = nil
    m.SubagentTypes = nil
    m.popFilter()
    m.switchResource(model.ResourceSessions)
```

**Step 6: Extend `isContentView()` and `contentMaxOffset()` and `View()`**

In `isContentView()`:
```go
func isContentView(rt model.ResourceType) bool {
    return rt == model.ResourcePluginItemDetail ||
        rt == model.ResourceMemoryDetail ||
        rt == model.ResourceSessionChat
}
```

In `contentMaxOffset()` switch, add:
```go
case model.ResourceSessionChat:
    contentStr = RenderSessionChat(m.SelectedTurns, m.SubagentTurns, m.SubagentTypes, m.contentWidth())
```

In `View()` switch, add:
```go
case model.ResourceSessionChat:
    contentStr = RenderSessionChat(m.SelectedTurns, m.SubagentTurns, m.SubagentTypes, m.contentWidth())
```

**Step 7: Add stub `RenderSessionChat` to `detail_render.go`**

```go
// RenderSessionChat renders a session's conversation as a chat timeline.
func RenderSessionChat(turns []model.Turn, subagentTurns [][]model.Turn, subagentTypes []model.AgentType, width int) string {
    return "" // stub — implemented in Task 7-9
}
```

**Step 8: Run tests**

```bash
go test ./internal/ui/... -run "TestDrilldownSessionsToSessionChat|TestSessionChatEsc|TestSessionChatIsContentView" -v
```
Expected: PASS.

**Step 9: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go internal/ui/detail_render.go
git commit -m "feat: wire session-chat navigation into AppModel"
```

---

### Task 5: Update menu hints for session-chat

**Files:**
- Modify: `internal/ui/menu.go`
- Modify: `internal/ui/menu_test.go`

**Step 1: Write failing test**

Add to `internal/ui/menu_test.go` (check existing test file first for patterns):
```go
func TestSessionChatMenuHints(t *testing.T) {
    items := ui.TableNavItems(model.ResourceSessionChat, false)
    keys := make(map[string]string)
    for _, it := range items {
        keys[it.Key] = it.Desc
    }
    if _, ok := keys["esc"]; !ok {
        t.Error("session-chat nav should include esc hint")
    }
}

func TestSessionsEnterHintIsChat(t *testing.T) {
    items := ui.TableNavItems(model.ResourceSessions, false)
    for _, it := range items {
        if it.Key == "enter" && it.Desc == "see agents" {
            t.Error("sessions enter hint should no longer say 'see agents'")
        }
    }
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestSessionChatMenuHints|TestSessionsEnterHintIsChat" -v
```

**Step 3: Update `TableNavItems` in `internal/ui/menu.go`**

Change the sessions enter hint:
```go
case model.ResourceSessions:
    items = append(items, MenuItem{Key: "enter", Desc: "view chat"})
```

Add session-chat esc hint in the esc switch:
```go
case model.ResourceSessionChat:
    items = append(items, MenuItem{Key: "esc", Desc: "see sessions"})
```

Update `TableUtilItems` to exclude filter for session-chat:
```go
func TableUtilItems(rt model.ResourceType) []MenuItem {
    switch rt {
    case model.ResourceMemoryDetail, model.ResourcePluginItemDetail, model.ResourceSessionChat:
        return nil
    }
    return []MenuItem{{Key: "/", Desc: "filter"}}
}
```

**Step 4: Run tests**

```bash
go test ./internal/ui/... -run "TestSessionChatMenuHints|TestSessionsEnterHintIsChat" -v
```
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ui/menu.go internal/ui/menu_test.go
git commit -m "feat: update menu hints for session-chat navigation"
```

---

### Task 6: Add bubble styles to `styles.go`

**Files:**
- Modify: `internal/ui/styles.go`

This task has no failing test — styles are visual. Just add the constants.

**Step 1: Add bubble styles**

In `internal/ui/styles.go`, add after existing styles:
```go
// Chat bubble border styles
StyleUserBubble = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(colorBlue).
    Padding(0, 1)

StyleClaudeBubble = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(colorGreen).
    Padding(0, 1)

StyleSubagentBubble = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(colorPurple).
    Padding(0, 1)

StyleChatThinking  = lipgloss.NewStyle().Foreground(colorGray)
StyleChatToolOK    = lipgloss.NewStyle().Foreground(colorGreen)
StyleChatToolErr   = lipgloss.NewStyle().Foreground(colorRed)
StyleChatToolName  = lipgloss.NewStyle().Foreground(colorBlue)
StyleChatTokens    = lipgloss.NewStyle().Foreground(colorGray)
StyleChatTimestamp = lipgloss.NewStyle().Foreground(colorDimGray)
StyleChatHeader    = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
```

**Step 2: Verify build**

```bash
go build ./...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add internal/ui/styles.go
git commit -m "feat: add chat bubble styles to styles.go"
```

---

### Task 7: `RenderSessionChat` — user bubbles

**Files:**
- Modify: `internal/ui/detail_render.go`
- Modify: `internal/ui/detail_render_test.go`

**Step 1: Write failing test**

Add to `internal/ui/detail_render_test.go`:
```go
func TestRenderSessionChat_UserBubble(t *testing.T) {
    turns := []model.Turn{
        {Role: "user", Text: "Hello, Claude!", Timestamp: time.Date(2026, 3, 2, 9, 13, 0, 0, time.UTC)},
    }
    got := ui.RenderSessionChat(turns, nil, nil, 80)
    if !strings.Contains(got, "Hello, Claude!") {
        t.Errorf("expected user text in output, got:\n%s", got)
    }
    if !strings.Contains(got, "You") {
        t.Errorf("expected 'You' label in user bubble, got:\n%s", got)
    }
    if !strings.Contains(got, "09:13") {
        t.Errorf("expected timestamp in user bubble, got:\n%s", got)
    }
}

func TestRenderSessionChat_MultilineUserBubble(t *testing.T) {
    turns := []model.Turn{
        {Role: "user", Text: "Requirement:\n- req1\n- req2", Timestamp: time.Now()},
    }
    got := ui.RenderSessionChat(turns, nil, nil, 80)
    if !strings.Contains(got, "req1") || !strings.Contains(got, "req2") {
        t.Errorf("expected multiline text preserved, got:\n%s", got)
    }
}

func TestRenderSessionChat_NilTurnsReturnsEmpty(t *testing.T) {
    got := ui.RenderSessionChat(nil, nil, nil, 80)
    if strings.TrimSpace(got) != "" {
        t.Errorf("expected empty output for nil turns, got %q", got)
    }
}
```

Note: add `"time"` import to test file if not already present.

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestRenderSessionChat_User|TestRenderSessionChat_Nil" -v
```
Expected: FAIL.

**Step 3: Implement user bubble rendering in `RenderSessionChat`**

Replace the stub in `internal/ui/detail_render.go`:

```go
func RenderSessionChat(turns []model.Turn, subagentTurns [][]model.Turn, subagentTypes []model.AgentType, width int) string {
    if len(turns) == 0 {
        return ""
    }

    userBubbleWidth := int(float64(width) * 0.70)
    if userBubbleWidth < 20 {
        userBubbleWidth = width - 4
    }
    if userBubbleWidth > width-4 {
        userBubbleWidth = width - 4
    }

    var sb strings.Builder
    subIdx := 0
    for _, turn := range turns {
        switch turn.Role {
        case "user":
            sb.WriteString(renderUserBubble(turn, userBubbleWidth, width))
            sb.WriteString("\n")
        case "assistant":
            sb.WriteString(renderClaudeBubble(turn, width))
            sb.WriteString("\n")
            // subagent bubbles interleaved after Task calls (Task 9)
            _ = subIdx
        }
    }
    return sb.String()
}

func renderUserBubble(turn model.Turn, bubbleWidth, fullWidth int) string {
    ts := StyleChatTimestamp.Render(turn.Timestamp.Format("15:04"))
    header := ts + " · " + StyleChatHeader.Render("You")
    content := header + "\n" + turn.Text
    bubble := StyleUserBubble.Width(bubbleWidth).Render(content)
    return lipgloss.PlaceHorizontal(fullWidth, lipgloss.Right, bubble)
}

func renderClaudeBubble(turn model.Turn, width int) string {
    return "" // implemented in Task 8
}
```

Add to imports in `detail_render.go`:
```go
import (
    "fmt"
    "os"
    "strings"

    "github.com/charmbracelet/lipgloss"

    "github.com/Curt-Park/claudeview/internal/model"
)
```

**Step 4: Run tests**

```bash
go test ./internal/ui/... -run "TestRenderSessionChat_User|TestRenderSessionChat_Nil" -v
```
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ui/detail_render.go internal/ui/detail_render_test.go
git commit -m "feat: render user bubbles in RenderSessionChat"
```

---

### Task 8: `RenderSessionChat` — Claude bubbles + thinking + tool calls

**Files:**
- Modify: `internal/ui/detail_render.go`
- Modify: `internal/ui/detail_render_test.go`

**Step 1: Write failing tests**

```go
func TestRenderSessionChat_ClaudeBubble(t *testing.T) {
    turns := []model.Turn{
        {
            Role:         "assistant",
            Text:         "네, 시작하겠습니다.",
            ModelName:    "claude-sonnet-4-6",
            InputTokens:  500,
            OutputTokens: 100,
            Timestamp:    time.Date(2026, 3, 2, 9, 14, 0, 0, time.UTC),
        },
    }
    got := ui.RenderSessionChat(turns, nil, nil, 120)
    if !strings.Contains(got, "네, 시작하겠습니다.") {
        t.Errorf("expected assistant text, got:\n%s", got)
    }
    if !strings.Contains(got, "Claude") {
        t.Errorf("expected 'Claude' label, got:\n%s", got)
    }
    if !strings.Contains(got, "09:14") {
        t.Errorf("expected timestamp, got:\n%s", got)
    }
    if !strings.Contains(got, "600") || !strings.Contains(got, "tok") {
        t.Errorf("expected token count (600 total), got:\n%s", got)
    }
}

func TestRenderSessionChat_ThinkingDim(t *testing.T) {
    turns := []model.Turn{
        {Role: "assistant", Text: "response", Thinking: "let me think about this carefully"},
    }
    got := ui.RenderSessionChat(turns, nil, nil, 120)
    if !strings.Contains(got, "think") {
        t.Errorf("expected thinking content in output, got:\n%s", got)
    }
}

func TestRenderSessionChat_ToolCallLines(t *testing.T) {
    input := json.RawMessage(`{"file_path":"internal/ui/app.go"}`)
    result := json.RawMessage(`"120 lines\nof content"`)
    turns := []model.Turn{
        {
            Role: "assistant",
            Text: "done",
            ToolCalls: []*model.ToolCall{
                {Name: "Read", Input: input, Result: result, IsError: false},
                {Name: "Bash", Input: json.RawMessage(`{"command":"make test"}`), IsError: true},
            },
        },
    }
    got := ui.RenderSessionChat(turns, nil, nil, 120)
    if !strings.Contains(got, "Read") {
        t.Errorf("expected 'Read' tool name, got:\n%s", got)
    }
    if !strings.Contains(got, "✓") {
        t.Errorf("expected ✓ for successful tool, got:\n%s", got)
    }
    if !strings.Contains(got, "✗") {
        t.Errorf("expected ✗ for failed tool, got:\n%s", got)
    }
}
```

Add `"encoding/json"` to test imports.

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestRenderSessionChat_Claude|TestRenderSessionChat_Think|TestRenderSessionChat_Tool" -v
```
Expected: FAIL.

**Step 3: Implement `renderClaudeBubble` and tool call lines**

```go
func renderClaudeBubble(turn model.Turn, width int) string {
    claudeWidth := width - 2 // border takes 2 cols

    ts := StyleChatTimestamp.Render(turn.Timestamp.Format("15:04"))
    modelShort := shortModel(turn.ModelName)
    header := "🤖 " + StyleChatHeader.Render("Claude") + " · " + StyleDim.Render(modelShort) + " · " + ts

    var parts []string
    parts = append(parts, header)

    if turn.Thinking != "" {
        thinking := turn.Thinking
        if len([]rune(thinking)) > 80 {
            thinking = string([]rune(thinking)[:77]) + "..."
        }
        parts = append(parts, StyleChatThinking.Render("···thinking: "+thinking+"···"))
    }

    if turn.Text != "" {
        parts = append(parts, "")
        parts = append(parts, turn.Text)
    }

    if len(turn.ToolCalls) > 0 {
        parts = append(parts, "")
        for _, tc := range turn.ToolCalls {
            parts = append(parts, renderToolLine(tc))
        }
    }

    totalTok := turn.InputTokens + turn.OutputTokens
    if totalTok > 0 {
        tokStr := StyleChatTokens.Render(fmt.Sprintf("░ %s tok", model.FormatTokenCount(totalTok)))
        parts = append(parts, "")
        parts = append(parts, lipgloss.PlaceHorizontal(claudeWidth-2, lipgloss.Right, tokStr))
    }

    content := strings.Join(parts, "\n")
    return StyleClaudeBubble.Width(claudeWidth).Render(content)
}

func renderToolLine(tc *model.ToolCall) string {
    name := StyleChatToolName.Render("▸ " + tc.Name)
    summary := tc.InputSummary()
    if len([]rune(summary)) > 40 {
        summary = string([]rune(summary)[:37]) + "..."
    }

    var outcome string
    if tc.IsError {
        outcome = StyleChatToolErr.Render("✗ error")
    } else {
        result := tc.ResultSummary()
        outcome = StyleChatToolOK.Render("→ "+result) + " " + StyleChatToolOK.Render("✓")
    }
    return "  " + name + "  " + StyleDim.Render(summary) + "  " + outcome
}

func shortModel(m string) string {
    lower := strings.ToLower(m)
    switch {
    case strings.Contains(lower, "opus"):
        return "opus"
    case strings.Contains(lower, "sonnet"):
        return "sonnet"
    case strings.Contains(lower, "haiku"):
        return "haiku"
    default:
        return m
    }
}
```

**Note on `model.FormatTokenCount`:** Add this helper to `internal/model/format.go` if not already present (check the file first). If it exists under a different name, use that instead. The function should format an int as "1.2k", "1.5M", etc. — this logic already exists in `session.go`'s `formatTokens` unexported function; expose it or duplicate it as a package-level function.

**Step 4: Run tests**

```bash
go test ./internal/ui/... -run "TestRenderSessionChat_Claude|TestRenderSessionChat_Think|TestRenderSessionChat_Tool" -v
```
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ui/detail_render.go internal/ui/detail_render_test.go internal/model/format.go
git commit -m "feat: render Claude bubbles, thinking, and tool call lines"
```

---

### Task 9: `RenderSessionChat` — subagent bubbles

**Files:**
- Modify: `internal/ui/detail_render.go`
- Modify: `internal/ui/detail_render_test.go`

**Step 1: Write failing test**

```go
func TestRenderSessionChat_SubagentBubble(t *testing.T) {
    taskInput := json.RawMessage(`{"description":"Explore codebase","subagent_type":"Explore"}`)
    mainTurns := []model.Turn{
        {
            Role: "assistant",
            Text: "Let me explore.",
            ToolCalls: []*model.ToolCall{
                {Name: "Task", Input: taskInput, IsError: false},
            },
        },
    }
    subTurns := [][]model.Turn{
        {
            {Role: "assistant", Text: "Found 42 files.", ModelName: "claude-sonnet-4-6"},
        },
    }
    subTypes := []model.AgentType{model.AgentTypeExplore}

    got := ui.RenderSessionChat(mainTurns, subTurns, subTypes, 120)
    if !strings.Contains(got, "Found 42 files.") {
        t.Errorf("expected subagent text in output, got:\n%s", got)
    }
    if !strings.Contains(got, "Explorer") || !strings.Contains(got, "🔍") {
        t.Errorf("expected subagent label with icon, got:\n%s", got)
    }
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/... -run TestRenderSessionChat_SubagentBubble -v
```
Expected: FAIL.

**Step 3: Implement subagent bubble rendering**

Update `RenderSessionChat` to count Task calls and interleave subagent bubbles:

```go
func RenderSessionChat(turns []model.Turn, subagentTurns [][]model.Turn, subagentTypes []model.AgentType, width int) string {
    if len(turns) == 0 {
        return ""
    }

    userBubbleWidth := int(float64(width) * 0.70)
    if userBubbleWidth < 20 {
        userBubbleWidth = width - 4
    }
    if userBubbleWidth > width-4 {
        userBubbleWidth = width - 4
    }

    var sb strings.Builder
    subIdx := 0
    for _, turn := range turns {
        switch turn.Role {
        case "user":
            sb.WriteString(renderUserBubble(turn, userBubbleWidth, width))
            sb.WriteString("\n")
        case "assistant":
            sb.WriteString(renderClaudeBubble(turn, width))
            sb.WriteString("\n")
            // Append subagent bubble for each Task tool call in this turn
            for _, tc := range turn.ToolCalls {
                if tc.Name == "Task" && subIdx < len(subagentTurns) {
                    agentType := model.AgentTypeGeneral
                    if subIdx < len(subagentTypes) {
                        agentType = subagentTypes[subIdx]
                    }
                    sb.WriteString(renderSubagentBubbles(subagentTurns[subIdx], agentType, width))
                    sb.WriteString("\n")
                    subIdx++
                }
            }
        }
    }
    return sb.String()
}

func renderSubagentBubbles(turns []model.Turn, agentType model.AgentType, width int) string {
    if len(turns) == 0 {
        return ""
    }
    indent := "  "
    subWidth := width - 6 // indent + border

    icon := subagentIcon(agentType)
    label := icon + " " + StyleChatHeader.Render(agentDisplayName(agentType))

    var parts []string
    for _, turn := range turns {
        if turn.Role != "assistant" {
            continue
        }
        ts := StyleChatTimestamp.Render(turn.Timestamp.Format("15:04"))
        header := label + " · " + ts

        var lines []string
        lines = append(lines, header)
        if turn.Text != "" {
            lines = append(lines, "")
            lines = append(lines, turn.Text)
        }
        for _, tc := range turn.ToolCalls {
            lines = append(lines, renderToolLine(tc))
        }
        totalTok := turn.InputTokens + turn.OutputTokens
        if totalTok > 0 {
            tokStr := StyleChatTokens.Render(fmt.Sprintf("░ %s tok", model.FormatTokenCount(totalTok)))
            lines = append(lines, "")
            lines = append(lines, lipgloss.PlaceHorizontal(subWidth-2, lipgloss.Right, tokStr))
        }

        bubble := StyleSubagentBubble.Width(subWidth).Render(strings.Join(lines, "\n"))
        // Indent the bubble
        for _, line := range strings.Split(bubble, "\n") {
            parts = append(parts, indent+line)
        }
        parts = append(parts, "")
    }
    if len(parts) == 0 {
        return ""
    }
    return strings.Join(parts, "\n")
}

func subagentIcon(t model.AgentType) string {
    switch t {
    case model.AgentTypeExplore:
        return "🔍"
    case model.AgentTypePlan:
        return "📋"
    case model.AgentTypeBash:
        return "💻"
    default:
        return "⚙️"
    }
}

func agentDisplayName(t model.AgentType) string {
    switch t {
    case model.AgentTypeExplore:
        return "Explorer"
    case model.AgentTypePlan:
        return "Planner"
    case model.AgentTypeBash:
        return "Bash"
    default:
        return "Agent"
    }
}
```

**Step 4: Run tests**

```bash
go test ./internal/ui/... -run TestRenderSessionChat -v
```
Expected: all PASS.

**Step 5: Commit**

```bash
git add internal/ui/detail_render.go internal/ui/detail_render_test.go
git commit -m "feat: render subagent bubbles in session-chat view"
```

---

### Task 10: Pre-completion checklist

**Files:** all changed files

**Step 1: Format**
```bash
make fmt
```

**Step 2: Lint**
```bash
make lint
```
Fix any issues.

**Step 3: Full test suite**
```bash
make test
```
Expected: all PASS.

**Step 4: Smoke-test with render-once**
```bash
go run . --render-once 2>/dev/null | head -60
```
Verify the app renders without panic.

**Step 5: Commit any fixes**
```bash
git add -A
git commit -m "chore: pre-completion fmt/lint/test fixes"
```

---

## Summary of new/changed files

| File | Change |
|------|--------|
| `internal/model/turn.go` | NEW — `model.Turn` struct |
| `internal/model/resource.go` | Add `ResourceSessionChat` |
| `internal/model/format.go` | Expose `FormatTokenCount` |
| `internal/ui/app.go` | DataProvider + AppModel fields + drillDown/navigateBack/isContentView/contentMaxOffset/View |
| `internal/ui/detail_render.go` | `RenderSessionChat` + helpers |
| `internal/ui/styles.go` | Bubble styles |
| `internal/ui/menu.go` | Nav hints |
| `cmd/root.go` | liveDataProvider + demoDataProvider `GetTurns` |
| `internal/ui/testhelpers_test.go` | mockDP `GetTurns` |
| `internal/ui/app_test.go` | Navigation tests |
| `internal/ui/detail_render_test.go` | Render tests |
| `internal/ui/menu_test.go` | Menu hint tests |
