package model

import "time"

// Project represents a Claude Code project directory.
type Project struct {
	Hash     string
	Path     string
	Sessions []*Session
	LastSeen time.Time
}

// SessionCount returns the number of sessions in this project.
func (p *Project) SessionCount() int {
	return len(p.Sessions)
}
