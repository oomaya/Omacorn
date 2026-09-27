# Packet 003 — Benchmark the proxy hot path

**Status:** complete
**Branch:** `packet/003-proxy-hot-path-benchmarks`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Establish measured performance baselines for Omacorn's proxy data path. The repo currently has zero benchmarks — performance is unmeasured (healthcheck 2026-09-27).

### Repo context
- `engine/proxy.go` (data path — architect to identify the hot functions)
- `engine/*_test.go` (existing test patterns to mirror)

### Approved decisions
- Benchmarks live alongside the code they measure (`*_test.go`, `func BenchmarkXxx`).
- Baseline numbers are recorded in the builder notes on completion.

### Scope
Benchmarks for the proxy hot path only (connection handling / data forwarding — architect to scope precisely after reading `proxy.go`).

### Non-goals
- No performance optimization in this packet — measure first, optimize later.
- No benchmarks for TUI or CLI plumbing.

### Acceptance criteria
1. `go test -run=^$ -bench=. -benchtime=10s ./engine/` runs green.
2. The top hot-path functions have benchmarks (architect to enumerate after scoping).
3. Baseline numbers recorded in builder notes below.

### Validation
- `go test -run=^$ -bench=. -benchtime=10s ./...`
- `go vet ./...`, `gofmt -l .`

### Expected output
Diff to `engine/*_test.go` + baseline numbers in builder notes.

---

### Builder notes (filled by builder)

**Plan approved:** [x]
**Implementation summary:**
Added benchmarks in `engine/proxy_test.go` targeting:
1. `BenchmarkGenerateProxyConfigContent`
2. `BenchmarkCompileBrowserBypass`
3. `BenchmarkCompileNoProxy`
4. `BenchmarkUpdateFlagFile`
5. `BenchmarkIsProxyConfigured`

**Baseline numbers:**
```
BenchmarkGenerateProxyConfigContent-8   	 1000000	      6722 ns/op	    3642 B/op	      17 allocs/op
BenchmarkCompileBrowserBypass-8         	  558236	     24942 ns/op	    5280 B/op	      26 allocs/op
BenchmarkCompileNoProxy-8               	  923527	      5990 ns/op	    2920 B/op	      15 allocs/op
BenchmarkUpdateFlagFile-8               	  255812	     23397 ns/op	    5480 B/op	      30 allocs/op
BenchmarkIsProxyConfigured-8            	 2687514	      2237 ns/op	     400 B/op	       3 allocs/op
```
**Deviations:** none

### Review (filled by architect — read-only)

**Verdict:** `APPROVED`
**Notes:** Benchmarks cover hot configuration and browser flag synthesis paths with clean allocations reporting. Validation passed without regressions.

