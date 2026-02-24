# Regression Tests Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `AppModel` integration + golden-snapshot regression tests covering the 11 UI bugs fixed in PR #3, plus a `--render-once` debug flag for visual inspection.

**Architecture:** All tests live in `internal/ui/` as `package ui_test`. They drive `AppModel.Update()` directly (no tea.Program needed — Update is a pure function). Golden snapshots use `AppModel.View()` saved to `testdata/`. A `--render-once` flag in `cmd/root.go` lets Claude dump one rendered frame to stdout for visual debugging.

**Tech Stack:** Go stdlib `testing`, `github.com/charmbracelet/bubbletea` KeyMsg types, existing `internal/model` and `internal/ui` packages.

---

### Task 1: Test helpers (mock provider + key sender)

**Files:**
- Create: `internal/ui/app_test_helpers_test.go`

**Step 1: Write the file**

```go
package ui_test

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// mockDP satisfies ui.DataProvider with empty stubs.
type mockDP struct{}

func (m *mockDP) GetProjects() any              { return []*model.Project{} }
func (m *mockDP) GetSessions(string) any        { return []*model.Session{} }
func (m *mockDP) GetAgents(string) any          { return []*model.Agent{} }
func (m *mockDP) GetTools(string) any           { return []*model.ToolCall{} }
func (m *mockDP) GetTasks(string) any           { return []*model.Task{} }
func (m *mockDP) GetPlugins() any               { return []*model.Plugin{} }
func (m *mockDP) GetMCPServers() any            { return []*model.MCPServer{} }
func (m *mockDP) CurrentProject() string        { return "" }
func (m *mockDP) CurrentSession() string        { return "" }
func (m *mockDP) CurrentAgent() string          { return "" }

// makeModel creates an AppModel pre-populated with rows, sized 120×40.
func makeModel(resource model.ResourceType, rows []ui.Row) ui.AppModel {
	app := ui.NewAppModel(&mockDP{}, resource)
	app.Width = 120
	app.Height = 40
	app.Table.SetRows(rows)
	return app
}

// key sends a rune key to the model and returns the updated AppModel.
func key(m ui.AppModel, k string) ui.AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	return next.(ui.AppModel)
}

// specialKey sends a special key (Enter, Esc, etc.) and returns updated AppModel.
func specialKey(m ui.AppModel, t tea.KeyType) ui.AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: t})
	return next.(ui.AppModel)
}

// ctrlKey sends a ctrl+<letter> key.
func ctrlKey(m ui.AppModel, k string) ui.AppModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	_ = k
	return next.(ui.AppModel)
}

// projectRow builds a Row whose Data is a *model.Project.
func projectRow(hash string) ui.Row {
	return ui.Row{
		Cells: []string{hash, "1", "0", "1d"},
		Data:  &model.Project{Hash: hash},
	}
}

// sessionRow builds a Row whose Data is a *model.Session.
func sessionRow(id string) ui.Row {
	return ui.Row{
		Cells: []string{id[:8], "claude-3", "active", "1", "5", "1.2k", "$0.01", "5m"},
		Data:  &model.Session{ID: id},
	}
}

// agentRow builds a Row whose Data is a *model.Agent.
func agentRow(id string) ui.Row {
	return ui.Row{
		Cells: []string{id, "main", "done", "3", "read file"},
		Data:  &model.Agent{ID: id},
	}
}

// nRows returns n generic rows (no meaningful Data).
func nRows(n int) []ui.Row {
	rows := make([]ui.Row, n)
	for i := range rows {
		rows[i] = ui.Row{Cells: []string{fmt.Sprintf("row-%d", i)}}
	}
	return rows
}
```

Note: add `"fmt"` to the import block.

**Step 2: Build to confirm no syntax errors**

```bash
go build ./internal/ui/...
```

Expected: no output (success).

**Step 3: Commit**

```bash
git add internal/ui/app_test_helpers_test.go
git commit -m "test: add AppModel test helpers (mockDP, key senders, row builders)"
```

---

### Task 2: j/k navigation regression (Bug 1)

**Files:**
- Create: `internal/ui/app_navigation_test.go`

**Step 1: Write tests**

```go
package ui_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

// TestJKNavigation — regression for Bug 1: cursor resets on every key press.
func TestJKNavigation(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(5))

	// j moves down
	m = key(m, "j")
	if m.Table.Selected != 1 {
		t.Errorf("after j: Selected=%d, want 1", m.Table.Selected)
	}
	m = key(m, "j")
	m = key(m, "j")
	if m.Table.Selected != 3 {
		t.Errorf("after 3×j: Selected=%d, want 3", m.Table.Selected)
	}

	// k moves up
	m = key(m, "k")
	if m.Table.Selected != 2 {
		t.Errorf("after k: Selected=%d, want 2", m.Table.Selected)
	}
}

func TestJKDoesNotGoOutOfBounds(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))

	// k at top — stays at 0
	m = key(m, "k")
	if m.Table.Selected != 0 {
		t.Errorf("k at top: Selected=%d, want 0", m.Table.Selected)
	}

	// j past bottom — stays at len-1
	m = key(m, "j")
	m = key(m, "j")
	m = key(m, "j")
	m = key(m, "j")
	if m.Table.Selected != 2 {
		t.Errorf("j past bottom: Selected=%d, want 2", m.Table.Selected)
	}
}
```

**Step 2: Run tests**

```bash
go test -race ./internal/ui/... -run TestJK -v
```

Expected: both tests PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_navigation_test.go
git commit -m "test: regression for j/k navigation (Bug 1)"
```

---

### Task 3: g/G and page navigation (Bug 9)

**Files:**
- Modify: `internal/ui/app_navigation_test.go` (append)

**Step 1: Append tests**

```go
func TestGotoTopBottom(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(10))

	// Move to middle first
	for i := 0; i < 5; i++ {
		m = key(m, "j")
	}
	if m.Table.Selected != 5 {
		t.Fatalf("setup: Selected=%d, want 5", m.Table.Selected)
	}

	// G goes to bottom
	m = key(m, "G")
	if m.Table.Selected != 9 {
		t.Errorf("G: Selected=%d, want 9", m.Table.Selected)
	}

	// g goes to top
	m = key(m, "g")
	if m.Table.Selected != 0 {
		t.Errorf("g: Selected=%d, want 0", m.Table.Selected)
	}
}

func TestPageUpDown(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(20))
	m.Table.Height = 10

	// ctrl+d moves down by ~half page
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = next.(ui.AppModel)
	if m.Table.Selected < 4 {
		t.Errorf("ctrl+d: Selected=%d, want >=4", m.Table.Selected)
	}

	// ctrl+u moves back up
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = next.(ui.AppModel)
	if m.Table.Selected > 1 {
		t.Errorf("ctrl+u: Selected=%d, want <=1", m.Table.Selected)
	}
}
```

Note: add `tea "github.com/charmbracelet/bubbletea"` and `"github.com/Curt-Park/claudeview/internal/ui"` to imports in the file.

**Step 2: Run**

```bash
go test -race ./internal/ui/... -run "TestGoto|TestPage" -v
```

Expected: PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_navigation_test.go
git commit -m "test: regression for g/G and ctrl+u/d page navigation (Bug 9)"
```

---

### Task 4: Drill-down context regression (Bug 4)

**Files:**
- Create: `internal/ui/app_drilldown_test.go`

**Step 1: Write tests**

```go
package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestDrillDownSetsProjectHash — regression for Bug 4.
func TestDrillDownSetsProjectHash(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{
		projectRow("proj-abc123"),
		projectRow("proj-def456"),
	})

	// Select second row then Enter
	m = key(m, "j")
	m = specialKey(m, tea.KeyEnter)

	if m.SelectedProjectHash != "proj-def456" {
		t.Errorf("SelectedProjectHash=%q, want %q", m.SelectedProjectHash, "proj-def456")
	}
	if m.Resource != model.ResourceSessions {
		t.Errorf("Resource=%q after drill-down, want sessions", m.Resource)
	}
}

func TestDrillDownSetsSessionID(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{
		sessionRow("aaaaaaaa-1111-2222-3333-444444444444"),
		sessionRow("bbbbbbbb-1111-2222-3333-444444444444"),
	})

	m = key(m, "j") // select second
	m = specialKey(m, tea.KeyEnter)

	want := "bbbbbbbb-1111-2222-3333-444444444444"
	if m.SelectedSessionID != want {
		t.Errorf("SelectedSessionID=%q, want %q", m.SelectedSessionID, want)
	}
	if m.Resource != model.ResourceAgents {
		t.Errorf("Resource=%q, want agents", m.Resource)
	}
}

func TestDrillDownSetsAgentID(t *testing.T) {
	m := makeModel(model.ResourceAgents, []ui.Row{
		agentRow("agent-xyz"),
	})

	m = specialKey(m, tea.KeyEnter)

	if m.SelectedAgentID != "agent-xyz" {
		t.Errorf("SelectedAgentID=%q, want %q", m.SelectedAgentID, "agent-xyz")
	}
	if m.Resource != model.ResourceTools {
		t.Errorf("Resource=%q, want tools", m.Resource)
	}
}

func TestDrillDownOnEmptyTableIsNoop(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{})
	before := m.Resource

	m = specialKey(m, tea.KeyEnter)

	if m.Resource != before {
		t.Errorf("drill-down on empty table changed resource to %q", m.Resource)
	}
}
```

**Step 2: Run**

```bash
go test -race ./internal/ui/... -run TestDrillDown -v
```

Expected: all PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_drilldown_test.go
git commit -m "test: regression for drill-down context passing (Bug 4)"
```

---

### Task 5: Breadcrumb Push/Pop regression (Bug 6)

**Files:**
- Create: `internal/ui/app_breadcrumb_test.go`

**Step 1: Write tests**

```go
package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestBreadcrumbPushOnDrillDown — regression for Bug 6.
func TestBreadcrumbPushOnDrillDown(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{projectRow("p1")})

	if len(m.Crumbs.Items) != 1 {
		t.Fatalf("initial crumbs=%v, want 1 item", m.Crumbs.Items)
	}

	m = specialKey(m, tea.KeyEnter)
	if len(m.Crumbs.Items) != 2 {
		t.Errorf("after drill-down crumbs=%v, want 2 items", m.Crumbs.Items)
	}
	if m.Crumbs.Items[1] != string(model.ResourceSessions) {
		t.Errorf("crumbs[1]=%q, want %q", m.Crumbs.Items[1], model.ResourceSessions)
	}

	// drill-down again
	m.Table.SetRows([]ui.Row{sessionRow("aaaaaaaa-0000-0000-0000-000000000000")})
	m = specialKey(m, tea.KeyEnter)
	if len(m.Crumbs.Items) != 3 {
		t.Errorf("after 2nd drill-down crumbs=%v, want 3 items", m.Crumbs.Items)
	}
}

func TestBreadcrumbPopOnBack(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{projectRow("p1")})
	m = specialKey(m, tea.KeyEnter) // drill into sessions (2 crumbs)

	m = specialKey(m, tea.KeyEsc)
	if len(m.Crumbs.Items) != 1 {
		t.Errorf("after Esc crumbs=%v, want 1 item", m.Crumbs.Items)
	}
	if m.Resource != model.ResourceProjects {
		t.Errorf("Resource=%q after back, want projects", m.Resource)
	}
}

func TestCommandSwitchResetsCrumbs(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{projectRow("p1")})
	m = specialKey(m, tea.KeyEnter) // 2 crumbs

	// :sessions command resets to single crumb
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	m = next.(ui.AppModel)
	for _, ch := range "sessions" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}
	m = specialKey(m, tea.KeyEnter)

	if len(m.Crumbs.Items) != 1 {
		t.Errorf("after :sessions crumbs=%v, want 1 item", m.Crumbs.Items)
	}
}
```

**Step 2: Run**

```bash
go test -race ./internal/ui/... -run TestBreadcrumb -v
go test -race ./internal/ui/... -run TestCommandSwitch -v
```

Expected: all PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_breadcrumb_test.go
git commit -m "test: regression for breadcrumb Push/Pop/Reset (Bug 6)"
```

---

### Task 6: Navigate-back clears context (Bug 4 extension)

**Files:**
- Modify: `internal/ui/app_drilldown_test.go` (append)

**Step 1: Append tests**

```go
func TestNavigateBackClearsContext(t *testing.T) {
	// Start at tools, simulate having come from agents
	m := makeModel(model.ResourceTools, []ui.Row{})
	m.SelectedAgentID = "agent-xyz"
	m.Crumbs.Items = []string{"projects", "sessions", "agents", "tools"}

	m = specialKey(m, tea.KeyEsc)

	if m.Resource != model.ResourceAgents {
		t.Errorf("Resource=%q, want agents", m.Resource)
	}
	if m.SelectedAgentID != "" {
		t.Errorf("SelectedAgentID=%q, want empty after back", m.SelectedAgentID)
	}
}

func TestNavigateBackFromSessionsClearsProjectHash(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{})
	m.SelectedProjectHash = "proj-abc"
	m.Crumbs.Items = []string{"projects", "sessions"}

	m = specialKey(m, tea.KeyEsc)

	if m.Resource != model.ResourceProjects {
		t.Errorf("Resource=%q, want projects", m.Resource)
	}
	if m.SelectedProjectHash != "" {
		t.Errorf("SelectedProjectHash=%q, want empty after back", m.SelectedProjectHash)
	}
}
```

**Step 2: Run**

```bash
go test -race ./internal/ui/... -run TestNavigateBack -v
```

Expected: PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_drilldown_test.go
git commit -m "test: regression for navigateBack context clearing (Bug 4 ext)"
```

---

### Task 7: Filter wiring regression (Bug 5)

**Files:**
- Create: `internal/ui/app_filter_test.go`

**Step 1: Write tests**

```go
package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

// TestFilterWiredToTable — regression for Bug 5.
func TestFilterWiredToTable(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(5))

	// Enter filter mode with /
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(ui.AppModel)

	// Type "foo"
	for _, ch := range "foo" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}

	if m.Table.Filter != "foo" {
		t.Errorf("Table.Filter=%q, want %q", m.Table.Filter, "foo")
	}
	if m.Log.Filter != "foo" {
		t.Errorf("Log.Filter=%q, want %q", m.Log.Filter, "foo")
	}
}

func TestFilterClearedOnEsc(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(5))

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(ui.AppModel)
	for _, ch := range "bar" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}

	// Esc clears filter
	m = specialKey(m, tea.KeyEsc)

	if m.Table.Filter != "" {
		t.Errorf("Table.Filter=%q after Esc, want empty", m.Table.Filter)
	}
}

func TestFilterBackspace(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m = next.(ui.AppModel)
	for _, ch := range "abc" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		m = next.(ui.AppModel)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(ui.AppModel)

	if m.Table.Filter != "ab" {
		t.Errorf("Table.Filter=%q after backspace, want %q", m.Table.Filter, "ab")
	}
}
```

**Step 2: Run**

```bash
go test -race ./internal/ui/... -run TestFilter -v
```

Expected: all PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_filter_test.go
git commit -m "test: regression for filter wiring to table/log (Bug 5)"
```

---

### Task 8: ViewMode transitions regression (Bugs 8 + new features)

**Files:**
- Create: `internal/ui/app_viewmode_test.go`

**Step 1: Write tests**

```go
package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestDKeyEntersModeDetail(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "d")
	if m.ViewMode != ui.ModeDetail {
		t.Errorf("ViewMode=%v after d, want ModeDetail", m.ViewMode)
	}
}

func TestLKeyEntersModeLog(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "l")
	if m.ViewMode != ui.ModeLog {
		t.Errorf("ViewMode=%v after l, want ModeLog", m.ViewMode)
	}
}

func TestYKeyEntersModeYAML(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "y")
	if m.ViewMode != ui.ModeYAML {
		t.Errorf("ViewMode=%v after y, want ModeYAML", m.ViewMode)
	}
}

func TestEscReturnsModeTable(t *testing.T) {
	for _, startMode := range []struct {
		name   string
		setup  func(ui.AppModel) ui.AppModel
	}{
		{"detail", func(m ui.AppModel) ui.AppModel { return key(m, "d") }},
		{"log",    func(m ui.AppModel) ui.AppModel { return key(m, "l") }},
		{"yaml",   func(m ui.AppModel) ui.AppModel { return key(m, "y") }},
	} {
		t.Run(startMode.name, func(t *testing.T) {
			m := startMode.setup(makeModel(model.ResourceSessions, nRows(3)))
			m = specialKey(m, tea.KeyEsc)
			if m.ViewMode != ui.ModeTable {
				t.Errorf("ViewMode=%v after Esc, want ModeTable", m.ViewMode)
			}
		})
	}
}

// TestDetailViewHLNotVerticalScroll — regression for Bug 8.
// h and l in detail view must NOT scroll (they're removed).
func TestDetailViewHLNotVerticalScroll(t *testing.T) {
	m := makeModel(model.ResourceSessions, nRows(3))
	m = key(m, "d") // enter detail mode

	// Pre-scroll down so offset > 0
	m.Detail.Lines = make([]string, 50)
	for i := range m.Detail.Lines {
		m.Detail.Lines[i] = "line"
	}
	m.Detail.Offset = 5

	before := m.Detail.Offset
	m = key(m, "h")
	if m.Detail.Offset != before {
		t.Errorf("h in detail changed Offset %d→%d (should be no-op)", before, m.Detail.Offset)
	}
	m = key(m, "l")
	if m.Detail.Offset != before {
		t.Errorf("l in detail changed Offset %d→%d (should be no-op)", before, m.Detail.Offset)
	}
}
```

**Step 2: Run**

```bash
go test -race ./internal/ui/... -run "TestDKey|TestLKey|TestYKey|TestEsc|TestDetail" -v
```

Expected: all PASS.

**Step 3: Commit**

```bash
git add internal/ui/app_viewmode_test.go
git commit -m "test: regression for view mode transitions and h/l detail fix (Bug 8)"
```

---

### Task 9: Golden snapshot tests (visual regression)

**Files:**
- Create: `internal/ui/testdata/` (directory)
- Create: `internal/ui/app_snapshot_test.go`

**Step 1: Write snapshot test**

```go
package ui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func init() {
	// Disable colour so snapshots are stable across terminals.
	lipgloss.SetColorProfile(termenv.Ascii)
}

// snapshotPath returns the golden file path for a test name.
func snapshotPath(name string) string {
	return filepath.Join("testdata", name+".txt")
}

// assertSnapshot compares View() output to a golden file.
// Set UPDATE_SNAPSHOTS=1 to regenerate golden files.
func assertSnapshot(t *testing.T, name string, m ui.AppModel) {
	t.Helper()
	got := stripAnsi(m.View())
	path := snapshotPath(name)

	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0644); err != nil {
			t.Fatalf("write snapshot %s: %v", path, err)
		}
		t.Logf("updated snapshot: %s", path)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("snapshot %s missing — run with UPDATE_SNAPSHOTS=1 to create it", path)
	}
	if got != string(want) {
		t.Errorf("snapshot %s mismatch:\ngot:\n%s\nwant:\n%s", name, got, string(want))
	}
}

// stripAnsi removes ANSI escape sequences for stable text comparison.
func stripAnsi(s string) string {
	// Use lipgloss.Width trick: render through plain profile already done in init.
	// For extra safety, strip any remaining ESC sequences.
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func TestSnapshotSessionsView(t *testing.T) {
	m := makeModel(model.ResourceSessions, []ui.Row{
		sessionRow("aaaaaaaa-0000-0000-0000-000000000000"),
		sessionRow("bbbbbbbb-0000-0000-0000-000000000000"),
	})
	m.Width = 120
	m.Height = 30
	assertSnapshot(t, "screen_sessions", m)
}

func TestSnapshotProjectsView(t *testing.T) {
	m := makeModel(model.ResourceProjects, []ui.Row{
		projectRow("Users-alice-my-awesome-project"),
		projectRow("Users-alice-another-project"),
	})
	m.Width = 120
	m.Height = 30
	assertSnapshot(t, "screen_projects", m)
}
```

Note: `init()` uses `termenv` — add import `"github.com/muesli/termenv"` (already in go.mod as indirect dep via lipgloss).

**Step 2: Generate golden files**

```bash
UPDATE_SNAPSHOTS=1 go test ./internal/ui/... -run TestSnapshot -v
```

Expected output:
```
--- LOG: updated snapshot: testdata/screen_sessions.txt
--- LOG: updated snapshot: testdata/screen_projects.txt
PASS
```

**Step 3: Verify snapshots pass without flag**

```bash
go test -race ./internal/ui/... -run TestSnapshot -v
```

Expected: PASS.

**Step 4: Commit**

```bash
git add internal/ui/app_snapshot_test.go internal/ui/testdata/
git commit -m "test: golden snapshot tests for sessions and projects views"
```

---

### Task 10: --render-once flag for visual debugging

**Files:**
- Modify: `cmd/root.go`

**Step 1: Add renderOnce flag**

In `cmd/root.go`, add to `init()`:
```go
rootCmd.Flags().BoolVar(&renderOnce, "render-once", false, "Render one frame to stdout and exit (for debugging)")
```

Add variable declaration near other vars:
```go
var renderOnce bool
```

In the `run()` function, after `root := newRootModel(appModel, dp)`, add:
```go
if renderOnce {
    // WindowSizeMsg to trigger layout
    root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
    fmt.Print(root.View())
    return nil
}
```

**Step 2: Test the flag**

```bash
go run . --demo --render-once
```

Expected: prints one frame of the TUI (with ANSI codes) and exits immediately. Claude can pipe this to a file and read it.

```bash
go run . --demo --render-once | cat > /tmp/screen.txt
# Claude reads /tmp/screen.txt
```

**Step 3: Build check**

```bash
go build ./...
go vet ./...
```

Expected: clean.

**Step 4: Commit**

```bash
git add cmd/root.go
git commit -m "feat: add --render-once flag for headless UI frame debugging"
```

---

### Task 11: Full verification and PR update

**Step 1: Run complete test suite**

```bash
go fmt ./...
go vet ./...
go test -race ./...
```

Expected: all packages pass, no race conditions.

**Step 2: Run lint**

```bash
mise exec -- golangci-lint run ./...
```

Expected: `0 issues`.

**Step 3: Verify render-once demo**

```bash
go run . --demo --render-once
```

Expected: TUI frame printed to terminal without hanging.

**Step 4: Push to existing PR branch**

```bash
git push origin fix/ui-navigation-and-bugs
```
