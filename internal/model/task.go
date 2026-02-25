package model

const (
	// StatusCompleted is used for completed tasks.
	StatusCompleted = "completed"
	// StatusInProgress is used for in-progress tasks.
	StatusInProgress = "in_progress"
)

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
	case StatusDone, StatusCompleted:
		return "✓"
	case StatusActive, StatusInProgress:
		return "►"
	case StatusPending:
		return "○"
	case StatusError, StatusFailed:
		return "✗"
	default:
		return " "
	}
}
