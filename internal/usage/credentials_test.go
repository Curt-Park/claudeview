package usage_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Curt-Park/claudeview/internal/usage"
)

func TestReadToken(t *testing.T) {
	dir := t.TempDir()
	creds := map[string]any{
		"claudeAiOauth": map[string]any{
			"accessToken": "test-token-abc",
		},
	}
	data, _ := json.Marshal(creds)
	path := filepath.Join(dir, ".credentials.json")
	os.WriteFile(path, data, 0600)

	token, err := usage.ReadToken(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "test-token-abc" {
		t.Errorf("got %q, want %q", token, "test-token-abc")
	}
}

func TestReadTokenMissingFile(t *testing.T) {
	_, err := usage.ReadToken("/nonexistent/.credentials.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadTokenEmptyToken(t *testing.T) {
	dir := t.TempDir()
	creds := map[string]any{"claudeAiOauth": map[string]any{"accessToken": ""}}
	data, _ := json.Marshal(creds)
	os.WriteFile(filepath.Join(dir, ".credentials.json"), data, 0600)

	_, err := usage.ReadToken(filepath.Join(dir, ".credentials.json"))
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
