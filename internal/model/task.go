package model

// Task represents a Claude Code task from ~/.claude/tasks/<session>/.
type Task struct {
	ID          string
	SessionID   string
	Subject     string
	Description string
	Status      Status
	Owner       string
	BlockedBy   []string
	Blocks      []string
	ActiveForm  string
}

// StatusIcon returns an emoji/char for the task status.
func (t *Task) StatusIcon() string {
	switch t.Status {
	case StatusDone, "completed":
		return "✓"
	case StatusActive, "in_progress":
		return "►"
	case StatusPending:
		return "○"
	case StatusError, StatusFailed:
		return "✗"
	default:
		return " "
	}
}
