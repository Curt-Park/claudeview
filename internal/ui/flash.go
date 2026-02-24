package ui

import "time"

// FlashLevel indicates the severity of a flash message.
type FlashLevel int

const (
	FlashInfo FlashLevel = iota
	FlashError
)

// FlashModel shows temporary status messages.
type FlashModel struct {
	Message   string
	Level     FlashLevel
	ExpiresAt time.Time
	Width     int
}

// Set sets a flash message with a duration.
func (f *FlashModel) Set(msg string, level FlashLevel, d time.Duration) {
	f.Message = msg
	f.Level = level
	f.ExpiresAt = time.Now().Add(d)
}

// Clear clears the flash message.
func (f *FlashModel) Clear() {
	f.Message = ""
}

// IsExpired returns true if the message has expired.
func (f *FlashModel) IsExpired() bool {
	if f.Message == "" {
		return true
	}
	return time.Now().After(f.ExpiresAt)
}

// View renders the flash bar.
func (f FlashModel) View() string {
	if f.Message == "" || f.IsExpired() {
		return StyleFlash.Width(f.Width).Render("")
	}
	if f.Level == FlashError {
		return StyleFlashError.Width(f.Width).Render(f.Message)
	}
	return StyleFlash.Width(f.Width).Render(f.Message)
}
