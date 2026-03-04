package model

import (
	"fmt"
	"time"
)

// FormatAge converts a duration into a human-friendly string (e.g. "5m", "2h", "3d").
func FormatAge(d time.Duration) string {
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

// FormatTokenCount formats a token count as "1.5k", "1.5M", etc.
func FormatTokenCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1000:
		return fmt.Sprintf("%dk", n/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// FormatSize formats a byte count as a human-readable size string.
func FormatSize(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case b >= mb:
		return fmt.Sprintf("%.1fMB", float64(b)/mb)
	case b >= kb:
		return fmt.Sprintf("%.1fKB", float64(b)/kb)
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}
