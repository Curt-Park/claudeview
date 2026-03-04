package ui_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Curt-Park/claudeview/internal/model"
	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestNavMoveDown(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(5))

	app = updateApp(app, keyMsg("j"))

	if app.Table.Selected != 1 {
		t.Errorf("expected Selected=1 after j, got %d", app.Table.Selected)
	}
}

func TestNavMoveUp(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(5))

	app = updateApp(app, keyMsg("j"))
	app = updateApp(app, keyMsg("j"))
	app = updateApp(app, keyMsg("k"))

	if app.Table.Selected != 1 {
		t.Errorf("expected Selected=1 after j/j/k, got %d", app.Table.Selected)
	}
}

func TestNavGotoTop(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(5))

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	app = updateApp(app, keyMsg("g"))

	if app.Table.Selected != 0 {
		t.Errorf("expected Selected=0 after G/g, got %d", app.Table.Selected)
	}
}

func TestNavPageDown(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(10))

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyCtrlD})

	if app.Table.Selected == 0 {
		t.Error("expected Selected>0 after ctrl+d")
	}
}

func TestDrilldownProjectsToSessions(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "3", "1", "2h"}, Data: p}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions after Enter, got %s", app.Resource)
	}
	if app.SelectedProjectHash != p.Hash {
		t.Errorf("expected SelectedProjectHash=%s, got %s", p.Hash, app.SelectedProjectHash)
	}
}

func TestDrilldownEscNavigatesBack(t *testing.T) {
	app := newApp(model.ResourceSessions)
	app.Table.SetRows(sessionRows(2))

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects after Esc from sessions, got %s", app.Resource)
	}
}

func TestFilterEscClears(t *testing.T) {
	app := newApp(model.ResourceProjects)
	app.Table.SetRows(projectRows(3))

	app = updateApp(app, keyMsg("/"))
	app = updateApp(app, keyMsg("q"))
	app = updateApp(app, keyMsg("q"))
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Filter.Input != "" {
		t.Errorf("expected filter cleared, got %q", app.Filter.Input)
	}
	if app.Table.Filter != "" {
		t.Errorf("expected table filter cleared, got %q", app.Table.Filter)
	}
}

func TestResizeHandled(t *testing.T) {
	app := newApp(model.ResourceProjects)

	app = updateApp(app, tea.WindowSizeMsg{Width: 80, Height: 24})

	if app.Width != 80 {
		t.Errorf("expected Width=80, got %d", app.Width)
	}
	if app.Height != 24 {
		t.Errorf("expected Height=24, got %d", app.Height)
	}
}

func TestDrilldownClearsFilter(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.Table.Filter = "proj" // matches the row so drilldown can proceed

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Table.Filter != "" {
		t.Errorf("expected Table.Filter cleared after drilldown, got %q", app.Table.Filter)
	}
	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions, got %s", app.Resource)
	}
}

func TestNavigateBackRestoresFilter(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.Table.Filter = "proj" // matches row so drilldown proceeds

	// Drill down to sessions — filter should be saved and cleared
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	if app.Table.Filter != "" {
		t.Errorf("expected filter cleared after drilldown, got %q", app.Table.Filter)
	}

	// Navigate back — filter should be restored
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects after Esc, got %s", app.Resource)
	}
	if app.Table.Filter != "proj" {
		t.Errorf("expected restored filter=%q, got %q", "proj", app.Table.Filter)
	}
}

func TestEscClearsRestoredFilter(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.Table.Filter = "proj" // matches row so drilldown proceeds

	// Drill down then back — filter is restored
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Table.Filter != "proj" {
		t.Fatalf("expected restored filter=%q, got %q", "proj", app.Table.Filter)
	}

	// Esc again — should clear the restored filter (not navigate back further)
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Table.Filter != "" {
		t.Errorf("expected filter cleared by Esc, got %q", app.Table.Filter)
	}
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected still on projects, got %s", app.Resource)
	}
}

func TestDrilldownPluginToDetail(t *testing.T) {
	p := &model.Plugin{Name: "superpowers", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""},
		Data:  p,
	}})

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
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""},
		Data:  p,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourcePlugins {
		t.Errorf("expected resource=plugins after Esc from plugin-detail, got %s", app.Resource)
	}
}

func TestDrilldownMemoryToDetail(t *testing.T) {
	m := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"MEMORY.md", "", "1 KB", "1h"},
		Data:  m,
	}})

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
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"MEMORY.md", "", "1 KB", "1h"},
		Data:  m,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourceMemory {
		t.Errorf("expected resource=memories after Esc from memory-detail, got %s", app.Resource)
	}
}

func TestEscFromPluginDetailRestoresFilter(t *testing.T) {
	p := &model.Plugin{Name: "superpowers", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""}, Data: p}})
	app.Table.Filter = "super"

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourcePlugins {
		t.Errorf("expected resource=plugins after Esc from plugin-detail, got %s", app.Resource)
	}
	if app.Table.Filter != "super" {
		t.Errorf("expected filter %q restored after Esc from plugin-detail, got %q", "super", app.Table.Filter)
	}
}

func TestFilterActivatesInPluginDetail(t *testing.T) {
	p := &model.Plugin{Name: "superpowers", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""},
		Data:  p,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // drill into plugin-detail

	app = updateApp(app, keyMsg("/"))

	if !app.Filter.Active {
		t.Error("expected filter to activate in plugin-detail, but it did not")
	}
}

func TestFilterTypingUpdatesPluginDetailView(t *testing.T) {
	p := &model.Plugin{Name: "superpowers", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"superpowers", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""},
		Data:  p,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	app = updateApp(app, keyMsg("/"))
	app = updateApp(app, keyMsg("d"))
	app = updateApp(app, keyMsg("e"))
	app = updateApp(app, keyMsg("b"))

	if app.Table.Filter != "deb" {
		t.Errorf("expected Table.Filter=%q after typing, got %q", "deb", app.Table.Filter)
	}
}

func TestFilterKeyIgnoredInMemoryDetail(t *testing.T) {
	m := &model.Memory{Name: "MEMORY.md", Path: "/tmp/MEMORY.md"}
	app := newApp(model.ResourceMemory)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"MEMORY.md", "", "1 KB", "1h"},
		Data:  m,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // drill into memory-detail

	app = updateApp(app, keyMsg("/"))

	if app.Filter.Active {
		t.Error("expected filter to remain inactive in memory-detail, but it was activated")
	}
}

func TestDrilldownPluginDetailToItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginDetail)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"skill", "my-skill"},
		Data:  pi,
	}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourcePluginItemDetail {
		t.Errorf("expected resource=plugin-item-detail after Enter on plugin item, got %s", app.Resource)
	}
	if app.SelectedPluginItem != pi {
		t.Errorf("expected SelectedPluginItem set after Enter")
	}
}

func TestEscFromPluginItemDetailReturnsToPluginDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginDetail)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"skill", "my-skill"},
		Data:  pi,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourcePluginDetail {
		t.Errorf("expected resource=plugin-detail after Esc from plugin-item-detail, got %s", app.Resource)
	}
}

func TestFilterKeyIgnoredInPluginItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginDetail)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"skill", "my-skill"},
		Data:  pi,
	}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // drill into plugin-item-detail

	app = updateApp(app, keyMsg("/"))

	if app.Filter.Active {
		t.Error("expected filter to remain inactive in plugin-item-detail, but it was activated")
	}
}

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

func TestMKeyBlockedInPluginDetail(t *testing.T) {
	p := &model.Plugin{Name: "myplugin", CacheDir: "/tmp"}
	app := newApp(model.ResourcePlugins)
	app.SelectedProjectHash = "proj-abc"
	app.Table.SetRows([]ui.Row{{Cells: []string{"myplugin", "1.0", "user", "enabled", "0", "0", "0", "0", "0", ""}, Data: p}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // → plugin-detail

	app = updateApp(app, keyMsg("m"))

	if app.Resource != model.ResourcePluginDetail {
		t.Errorf("expected to stay on plugin-detail after m, got %s", app.Resource)
	}
}

func TestMKeyBlockedInPluginItemDetail(t *testing.T) {
	pi := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: "/tmp"}
	app := newApp(model.ResourcePluginDetail)
	app.SelectedProjectHash = "proj-abc"
	app.Table.SetRows([]ui.Row{{Cells: []string{"skill", "my-skill"}, Data: pi}})
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter}) // → plugin-item-detail

	app = updateApp(app, keyMsg("m"))

	if app.Resource != model.ResourcePluginItemDetail {
		t.Errorf("expected to stay on plugin-item-detail after m, got %s", app.Resource)
	}
}

func TestDrilldownSessionsToSessionChat(t *testing.T) {
	s := &model.Session{ID: "sess-abc123", FilePath: "/tmp/fake.jsonl"}
	app := newApp(model.ResourceSessions)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"", s.ShortID(), "topic", "2", "10", "1k", "1h"},
		Data:  s,
	}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceHistory {
		t.Errorf("expected resource=session-chat after Enter on sessions, got %s", app.Resource)
	}
}

func TestSessionChatEscReturnsToSessions(t *testing.T) {
	app := newApp(model.ResourceHistory)

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions after Esc from session-chat, got %s", app.Resource)
	}
}

func TestSessionChatIsTableView(t *testing.T) {
	app := newApp(model.ResourceHistory)
	app.Width = termWidth
	app.Height = termHeight
	items := []ui.ChatItem{
		{Turn: model.Turn{Role: "user", Text: "hello"}},
		{Turn: model.Turn{Role: "assistant", Text: "hi"}},
	}
	app.ChatItems = items
	app.Table.SetRows(chatItemRows(items))

	app = updateApp(app, keyMsg("j"))

	if app.Table.Selected != 1 {
		t.Errorf("expected Table.Selected=1 after j in session-chat table, got %d", app.Table.Selected)
	}
}

func TestContentOffsetFieldExists(t *testing.T) {
	app := newApp(model.ResourcePluginItemDetail)
	if app.ContentOffset != 0 {
		t.Errorf("expected ContentOffset=0, got %d", app.ContentOffset)
	}
}

func TestScrollDownInPluginItemDetail(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.Width = termWidth
	app.Height = termHeight
	tmpFile := filepath.Join(t.TempDir(), "mem.md")
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&sb, "line%d\n", i)
	}
	if err := os.WriteFile(tmpFile, []byte(strings.TrimRight(sb.String(), "\n")), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	app.SelectedMemory = &model.Memory{Path: tmpFile}

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
	app := newApp(model.ResourceMemoryDetail)
	app.Width = termWidth
	app.Height = termHeight
	tmpFile := filepath.Join(t.TempDir(), "mem.md")
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&sb, "line%d\n", i)
	}
	if err := os.WriteFile(tmpFile, []byte(strings.TrimRight(sb.String(), "\n")), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	app.SelectedMemory = &model.Memory{Path: tmpFile}
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

func TestScrollGotoBottomGoesToMax(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.Width = termWidth
	app.Height = termHeight
	tmpFile := filepath.Join(t.TempDir(), "mem.md")
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&sb, "line%d\n", i)
	}
	if err := os.WriteFile(tmpFile, []byte(strings.TrimRight(sb.String(), "\n")), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	app.SelectedMemory = &model.Memory{Path: tmpFile}

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})

	max := app.ContentOffset
	if max <= 0 {
		t.Errorf("expected ContentOffset > 0 after G with 50-line content, got %d", max)
	}
	// G is idempotent at the bottom.
	app2 := updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if app2.ContentOffset != max {
		t.Errorf("expected G idempotent at bottom: got %d, want %d", app2.ContentOffset, max)
	}
}

func TestScrollCtrlDCapsAtMax(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.Width = termWidth
	app.Height = termHeight
	tmpFile := filepath.Join(t.TempDir(), "mem.md")
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&sb, "line%d\n", i)
	}
	if err := os.WriteFile(tmpFile, []byte(strings.TrimRight(sb.String(), "\n")), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	app.SelectedMemory = &model.Memory{Path: tmpFile}

	// Scroll to bottom via G.
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	max := app.ContentOffset

	// ctrl+d at bottom must not exceed max.
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyCtrlD})
	if app.ContentOffset != max {
		t.Errorf("expected ContentOffset capped at %d after ctrl+d at bottom, got %d", max, app.ContentOffset)
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

func TestViewAppliesContentOffset(t *testing.T) {
	app := newApp(model.ResourceMemoryDetail)
	app.Width = termWidth
	app.Height = termHeight
	app.ContentOffset = 2

	tmpFile := filepath.Join(t.TempDir(), "mem.md")
	// Build enough lines to exceed contentHeight so scrolling is not capped to 0.
	// ContentHeight ≈ termHeight - chrome (~8 rows), so 50 lines is always enough.
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		fmt.Fprintf(&sb, "line%d\n", i)
	}
	content := strings.TrimRight(sb.String(), "\n")
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

func TestJumpPreservesFilterStack(t *testing.T) {
	p := &model.Project{Hash: "proj-abc123"}
	s := &model.Session{ID: "sess-xyz789"}
	app := newApp(model.ResourceProjects)
	app.Table.SetRows([]ui.Row{{Cells: []string{p.Hash, "1", "1h"}, Data: p}})
	app.SelectedProjectHash = p.Hash
	app.Table.Filter = "proj"

	// Drill into sessions — filterStack=["proj"], filter=""
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	app.Table.SetRows([]ui.Row{{Cells: []string{"", s.ShortID(), "topic", "1", "1", "sonnet:1k", "1h"}, Data: s}})

	// Jump to plugins — saves {Filter:"", FilterStack:["proj"]}
	app = updateApp(app, keyMsg("p"))
	if app.Resource != model.ResourcePlugins {
		t.Fatalf("expected resource=plugins after p, got %s", app.Resource)
	}

	// Esc back to sessions — restores filter="" and filterStack=["proj"]
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceSessions {
		t.Fatalf("expected resource=sessions after Esc from plugins, got %s", app.Resource)
	}
	if app.Table.Filter != "" {
		t.Errorf("expected filter=%q after jump-back, got %q", "", app.Table.Filter)
	}

	// Esc from sessions — popFilter restores "proj", go to projects
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceProjects {
		t.Errorf("expected resource=projects, got %s", app.Resource)
	}
	if app.Table.Filter != "proj" {
		t.Errorf("expected filter=%q after esc from sessions, got %q", "proj", app.Table.Filter)
	}
}

func TestSessionChat_FollowMode_InitialDrillDown(t *testing.T) {
	s := &model.Session{ID: "sess-abc123", FilePath: "/tmp/fake.jsonl"}
	app := newApp(model.ResourceSessions)
	app.Table.SetRows([]ui.Row{{
		Cells: []string{"", s.ShortID(), "topic", "2", "10", "1k", "1h"},
		Data:  s,
	}})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceHistory {
		t.Fatalf("expected resource=session-chat, got %s", app.Resource)
	}
	if !app.ChatFollow {
		t.Error("expected ChatFollow=true after drill-down into session-chat")
	}
}

func TestSessionChat_FollowMode_ScrollUp_DisablesFollow(t *testing.T) {
	app := newApp(model.ResourceHistory)
	app.ChatFollow = true
	items := []ui.ChatItem{
		{Turn: model.Turn{Role: "user", Text: "hello"}},
		{Turn: model.Turn{Role: "assistant", Text: "hi"}},
	}
	app.ChatItems = items
	app.Table.SetRows(chatItemRows(items))
	app.Table.Selected = 1

	app = updateApp(app, keyMsg("k"))

	if app.ChatFollow {
		t.Error("expected ChatFollow=false after scrolling up with k")
	}
}

func TestSessionChat_FollowMode_G_EnablesFollow(t *testing.T) {
	app := newApp(model.ResourceHistory)
	app.ChatFollow = false
	items := []ui.ChatItem{
		{Turn: model.Turn{Role: "user", Text: "hello"}},
		{Turn: model.Turn{Role: "assistant", Text: "hi"}},
	}
	app.ChatItems = items
	app.Table.SetRows(chatItemRows(items))
	app.Table.Selected = 0

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})

	if !app.ChatFollow {
		t.Error("expected ChatFollow=true after pressing G in session-chat")
	}
}

func TestSessionChat_NavigateBack_ClearsState(t *testing.T) {
	app := newApp(model.ResourceHistory)
	app.SelectedSessionID = "sess-abc123"
	app.SelectedSessionFilePath = "/tmp/fake.jsonl"
	app.SelectedSessionSubagentDir = "/tmp/subagents"
	app.ChatFollow = true

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})

	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions after Esc from session-chat, got %s", app.Resource)
	}
	if app.SelectedSessionFilePath != "" {
		t.Errorf("expected SelectedSessionFilePath cleared, got %q", app.SelectedSessionFilePath)
	}
	if app.SelectedSessionSubagentDir != "" {
		t.Errorf("expected SelectedSessionSubagentDir cleared, got %q", app.SelectedSessionSubagentDir)
	}
	if app.ChatFollow {
		t.Error("expected ChatFollow=false after navigating back from session-chat")
	}
}

func TestSlugGroupDrillDown(t *testing.T) {
	s1 := &model.Session{ID: "sess-1", Slug: "fizzy-stallman", FilePath: "/tmp/s1.jsonl"}
	s2 := &model.Session{ID: "sess-2", Slug: "fizzy-stallman", FilePath: "/tmp/s2.jsonl"}
	// s2 is the representative with GroupSessions set (as GroupSessionsBySlug would produce)
	s2.GroupSessions = []*model.Session{s1, s2}
	s3 := &model.Session{ID: "sess-3", FilePath: "/tmp/s3.jsonl"} // no slug
	app := newApp(model.ResourceSessions)
	app.Table.SetRows([]ui.Row{
		{Cells: []string{"fizzy-stallman", s2.GroupNameCell(), "topic2", "5", "3", "3k", "30m"}, Data: s2},
		{Cells: []string{"", s3.ShortID(), "topic3", "1", "1", "500", "5m"}, Data: s3},
	})

	// Drill into s2 (group representative) — should populate SlugSessions
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceHistory {
		t.Fatalf("expected resource=history, got %s", app.Resource)
	}
	if len(app.SlugSessions) != 2 {
		t.Fatalf("expected 2 SlugSessions, got %d", len(app.SlugSessions))
	}
	if app.SlugSessions[0].ID != "sess-1" || app.SlugSessions[1].ID != "sess-2" {
		t.Errorf("expected SlugSessions=[sess-1,sess-2], got [%s,%s]", app.SlugSessions[0].ID, app.SlugSessions[1].ID)
	}
}

func TestSlugGroupDrillDown_NoSlug(t *testing.T) {
	s := &model.Session{ID: "sess-solo", FilePath: "/tmp/solo.jsonl"} // no slug
	app := newApp(model.ResourceSessions)
	app.Table.SetRows([]ui.Row{
		{Cells: []string{"", s.ShortID(), "topic", "2", "1", "1k", "1h"}, Data: s},
	})

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	if app.Resource != model.ResourceHistory {
		t.Fatalf("expected resource=history, got %s", app.Resource)
	}
	if app.SlugSessions != nil {
		t.Errorf("expected nil SlugSessions for no-slug session, got %d", len(app.SlugSessions))
	}
}

func TestSlugGroupNavigateBack_ClearsState(t *testing.T) {
	s1 := &model.Session{ID: "sess-1", Slug: "fizzy-stallman", FilePath: "/tmp/s1.jsonl"}
	s2 := &model.Session{ID: "sess-2", Slug: "fizzy-stallman", FilePath: "/tmp/s2.jsonl"}
	s2.GroupSessions = []*model.Session{s1, s2}
	app := newApp(model.ResourceSessions)
	app.Table.SetRows([]ui.Row{
		{Cells: []string{"fizzy-stallman", s2.GroupNameCell(), "topic2", "5", "2", "3k", "30m"}, Data: s2},
	})

	// Drill into slug group
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})
	if len(app.SlugSessions) != 2 {
		t.Fatalf("expected 2 SlugSessions after drill-down, got %d", len(app.SlugSessions))
	}

	// Navigate back
	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEsc})
	if app.Resource != model.ResourceSessions {
		t.Errorf("expected resource=sessions after Esc, got %s", app.Resource)
	}
	if app.SlugSessions != nil {
		t.Errorf("expected SlugSessions cleared after navigate-back, got %d", len(app.SlugSessions))
	}
}

func TestDividerSkippedDuringNavigation(t *testing.T) {
	normal1 := ui.ChatItem{Turn: model.Turn{Role: "user", Text: "hello"}}
	divider := ui.ChatItem{IsDivider: true, DividerLabel: "── session 2/2 ──"}
	normal2 := ui.ChatItem{Turn: model.Turn{Role: "assistant", Text: "hi"}}
	items := []ui.ChatItem{normal1, divider, normal2}
	app := newApp(model.ResourceHistory)
	app.ChatItems = items
	app.Table.SetRows(chatItemRows(items))
	app.Table.Selected = 0

	// j should skip divider (index 1) and land on index 2
	app = updateApp(app, keyMsg("j"))
	if app.Table.Selected != 2 {
		t.Errorf("expected Selected=2 after j (skipping divider), got %d", app.Table.Selected)
	}

	// k should skip divider (index 1) and land on index 0
	app = updateApp(app, keyMsg("k"))
	if app.Table.Selected != 0 {
		t.Errorf("expected Selected=0 after k (skipping divider), got %d", app.Table.Selected)
	}
}

func TestDividerSkippedOnDrillDown(t *testing.T) {
	divider := ui.ChatItem{IsDivider: true, DividerLabel: "── session 2/2 ──"}
	normal := ui.ChatItem{Turn: model.Turn{Role: "user", Text: "hello"}}
	app := newApp(model.ResourceHistory)
	app.ChatItems = []ui.ChatItem{normal, divider}
	app.Table.SetRows(chatItemRows([]ui.ChatItem{normal, divider}))
	app.Table.Selected = 1 // select the divider row

	app = updateApp(app, tea.KeyMsg{Type: tea.KeyEnter})

	// Should remain on history (divider is not drillable)
	if app.Resource != model.ResourceHistory {
		t.Errorf("expected to stay on history when entering on divider, got %s", app.Resource)
	}
}

func TestPluginItemDetailViewHidesJumpHints(t *testing.T) {
	app := newApp(model.ResourcePluginItemDetail)
	app.Width = termWidth
	app.Height = termHeight
	app.Info.MemoriesActive = true

	rendered := app.View()

	if strings.Contains(rendered, "plugins") {
		t.Error("expected '<p> plugins' hint to be hidden in plugin-item-detail view")
	}
	if strings.Contains(rendered, "memories") {
		t.Error("expected '<m> memories' hint to be hidden in plugin-item-detail view")
	}
}
