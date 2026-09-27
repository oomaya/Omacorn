# Packet 005 — Update() dispatcher coverage

**Status:** active
**Branch:** `packet/005-update-dispatcher-coverage`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Close the coverage gap in the TUI's nerve center: `Update()` in `tui/model.go` sits at 51.5% after packet 004 (delivery report 2026-09-27). Unit-test the unhandled message branches.

### Repo context
- `tui/model.go` (`Update` function — architect to enumerate uncovered branches from the coverage report)
- `tui/model_test.go` (test patterns from packet 004 to mirror)

### Approved decisions
- Unit tests for message handling; golden tests remain packet 004's domain.

### Scope
`Update()` in `tui/model.go` only.

### Non-goals
- No behavior changes to `Update`. No view changes. No new message types.

### Acceptance criteria
1. All currently-uncovered branches of `Update()` identified and listed in builder notes.
2. Each message branch covered: valid messages, edge cases (unknown/nil messages).
3. `Update` coverage ≥80% (architect to confirm target after scoping).

### Validation
- `go test ./tui/ -coverprofile=/tmp/cov.out && go tool cover -func=/tmp/cov.out | grep -E "Update|total"`
- `go vet ./...`, `gofmt -l .`

### Expected output
Diff to `tui/model_test.go` (+ testdata if needed). Coverage before/after in builder notes.

---

### Builder notes (filled by builder)

**Plan approved:** [x]
**Implementation summary:**
Added `TestUpdateUncoveredBranches` in `tui/model_test.go`. Command stubs and a temp `HOME` sit in front of `systemctl`, `sudo`, `flatpak`, and `gecit`, so the handoff keys do not touch the live session. Returned probe commands are not executed.

**Uncovered branches found:**
Cover profile before this packet (33 of 68 statements in `Update`):
- `tab` landing on the logs tab (`activeTab == 1`, `loadLogsCmd`)
- `s` success, `s` error, and the status refresh
- `g` hypervisor warning, bare-metal handoff, start success, start error, and the status refresh
- `x` stop-all
- `p` error, enabled, and disabled, plus the status refresh
- `c` cleanup success, cleanup error, and the status refresh

Also exercised, per the packet, where Go cover had no separate else block: `ctrl+c`, a second `t` while already probing, `nil`, and an unknown message (dashboard and logs tab).

**Coverage before/after:**
- `Update`: 51.5% → 100.0%
- `tui` package: 64.0% → 83.1%
- Repo total (`go test ./... -coverprofile`): 41.3% → 45.4%

Validation: `gofmt -l .` clean, `go vet ./...` clean, `go test ./...` pass. `go test -race ./tui/` pass.

**Deviations:** none

### Review (filled by architect — read-only)

**Verdict:** `APPROVED` | `CHANGES REQUESTED`
**Notes:**
