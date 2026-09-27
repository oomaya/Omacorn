# Architect Review: Packets 005 & 006

**Date:** 2026-09-27  
**Architect:** Antigravity  
**Builder:** Grok  
**Branches Reviewed:** `packet/005-update-dispatcher-coverage`, `packet/006-engine-coverage-gaps`  
**Integration Target:** `master` (Merged at `12eae3e`)

---

## 1. Executive Verdicts

| Packet | Scope | Statement Coverage Impact | Verdict | Merge Commit |
| :--- | :--- | :--- | :---: | :---: |
| **005** | `tui/model.go: Update()` dispatcher branches | `Update`: 51.5% → **100.0%**<br>`tui` pkg: 64.0% → **83.1%** | `APPROVED` | [`5a8a3a6`](file:///home/rand/Omacorn/tui/model_test.go) |
| **006** | `main.go` & `engine/` top 5 coverage gaps | `main` pkg: 1.5% → **59.1%**<br>`engine` pkg: 60.9% → **70.9%** | `APPROVED` | [`12eae3e`](file:///home/rand/Omacorn/main_test.go) |
| **Total** | Repository Total Coverage | Baseline: 41.3% → Post-005/006: **69.5%** | `APPROVED` | Clean Merge |

---

## 2. Packet 005 Assessment: `Update()` Dispatcher Coverage

### Verification & Findings
1. **100% Statement Coverage**: Every branch within `Update()` in [tui/model.go](file:///home/rand/Omacorn/tui/model.go#L102) is now executed.
2. **Branch Coverage Detail**:
   - Tab switching to logs (`activeTab == 1`, returning `loadLogsCmd`).
   - Hotkey dispatches: `s` (start spoofdpi, error handling, status refresh), `g` (gecit start, virt warnings, bare-metal bypass, error handling), `x` (stop-all), `p` (proxy toggle enable/disable/error), and `c` (emergency cleanup success/error).
   - Guardrails: `ctrl+c` quit dispatch, double-press `t` probe latching, nil message safety, and unknown message handling across tabs.
3. **Environment Isolation**:
   - `isolatePilotCommands` creates temporary executables ahead of `$PATH` (`systemctl`, `sudo`, `journalctl`, `flatpak`, `gecit`, `getcap`) and redirects `$HOME` to `t.TempDir()`.
   - Host session is completely protected from actual service restarts or proxy toggling during test runs.
4. **Tooling & Race Checks**:
   - `go test -v -race ./tui/` passed cleanly (0 data races).
   - `gofmt` and `go vet` clean.

---

## 3. Packet 006 Assessment: Engine & Main Coverage Gaps

### Verification & Findings
1. **Top 5 Gap Closure**:
   - `handleExclude`: 0% → **100.0%** (56/56 statements)
   - `handleUpdate`: 0% → **100.0%** (45/45 statements)
   - `handleStatus`: 0% → **100.0%** (32/32 statements)
   - `handleStart`: 0% → **100.0%** (25/25 statements)
   - `InstallSpoofDPI`: 0% → **92.0%** (23/25 statements)
2. **Process & Subprocess Isolation**:
   - Tests exercising `os.Exit` paths ([TestHandleUpdateScriptExit](file:///home/rand/Omacorn/main_test.go#L368), [TestHandleUpdateScriptNotExecutable](file:///home/rand/Omacorn/main_test.go#L394)) run cleanly via re-exec subprocess helpers (`OMACORN_WANT_HELPER=1`).
   - Mock binaries (`systemctl`, `getcap`, `spoofdpi`) and isolated directory structures thoroughly validate unit file generation, capability detection, and fallback paths.
3. **Tooling & Race Checks**:
   - `go test -v -race ./...` passed across all packages.
   - `gofmt` and `go vet` clean.

---

## 4. Architect Flag Deep-Dive: `handleUpdate` vs `scripts/update.sh`

### Architectural Review Question
> *"One flag for the architect: `handleUpdate` never runs the repo's `scripts/update.sh`. Might be intentional, might be drift — but the architecture map should say which. That's a review-step question, not a blocker."*

### Architectural Analysis & Root Cause
1. **Intended Design**:
   Omacorn's update mechanism was designed with a dual-tier strategy:
   - **Tier 1 (Source Repository Clone)**: When running from a developer or git-managed install, `handleUpdate()` delegates to [scripts/update.sh](file:///home/rand/Omacorn/scripts/update.sh). The script handles fast-forward git pulls (`git pull --ff-only origin master`), Go binary compilation, systemd daemon hot-reloading, and proxy re-synchronization.
   - **Tier 2 (Pre-compiled Binary Fallback)**: When running as a standalone binary without a local Git checkout, `handleUpdate()` falls back to downloading the latest pre-compiled static release from GitHub releases via `curl`.

2. **The Cause of Drift**:
   In [main.go:handleUpdate](file:///home/rand/Omacorn/main.go#L435-L442):
   ```go
   candidatePaths := []string{
       ".",
       filepath.Join(home, "Projects", "Omacorn"),
       filepath.Join(home, "Projects", "omacorn"),
       filepath.Join(home, "src", "Omacorn"),
       filepath.Join(home, "src", "omacorn"),
       filepath.Join(home, ".local", "src", "Omacorn"),
   }
   ```
   - On this machine, the repository is located at `~/Omacorn` (`$HOME/Omacorn`).
   - `~/Omacorn` is **not** in `candidatePaths`.
   - Consequently, running `omacorn update` from any directory other than the repo root (where `.` would match) fails to detect the local git repository and silently drifts into the Tier 2 curl fallback!
   - In Packet 006 unit tests, Grok purposefully used `chdirEmpty(t)` in `TestHandleUpdateCurlPaths` so that the test runner (which executes at the repo root where `.` exists) would not accidentally invoke the actual `scripts/update.sh`. `TestHandleUpdateScriptSuccess` explicitly validated the script execution path using a mocked `scripts/update.sh`.

3. **Architectural Resolution & Recommended Action**:
   - The design is **intentional**, but the path resolution logic suffered from **candidate path drift**.
   - **Follow-up Action**: Add `filepath.Join(home, "Omacorn")` and/or dynamic detection via `git rev-parse --show-toplevel` and executable directory lookup (`os.Executable()`).
   - Documented in `.agents/ARCHITECTURE.md` as non-blocking technical debt.

---

## 5. Current Repository Coverage Matrix (Post-Merge)

```
Package         Statements Covered / Total      Percentage
----------------------------------------------------------
dpipe (main)             81 / 137                 59.1%
dpipe/engine            144 / 203                 70.9%
dpipe/tui                98 / 118                 83.1%
----------------------------------------------------------
Repository Total        323 / 465                 69.5%
```

---

## 6. Pipeline Progression & Next Steps

1. **Packet 005 Task Record**: Completed, approved, and merged ([.agents/tasks/005-update-dispatcher-coverage.md](file:///home/rand/Omacorn/.agents/tasks/005-update-dispatcher-coverage.md)).
2. **Packet 006 Task Record**: Completed, approved, and merged ([.agents/tasks/006-engine-coverage-gaps.md](file:///home/rand/Omacorn/.agents/tasks/006-engine-coverage-gaps.md)).
3. **Clean Master Branch**: Master is now verified green at commit `12eae3e`.
4. **Recommended Next Packet**: Packet 007 — `update.sh` discovery robustness (`candidatePaths` fix) or remaining gap coverage for `ProbeURL` / `launchTUI` / `handleRunClean`.
