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

func TestProjectActiveSessions(t *testing.T) {
	active := &model.Session{Status: model.StatusActive}
	ended := &model.Session{Status: model.StatusEnded}
	done := &model.Session{Status: model.StatusDone}

	tests := []struct {
		name     string
		sessions []*model.Session
		wantLen  int
	}{
		{"no sessions", nil, 0},
		{"mixed status active and ended", []*model.Session{active, ended, done}, 1},
		{"all active", []*model.Session{active, active}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &model.Project{Sessions: tc.sessions}
			got := p.ActiveSessions()
			if len(got) != tc.wantLen {
				t.Errorf("ActiveSessions() len = %d, want %d", len(got), tc.wantLen)
			}
			for _, s := range got {
				if s.Status != model.StatusActive {
					t.Errorf("ActiveSessions() returned session with status %q, want active", s.Status)
				}
			}
		})
	}
}
