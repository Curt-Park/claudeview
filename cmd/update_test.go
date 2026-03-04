package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchLatestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	defer srv.Close()

	old := releaseURL
	releaseURL = srv.URL
	defer func() { releaseURL = old }()

	tag, err := fetchLatestRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.2.3" {
		t.Fatalf("got tag %q, want %q", tag, "v1.2.3")
	}
}

func TestFetchLatestReleaseHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	old := releaseURL
	releaseURL = srv.URL
	defer func() { releaseURL = old }()

	_, err := fetchLatestRelease()
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestDownloadAndReplace(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new-binary-content"))
	}))
	defer srv.Close()

	old := downloadURL
	downloadURL = srv.URL + "/%s/claudeview-%s-%s"
	defer func() { downloadURL = old }()

	// Create a fake current binary.
	dir := t.TempDir()
	binPath := filepath.Join(dir, "claudeview")
	if err := os.WriteFile(binPath, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := downloadAndReplace("v1.2.3", binPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new-binary-content" {
		t.Fatalf("got %q, want %q", data, "new-binary-content")
	}

	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("got permissions %o, want 0755", info.Mode().Perm())
	}
}

func TestRunSelfUpdateAlreadyCurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	defer srv.Close()

	oldRelease := releaseURL
	releaseURL = srv.URL
	defer func() { releaseURL = oldRelease }()

	oldVersion := AppVersion
	AppVersion = "v1.0.0"
	defer func() { AppVersion = oldVersion }()

	// Should return nil without attempting a download.
	if err := runSelfUpdate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
