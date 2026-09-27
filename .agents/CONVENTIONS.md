# CONVENTIONS.md — Omacorn

> Seed draft — architect to confirm with Master (packet 002).

## Code
- `gofmt -l .` must be clean.
- `go vet ./...` passes (wired into CI).
- `go test ./...` green — new engine logic ships with `_test.go` coverage (the repo already does this; keep it).

## Packets & branches
- Packets: `.agents/tasks/NNN-slug.md` → completed move to `.agents/tasks/done/`
- Branches: `packet/NNN-slug` (one packet = one branch = one worktree)
- One packet = one PR-sized change.

## Commits
- Proposed: Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`). *[Master to confirm — packet 002.]*

## Reviews
- Reviewers are read-only. Remediation becomes a new bounded packet, never a fix-commit.
- Verdicts: `APPROVED` / `CHANGES REQUESTED`, filed in `.agents/reviews/NNN-slug.md`.

## Definition of done
- [ ] Acceptance criteria met · [ ] `gofmt`/`vet`/`test` green · [ ] no out-of-scope changes in diff · [ ] review APPROVED · [ ] packet archived
