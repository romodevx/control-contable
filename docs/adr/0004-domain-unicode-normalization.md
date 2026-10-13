# ADR-0004 — The domain may import `golang.org/x/text/unicode/norm`

| Field | Value |
|---|---|
| Status | ACCEPTED (tech-lead, 2026-10-13, PR #35) |
| Date | 2026-10-13 |
| Related | [ADR-0001 layout and Dependency Rule](0001-stack.md), [PLAN M1-04](../project-specs/PLAN-control-contable.md), [P0-30 security gate](../reviews/P0-30-security-gate.md) |
| Owner | backend-sr |

## Context

ADR-0001 says `domain/` "imports stdlib only", and `internal/archtest` plus the depguard
`domain` rule enforce it. M1-04 puts the password policy in `internal/auth/domain`, and
NIST SP 800-63B-4 asks verifiers to normalize passwords with NFKC (or NFKD) before hashing,
so the same characters typed on different keyboards or platforms hash the same. The Go
standard library has no Unicode normalization; `golang.org/x/text/unicode/norm` is the Go
team's implementation and is already in the module graph (indirect, v0.43.0).

## Options

- **A. Allow exactly `golang.org/x/text/unicode/norm` in the domain (chosen).** A pure
  algorithm package from the Go project: no I/O, no globals to configure, no framework. The
  policy (normalize, then count characters, then check the blocklist) stays in one place
  and is unit-tested end to end.
- **B. Inject a normalizer port** (`func(string) string`) from an adapter. The domain stays
  stdlib-only, but the policy's correctness then depends on wiring: a test or caller that
  passes the identity function silently skips NFKC, and length is counted on the wrong
  form.
- **C. Normalize in the HTTP adapter.** Same wiring risk as B, and login and registration
  could drift apart.

## Decision

Option A. The exception is the **exact import path** `golang.org/x/text/unicode/norm`:

- `internal/archtest`: `domainExternalAllowlist`, with cases proving that
  `golang.org/x/text/cases` and a look-alike path are still rejected, and that pure
  platform packages (`money`) stay stdlib-only.
- `.golangci.yml` depguard `domain` rule: `golang.org/x/text/unicode/norm$` (anchored).

Any further external package in the domain needs its own ADR.

## Consequences

- `golang.org/x/text` becomes a direct dependency; govulncheck already covers it.
- The Unicode tables of `x/text` (pinned version) define NFKC. A version bump could in
  theory change the normal form of newly assigned characters; NFKC is stable for
  assigned characters under Unicode's normalization stability policy, so stored hashes
  keep verifying.

## Review

Accepted by tech-lead on 2026-10-13 (PR #35). The exception is one exact import path, its
only transitive import is `golang.org/x/text/transform` (also pure, no I/O), and both guards
were probed: a temporary `golang.org/x/text/cases` import in `internal/auth/domain` failed
depguard (`not allowed from list 'domain'`) and `TestBackendModuleRespectsDependencyRule`.
Options B and C are rejected for the reason given above: NFKC is part of the password
policy, and a port would make that policy depend on correct wiring.
