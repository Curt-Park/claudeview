package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestProjectSessionCount(t *testing.T) {
	tests := []struct {
		name     string
		sessions []*model.Session
		want     int
	}{
		{"zero sessions", nil, 0},
		{"one session", []*model.Session{{}}, 1},
		{"three sessions", []*model.Session{{}, {}, {}}, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Project{Sessions: tc.sessions}
			if got := p.SessionCount(); got != tc.want {
				t.Errorf("SessionCount() = %d, want %d", got, tc.want)
			}
		})
	}
}
