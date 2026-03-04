package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// releaseURL and downloadURL are vars so tests can override them.
var (
	releaseURL  = "https://api.github.com/repos/Curt-Park/claudeview/releases/latest"
	downloadURL = "https://github.com/Curt-Park/claudeview/releases/download/%s/claudeview-%s-%s"
)

var updateFlag bool

func init() {
	rootCmd.Flags().BoolVar(&updateFlag, "update", false, "Self-update to the latest GitHub release")

	origRunE := rootCmd.RunE
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if updateFlag {
			return runSelfUpdate()
		}
		return origRunE(cmd, args)
	}
}

// githubRelease is the minimal JSON shape from the GitHub releases API.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// fetchLatestRelease queries GitHub for the latest release tag.
func fetchLatestRelease() (string, error) {
	resp, err := http.Get(releaseURL) //nolint:gosec,noctx
	if err != nil {
		return "", fmt.Errorf("fetching latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("decoding release JSON: %w", err)
	}
	return rel.TagName, nil
}

// currentBinaryPath returns the resolved path of the running executable.
func currentBinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// downloadAndReplace fetches the platform binary for tag and atomically
// replaces the file at binPath.
func downloadAndReplace(tag, binPath string) error {
	url := fmt.Sprintf(downloadURL, tag, runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	// Preserve original permissions.
	info, err := os.Stat(binPath)
	if err != nil {
		return fmt.Errorf("stat current binary: %w", err)
	}

	// Write to temp file in same directory for atomic rename.
	dir := filepath.Dir(binPath)
	tmp, err := os.CreateTemp(dir, "claudeview-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Clean up temp file on error; clear path on success to keep new binary.
	defer func() {
		if tmpPath != "" {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	if err := os.Rename(tmpPath, binPath); err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}

	tmpPath = "" // prevent deferred removal of the new binary
	return nil
}

// runSelfUpdate checks for the latest release and updates if needed.
func runSelfUpdate() error {
	fmt.Println("Checking for updates...")

	tag, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	if AppVersion != "dev" && AppVersion == tag {
		fmt.Printf("Already up to date (%s).\n", AppVersion)
		return nil
	}

	fmt.Printf("Updating %s → %s...\n", AppVersion, tag)

	binPath, err := currentBinaryPath()
	if err != nil {
		return fmt.Errorf("locating binary: %w", err)
	}

	if err := downloadAndReplace(tag, binPath); err != nil {
		return err
	}

	fmt.Printf("Updated to %s.\n", tag)
	return nil
}
