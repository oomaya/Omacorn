# Packet 004 — Golden tests for the TUI

**Status:** complete
**Branch:** `packet/004-tui-golden-tests`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Cover the `tui/` package, currently at 0% coverage (`model.go`: `loadLogsCmd`, `runProbesCmd`, `Update`; `views.go`: `View`, `renderDashboard`, `renderLogs`, `renderDiagnostics`). Total repo coverage is 27.6% (healthcheck 2026-09-27) — this packet attacks the biggest single gap.

### Repo context
- `tui/model.go`, `tui/views.go`, `tui/style.go`
- Note: `github.com/charmbracelet/x/exp/golden` is already a dependency — use it for view golden tests.

### Approved decisions
- Golden-file tests for view rendering; direct unit tests for `Update` message handling.

### Scope
`tui/` package only.

### Non-goals
- No TUI behavior changes. No engine coverage work (separate packet).

### Acceptance criteria
1. `Update` message handling covered for the main message types.
2. Golden tests for `View`, `renderDashboard`, `renderLogs`, `renderDiagnostics`.
3. `tui/` coverage materially improved (architect to set the target after scoping).

### Validation
- `go test ./tui/ -coverprofile=/tmp/cov.out && go tool cover -func=/tmp/cov.out | tail -3`
- `go vet ./...`, `gofmt -l .`

### Expected output
Diff to `tui/*_test.go` + testdata golden files.

---

### Builder notes (filled by builder)

**Plan approved:** [x]
**Implementation summary:**
Created `tui/model_test.go` and associated golden files under `tui/testdata/`:
1. Golden snapshot tests for `View()`, `renderDashboard()`, `renderLogs()`, `renderDiagnostics()`.
2. Unit test coverage for `Init()` and `Update()` with `WindowSizeMsg`, `logsMsg`, `probeResultMsg`, `tickMsg`, navigation keys `1`-`3`/`tab`, probe trigger `t`, and quit key `q`.

**Test results:**
- `tui/` package statement coverage jumped from 0.0% to 64.0%.
- Total repository coverage increased from 27.6% to 40.4%.
- All tests pass (`go test -v ./...` green).
- `gofmt -l .` clean and `go vet ./...` clean.

**Deviations:** none

### Review (filled by architect — read-only)

**Verdict:** `APPROVED`
**Notes:** Golden tests established for all view modes using `github.com/charmbracelet/x/exp/golden`. Core update message dispatcher covered, elevating package coverage from 0% to 64%.

