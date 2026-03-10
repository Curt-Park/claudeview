package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestResourceHistoryExists(t *testing.T) {
	if model.ResourceHistory == "" {
		t.Error("ResourceHistory should be non-empty")
	}
	if string(model.ResourceHistory) != "history" {
		t.Errorf("expected 'history', got %q", model.ResourceHistory)
	}
}
