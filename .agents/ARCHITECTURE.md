# ARCHITECTURE.md — Omacorn

> **DRAFT — inferred from the repository file tree on 2026-09-27. The architect MUST verify and correct every claim below against the actual code (packet 001) before this document is trusted.**

## What it is
Omacorn is a Go-based local proxy tool for SNI filter bypassing (see `reference/SNI-filter-bypassing-app.md` for background), with a terminal UI, install/update scripts, and CI + release automation.

## Layout
- `main.go` — entry point. *[Verify: what it wires together.]*
- `engine/` — core logic
  - `engine.go` — *[Verify.]*
  - `proxy.go` — proxy implementation *[Verify.]*
  - `spoofdpi.go` — SpoofDPI technique integration *[Verify.]*
  - `exclusion.go` — exclusion-list handling *[Verify.]*
  - `gecit.go` — *[Verify. "Geçit" is Turkish for "gateway/passage".]*
- `config/20-omacorn-proxy.conf` — proxy configuration
- `tui/` — terminal UI: `model.go`, `views.go`, `style.go` *[Verify framework — likely Bubble Tea.]*
- `scripts/` — `install.sh`, `update.sh`, `test_spoofdpi.sh`
- `reference/` — background reading, not code

## Build & test
- `Makefile` *[Verify targets.]*
- `go test ./...` — engine tests (`engine_test.go`, `exclusion_test.go`, `proxy_test.go`) + `main_test.go`
- `go vet ./...` — *[Confirm wired into CI.]*
- CI: `.github/workflows/ci.yml`, `.github/workflows/release.yml`

## Dependencies
*[Verify from `go.mod` — list the notable ones.]*

## Verification checklist (packet 001)
- [ ] One-line verified responsibility for every file under Layout
- [ ] TUI framework confirmed
- [ ] Makefile targets listed
- [ ] Notable `go.mod` dependencies noted
- [ ] This draft notice removed
