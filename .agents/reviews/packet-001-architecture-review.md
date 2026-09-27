# Masterpiece Delivery: Packet 001 Architecture Contract

**Date:** 2026-09-27  
**Architect:** Antigravity  
**Artifact:** [.agents/ARCHITECTURE.md](file:///home/rand/Omacorn/.agents/ARCHITECTURE.md)  
**Status:** `APPROVED & MERGED` ([master](file:///home/rand/Omacorn/main.go))

---

## 1. Executive Summary

Packet 001—the foundational contract of the repository—has been verified, finalized, approved, and merged into `master`. All provisional draft notices and unresolved verification markers have been eliminated. The document now serves as the authoritative architectural blueprint across all subsystems: the entry point dispatcher, the dual-pilot (PF/PM) mutual exclusion engines, the split-tunneling exclusion engine, the cyberpunk Bubble Tea TUI, the build automation matrix, and the zero-CGO runtime dependencies.

---

## 2. Verified Subsystem Boundaries

| Subsystem | Primary Files | Verified Architectural Responsibility |
| :--- | :--- | :--- |
| **CLI & Dispatcher** | [main.go](file:///home/rand/Omacorn/main.go), [main_test.go](file:///home/rand/Omacorn/main_test.go) | Root command dispatcher, headless flags, Bubble Tea launcher, and candidate path discovery (`.` → git toplevel → binary walk → `~/Omacorn` → fallbacks). |
| **Pilot 1 (Main Pilot / PF)** | [engine/spoofdpi.go](file:///home/rand/Omacorn/engine/spoofdpi.go) | User-space bypass daemon on `127.0.0.1:8080`, low-TTL decoy injection via `CAP_NET_RAW`, hypervisor-safe user-space TCP fragmentation fallback. |
| **Pilot 2 (Co-Pilot / PM)** | [engine/gecit.go](file:///home/rand/Omacorn/engine/gecit.go) | Kernel-space eBPF `sock_ops` engine attached to root cgroups, hardened MSS clamping (`--mss 88`, restored after 600B), bare-metal performance. |
| **Desktop Proxy & Split-Tunneling** | [engine/proxy.go](file:///home/rand/Omacorn/engine/proxy.go), [engine/exclusion.go](file:///home/rand/Omacorn/engine/exclusion.go) | Automatic `environment.d` session bus sync, Chromium/Brave flags sync, user exclusion persistence, `all_proxy` omission to preserve MTProto. |
| **Cyberpunk TUI** | [tui/model.go](file:///home/rand/Omacorn/tui/model.go), [tui/views.go](file:///home/rand/Omacorn/tui/views.go), [tui/style.go](file:///home/rand/Omacorn/tui/style.go) | Bubble Tea state engine, Lip Gloss styling, telemetry dashboards, scrollable `journalctl` log viewports, and golden file snapshot test suite. |
| **Fleet Automation & CI/CD** | [scripts/update.sh](file:///home/rand/Omacorn/scripts/update.sh), [Makefile](file:///home/rand/Omacorn/Makefile), [.github/workflows/](file:///home/rand/Omacorn/.github/workflows/) | Cross-distro installer, tier 1 hot-reload maintenance script, multi-arch cross-compilation matrix (`amd64` / `arm64`), automated GitHub Releases. |

---

## 3. Acceptance Criteria Audit

1. **Every file listed under Layout verified:** All components in `main.go`, `engine/`, `tui/`, `config/`, `scripts/`, and `reference/` are explicitly audited and mapped.
2. **Zero `[Verify ...]` markers remaining:** Audited via search; all claims verified against Go source.
3. **Draft notice and checklist removed:** Clean, production-grade living architecture manual.
4. **Healthcheck & Safety:** Full test suite clean under race detector (`go test -race ./...`), `go vet` clean, `gofmt` clean.

---

## 4. Git Integration Record

- Branch `packet/001-verify-architecture-map` committed at `9955fd4`.
- Merged into `master` via non-fast-forward merge.
- Working tree clean and ahead of remote by 15 commits.
