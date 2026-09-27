# Packet 001 — Verify and complete the architecture map

**Status:** complete
**Branch:** `packet/001-verify-architecture-map`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Verify every claim in `.agents/ARCHITECTURE.md` (currently a DRAFT inferred from the file tree) against the actual code, correct it, and remove the draft notice. This document is the foundation all future packets build on — it must be accurate.

### Repo context
- `.agents/ARCHITECTURE.md` (the draft to verify)
- `main.go`, `engine/`, `tui/`, `config/`, `scripts/`, `Makefile`, `go.mod`

### Approved decisions
- None yet — this packet creates the baseline.

### Scope
The `ARCHITECTURE.md` document only.

### Non-goals
- No code changes. No refactoring. No new docs beyond `ARCHITECTURE.md`.

### Acceptance criteria
1. Every file listed under Layout has a verified one-line responsibility description.
2. All `[Verify ...]` markers resolved.
3. The draft notice and verification checklist removed.

### Validation
- Read `.agents/ARCHITECTURE.md` — zero remaining "verify" markers.

### Expected output
Diff to `.agents/ARCHITECTURE.md` only.

---

### Builder notes (filled by builder)

**Plan approved:** [x]
**Implementation summary:**
Audited and verified all files under Layout against the active codebase:
1. Entry point & dispatcher: `main.go` and `main_test.go` responsibilities verified (CLI dispatcher, Bubble Tea launcher, dual-tier candidate discovery).
2. Engine subsystem: verified responsibilities across `engine.go`, `proxy.go`, `spoofdpi.go`, `exclusion.go`, `gecit.go`, and their test suites.
3. Terminal UI: verified Bubble Tea model, Lip Gloss styling, views, and golden snapshot test infrastructure.
4. Build & CI: documented `Makefile` targets (`all`, `build`, `test`, `lint`, `static`, `cross`, `install`, `update`, `clean`), CI workflows (`.github/workflows/ci.yml`, `release.yml`), and `go.mod` dependencies (Bubble Tea, Lip Gloss, Bubbles, golden).
5. Architecture model: formalized Pilot & Co-Pilot (PF/PM) mutual exclusion model and desktop `environment.d` integration.
6. Removed draft notice and verification checklist. Zero "verify" markers remain.

### Review (filled by architect — read-only)

**Verdict:** `APPROVED`
**Notes:**
All layout files, subsystem boundaries, build targets, and runtime dependencies verified against the active codebase. DRAFT notice removed. The foundational codebase map is now complete and authoritative.
