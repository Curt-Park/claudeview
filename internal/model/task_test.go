package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestTaskStatusIcon(t *testing.T) {
	tests := []struct {
		name   string
		status model.Status
		want   string
	}{
		{"done", model.StatusDone, "\u2713"},
		{"completed alias", "completed", "\u2713"},
		{"in_progress alias", "in_progress", "\u25ba"},
		{"active", model.StatusActive, "\u25ba"},
		{"pending", model.StatusPending, "\u25cb"},
		{"error", model.StatusError, "\u2717"},
		{"failed", model.StatusFailed, "\u2717"},
		{"unknown returns space", "something_else", " "},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := &model.Task{Status: tc.status}
			if got := task.StatusIcon(); got != tc.want {
				t.Errorf("StatusIcon() = %q, want %q", got, tc.want)
			}
		})
	}
}
