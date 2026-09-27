# Packet 001 — Verify and complete the architecture map

**Status:** queue
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
