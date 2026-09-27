# Packet 006 — Engine coverage gaps

**Status:** active
**Branch:** `packet/006-engine-coverage-gaps`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Repo total coverage is 41.3% after packet 004 (delivery report 2026-09-27), driven mostly by `tui/` work. Measure full per-function coverage of `engine/` and `main.go`, then close the largest gaps.

### Repo context
- `engine/` (`engine.go`, `proxy.go`, `exclusion.go`, `gecit.go`, `spoofdpi.go`)
- `main.go`, `main_test.go`
- `engine/*_test.go` (existing patterns to mirror)

### Approved decisions
- Tests mirror existing `engine/*_test.go` patterns.

### Scope
`engine/` + `main.go` test coverage only.

### Non-goals
- No behavior changes. No TUI work (packets 004/005). No benchmarks (packet 003).

### Acceptance criteria
1. Full per-function coverage report produced and attached to builder notes.
2. Tests added for the largest uncovered functions (architect to prioritize top 5 by statement count after measuring).
3. Repo total coverage improved — before/after recorded in builder notes.

### Validation
- `go test ./... -coverprofile=/tmp/cov.out && go tool cover -func=/tmp/cov.out | tail -5`
- `go vet ./...`, `gofmt -l .`

### Expected output
Diff to `engine/*_test.go`, `main_test.go`. Coverage report + before/after in builder notes.

---

### Builder notes (filled by builder)

**Plan approved:** [x]
**Implementation summary:**
Measured uncovered statements on `d3aa6c7` and tested the top 5: `handleExclude` (56), `handleUpdate` (45), `handleStatus` (32), `handleStart` (25), `InstallSpoofDPI` (25). New tests live in `main_test.go` and `engine/spoofdpi_test.go`. `HOME` is a temp dir, `PATH` leads with stubs, and `handleUpdate` runs from an empty cwd so `scripts/update.sh` in this repo is never executed. `os.Exit` paths run in a subprocess.

**Coverage report (top gaps):**
Baseline uncovered-statement ranking for `engine/` + `main.go` (same `go test ./... -coverprofile` method as the 41.3% figure). Top 5 are marked.

```
uncov  function
   56  main.go handleExclude          <- top 5
   45  main.go handleUpdate           <- top 5
   32  main.go handleStatus           <- top 5
   25  main.go handleStart            <- top 5
   25  engine/spoofdpi.go InstallSpoofDPI <- top 5
   23  main.go handleRunClean
   22  main.go main
   18  engine/engine.go ProbeURL
   16  main.go handleProxy
   16  engine/gecit.go InstallGecit
   15  main.go handleTest
   13  main.go handleInstall
   12  engine/gecit.go EmergencyCleanup
   10  engine/spoofdpi.go FindSpoofDPIBinary
```

Per-function result after this packet (`engine/` + `main.go`):

```
100.0%   3/3    engine/engine.go RunCmd
 58.3%   7/12   engine/engine.go DetectHypervisor
100.0%   2/2    engine/engine.go IsSpoofDPIRunning
100.0%   2/2    engine/engine.go IsGecitRunning
 75.0%  24/32   engine/engine.go GetSystemStatus
  0.0%   0/18   engine/engine.go ProbeURL
100.0%  10/10   engine/engine.go GetProbeTarget
 75.0%   3/4    engine/exclusion.go getExclusionsConfigPath
 87.5%  14/16   engine/exclusion.go LoadUserExclusions
 86.7%  13/15   engine/exclusion.go SaveUserExclusions
100.0%  14/14   engine/exclusion.go GetAllExclusions
 85.7%  12/14   engine/exclusion.go AddExclusion
 81.2%  13/16   engine/exclusion.go RemoveExclusion
 66.7%   4/6    engine/exclusion.go ResetExclusions
100.0%   2/2    engine/exclusion.go CompileNoProxy
100.0%  11/11   engine/exclusion.go CompileBrowserBypass
  0.0%   0/16   engine/gecit.go InstallGecit
  0.0%   0/6    engine/gecit.go StartGecit
  0.0%   0/6    engine/gecit.go StopGecit
  0.0%   0/12   engine/gecit.go EmergencyCleanup
100.0%   2/2    engine/proxy.go GenerateProxyConfigContent
 75.0%   3/4    engine/proxy.go getEnvDPath
 83.3%   5/6    engine/proxy.go IsProxyConfigured
 88.9%  16/18   engine/proxy.go updateFlagFile
 84.8%  28/33   engine/proxy.go syncBrowserFlags
 83.3%  15/18   engine/proxy.go EnableSystemProxy
 93.3%  14/15   engine/proxy.go DisableSystemProxy
100.0%   5/5    engine/proxy.go ToggleSystemProxy
  0.0%   0/3    engine/proxy.go SyncProxyIfEnabled
 90.0%   9/10   engine/spoofdpi.go FindSpoofDPIBinary
 92.0%  23/25   engine/spoofdpi.go InstallSpoofDPI
  0.0%   0/1    engine/spoofdpi.go StartSpoofDPI
  0.0%   0/1    engine/spoofdpi.go StopSpoofDPI
  0.0%   0/22   main.go main
100.0%   4/4    main.go getBinName
  0.0%   0/5    main.go launchTUI
  0.0%   0/2    main.go printUsage
  0.0%   0/13   main.go handleInstall
100.0%  25/25   main.go handleStart
  0.0%   0/6    main.go handleStop
100.0%  32/32   main.go handleStatus
  0.0%   0/16   main.go handleProxy
  0.0%   0/15   main.go handleTest
  0.0%   0/6    main.go handleLogs
  0.0%   0/4    main.go handleCleanup
100.0%  56/56   main.go handleExclude
  0.0%   0/23   main.go handleRunClean
100.0%  45/45   main.go handleUpdate
```

`StartGecit`, `StopGecit`, `StartSpoofDPI`, `StopSpoofDPI`, and `SyncProxyIfEnabled` stay at 0% in this profile because `go test ./... -coverprofile` counts a package only from its own tests. `handleStart` and `handleExclude` do call them. They were outside the top 5.

**Coverage before/after:**
- Repo total: 41.3% → 65.4%
- `main` package: 1.5% → 59.1%
- `engine` package: 60.9% → 70.9%
- Top 5 after: `handleExclude` 100%, `handleUpdate` 100%, `handleStatus` 100%, `handleStart` 100%, `InstallSpoofDPI` 92.0% (23/25)

Validation: `gofmt -l .` clean, `go vet ./...` clean, `go test ./...` pass. `go test -race . ./engine/` pass.

**Deviations:** Two `InstallSpoofDPI` statements stay uncovered on this VMware guest without a production change: the `os.UserHomeDir` error return, and the `!isVirt && hasRawCap` decoy `ExecStart` (`systemd-detect-virt` reports `vmware`, and `/sys/class/dmi/id/sys_vendor` is `VMware, Inc.`). `FindSpoofDPIBinary`'s `/usr/local/bin/spoofdpi` success return is also uncovered because that file is absent (9/10).

### Review (filled by architect — read-only)

**Verdict:** `APPROVED` | `CHANGES REQUESTED`
**Notes:**
