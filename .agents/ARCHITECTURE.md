# ARCHITECTURE.md — Omacorn

## What it is
Omacorn is a sovereign, local packet-shaping daemon and cyberpunk Bubble Tea TUI for Linux that evades DPI (Deep Packet Inspection) and SNI-based middlebox filtering without remote proxy servers, third-party VPN hops, or browser extensions. It coordinates two evasion engines under an aviation-grade Pilot Flying / Pilot Monitoring (PF/PM) mutual exclusion model, automatically synchronizes desktop proxy exemptions via `environment.d` and browser flags, and provides a zero-dependency static binary with self-update capabilities.

## Architecture Model: Pilot & Co-Pilot (PF/PM)
Omacorn prevents packet double-mangling through strict mutual exclusion:
1. **Pilot 1 (SpoofDPI) — Main Pilot / Pilot Flying (PF):**
   - User-space daemon listening on `127.0.0.1:8080`.
   - Injects out-of-window low-TTL (`TTL=8`) fake ClientHello decoy packets via `CAP_NET_RAW`.
   - Hypervisor-safe: operates cleanly in virtual machines (VMware/QEMU) by falling back to user-space TCP fragmentation when raw socket capabilities or bare-metal interfaces are unavailable.
2. **Pilot 2 (Gecit) — Co-Pilot / Standby (PM):**
   - Kernel-space eBPF `sock_ops` engine attached to root cgroup.
   - Synchronous packet manipulation with hardened MSS clamping (`--mss 88`, restored after 600 bytes) to prevent hypervisor/GPU memory panics.
   - Dedicated for bare-metal performance with root privileges (`CAP_BPF`, `CAP_PERFMON`, `CAP_SYS_ADMIN`).

## Layout & File Responsibilities

### Entry Point & Dispatcher
- `main.go` — Root CLI command dispatcher (`start`, `stop`, `status`, `proxy`, `test`, `logs`, `install`, `cleanup`, `exclude`, `run-clean`, `update`), interactive Bubble Tea TUI launcher, and dual-tier update discovery engine.
- `main_test.go` — Unit tests for CLI handlers, subprocess execution, status formatters, and candidate discovery paths.

### Engine Subsystem (`engine/`)
- `engine/engine.go` — Core system inspection primitives: command execution (`RunCmd`), hypervisor/virtualization detection (`DetectHypervisor`), engine process status probes (`IsSpoofDPIRunning`, `IsGecitRunning`, `GetSystemStatus`), and HTTP latency probing (`ProbeURL`, `GetProbeTarget`).
- `engine/proxy.go` — Desktop proxy integration manager: generates `~/.config/environment.d/20-omacorn-proxy.conf`, synchronizes browser runtime flags (Brave, Chromium, Flatpak sandboxes), and toggles desktop environment proxy settings.
- `engine/spoofdpi.go` — Pilot 1 manager: locates `spoofdpi` binaries, inspects network capabilities, generates and writes systemd user service units (`~/.config/systemd/user/dpipe-spoofdpi.service`), and controls daemon lifecycle.
- `engine/exclusion.go` — Split-tunneling exclusion subsystem: loads and persists user bypass rules (`~/.config/omacorn/exclusions.conf`), merges built-in network defaults (Google Auth, Telegram MTProto, RFC 1918 subnets), and compiles environment `no_proxy` strings and browser `--proxy-bypass-list` arguments.
- `engine/gecit.go` — Pilot 2 manager: controls root-level `dpipe-gecit.service` systemd unit, manages eBPF attachment lifecycle, and executes emergency network restoration (`EmergencyCleanup`).
- `engine/engine_test.go` — Unit tests for target probe resolution, command runners, and hypervisor detection.
- `engine/exclusion_test.go` — Unit tests for exclusion persistence, file formatting, and proxy list compilation.
- `engine/proxy_test.go` — Unit tests for desktop environment files, browser flags, hot-path proxy configuration benchmarks, and state toggling.
- `engine/spoofdpi_test.go` — Unit tests for SpoofDPI unit generation, virtualization detection branching, and binary discovery.

### Terminal UI Subsystem (`tui/`)
- `tui/model.go` — Bubble Tea state model (`Model`): manages active navigation tabs, status polling ticks, probe latches, and keyboard shortcut event dispatching.
- `tui/views.go` — Terminal UI presentation layer: renders cyber-styled headers, system flight board, telemetry panels, scrollable `journalctl` log viewports, and security diagnostic tables.
- `tui/style.go` — Styling design tokens and Lip Gloss layout helpers (Tokyo Night palette, status badges, neon accents, bordered boxes).
- `tui/model_test.go` — Unit tests for Bubble Tea initialization, message update dispatching, and golden file snapshot tests for dashboard/logs/diagnostics views.
- `tui/testdata/` — Committed golden snapshot files for deterministic terminal UI rendering regression testing.

### Configuration & Templates (`config/`)
- `config/20-omacorn-proxy.conf` — Static template declaring default `http_proxy`, `https_proxy`, and `no_proxy` variables (deliberately omitting `all_proxy` to preserve MTProto and raw socket stability).

### Fleet Maintenance & Tooling (`scripts/`)
- `scripts/install.sh` — Universal multi-distro installer script (pacman/dnf/apt): provisions prerequisites, builds or downloads binaries, writes systemd units, and configures user paths.
- `scripts/update.sh` — Tier 1 fleet maintenance script: pulls latest Git commits (`git pull --ff-only`), builds static binary, hot-reloads running daemons, and verifies connectivity.
- `scripts/test_spoofdpi.sh` — Diagnostic script testing live SpoofDPI proxy execution against target URLs.

### Documentation & Reference (`reference/`)
- `reference/SNI-filter-bypassing-app.md` — Conceptual design notes, DPI bypass theory, decoy injection analysis, and early architecture history.

## Build, Test & CI Tooling
- `Makefile` — Build automation targets:
  - `make build`: Local compilation with version ldflags.
  - `make test`: Executes test suite with race detector (`go test -v -race ./...`).
  - `make lint`: Validates formatting and code correctness (`go vet ./...` + `gofmt -l .`).
  - `make static`: Builds standalone CGO-free static binary.
  - `make cross`: Cross-compiles stripped binaries for `linux/amd64` and `linux/arm64`.
  - `make install`: Installs binary and `dpipe` alias to `~/.local/bin/` using `install -Dm755`.
  - `make update`: Invokes `./scripts/update.sh`.
  - `make clean`: Cleans build artifacts and test binaries.
- `go test ./...` — Unit and integration tests covering `dpipe` (main), `dpipe/engine`, and `dpipe/tui` (>70% statement coverage).
- `go vet ./...` — Code analysis tool enforced locally and in CI.
- GitHub Actions CI (`.github/workflows/ci.yml`) — Validates `gofmt`, runs `go vet`, executes race tests, and builds cross-architecture release matrix on every PR and push to master.
- GitHub Actions CD (`.github/workflows/release.yml`) — Generates tagged release binaries (`v*`), computes SHA256 checksums, and publishes assets to GitHub Releases.

## Dependencies
- Go 1.24+ standard library (zero CGO runtime dependencies).
- `github.com/charmbracelet/bubbletea` (v1.3.10) — Terminal UI runtime event loop and state management framework.
- `github.com/charmbracelet/lipgloss` (v1.1.0) — Terminal CSS styling, color tokens, and layout composition.
- `github.com/charmbracelet/bubbles` (v1.0.0) — TUI components (viewport for log streaming).
- `github.com/charmbracelet/x/exp/golden` — Golden snapshot testing assertion library.
