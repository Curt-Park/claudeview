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

	t.Run("counts skill subdirectories", func(t *testing.T) {
		base := makeTempDir(t)
		dir := filepath.Join(base, "skills")
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		// Create subdirectories (each skill is a subdir)
		if err := os.Mkdir(filepath.Join(dir, "skill1"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(dir, "skill2"), 0o755); err != nil {
			t.Fatal(err)
		}
		// Files should not be counted
		writeFile(t, filepath.Join(dir, "README.md"))
		if got := model.CountSkills(base); got != 2 {
			t.Errorf("CountSkills() = %d, want 2 (only subdirs)", got)
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

func TestContentDirFallback(t *testing.T) {
	base := makeTempDir(t)
	// No "plugin/" subdir — should return base itself
	skills := filepath.Join(base, "skills")
	if err := os.Mkdir(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(skills, "my-skill"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := model.CountSkills(base); got != 1 {
		t.Errorf("CountSkills() = %d, want 1 (from root)", got)
	}
}

func TestContentDirWithPluginSubdir(t *testing.T) {
	base := makeTempDir(t)
	// Create "plugin/" subdir with its own skills/
	pluginSub := filepath.Join(base, "plugin")
	if err := os.Mkdir(pluginSub, 0o755); err != nil {
		t.Fatal(err)
	}
	subSkills := filepath.Join(pluginSub, "skills")
	if err := os.Mkdir(subSkills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(subSkills, "skill-a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(subSkills, "skill-b"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Root-level skills/ should be ignored when plugin/ exists
	rootSkills := filepath.Join(base, "skills")
	if err := os.Mkdir(rootSkills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(rootSkills, "decoy"), 0o755); err != nil {
		t.Fatal(err)
	}

	if got := model.CountSkills(base); got != 2 {
		t.Errorf("CountSkills() = %d, want 2 (from plugin/ subdir)", got)
	}
}

func TestCountSkillsCountsSubdirs(t *testing.T) {
	base := makeTempDir(t)
	skills := filepath.Join(base, "skills")
	if err := os.Mkdir(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	// 3 skill directories
	for _, name := range []string{"brainstorming", "debugging", "tdd"} {
		if err := os.Mkdir(filepath.Join(skills, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Files should not count
	writeFile(t, filepath.Join(skills, "index.md"))

	if got := model.CountSkills(base); got != 3 {
		t.Errorf("CountSkills() = %d, want 3 subdirs", got)
	}
}

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
	writeFile(t, filepath.Join(dir, "README.md")) // files should be ignored

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
	for _, name := range got {
		if filepath.Ext(name) == ".md" {
			t.Errorf("expected .md extension stripped, got %q", name)
		}
	}
}

func TestListSkillsMissingDir(t *testing.T) {
	got := model.ListSkills("/nonexistent/path/xyz")
	if got != nil {
		t.Errorf("expected nil for missing dir, got %v", got)
	}
}

func TestListHooks(t *testing.T) {
	base := makeTempDir(t)
	dir := filepath.Join(base, "hooks")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"hooks":{"PreToolUse":[],"PostToolUse":[]}}`
	if err := os.WriteFile(filepath.Join(dir, "hooks.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := model.ListHooks(base)
	if len(got) != 2 {
		t.Fatalf("ListHooks() = %v, want 2 items", got)
	}
}

func TestListAgents(t *testing.T) {
	base := makeTempDir(t)
	dir := filepath.Join(base, "agents")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "code-reviewer.md"))
	got := model.ListAgents(base)
	if len(got) != 1 {
		t.Fatalf("ListAgents() = %v, want 1 item", got)
	}
	if got[0] != "code-reviewer" {
		t.Errorf("got %q, want %q", got[0], "code-reviewer")
	}
}

func TestListMCPs(t *testing.T) {
	base := makeTempDir(t)
	content := `{"mcpServers":{"my-server":{},"other-server":{}}}`
	if err := os.WriteFile(filepath.Join(base, ".mcp.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got := model.ListMCPs(base)
	if len(got) != 2 {
		t.Fatalf("ListMCPs() = %v, want 2 items", got)
	}
}

func TestCountMCPs(t *testing.T) {
	t.Run("missing file returns 0", func(t *testing.T) {
		if got := model.CountMCPs("/nonexistent/path"); got != 0 {
			t.Errorf("CountMCPs() = %d, want 0", got)
		}
	})

	t.Run("parses mcpServers count", func(t *testing.T) {
		base := makeTempDir(t)
		content := `{"mcpServers":{"server1":{},"server2":{},"server3":{}}}`
		if err := os.WriteFile(filepath.Join(base, ".mcp.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := model.CountMCPs(base); got != 3 {
			t.Errorf("CountMCPs() = %d, want 3", got)
		}
	})

	t.Run("empty mcpServers returns 0", func(t *testing.T) {
		base := makeTempDir(t)
		content := `{"mcpServers":{}}`
		if err := os.WriteFile(filepath.Join(base, ".mcp.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := model.CountMCPs(base); got != 0 {
			t.Errorf("CountMCPs() = %d, want 0", got)
		}
	})
}
