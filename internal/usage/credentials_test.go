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
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("failed to marshal creds: %v", err)
	}
	path := filepath.Join(dir, ".credentials.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write credentials file: %v", err)
	}

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
	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("failed to marshal creds: %v", err)
	}
	path := filepath.Join(dir, ".credentials.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write credentials file: %v", err)
	}

	_, err = usage.ReadToken(path)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestReadTokenMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".credentials.json")
	if err := os.WriteFile(path, []byte("{bad json"), 0600); err != nil {
		t.Fatalf("failed to write credentials file: %v", err)
	}

	_, err := usage.ReadToken(path)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
