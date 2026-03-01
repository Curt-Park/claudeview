# Content View Scrolling + Sub-View Key Blocking Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `plugin-item-detail` and `memory-detail` scrollable with j/k/G/g/ctrl+d/u, and block `p`/`m` jump keys in all plugin/memory sub-views.

**Architecture:** Add `ContentOffset int` to `AppModel`. In `updateTable()`, branch: content views route movement keys to `updateContentScroll()`; table views keep the existing `m.Table.Update()` path. `View()` slices rendered lines by offset. Block `p`/`m` when `isSubView()` is true. Reset ContentOffset on all navigation transitions.

**Tech Stack:** Go, Bubble Tea (`github.com/charmbracelet/bubbletea`), no new dependencies.

---

### Task 1: Block `p`/`m` in plugin/memory sub-views

**Files:**
- Modify: `internal/ui/app.go` (lines 161–168, `p`/`m` key handlers)
- Test: `internal/ui/app_test.go`

**Step 1: Write the failing tests**

Add to `internal/ui/app_test.go`:

```go
func TestPKeyBlockedInPluginDetail(t *testing.T) {
	p := &model.Plugin{Name: "myplugin", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{Cells: []string{"myplugin", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""}, Data: p}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // → plugin-detail

	app = updateApp(app, keyMsg("p"))

	if app.Resource != model.ResourcePluginDetail {
		t.Errorf("expected to stay on plugin-detail after p, got %s", app.Resource)
	}
}

func TestPKeyBlockedInPluginItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginDetail)
	app.Table.SetRows([]ui.Row{{Cells: []string{"skill", "my-skill"}, Data: pi}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // → plugin-item-detail

	app = updateApp(app, keyMsg("p"))

	if app.Resource != model.ResourcePluginItemDetail {
		t.Errorf("expected to stay on plugin-item-detail after p, got %s", app.Resource)
	}
}

func TestMKeyBlockedInMemoryDetail(t *testing.T) {
	mem := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.SelectedProjectHash = "proj-abc"
	app.Table.SetRows([]ui.Row{{Cells: []string{"MEMORY.md", "", "1 KB", "1h"}, Data: mem}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // → memory-detail

	app = updateApp(app, keyMsg("m"))

	if app.Resource != model.ResourceMemoryDetail {
		t.Errorf("expected to stay on memory-detail after m, got %s", app.Resource)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestPKeyBlockedIn|TestMKeyBlockedIn" -v
```

Expected: FAIL — `p` currently jumps away from sub-views.

**Step 3: Add `isSubView` helper and guard the keys**

In `internal/ui/app.go`, add after the `AppModel` struct definition (before `NewAppModel`):

```go
// isSubView returns true for views nested under plugins or memory
// (plugin-detail, plugin-item-detail, memory-detail).
// p/m jump keys are blocked in these views to preserve navigation context.
func isSubView(rt model.ResourceType) bool {
	return rt == model.ResourcePluginDetail ||
		rt == model.ResourcePluginItemDetail ||
		rt == model.ResourceMemoryDetail
}
```

Replace the `"p"` and `"m"` cases in `Update()`:

```go
case "p":
	if !isSubView(m.Resource) {
		m.jumpTo(model.ResourcePlugins)
	}
	return m, highlightCmd
case "m":
	if m.SelectedProjectHash != "" && !isSubView(m.Resource) {
		m.jumpTo(model.ResourceMemory)
	}
	return m, highlightCmd
```

**Step 4: Run tests to verify they pass**

```bash
go test ./internal/ui/... -run "TestPKeyBlockedIn|TestMKeyBlockedIn" -v
```

Expected: PASS

**Step 5: Run full test suite**

```bash
make test
```

Expected: all pass.

**Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: block p/m jump keys in plugin/memory sub-views"
```

---

### Task 2: Add `ContentOffset` field + `isContentView` helper

**Files:**
- Modify: `internal/ui/app.go`
- Test: `internal/ui/app_test.go`

**Step 1: Write a failing test**

Add to `internal/ui/app_test.go`:

```go
func TestContentOffsetFieldExists(t *testing.T) {
	app := newApp(model.ResourcePluginItemDetail)
	// Field must be accessible and zero-valued by default
	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset=0, got %d", app.ContentOffset)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/... -run "TestContentOffsetFieldExists" -v
```

Expected: FAIL — compile error, field doesn't exist.

**Step 3: Add `ContentOffset` to `AppModel` and `isContentView` helper**

In `internal/ui/app.go`, add `ContentOffset int` to `AppModel`:

```go
// AppModel is the top-level Bubble Tea model.
type AppModel struct {
	// Layout
	Width  int
	Height int

	// Chrome components
	Info   InfoModel
	Menu   MenuModel
	Crumbs CrumbsModel
	Flash  FlashModel
	Filter FilterModel

	// Content
	Resource      model.ResourceType
	Table         TableView
	ContentOffset int // scroll offset for content-only views

	// Navigation context (set on drill-down)
	SelectedProjectHash string
	SelectedSessionID   string
	SelectedAgentID     string
	SelectedPlugin      *model.Plugin
	SelectedPluginItem  *model.PluginItem
	SelectedMemory      *model.Memory

	// Data providers (injected from outside)
	DataProvider DataProvider

	// Animation tick counter
	tick int

	// Filter mode flag
	inFilter bool

	// filterStack saves parent-view filters across drill-downs
	filterStack []string

	// State saved before a t/p/m jump (for esc-to-restore)
	jumpFrom *jumpFromState
}
```

Also add below `isSubView`:

```go
// isContentView returns true for views that render flat text (not a table).
// These views use ContentOffset for scrolling instead of Table navigation.
func isContentView(rt model.ResourceType) bool {
	return rt == model.ResourcePluginItemDetail || rt == model.ResourceMemoryDetail
}
```

**Step 4: Run test**

```bash
go test ./internal/ui/... -run "TestContentOffsetFieldExists" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: add ContentOffset field and isContentView helper to AppModel"
```

---

### Task 3: Implement `updateContentScroll()` + wire into `updateTable()`

**Files:**
- Modify: `internal/ui/app.go`
- Test: `internal/ui/app_test.go`

**Step 1: Write failing tests**

Add to `internal/ui/app_test.go`:

```go
func TestScrollDownInPluginItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginItemDetail)
	app.Table.SetRows([]ui.Row{{Cells: []string{"skill", "my-skill"}, Data: pi}})
	app.SelectedPluginItem = pi

	app = updateApp(app, keyMsg("j"))

	if app.ContentOffset != 1 {
		t.Errorf("expected ContentOffset=1 after j, got %d", app.ContentOffset)
	}
}

func TestScrollUpFloorInPluginItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginItemDetail)
	app.SelectedPluginItem = pi
	app.ContentOffset = 0

	app = updateApp(app, keyMsg("k"))

	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset=0 (floor) after k at 0, got %d", app.ContentOffset)
	}
}

func TestScrollUpDecrementsInPluginItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginItemDetail)
	app.SelectedPluginItem = pi
	app.ContentOffset = 5

	app = updateApp(app, keyMsg("k"))

	if app.ContentOffset != 4 {
		t.Errorf("expected ContentOffset=4 after k from 5, got %d", app.ContentOffset)
	}
}

func TestScrollGotoTopInContentView(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.ContentOffset = 42

	app = updateApp(app, keyMsg("g"))

	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset=0 after g, got %d", app.ContentOffset)
	}
}

func TestScrollGotoBottomSetsLargeValue(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.ContentOffset = 0

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})

	if app.ContentOffset <= 100 {
		t.Errorf("expected large ContentOffset after G, got %d", app.ContentOffset)
	}
}

func TestScrollJDoesNotMoveTableInContentView(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.Table.SetRows([]ui.Row{
		{Cells: []string{"a"}, Data: nil},
		{Cells: []string{"b"}, Data: nil},
	})
	app.Table.Selected = 0

	app = updateApp(app, keyMsg("j"))

	if app.Table.Selected != 0 {
		t.Errorf("expected Table.Selected unchanged (0) in content view after j, got %d", app.Table.Selected)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestScroll" -v
```

Expected: FAIL — `ContentOffset` stays 0, `Table.Selected` moves.

**Step 3: Add `updateContentScroll()` and branch in `updateTable()`**

Add the new method to `internal/ui/app.go`:

```go
// updateContentScroll handles movement keys for content-only views (plugin-item-detail,
// memory-detail). It adjusts ContentOffset; View() caps it to actual content length.
func (m *AppModel) updateContentScroll(msg tea.KeyMsg) {
	half := m.contentHeight() / 2
	switch msg.String() {
	case "j":
		m.ContentOffset++
	case "k":
		if m.ContentOffset > 0 {
			m.ContentOffset--
		}
	case "G":
		m.ContentOffset = 1 << 30 // View() caps to actual max
	case "g":
		m.ContentOffset = 0
	case "ctrl+d":
		m.ContentOffset += half
	case "ctrl+u":
		if m.ContentOffset >= half {
			m.ContentOffset -= half
		} else {
			m.ContentOffset = 0
		}
	}
}
```

Replace the `default` branch in `updateTable()`:

```go
default:
	if isContentView(m.Resource) {
		m.updateContentScroll(msg)
	} else {
		m.Table.Update(msg)
	}
```

**Step 4: Run tests**

```bash
go test ./internal/ui/... -run "TestScroll" -v
```

Expected: PASS

**Step 5: Run full suite**

```bash
make test
```

Expected: all pass.

**Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: implement content view scrolling via ContentOffset"
```

---

### Task 4: Apply `ContentOffset` in `View()`

**Files:**
- Modify: `internal/ui/app.go` (`View()` method, lines 425–432)
- Test: `internal/ui/app_test.go`

**Step 1: Write a failing test**

The test creates a memory-detail with enough lines and verifies the rendered output starts at the offset line.

Add to `internal/ui/app_test.go`:

```go
func TestViewAppliesContentOffset(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.Width = termWidth
	app.Height = termHeight
	app.ContentOffset = 2

	// Build a memory with known multi-line content
	tmpFile := filepath.Join(t.TempDir(), "mem.md")
	content := "line0\nline1\nline2\nline3\nline4"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	app.SelectedMemory = &model.Memory{Path: tmpFile}

	rendered := app.View()

	if strings.Contains(rendered, "line0") {
		t.Error("expected line0 to be scrolled off, but it appears in output")
	}
	if !strings.Contains(rendered, "line2") {
		t.Error("expected line2 (at offset 2) to appear in output")
	}
}
```

Also add required imports to the test file if not already present:
```go
import (
    "os"
    "path/filepath"
    "strings"
    ...
)
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/ui/... -run "TestViewAppliesContentOffset" -v
```

Expected: FAIL — offset is ignored, line0 still appears.

**Step 3: Update `View()` to apply ContentOffset**

In `internal/ui/app.go`, update the content slicing block (currently lines 425–432):

```go
rawLines := strings.Split(strings.TrimRight(contentStr, "\n"), "\n")
limit := m.contentHeight()
// For content-only views, apply scroll offset (capped to actual max).
if isContentView(m.Resource) {
	maxOffset := len(rawLines) - limit
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := m.ContentOffset
	if offset > maxOffset {
		offset = maxOffset
	}
	if offset > 0 {
		rawLines = rawLines[offset:]
	}
}
if len(rawLines) > limit {
	rawLines = rawLines[:limit]
}
for len(rawLines) < limit {
	rawLines = append(rawLines, "")
}
```

**Step 4: Run test**

```bash
go test ./internal/ui/... -run "TestViewAppliesContentOffset" -v
```

Expected: PASS

**Step 5: Run full suite**

```bash
make test
```

Expected: all pass.

**Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: apply ContentOffset in View() for content view scrolling"
```

---

### Task 5: Reset `ContentOffset` on navigation transitions

**Files:**
- Modify: `internal/ui/app.go` (`drillInto`, `navigateBack`, `jumpTo`)
- Test: `internal/ui/app_test.go`

**Step 1: Write failing tests**

Add to `internal/ui/app_test.go`:

```go
func TestContentOffsetResetOnDrillInto(t *testing.T) {
	mem := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.Table.SetRows([]ui.Row{{Cells: []string{"MEMORY.md", "", "1 KB", "1h"}, Data: mem}})
	app.ContentOffset = 99 // simulate stale offset

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // drillInto memory-detail

	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset reset to 0 after drillInto, got %d", app.ContentOffset)
	}
}

func TestContentOffsetResetOnNavigateBack(t *testing.T) {
	mem := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.Table.SetRows([]ui.Row{{Cells: []string{"MEMORY.md", "", "1 KB", "1h"}, Data: mem}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // → memory-detail
	app.ContentOffset = 15

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc}) // navigateBack → memory

	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset reset to 0 after navigateBack, got %d", app.ContentOffset)
	}
}

func TestContentOffsetResetOnJumpTo(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.ContentOffset = 7

	app = updateApp(app, keyMsg("p")) // jumpTo plugins

	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset reset to 0 after jumpTo, got %d", app.ContentOffset)
	}
}
```

**Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/... -run "TestContentOffsetReset" -v
```

Expected: FAIL — ContentOffset carries over.

**Step 3: Add resets to navigation methods**

In `drillInto()`:
```go
func (m *AppModel) drillInto(rt model.ResourceType) {
	m.filterStack = append(m.filterStack, m.Table.Filter)
	m.Table.Filter = ""
	m.Resource = rt
	m.Crumbs.Push(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.ContentOffset = 0
	m.refreshMenu()
}
```

In `navigateBack()`, add `m.ContentOffset = 0` before the `return` in the jump-restore branch and at the top of the hierarchy branch:

```go
func (m *AppModel) navigateBack() {
	// Flat resources jumped to via t/p/m: restore previous state
	switch m.Resource {
	case model.ResourcePlugins, model.ResourceMemory:
		if m.jumpFrom != nil {
			m.Resource = m.jumpFrom.Resource
			m.SelectedProjectHash = m.jumpFrom.SelectedProjectHash
			m.SelectedSessionID = m.jumpFrom.SelectedSessionID
			m.SelectedAgentID = m.jumpFrom.SelectedAgentID
			m.Crumbs = m.jumpFrom.Crumbs
			m.Table.Filter = m.jumpFrom.Filter
			m.Filter.Input = m.jumpFrom.Filter
			m.filterStack = m.jumpFrom.FilterStack
			m.jumpFrom = nil
			m.ContentOffset = 0
			m.refreshMenu()
		} else {
			m.Resource = model.ResourceProjects
			m.SelectedProjectHash = ""
			m.SelectedSessionID = ""
			m.SelectedAgentID = ""
			m.Crumbs.Reset(string(model.ResourceProjects))
			m.ContentOffset = 0
			m.refreshMenu()
		}
		return
	}
	// Navigate up the resource hierarchy
	m.ContentOffset = 0
	switch m.Resource {
	case model.ResourceAgents:
		m.SelectedSessionID = ""
		m.popFilter()
		m.switchResource(model.ResourceSessions)
	case model.ResourceSessions:
		m.SelectedProjectHash = ""
		m.popFilter()
		m.switchResource(model.ResourceProjects)
	case model.ResourcePluginDetail:
		m.popFilter()
		m.switchResource(model.ResourcePlugins)
	case model.ResourcePluginItemDetail:
		m.popFilter()
		m.switchResource(model.ResourcePluginDetail)
	case model.ResourceMemoryDetail:
		m.popFilter()
		m.switchResource(model.ResourceMemory)
	}
}
```

In `jumpTo()`, add `m.ContentOffset = 0` after `m.Table.Filter = ""`:

```go
func (m *AppModel) jumpTo(rt model.ResourceType) {
	m.jumpFrom = &jumpFromState{
		Resource:            m.Resource,
		SelectedProjectHash: m.SelectedProjectHash,
		SelectedSessionID:   m.SelectedSessionID,
		SelectedAgentID:     m.SelectedAgentID,
		Crumbs:              m.Crumbs,
		Filter:              m.Table.Filter,
		FilterStack:         m.filterStack,
	}
	m.Resource = rt
	m.filterStack = nil
	m.refreshMenu()
	m.Crumbs.Reset(string(rt))
	m.Filter.Deactivate()
	m.Filter.Input = ""
	m.Table.Filter = ""
	m.ContentOffset = 0
}
```

**Step 4: Run tests**

```bash
go test ./internal/ui/... -run "TestContentOffsetReset" -v
```

Expected: PASS

**Step 5: Run full suite + lint**

```bash
make fmt && make lint && make test
```

Expected: 0 issues, all pass.

**Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: reset ContentOffset on all navigation transitions"
```

---

### Task 6: Final verification

**Step 1: Run the complete suite one more time**

```bash
make fmt && make lint && make test
```

Expected output:
```
0 issues.
ok  github.com/Curt-Park/claudeview/internal/model    ...
ok  github.com/Curt-Park/claudeview/internal/ui       ...
ok  github.com/Curt-Park/claudeview/internal/config   ...
ok  github.com/Curt-Park/claudeview/internal/transcript ...
```

**Step 2: Manual smoke test (optional)**

```bash
go run . --demo
```

Navigate: plugins → Enter (plugin-detail) → Enter on a skill → press j/k to scroll content → press p (should stay on plugin-item-detail) → Esc back.

Navigate: project → memories → Enter on a memory → press j/k to scroll → press m (should stay on memory-detail) → Esc back.
