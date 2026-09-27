# Packet 007 — Fix update.sh discovery (candidatePaths)

**Status:** active
**Branch:** `packet/007-update-discovery-fix`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Fix the update-tier drift documented in `.agents/ARCHITECTURE.md` (architect review 2026-09-27): `handleUpdate` in `main.go` fails to detect the local git checkout unless run from the repo root, because `candidatePaths` omits `~/Omacorn` — the actual repo location on this machine. Result: silent fallback to the Tier 2 curl path.

### Repo context
- `main.go` (`handleUpdate`, ~L435-442 — `candidatePaths`)
- `scripts/update.sh` (Tier 1 script — do not modify)
- `main_test.go` (`TestHandleUpdateScriptSuccess`, `TestHandleUpdateCurlPaths`, `chdirEmpty` patterns from packet 006 to extend)

### Approved decisions
- Dual-tier design is intentional and stays: Tier 1 (git checkout → `scripts/update.sh`), Tier 2 (curl fallback for standalone binaries).
- Candidate fix per architect: add `filepath.Join(home, "Omacorn")`; consider dynamic detection via `git rev-parse --show-toplevel` and `os.Executable()` lookup. Architect to choose the final approach in the plan.

### Scope
Path-discovery logic in `handleUpdate` only.

### Non-goals
- No changes to `scripts/update.sh` itself. No changes to Tier 2 curl behavior. No other `main.go` handlers.

### Acceptance criteria
1. Running `omacorn update` from any directory, with the checkout at `~/Omacorn` (or any detected git root), takes the Tier 1 script path.
2. Tier 2 curl fallback still works when no checkout is found.
3. Both tiers covered by tests — extend the packet-006 `chdirEmpty` / mock-script patterns.
4. **Human gate:** Master approves the builder's plan before implementation. This is the self-update mechanism — a bug here bricks the update path.

### Validation
- `go test -run 'TestHandleUpdate' ./... -v`
- `go test ./...`, `go vet ./...`, `gofmt -l .`

### Expected output
Diff to `main.go` + `main_test.go`. Manual verification notes: which directories the update was tested from, and which tier each took.

---

### Builder notes (filled by builder)

**Plan approved:** [x] (Master — full search order)
**Implementation summary:**
`handleUpdate` now asks `updateScriptCandidates` for directories and runs the first `scripts/update.sh` that is a real file. The curl body is unchanged. `scripts/update.sh` was not edited.

**Approach chosen:**
Search order, first hit wins, duplicates removed:
1. `.`
2. `git rev-parse --show-toplevel` from cwd
3. Walk up from the executable until `scripts/update.sh` exists
4. `~/Omacorn`, then the previous `~/Projects`, `~/src`, and `~/.local/src` entries

**Manual verification (cwd → tier taken):**
Fake checkouts only. The real `~/Omacorn/scripts/update.sh` was not executed (it git-pulls and rebuilds).

| cwd | Checkout present | Tier taken |
|---|---|---|
| empty temp dir | `$HOME/Omacorn/scripts/update.sh` | Tier 1, home script (`tier1-home`) |
| temp dir that itself has `scripts/update.sh` | that dir, and also `$HOME/Omacorn` | Tier 1, cwd script (`tier1-cwd`) |
| `repo/nested` inside a temp git repo | script at the git root | Tier 1, git root (`tier1-git`) |
| empty temp dir | script beside a stand-in executable | Tier 1, executable walk (`tier1-exe`) |
| empty temp dir | none | Tier 2 curl fallback |

`go test -run 'TestHandleUpdate' -v` pass. `gofmt -l .` clean, `go vet ./...` clean, `go test ./...` pass.

**Deviations:** none

### Review (filled by architect — read-only)

**Verdict:** `APPROVED` | `CHANGES REQUESTED`
**Notes:**
