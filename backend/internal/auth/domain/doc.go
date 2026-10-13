// Package domain holds the auth module's value objects and domain errors:
//
//   - Email: trimmed, ASCII-lowercased address (the same normalization as the
//     email_hash blind index); non-ASCII is rejected; RFC 5321 length limits.
//   - Password: NFKC-normalized secret under the NIST SP 800-63B policy
//     (12 to 1024 bytes, no composition rules, common-password blocklist).
//   - RefreshToken: "<user_uuid>.<secret>" with a 256-bit secret; only its
//     TokenHash is persisted, compared in constant time.
//   - VerificationCode: 6 decimal digits drawn without modulo bias.
//
// Randomness is injected as an io.Reader (crypto/rand.Reader in production)
// so tests are deterministic. The package imports the standard library and
// golang.org/x/text/unicode/norm only (ADR-0004). Secret-holding types render
// as REDACTED in fmt and slog, and errors never carry the input.
package domain
