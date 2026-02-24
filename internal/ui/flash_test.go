package ui_test

import (
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/ui"
)

func TestFlashSetAndExpiry(t *testing.T) {
	f := ui.FlashModel{}
	f.Set("hello", ui.FlashInfo, 50*time.Millisecond)

	if f.Message != "hello" {
		t.Errorf("Message = %q, want %q", f.Message, "hello")
	}
	if f.IsExpired() {
		t.Error("expected not expired immediately after Set()")
	}

	time.Sleep(60 * time.Millisecond)
	if !f.IsExpired() {
		t.Error("expected expired after duration elapsed")
	}
}

func TestFlashClear(t *testing.T) {
	f := ui.FlashModel{}
	f.Set("msg", ui.FlashInfo, time.Minute)
	f.Clear()
	if f.Message != "" {
		t.Errorf("Message after Clear = %q, want empty", f.Message)
	}
	if !f.IsExpired() {
		t.Error("empty message should be considered expired")
	}
}

func TestFlashLevels(t *testing.T) {
	f := ui.FlashModel{}
	f.Set("error msg", ui.FlashError, time.Minute)
	if f.Level != ui.FlashError {
		t.Errorf("Level = %v, want FlashError", f.Level)
	}
}
