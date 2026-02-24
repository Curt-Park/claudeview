package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Curt-Park/claudeview/internal/config"
)

func TestLoadTasks_Valid(t *testing.T) {
	dir := t.TempDir()
	sessionID := "session-abc123"
	taskDir := filepath.Join(dir, "tasks", sessionID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}

	tasks := []string{
		`{"id":"1","subject":"Explore","status":"completed","blockedBy":[]}`,
		`{"id":"2","subject":"Implement","status":"in_progress","blockedBy":["1"]}`,
	}
	for i, content := range tasks {
		name := filepath.Join(taskDir, filepath.Base(filepath.Join(taskDir, string(rune('a'+i))+".json")))
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result, err := config.LoadTasks(dir, sessionID)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len = %d, want 2", len(result))
	}
}

func TestLoadTasks_MissingDirectory(t *testing.T) {
	result, err := config.LoadTasks(t.TempDir(), "nonexistent-session")
	if err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestLoadTasks_SkipsNonJSON(t *testing.T) {
	dir := t.TempDir()
	sessionID := "session-xyz"
	taskDir := filepath.Join(dir, "tasks", sessionID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}

	// One valid JSON, one non-JSON file
	if err := os.WriteFile(filepath.Join(taskDir, "task1.json"), []byte(`{"id":"1","subject":"Test","status":"pending"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "notes.txt"), []byte("not a task"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := config.LoadTasks(dir, sessionID)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("len = %d, want 1 (non-JSON should be skipped)", len(result))
	}
}

func TestLoadTasks_SkipsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	sessionID := "session-bad"
	taskDir := filepath.Join(dir, "tasks", sessionID)
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(taskDir, "bad.json"), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taskDir, "good.json"), []byte(`{"id":"1","subject":"OK","status":"pending"}`), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := config.LoadTasks(dir, sessionID)
	if err != nil {
		t.Fatalf("LoadTasks: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("len = %d, want 1 (malformed JSON should be skipped)", len(result))
	}
}
