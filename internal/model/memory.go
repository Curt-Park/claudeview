package model

import "time"

// Memory represents a single memory file in a project's memory/ directory.
type Memory struct {
	Name    string
	Path    string
	Title   string // first # heading in the file, empty if none
	Size    int64
	ModTime time.Time
}

// SizeStr formats the file size in a human-readable form.
func (m *Memory) SizeStr() string {
	return FormatSize(m.Size)
}

// LastModified returns the file modification time as a human-readable age string.
func (m *Memory) LastModified() string {
	return FormatAge(time.Since(m.ModTime))
}
