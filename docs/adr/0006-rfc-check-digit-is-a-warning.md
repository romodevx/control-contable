# ADR-0006 — An RFC check digit mismatch is a warning, not a validation error

| Field | Value |
|---|---|
| Status | ACCEPTED (coordinator decision D-RFC, applied in PR #54, M2-04) |
| Date | 2026-10-13 |
| Related | [PRD US-T1 and Q9](../project-specs/PRD.md), `backend/internal/taxpayers/domain/rfc.go`, contract `TaxpayerCreate.rfc` |
| Owner | backend-sr / tech-lead |

## Context

PRD Q9 answered "format + check digit validated" for the RFC. The check digit is the last
character of the homoclave, computed with the SAT algorithm published as Anexo 3 of the
"Algoritmo para generar el RFC con homoclave para personas físicas y morales" (released
under transparency request 0610100162005). python-stdnum, which implements the same
algorithm, reports that about 1.5% of the RFCs the SAT has actually issued do not match it
and leaves the check off by default. The SAT, not the published algorithm, is the
authority on which RFCs exist. Rejecting a mismatch would lock out a real taxpayer with no
way to fix their own RFC.

## Decision

1. `domain.NewRFC` still rejects, in order: generic RFCs (`ErrRFCGeneric`), a wrong shape or
   charset (`ErrRFCMalformed`) and a non-calendar YYMMDD (`ErrRFCInvalidDate`).
   `NewRFCFor` still rejects a length that contradicts the declared person type.
2. The check digit is still computed (Anexo 3), but a mismatch parses successfully. The
   value object exposes it as `RFC.CheckDigitValid() bool`. There is no
   `ErrRFCCheckDigit`.
3. Surfacing the warning (a response field or UI hint asking the user to double-check the
   RFC) belongs to the use case and HTTP adapter (M2-05/M2-07). Any new response field is
   a contract change made spec-first.

## Consequences

- PRD Q9's "check digit validated" now reads "check digit computed and reported".
- A typo that only changes the homoclave can be stored. The display name, the person type
  and the warning are the user's cues; the RFC is never verified against the SAT anyway.
- Tests keep both directions: synthetic RFCs with a correct digit report `true`, and the
  same RFCs with a wrong digit parse and report `false`. Fixtures are synthetic only.
