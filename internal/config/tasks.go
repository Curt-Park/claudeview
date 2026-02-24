package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TaskEntry represents a task from ~/.claude/tasks/<session>/*.json.
type TaskEntry struct {
	ID          string   `json:"id"`
	Subject     string   `json:"subject"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Owner       string   `json:"owner"`
	BlockedBy   []string `json:"blockedBy"`
	Blocks      []string `json:"blocks"`
	ActiveForm  string   `json:"activeForm"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// LoadTasks loads all tasks for a session from ~/.claude/tasks/<session>/.
func LoadTasks(claudeDir, sessionID string) ([]TaskEntry, error) {
	dir := filepath.Join(claudeDir, "tasks", sessionID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []TaskEntry
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var t TaskEntry
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
