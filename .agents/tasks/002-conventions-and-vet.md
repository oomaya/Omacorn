# Packet 002 — Confirm conventions and wire go vet into CI

**Status:** queue
**Branch:** `packet/002-conventions-and-vet`
**Author (architect):** Antigravity
**Date:** 2026-09-27

### Objective
Confirm `.agents/CONVENTIONS.md` with Master (especially the commit-style question), then ensure `go vet ./...` runs in CI so the convention is enforced, not just documented.

### Repo context
- `.agents/CONVENTIONS.md` (seed draft)
- `.github/workflows/ci.yml`, `Makefile`

### Approved decisions
- Validation trio for this repo: `gofmt -l .` clean, `go vet ./...`, `go test ./...`.

### Scope
`CONVENTIONS.md` confirmation + CI/Makefile changes to run `go vet`.

### Non-goals
- No changes to application code. No new lint tooling beyond `go vet`.

### Acceptance criteria
1. `CONVENTIONS.md` confirmed — the commit-style question resolved with Master (kick back to Master if undecided; do not guess).
2. `go vet ./...` runs in CI (directly or via a `make` target invoked by CI).
3. `gofmt` cleanliness is checked in CI or documented as a pre-commit expectation.

### Validation
- `go vet ./...`, `gofmt -l .`, `go test ./...`
- Inspect the CI workflow diff.

### Expected output
Diff to `.agents/CONVENTIONS.md`, `.github/workflows/ci.yml`, `Makefile`.
