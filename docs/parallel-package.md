---
title: "Parallel Package (internal/parallel)"
type: component
tags: [concurrency, utilities, internals]
---

# Parallel Package — `internal/parallel`

Generic concurrent map helper. Replaced 8 identical `errgroup` boilerplate blocks scattered across [[cmd-package]] and [[transcript-package]].

## Files

| File         | Purpose                                      |
|--------------|----------------------------------------------|
| `map.go`     | `Map[T, R]` — concurrent slice transformation |
| `map_test.go`| 4 correctness tests                          |

## API

```go
func Map[T any, R any](items []T, fn func(T) R) []R
```

Applies `fn` to each item concurrently. Results are returned **in input order**. Uses `errgroup` (golang.org/x/sync) limited to `runtime.NumCPU()` goroutines. `fn` is expected to handle its own errors — return a zero value on failure; `Map` never surfaces errors.

## Why It Exists

All 8 replaced sites followed the same pattern:

```go
results := make([]R, len(items))
g, _ := errgroup.WithContext(context.Background())
g.SetLimit(runtime.NumCPU())
for i, item := range items {
    g.Go(func() error {
        results[i] = fn(item)
        return nil
    })
}
g.Wait()
```

`parallel.Map` collapses that into a one-liner at each call site.

## Tests

- Basic correctness: `[]int{1,2,3,4,5}` → doubles
- Empty slice: returns empty (not nil)
- Order preservation: 100-element identity map
- Type variety: `[]string` → `[]int` (length mapping)

## Related

- [[architecture]] — listed in internal packages table
- [[transcript-package]] — `ScanProjects` uses `parallel.Map` for concurrent directory scanning
- [[provider-package]] — `sessionFromInfo`, `parseAgentsFromSession`, `GetProjects`, `GetSessions` all use `parallel.Map`
- [[cmd-package]] — `loadDataAsync` uses `parallel.Map` for slug-group and subagent turn loading
