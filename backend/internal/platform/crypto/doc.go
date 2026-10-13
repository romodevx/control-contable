// Package crypto holds the application-level cryptography of ADR-0001:
//
//   - PasswordHasher: argon2id password hashes in PHC string format, with
//     constant-time verification and rehash-needed detection so the cost
//     parameters can be raised later.
//   - Keyring: AES-256-GCM encryption of the RFC. Every ciphertext starts with
//     the key_id that sealed it, so keys can be rotated without a flag day.
//   - BlindIndex: HMAC-SHA256 peppered hashes (rfc_hash, email_hash) over a
//     normalized value, for equality lookups and unique indexes.
//
// The package never logs, and its errors never carry plaintext, passwords,
// keys or peppers. Types that hold secrets render as REDACTED in fmt and slog.
package crypto

import "errors"

// Sentinel errors. Callers match them with errors.Is.
var (
	// ErrInvalidParams reports argon2id parameters outside the accepted bounds.
	ErrInvalidParams = errors.New("crypto: invalid argon2id parameters")
	// ErrMalformedHash reports a stored password hash that is not a supported
	// argon2id PHC string.
	ErrMalformedHash = errors.New("crypto: malformed password hash")
	// ErrInvalidKey reports a keyring built with a bad key or key id.
	ErrInvalidKey = errors.New("crypto: invalid encryption key")
	// ErrUnknownKeyID reports a ciphertext sealed with a key the keyring lacks.
	ErrUnknownKeyID = errors.New("crypto: unknown key id")
	// ErrMalformedCiphertext reports a ciphertext too short to be valid.
	ErrMalformedCiphertext = errors.New("crypto: malformed ciphertext")
	// ErrDecrypt reports a ciphertext that fails authentication: tampered
	// data, the wrong key or the wrong additional data.
	ErrDecrypt = errors.New("crypto: decryption failed")
	// ErrInvalidPepper reports a pepper shorter than MinPepperSize.
	ErrInvalidPepper = errors.New("crypto: invalid pepper")
	// ErrEmptyInput reports a value that is empty after normalization.
	ErrEmptyInput = errors.New("crypto: empty input")
	// ErrInvalidRFC reports a value that is not a 12- or 13-character RFC
	// after normalization.
	ErrInvalidRFC = errors.New("crypto: invalid rfc")
	// ErrPasswordTooLong reports a password above MaxPasswordBytes.
	ErrPasswordTooLong = errors.New("crypto: password too long")
)

// redacted is how every secret-holding type renders in fmt and slog.
const redacted = "REDACTED"
