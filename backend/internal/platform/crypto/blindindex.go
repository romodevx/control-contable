package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"
)

const (
	// MinPepperSize is the shortest accepted pepper: the HMAC-SHA256 output size.
	MinPepperSize = sha256.Size
	// HashSize is the length in bytes of a blind index hash.
	HashSize = sha256.Size
	// ScopeSize is the length in bytes of a HashScoped scope: a UUID.
	ScopeSize = 16
)

// BlindIndex computes HMAC-SHA256(pepper, normalize(value)): a deterministic,
// keyed hash for equality lookups and unique indexes (rfc_hash, email_hash)
// that cannot be reversed by brute force without the pepper. Use a different
// pepper per purpose. It is safe for concurrent use.
type BlindIndex struct {
	pepper    []byte
	normalize func(string) (string, error)
}

// NewRFCIndex returns the rfc_hash index: the RFC is trimmed, its ASCII
// letters and ñ are upper-cased, and the result must be a 12-character
// (persona moral) or 13-character (persona física) RFC, else ErrInvalidRFC.
func NewRFCIndex(pepper []byte) (*BlindIndex, error) {
	return newBlindIndex(pepper, normalizeRFC)
}

// NewEmailIndex returns the email_hash index: the email is trimmed and its
// ASCII letters are lower-cased. Non-ASCII runes are kept as they are, so no
// Unicode case mapping (e.g. KELVIN SIGN to "k") can collide with an ASCII
// address.
func NewEmailIndex(pepper []byte) (*BlindIndex, error) {
	return newBlindIndex(pepper, normalizeEmail)
}

// NewTokenIndex returns the token_hmac index of push tokens
// (MODULE-reminders §2.3): the token is opaque and case-sensitive, so it is
// hashed exactly as given; only an empty token is refused.
func NewTokenIndex(pepper []byte) (*BlindIndex, error) {
	return newBlindIndex(pepper, verbatim)
}

func newBlindIndex(pepper []byte, normalize func(string) (string, error)) (*BlindIndex, error) {
	if len(pepper) < MinPepperSize {
		return nil, fmt.Errorf("%w: %d bytes, want at least %d", ErrInvalidPepper, len(pepper), MinPepperSize)
	}
	return &BlindIndex{pepper: bytes.Clone(pepper), normalize: normalize}, nil
}

// Hash returns the HashSize-byte blind index of value. A value that is empty
// after normalization is ErrEmptyInput, so no two blank inputs ever collide.
func (b *BlindIndex) Hash(value string) ([]byte, error) {
	return b.hash(nil, value)
}

// HashScoped returns HMAC-SHA256(pepper, scope || normalize(value)): the
// blind index of value within one scope, e.g. rfc_hash keyed by the owning
// user's id (MODULE-taxpayers S-2), so equal values of two scopes cannot be
// linked. scope must be exactly ScopeSize bytes (the 16 bytes of a UUID):
// a fixed width keeps scope || value unambiguous. Otherwise ErrInvalidScope.
func (b *BlindIndex) HashScoped(scope []byte, value string) ([]byte, error) {
	if len(scope) != ScopeSize {
		return nil, fmt.Errorf("%w: %d bytes, want %d", ErrInvalidScope, len(scope), ScopeSize)
	}
	return b.hash(scope, value)
}

func (b *BlindIndex) hash(scope []byte, value string) ([]byte, error) {
	normalized, err := b.normalize(value)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, b.pepper)
	mac.Write(scope) // hash.Hash.Write never returns an error
	mac.Write([]byte(normalized))
	return mac.Sum(nil), nil
}

// Format redacts the index for every fmt verb (%v, %+v, %#v, %s), so
// secret material never reaches a log or an error message.
func (b BlindIndex) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the index in slog.
func (b BlindIndex) LogValue() slog.Value { return slog.StringValue(redacted) }

// rfcPattern is the RFC shape: 3 (moral) or 4 (física) letters from
// A-Z, Ñ and &, a yymmdd date, and a 3-character homoclave. Ñ must be the
// precomposed U+00D1; a decomposed N + U+0303 does not match.
var rfcPattern = regexp.MustCompile(`^[A-ZÑ&]{3,4}[0-9]{6}[A-Z0-9]{3}$`)

func normalizeRFC(rfc string) (string, error) {
	normalized := strings.Map(upperRFCRune, strings.TrimSpace(rfc))
	if normalized == "" {
		return "", ErrEmptyInput
	}
	if !rfcPattern.MatchString(normalized) {
		return "", ErrInvalidRFC
	}
	return normalized, nil
}

// upperRFCRune upper-cases only a-z and ñ. strings.ToUpper would also map
// runes such as U+017F (long s) to "S" and U+0131 (dotless i) to "I".
func upperRFCRune(r rune) rune {
	switch {
	case r >= 'a' && r <= 'z':
		return r - 'a' + 'A'
	case r == 'ñ':
		return 'Ñ'
	}
	return r
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.Map(lowerASCIIRune, strings.TrimSpace(email))
	if normalized == "" {
		return "", ErrEmptyInput
	}
	return normalized, nil
}

// verbatim keeps an opaque value as given; only an empty one is refused.
func verbatim(value string) (string, error) {
	if value == "" {
		return "", ErrEmptyInput
	}
	return value, nil
}

// lowerASCIIRune lower-cases only A-Z. strings.ToLower would also map U+212A
// (KELVIN SIGN) to "k" and U+0130 (dotted capital I) to "i̇".
func lowerASCIIRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}
