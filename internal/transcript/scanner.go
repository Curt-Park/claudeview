package transcript

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ProjectInfo holds metadata about a discovered project directory.
type ProjectInfo struct {
	Hash     string
	Path     string
	Sessions []SessionInfo
	LastSeen time.Time
}

// SessionInfo holds metadata about a discovered session JSONL file.
type SessionInfo struct {
	ID          string
	FilePath    string
	SubagentDir string // path to subagents/ directory if it exists
	ModTime     time.Time
}

// ScanProjects scans ~/.claude/projects/ and returns all projects.
func ScanProjects(claudeDir string) ([]ProjectInfo, error) {
	projectsDir := filepath.Join(claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var projects []ProjectInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		hash := e.Name()
		projectPath := filepath.Join(projectsDir, hash)
		info, err := e.Info()
		if err != nil {
			continue
		}

		sessions, err := ScanSessions(projectPath)
		if err != nil {
			continue
		}

		lastSeen := info.ModTime()
		for _, s := range sessions {
			if s.ModTime.After(lastSeen) {
				lastSeen = s.ModTime
			}
		}

		projects = append(projects, ProjectInfo{
			Hash:     hash,
			Path:     projectPath,
			Sessions: sessions,
			LastSeen: lastSeen,
		})
	}

	// Sort by last activity descending
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastSeen.After(projects[j].LastSeen)
	})

	return projects, nil
}

// scanJSONLFiles reads a directory for .jsonl files and returns SessionInfo entries.
// If descending is true, entries are sorted newest-first; otherwise oldest-first.
// If includeSubagentDir is true, each entry's SubagentDir is populated.
func scanJSONLFiles(dir string, descending bool, includeSubagentDir bool) ([]SessionInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []SessionInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".jsonl")
		filePath := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}

		subagentDir := ""
		if includeSubagentDir {
			candidate := filepath.Join(dir, id, "subagents")
			if _, err := os.Stat(candidate); err == nil {
				subagentDir = candidate
			}
		}

		sessions = append(sessions, SessionInfo{
			ID:          id,
			FilePath:    filePath,
			SubagentDir: subagentDir,
			ModTime:     info.ModTime(),
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		if descending {
			return sessions[i].ModTime.After(sessions[j].ModTime)
		}
		return sessions[i].ModTime.Before(sessions[j].ModTime)
	})

	return sessions, nil
}

// ScanSessions scans a project directory for session JSONL files.
func ScanSessions(projectPath string) ([]SessionInfo, error) {
	return scanJSONLFiles(projectPath, true, true)
}

// ScanSubagents scans the subagents/ directory of a session.
func ScanSubagents(subagentDir string) ([]SessionInfo, error) {
	if subagentDir == "" {
		return nil, nil
	}
	return scanJSONLFiles(subagentDir, false, false)
}

// CountSubagents returns the number of subagent JSONL files in the given directory.
func CountSubagents(subagentDir string) int {
	if subagentDir == "" {
		return 0
	}
	entries, err := os.ReadDir(subagentDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonl") {
			count++
		}
	}
	return count
}
