package model_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestListPluginItems(t *testing.T) {
	base := makeTempDir(t)

	// Create a skill
	skillDir := filepath.Join(base, "skills", "debug")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(skillDir, "debug.md"))

	// Create a command
	cmdDir := filepath.Join(base, "commands")
	if err := os.Mkdir(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(cmdDir, "commit.md"))

	items := model.ListPluginItems(base)
	if len(items) != 2 {
		t.Fatalf("ListPluginItems() returned %d items, want 2", len(items))
	}
	categories := make(map[string]bool)
	for _, item := range items {
		categories[item.Category] = true
		if item.CacheDir != base {
			t.Errorf("expected CacheDir=%q, got %q", base, item.CacheDir)
		}
	}
	if !categories["skill"] {
		t.Error("expected a skill item")
	}
	if !categories["command"] {
		t.Error("expected a command item")
	}
}

func TestReadPluginItemContent_Skill(t *testing.T) {
	base := makeTempDir(t)
	skillDir := filepath.Join(base, "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "my-skill.md"), []byte("skill content"), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: base}
	got := model.ReadPluginItemContent(item)
	if got != "skill content" {
		t.Errorf("ReadPluginItemContent() = %q, want %q", got, "skill content")
	}
}

func TestReadPluginItemContent_Command(t *testing.T) {
	base := makeTempDir(t)
	cmdDir := filepath.Join(base, "commands")
	if err := os.Mkdir(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "commit.md"), []byte("commit docs"), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "commit", Category: "command", CacheDir: base}
	got := model.ReadPluginItemContent(item)
	if got != "commit docs" {
		t.Errorf("ReadPluginItemContent() = %q, want %q", got, "commit docs")
	}
}

func TestReadPluginItemContent_MCPFromJSON(t *testing.T) {
	base := makeTempDir(t)
	content := `{"mcpServers":{"my-server":{"command":"npx","args":["server"]}}}`
	if err := os.WriteFile(filepath.Join(base, ".mcp.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "my-server", Category: "mcp", CacheDir: base}
	got := model.ReadPluginItemContent(item)
	if !strings.Contains(got, "npx") {
		t.Errorf("expected MCP content to contain server config, got %q", got)
	}
}

func TestReadPluginItemContent_MCPNormalizesIndentation(t *testing.T) {
	base := makeTempDir(t)
	// Indented JSON: the "my-server" value has extra leading spaces inherited from
	// the parent structure. This mirrors real plugin .mcp.json files and was the
	// source of the over-indentation bug.
	content := "{\n  \"mcpServers\": {\n    \"my-server\": {\n      \"command\": \"npx\",\n      \"args\": [\"server\"]\n    }\n  }\n}"
	if err := os.WriteFile(filepath.Join(base, ".mcp.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "my-server", Category: "mcp", CacheDir: base}
	got := model.ReadPluginItemContent(item)
	want := "{\n  \"command\": \"npx\",\n  \"args\": [\n    \"server\"\n  ]\n}"
	if got != want {
		t.Errorf("ReadPluginItemContent() indentation not normalized\ngot:  %q\nwant: %q", got, want)
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

func TestReadHookCommandScripts_DirectScriptPath(t *testing.T) {
	base := makeTempDir(t)
	hooksDir := filepath.Join(base, "hooks")
	if err := os.Mkdir(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	scriptContent := "#!/bin/bash\necho stop"
	scriptPath := filepath.Join(hooksDir, "stop-hook.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatal(err)
	}

	hookJSON := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"${CLAUDE_PLUGIN_ROOT}/hooks/stop-hook.sh"}]}]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "Stop", Category: "hook", CacheDir: base}
	scripts := model.ReadHookCommandScripts(item)

	if len(scripts) != 1 {
		t.Fatalf("ReadHookCommandScripts() returned %d scripts, want 1", len(scripts))
	}
	if scripts[0].Path != scriptPath {
		t.Errorf("Path = %q, want %q", scripts[0].Path, scriptPath)
	}
	if !strings.Contains(scripts[0].Content, "echo stop") {
		t.Errorf("Content = %q, expected to contain %q", scripts[0].Content, "echo stop")
	}
}

func TestReadHookCommandScripts_BashInvocationExtractsScript(t *testing.T) {
	base := makeTempDir(t)
	hooksDir := filepath.Join(base, "hooks")
	scriptsDir := filepath.Join(base, "scripts")
	if err := os.Mkdir(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(scriptsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	scriptContent := "#!/bin/bash\necho session-start"
	scriptPath := filepath.Join(scriptsDir, "launcher.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatal(err)
	}

	hookJSON := `{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"bash \"${CLAUDE_PLUGIN_ROOT}/scripts/launcher.sh\" hook session-start"}]}]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "SessionStart", Category: "hook", CacheDir: base}
	scripts := model.ReadHookCommandScripts(item)

	if len(scripts) != 1 {
		t.Fatalf("ReadHookCommandScripts() returned %d scripts, want 1", len(scripts))
	}
	if !strings.Contains(scripts[0].Content, "echo session-start") {
		t.Errorf("Content = %q, expected to contain %q", scripts[0].Content, "echo session-start")
	}
}

func TestReadHookCommandScripts_NonHookReturnsNil(t *testing.T) {
	item := &model.PluginItem{Name: "my-skill", Category: "skill", CacheDir: t.TempDir()}
	scripts := model.ReadHookCommandScripts(item)
	if scripts != nil {
		t.Errorf("ReadHookCommandScripts() = %v, want nil for non-hook category", scripts)
	}
}

func TestReadHookCommandScripts_PluginSubdirLayout(t *testing.T) {
	base := makeTempDir(t)
	pluginSub := filepath.Join(base, "plugin")
	hooksDir := filepath.Join(pluginSub, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	scriptContent := "#!/bin/bash\necho plugin-subdir"
	scriptPath := filepath.Join(hooksDir, "hook.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		t.Fatal(err)
	}

	// ${CLAUDE_PLUGIN_ROOT} should resolve to the plugin/ subdir (contentDir result)
	hookJSON := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"${CLAUDE_PLUGIN_ROOT}/hooks/hook.sh"}]}]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "Stop", Category: "hook", CacheDir: base}
	scripts := model.ReadHookCommandScripts(item)

	if len(scripts) != 1 {
		t.Fatalf("ReadHookCommandScripts() returned %d scripts, want 1 (plugin/ layout)", len(scripts))
	}
	if !strings.Contains(scripts[0].Content, "echo plugin-subdir") {
		t.Errorf("Content = %q, expected to contain %q", scripts[0].Content, "echo plugin-subdir")
	}
}

func TestReadHookCommandScripts_MissingScriptSkipped(t *testing.T) {
	base := makeTempDir(t)
	hooksDir := filepath.Join(base, "hooks")
	if err := os.Mkdir(hooksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	hookJSON := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"${CLAUDE_PLUGIN_ROOT}/hooks/nonexistent.sh"}]}]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(hookJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	item := &model.PluginItem{Name: "Stop", Category: "hook", CacheDir: base}
	scripts := model.ReadHookCommandScripts(item)
	if len(scripts) != 0 {
		t.Errorf("ReadHookCommandScripts() returned %d scripts, want 0 for missing script", len(scripts))
	}
}
