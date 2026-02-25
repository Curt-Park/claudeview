package model_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "plugin-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile %q: %v", path, err)
	}
}

func TestCountSkills(t *testing.T) {
	t.Run("missing dir returns 0", func(t *testing.T) {
		if got := model.CountSkills("/nonexistent/path/xyz"); got != 0 {
			t.Errorf("CountSkills() = %d, want 0 for missing dir", got)
		}
	})

	t.Run("dir with md files counts them", func(t *testing.T) {
		base := makeTempDir(t)
		dir := filepath.Join(base, "skills")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "skill1.md"))
		writeFile(t, filepath.Join(dir, "skill2.md"))
		if got := model.CountSkills(base); got != 2 {
			t.Errorf("CountSkills() = %d, want 2", got)
		}
	})

	t.Run("mixed extensions counts only md files", func(t *testing.T) {
		base := makeTempDir(t)
		dir := filepath.Join(base, "skills")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "skill1.md"))
		writeFile(t, filepath.Join(dir, "skill2.txt"))
		writeFile(t, filepath.Join(dir, "skill3.sh"))
		if got := model.CountSkills(base); got != 1 {
			t.Errorf("CountSkills() = %d, want 1 (only .md)", got)
		}
	})
}

func TestCountCommands(t *testing.T) {
	t.Run("missing dir returns 0", func(t *testing.T) {
		if got := model.CountCommands("/nonexistent/path/xyz"); got != 0 {
			t.Errorf("CountCommands() = %d, want 0 for missing dir", got)
		}
	})

	t.Run("dir with md files counts them", func(t *testing.T) {
		base := makeTempDir(t)
		dir := filepath.Join(base, "commands")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "cmd1.md"))
		writeFile(t, filepath.Join(dir, "cmd2.md"))
		writeFile(t, filepath.Join(dir, "cmd3.md"))
		if got := model.CountCommands(base); got != 3 {
			t.Errorf("CountCommands() = %d, want 3", got)
		}
	})

	t.Run("mixed extensions counts only md files", func(t *testing.T) {
		base := makeTempDir(t)
		dir := filepath.Join(base, "commands")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "cmd1.md"))
		writeFile(t, filepath.Join(dir, "cmd2.json"))
		if got := model.CountCommands(base); got != 1 {
			t.Errorf("CountCommands() = %d, want 1 (only .md)", got)
		}
	})
}

func TestCountHooks(t *testing.T) {
	t.Run("missing dir returns 0", func(t *testing.T) {
		if got := model.CountHooks("/nonexistent/path/xyz"); got != 0 {
			t.Errorf("CountHooks() = %d, want 0 for missing dir", got)
		}
	})

	t.Run("dir with any extension files counts all", func(t *testing.T) {
		base := makeTempDir(t)
		dir := filepath.Join(base, "hooks")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dir, "hook1.sh"))
		writeFile(t, filepath.Join(dir, "hook2.js"))
		writeFile(t, filepath.Join(dir, "hook3.py"))
		if got := model.CountHooks(base); got != 3 {
			t.Errorf("CountHooks() = %d, want 3", got)
		}
	})
}
