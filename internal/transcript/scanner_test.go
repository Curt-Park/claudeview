package transcript_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/transcript"
)

func TestScanProjects(t *testing.T) {
	// Create temp claude dir structure
	dir := t.TempDir()
	projectsDir := filepath.Join(dir, "projects")

	// Create two project dirs with session files
	for _, proj := range []string{"proj-hash-1", "proj-hash-2"} {
		projDir := filepath.Join(projectsDir, proj)
		if err := os.MkdirAll(projDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Create a session file
		sessionFile := filepath.Join(projDir, "session-abc123.jsonl")
		if err := os.WriteFile(sessionFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		// Touch to set mod time
		now := time.Now()
		os.Chtimes(sessionFile, now, now)
	}

	projects, err := transcript.ScanProjects(dir)
	if err != nil {
		t.Fatalf("ScanProjects failed: %v", err)
	}

	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}

	for _, p := range projects {
		if len(p.Sessions) != 1 {
			t.Errorf("project %s: expected 1 session, got %d", p.Hash, len(p.Sessions))
		}
	}
}

func TestScanProjectsEmpty(t *testing.T) {
	dir := t.TempDir()
	projects, err := transcript.ScanProjects(dir)
	if err != nil {
		t.Fatalf("ScanProjects on empty dir failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestScanSubagents(t *testing.T) {
	dir := t.TempDir()
	subagentDir := filepath.Join(dir, "subagents")
	if err := os.MkdirAll(subagentDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create two subagent files
	for _, name := range []string{"agent-abc123.jsonl", "agent-def456.jsonl"} {
		if err := os.WriteFile(filepath.Join(subagentDir, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	agents, err := transcript.ScanSubagents(subagentDir)
	if err != nil {
		t.Fatalf("ScanSubagents failed: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("expected 2 subagents, got %d", len(agents))
	}
}
