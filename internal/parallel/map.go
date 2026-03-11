package parallel

import (
	"context"
	"runtime"

	"golang.org/x/sync/errgroup"
)

// Map applies fn to each item in items concurrently and returns results in order.
// It uses up to runtime.NumCPU() goroutines.
func Map[T any, R any](items []T, fn func(T) R) []R {
	results := make([]R, len(items))
	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(runtime.NumCPU())
	for i, item := range items {
		g.Go(func() error {
			results[i] = fn(item)
			return nil
		})
	}
	g.Wait() //nolint:errcheck
	return results
}
