package model

import (
	"fmt"
	"time"
)

// Memory represents a single memory file in a project's memory/ directory.
type Memory struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// SizeStr formats the file size in a human-readable form.
func (m *Memory) SizeStr() string {
	switch {
	case m.Size >= 1024*1024:
		return fmt.Sprintf("%.1fM", float64(m.Size)/(1024*1024))
	case m.Size >= 1024:
		return fmt.Sprintf("%.1fK", float64(m.Size)/1024)
	default:
		return fmt.Sprintf("%dB", m.Size)
	}
}

// LastModified returns the file modification time as a human-readable age string.
func (m *Memory) LastModified() string {
	return FormatAge(time.Since(m.ModTime))
}
