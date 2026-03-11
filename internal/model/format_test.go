package model_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/model"
)

func TestFormatTokenInOutCache(t *testing.T) {
	tests := []struct {
		in, cache, out int
		want           string
	}{
		{50000, 7400000, 26000, "50k+7.4M/26k"},
		{243000, 0, 26000, "243k/26k"}, // no cache: omit +section
		{0, 0, 0, "0/0"},
	}
	for _, tt := range tests {
		got := model.FormatTokenInOutCache(tt.in, tt.cache, tt.out)
		if got != tt.want {
			t.Errorf("FormatTokenInOutCache(%d,%d,%d) = %q, want %q",
				tt.in, tt.cache, tt.out, got, tt.want)
		}
	}
}
