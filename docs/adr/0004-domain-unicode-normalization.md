# ADR-0004 — The domain may import `golang.org/x/text/unicode/norm`

| Field | Value |
|---|---|
| Status | ACCEPTED (tech-lead, 2026-10-13, PR #35), with [Amendment 1](#amendment-1-user-text-refuses-invisible-characters-0140) (0.14.0) |
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

## Amendment 1: user text refuses invisible characters (0.14.0)

**Context.** The ANN-25 security gate (PR #153) found that the annual labels accepted
Unicode format characters (Cf): `unicode.IsControl` only refuses Cc. A bidi override
(U+202E) or a zero-width character can make a label read differently from what is stored.
The same gap existed in every free-text field a user types, and the income and expense
descriptions refused nothing at all (a NUL reached the database and failed there).

**Decision.**

1. One rule for user text, `internal/platform/usertext.HasInvisible`: no control (Cc, line
   breaks and tabs included: every field is single-line), format (Cf: zero-width
   characters, bidi marks, embeddings, overrides and isolates, the BOM, the soft hyphen,
   tags) or line/paragraph separator (Zl, Zp) character. Combining marks and variation
   selectors (Mn) stay allowed (accents, emoji). Applied after NFC and trimming to: the
   account `name` (`verifyEmail`, `updateMe`), the taxpayer `display_name`, the income
   and expense `description`, `employer_label` and `student_label`. Passwords are out of
   scope (never displayed), and so are values with their own grammar (RFC, ids, codes).
2. The answer is `400 validation_failed` with a `FieldError` on the field and the code
   `invisible_characters`; the message is each module's, in its language.
3. `usertext` is stdlib-only and joins `money` as a pure platform package a domain may
   import (`internal/archtest` `domainPlatformAllowlist`).
4. The contract states the rule in each field's description. It is not a JSON Schema
   `pattern`: OpenAPI patterns are ECMA-262, where `\p{Cf}` needs the `u` flag that
   generated clients do not set, and kin-openapi would answer a generic `pattern` code
   instead of `invisible_characters`.

**Consequences.** Tightening validation on frozen nodes is a breaking change
(CONTRACT.md), hence 0.14.0. No shipped client sends these characters on purpose: the
mobile fields are single-line `TextField`s. Rows stored before 0.14.0 are not rewritten.
An edit that sends the field must remove the characters; the annual updates re-check the
merged row, so there any edit names the stored label until it is fixed. The database CHECKs still enforce trim, length and
NFC only (PostgreSQL regular expressions have no Unicode category classes); the
application rule is the boundary.
