package view

import (
	"fmt"
	"time"
)

func truncateHash(hash string) string {
	// Project hashes can be long path-encoded strings; truncate for display
	if len(hash) > 50 {
		return "…" + hash[len(hash)-49:]
	}
	return hash
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
