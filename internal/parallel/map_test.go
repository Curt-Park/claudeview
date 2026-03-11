package parallel_test

import (
	"testing"

	"github.com/Curt-Park/claudeview/internal/parallel"
)

func TestMapBasic(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	got := parallel.Map(items, func(n int) int { return n * 2 })
	want := []int{2, 4, 6, 8, 10}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("result[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestMapEmpty(t *testing.T) {
	got := parallel.Map([]string{}, func(s string) int { return len(s) })
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestMapPreservesOrder(t *testing.T) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}
	got := parallel.Map(items, func(n int) int { return n })
	for i, v := range got {
		if v != i {
			t.Errorf("result[%d] = %d, want %d", i, v, i)
		}
	}
}

func TestMapStrings(t *testing.T) {
	items := []string{"hello", "world", "foo"}
	got := parallel.Map(items, func(s string) int { return len(s) })
	want := []int{5, 5, 3}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("result[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
