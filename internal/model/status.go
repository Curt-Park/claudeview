package model

// Status represents the current state of an agent or task.
type Status string

const (
	StatusActive    Status = "active"
	StatusThinking  Status = "thinking"
	StatusReading   Status = "reading"
	StatusExecuting Status = "executing"
	StatusDone      Status = "done"
	StatusEnded     Status = "ended"
	StatusError     Status = "error"
	StatusFailed    Status = "failed"
	StatusRunning   Status = "running"
	StatusPending   Status = "pending"

	// StatusCompleted is used for completed agents/tasks.
	StatusCompleted Status = "completed"
)
