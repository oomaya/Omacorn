# CONVENTIONS.md — Omacorn

Official engineering conventions for the Omacorn repository, maintained by the Architect and enforced across all builder packets.

## Code Standards
- `gofmt -l .` must be clean across all Go files.
- `go vet ./...` must pass with zero warnings (enforced in Makefile and GitHub Actions CI).
- `go test -race ./...` must pass with zero data races.
- New features or bug fixes must include unit or golden regression tests.
- Zero CGO runtime dependencies (`CGO_ENABLED=0` static builds).

## Host-Neutrality Invariant (Anti-Environment Assumption Law)
- Tests must **never assert on host-specific hardware, hypervisor types (`vmware`, `kvm`, `microsoft`), usernames, or home paths**.
- Always query dynamic detection helpers (`engine.DetectHypervisor()`, `$HOME`, `$XDG_CONFIG_HOME`) or use hermetic test fixtures.
- Test suites must execute and pass identically on bare metal (Arch/CachyOS), virtual machines (VMware/QEMU), and CI cloud runners (Azure/GitHub Actions).

## Packets & Worktrees
- Task Packets: `.agents/tasks/NNN-slug.md` (completed packets move to `tasks/done/`).
- Task Scaffolding Lifecycle: The Architect must create and commit the task packet scaffold to `master` **before** the builder checks out an isolated worktree branch. This prevents untracked root collisions during `git merge`.
- Branches & Worktrees: `packet/NNN-slug` (one packet = one branch = one isolated worktree).
- One packet = one PR-sized, bounded scope with explicit non-goals.
- Worktrees operate in non-overlapping filesystem domains to ensure deterministic, zero-conflict merges.

## Commits
- Standard: Conventional Commits.
  - `feat:` — User-facing feature or CLI command addition.
  - `fix:` — Bug fix or behavioral correction.
  - `test:` — Test suite additions, golden snapshots, or coverage fixes.
  - `docs:` — Documentation, architectural reviews, or task packet updates.
  - `chore:` — Dependency updates, tooling, or release version bumps.

## Reviews & Verification
- Reviewers are strictly read-only. Remediation becomes a new bounded packet, never an ad-hoc fix-commit during review.
- Verdicts: `APPROVED` or `CHANGES REQUESTED`, recorded in `.agents/reviews/NNN-slug.md`.
- Remote Verification Tollgate: The Architect must poll and confirm GitHub Actions CI (`gh run list` / `gh run watch`) post-push before certifying milestone completion.

## Definition of Done
- [ ] Packet acceptance criteria fully satisfied.
- [ ] Non-goals strictly respected (zero scope creep in diff).
- [ ] Host-neutrality verified (runs cleanly on bare metal, VMs, and CI).
- [ ] `gofmt`, `go vet`, and `go test -race ./...` clean locally.
- [ ] Architect review `APPROVED`.
- [ ] Merged cleanly into `master` and verified green in GitHub Actions CI.
- [ ] Task packet moved to `.agents/tasks/done/`.
