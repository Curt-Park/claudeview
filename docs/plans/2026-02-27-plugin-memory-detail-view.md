# Plugin & Memory Detail View Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Pressing enter on a plugin or memory row opens a full-screen detail view; esc returns to the list.

**Architecture:** Add two new `ResourceType` constants (`plugin-detail`, `memory-detail`) and reuse the existing `drillDown`/`drillInto`/`navigateBack` navigation chain. Store the selected `*model.Plugin` or `*model.Memory` on `AppModel`. `View()` switches rendering based on resource type.

**Tech Stack:** Go, Bubble Tea, Lip Gloss — no new dependencies.

---

### Task 1: Add ResourceType constants

**Files:**
- Modify: `internal/model/resource.go`

**Step 1: Add the two new constants**

In `internal/model/resource.go`, add after `ResourceMemory`:

```go
ResourcePluginDetail ResourceType = "plugin-detail"
ResourceMemoryDetail ResourceType = "memory-detail"
```

**Step 2: Build to confirm no errors**

```bash
make build
```
Expected: success, no errors.

**Step 3: Commit**

```bash
git add internal/model/resource.go
git commit -m "feat: add ResourcePluginDetail and ResourceMemoryDetail types"
```

---

### Task 2: Add List* functions to model/plugin.go

**Files:**
- Modify: `internal/model/plugin.go`
- Test: `internal/model/plugin_test.go`

These functions return item names (filenames without extension, or directory names) for each plugin content category. They mirror the existing `Count*` functions but return `[]string` instead of `int`.

**Step 1: Write failing tests**

Add to `internal/model/plugin_test.go`:

```go
func TestListSkills(t *testing.T) {
	base := makeTempDir(t)
	dir := filepath.Join(base, "skills")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "brainstorming"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "debugging"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "README.md")) // should be ignored

	got := model.ListSkills(base)
	if len(got) != 2 {
		t.Fatalf("ListSkills() = %v, want 2 items", got)
	}
}

func TestListCommands(t *testing.T) {
	base := makeTempDir(t)
	dir := filepath.Join(base, "commands")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "commit.md"))
	writeFile(t, filepath.Join(dir, "review-pr.md"))

	got := model.ListCommands(base)
	if len(got) != 2 {
		t.Fatalf("ListCommands() = %v, want 2 items", got)
	}
	// names should have .md stripped
	for _, name := range got {
		if filepath.Ext(name) == ".md" {
			t.Errorf("expected extension stripped, got %q", name)
		}
	}
}

func TestListSkillsMissingDir(t *testing.T) {
	got := model.ListSkills("/nonexistent/path/xyz")
	if got != nil {
		t.Errorf("expected nil for missing dir, got %v", got)
	}
}
```

**Step 2: Run tests to confirm they fail**

```bash
go test ./internal/model/... -run "TestListSkills|TestListCommands" -v
```
Expected: FAIL — `model.ListSkills undefined`.

**Step 3: Implement List* functions in `internal/model/plugin.go`**

Add after the `Count*` functions:

```go
// ListSkills returns the names of skill subdirectories.
func ListSkills(cacheDir string) []string {
	return listDirNames(filepath.Join(contentDir(cacheDir), "skills"))
}

// ListCommands returns command names (filenames without .md extension).
func ListCommands(cacheDir string) []string {
	return listFileStems(filepath.Join(contentDir(cacheDir), "commands"), ".md")
}

// ListHooks returns hook event names from hooks.json or filenames.
func ListHooks(cacheDir string) []string {
	hooksDir := filepath.Join(contentDir(cacheDir), "hooks")
	jsonPath := filepath.Join(hooksDir, "hooks.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		var wrapper struct {
			Hooks map[string]json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(data, &wrapper); err == nil {
			names := make([]string, 0, len(wrapper.Hooks))
			for k := range wrapper.Hooks {
				names = append(names, k)
			}
			sort.Strings(names)
			return names
		}
	}
	return listFileStems(hooksDir, "")
}

// ListAgents returns agent names (filenames without .md extension).
func ListAgents(cacheDir string) []string {
	return listFileStems(filepath.Join(contentDir(cacheDir), "agents"), ".md")
}

// ListMCPs returns MCP server names from .mcp.json or plugin.json.
func ListMCPs(cacheDir string) []string {
	type mcpWrapper struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	candidates := []string{
		filepath.Join(contentDir(cacheDir), ".mcp.json"),
		filepath.Join(cacheDir, ".claude-plugin", "plugin.json"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var w mcpWrapper
		if err := json.Unmarshal(data, &w); err == nil && len(w.MCPServers) > 0 {
			names := make([]string, 0, len(w.MCPServers))
			for k := range w.MCPServers {
				names = append(names, k)
			}
			sort.Strings(names)
			return names
		}
	}
	return nil
}

// listDirNames returns names of subdirectories in dir.
func listDirNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// listFileStems returns filenames in dir with the given extension stripped.
// If ext is empty, all filenames are returned as-is.
func listFileStems(dir, ext string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if ext != "" {
			if filepath.Ext(name) != ext {
				continue
			}
			name = strings.TrimSuffix(name, ext)
		}
		names = append(names, name)
	}
	return names
}
```

Also add `"sort"` and `"strings"` to the import block if not already present.

**Step 4: Run tests to confirm they pass**

```bash
go test ./internal/model/... -run "TestListSkills|TestListCommands" -v
```
Expected: PASS.

**Step 5: Run full test suite**

```bash
make test
```
Expected: all pass.

**Step 6: Commit**

```bash
git add internal/model/plugin.go internal/model/plugin_test.go
git commit -m "feat: add List* functions to model/plugin for detail view content"
```

---

### Task 3: Add RenderPluginDetail to internal/view/plugins.go

**Files:**
- Modify: `internal/view/plugins.go`

This function takes a `*model.Plugin` and returns a plain-text string rendering of the plugin's content sections. It does NOT interact with Bubble Tea — it is a pure string builder.

**Step 1: Implement `RenderPluginDetail`**

Add to `internal/view/plugins.go`:

```go
// RenderPluginDetail renders the detail content for a plugin as a plain string.
// Each non-empty category is shown as a labeled section.
func RenderPluginDetail(p *model.Plugin, width int) string {
	if p == nil {
		return ""
	}
	header := ui.StyleTitle.Render(p.Name) + "  " +
		ui.StyleDim.Render(p.Version) + "  " +
		ui.StyleDim.Render(p.Scope)

	sections := []struct {
		label string
		items []string
	}{
		{"Skills", model.ListSkills(p.CacheDir)},
		{"Commands", model.ListCommands(p.CacheDir)},
		{"Hooks", model.ListHooks(p.CacheDir)},
		{"Agents", model.ListAgents(p.CacheDir)},
		{"MCPs", model.ListMCPs(p.CacheDir)},
	}

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	for _, s := range sections {
		if len(s.items) == 0 {
			continue
		}
		label := ui.StyleHeader.Render(s.label + ":")
		sb.WriteString(label + "\n")
		for _, item := range s.items {
			sb.WriteString("  " + item + "\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
```

Note: `ui.StyleHeader` and `ui.StyleDim` — check `internal/ui/styles.go` for existing style names and use matching ones (e.g. `StyleTitle`, `StyleDone`, `StyleRunning`). Add a `StyleDim` if it doesn't exist, or reuse the closest equivalent.

**Step 2: Build**

```bash
make build
```
Expected: success.

**Step 3: Commit**

```bash
git add internal/view/plugins.go
git commit -m "feat: add RenderPluginDetail for plugin content display"
```

---

### Task 4: Add RenderMemoryDetail to internal/view/memories.go

**Files:**
- Modify: `internal/view/memories.go`

**Step 1: Implement `RenderMemoryDetail`**

Add to `internal/view/memories.go`:

```go
// RenderMemoryDetail reads and returns the raw content of a memory file.
func RenderMemoryDetail(m *model.Memory) string {
	if m == nil {
		return ""
	}
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}
	return string(data)
}
```

Add `"fmt"` and `"os"` to the import block.

**Step 2: Build**

```bash
make build
```
Expected: success.

**Step 3: Commit**

```bash
git add internal/view/memories.go
git commit -m "feat: add RenderMemoryDetail for memory file content display"
```

---

### Task 5: Extend AppModel — fields, drillDown, navigateBack

**Files:**
- Modify: `internal/ui/app.go`
- Test: `internal/ui/app_test.go`

**Step 1: Write failing tests**

Add to `internal/ui/app_test.go`:

```go
func TestDrilldownPluginToDetail(t *testing.T) {
	p := &model.Plugin{Name: "superpowers", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""}, Data: p}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourcePluginDetail {
		t.Errorf("expected resource=plugin-detail after Enter on plugin, got %s", app.Resource)
	}
	if app.SelectedPlugin != p {
		t.Errorf("expected SelectedPlugin set after Enter")
	}
}

func TestEscFromPluginDetailReturnsToPlugins(t *testing.T) {
	p := &model.Plugin{Name: "superpowers", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""}, Data: p}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // go to detail

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourcePlugins {
		t.Errorf("expected resource=plugins after Esc from plugin-detail, got %s", app.Resource)
	}
}

func TestDrilldownMemoryToDetail(t *testing.T) {
	m := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.Table.SetRows([]ui.Row{{Cells: []string{"MEMORY.md", "", "1 KB", "1h"}, Data: m}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceMemoryDetail {
		t.Errorf("expected resource=memory-detail after Enter on memory, got %s", app.Resource)
	}
	if app.SelectedMemory != m {
		t.Errorf("expected SelectedMemory set after Enter")
	}
}

func TestEscFromMemoryDetailReturnsToMemories(t *testing.T) {
	m := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.Table.SetRows([]ui.Row{{Cells: []string{"MEMORY.md", "", "1 KB", "1h"}, Data: m}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // go to detail

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourceMemory {
		t.Errorf("expected resource=memories after Esc from memory-detail, got %s", app.Resource)
	}
}
```

**Step 2: Run tests to confirm they fail**

```bash
go test ./internal/ui/... -run "TestDrilldownPlugin|TestEscFromPlugin|TestDrilldownMemory|TestEscFromMemory" -v
```
Expected: FAIL — fields and cases don't exist yet.

**Step 3: Add fields to AppModel in `internal/ui/app.go`**

In the `AppModel` struct, add after `SelectedAgentID`:

```go
SelectedPlugin *model.Plugin
SelectedMemory *model.Memory
```

**Step 4: Extend drillDown()**

In `drillDown()`, add cases after the existing `ResourceSessions` case:

```go
case model.ResourcePlugins:
    if p, ok := row.Data.(*model.Plugin); ok {
        m.SelectedPlugin = p
    }
    m.drillInto(model.ResourcePluginDetail)
case model.ResourceMemory:
    if mem, ok := row.Data.(*model.Memory); ok {
        m.SelectedMemory = mem
    }
    m.drillInto(model.ResourceMemoryDetail)
```

**Step 5: Extend navigateBack()**

In `navigateBack()`, add cases for the two new detail types. Look at the existing switch — find where `ResourcePlugins` and `ResourceMemory` handle the `jumpFrom` restore, and add:

```go
case model.ResourcePluginDetail:
    m.switchResource(model.ResourcePlugins)
case model.ResourceMemoryDetail:
    m.switchResource(model.ResourceMemory)
```

**Step 6: Run tests to confirm they pass**

```bash
go test ./internal/ui/... -run "TestDrilldownPlugin|TestEscFromPlugin|TestDrilldownMemory|TestEscFromMemory" -v
```
Expected: PASS.

**Step 7: Run full test suite**

```bash
make test
```
Expected: all pass.

**Step 8: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "feat: add plugin/memory detail drilldown and esc navigation"
```

---

### Task 6: Update View() to render detail content

**Files:**
- Modify: `internal/ui/app.go`

The current `View()` always calls `m.Table.View()` for content. We need to branch on the resource type and pass the content lines through the same table-height limiting code.

**Step 1: Replace the content rendering block in `View()`**

Find this in `View()`:
```go
// --- 3. Content ---
contentStr := m.Table.View()
```

Replace with:

```go
// --- 3. Content ---
var contentStr string
switch m.Resource {
case model.ResourcePluginDetail:
    contentStr = view.RenderPluginDetail(m.SelectedPlugin, m.contentWidth())
case model.ResourceMemoryDetail:
    contentStr = view.RenderMemoryDetail(m.SelectedMemory)
default:
    contentStr = m.Table.View()
}
```

Add the `view` package import: `"github.com/Curt-Park/claudeview/internal/view"`.

**Step 2: Build**

```bash
make build
```
Expected: success.

**Step 3: Smoke-test manually**

```bash
./bin/claudeview --demo
```

Navigate to plugins with `p`, select an item, press enter. Confirm a detail view appears. Press esc to return.

**Step 4: Run full test suite**

```bash
make test
```
Expected: all pass.

**Step 5: Commit**

```bash
git add internal/ui/app.go
git commit -m "feat: render plugin/memory detail content in View()"
```

---

### Task 7: Update menu hints

**Files:**
- Modify: `internal/ui/menu.go`

**Step 1: Add detail cases to `TableNavItems()`**

Find the `switch rt` block in `TableNavItems`. Add:

```go
case model.ResourcePlugins:
    items = append(items, MenuItem{Key: "enter", Desc: "detail"})
case model.ResourceMemory:
    items = append(items, MenuItem{Key: "enter", Desc: "detail"})
```

And in the `esc` section:

```go
case model.ResourcePluginDetail, model.ResourceMemoryDetail:
    items = append(items, MenuItem{Key: "esc", Desc: "back"})
```

**Step 2: Build and run tests**

```bash
make build && make test
```
Expected: success.

**Step 3: Commit**

```bash
git add internal/ui/menu.go
git commit -m "feat: add enter/esc menu hints for plugin and memory detail views"
```

---

### Task 8: Verify conventions and push

**Step 1: Run fmt, lint, test**

```bash
make fmt && make lint && make test
```
Expected: all clean.

**Step 2: Manual smoke test with real data**

```bash
./bin/claudeview
```

- Navigate to plugins (`p`), select a plugin, press enter → detail view shows Skills/Commands/Hooks/Agents/MCPs.
- Press esc → back to plugins list.
- Navigate to a project, then memories (`m`), select a memory file, press enter → file content shown.
- Press esc → back to memories list.

**Step 3: Push**

```bash
git push origin feat/ui-ux-improvements
```
