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

// ScanSessions scans a project directory for session JSONL files.
func ScanSessions(projectPath string) ([]SessionInfo, error) {
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil, err
	}

	var sessions []SessionInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".jsonl")
		filePath := filepath.Join(projectPath, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}

		subagentDir := filepath.Join(projectPath, id, "subagents")
		if _, err := os.Stat(subagentDir); os.IsNotExist(err) {
			subagentDir = ""
		}

		sessions = append(sessions, SessionInfo{
			ID:          id,
			FilePath:    filePath,
			SubagentDir: subagentDir,
			ModTime:     info.ModTime(),
		})
	}

	// Sort by mod time descending
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModTime.After(sessions[j].ModTime)
	})

	return sessions, nil
}

// ScanSubagents scans the subagents/ directory of a session.
func ScanSubagents(subagentDir string) ([]SessionInfo, error) {
	if subagentDir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(subagentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var agents []SessionInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".jsonl")
		filePath := filepath.Join(subagentDir, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}
		agents = append(agents, SessionInfo{
			ID:       id,
			FilePath: filePath,
			ModTime:  info.ModTime(),
		})
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ModTime.Before(agents[j].ModTime)
	})

	return agents, nil
}
