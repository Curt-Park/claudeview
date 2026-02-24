package ui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/muesli/termenv"

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
