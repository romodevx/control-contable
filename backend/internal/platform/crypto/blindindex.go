package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

const (
	// MinPepperSize is the shortest accepted pepper: the HMAC-SHA256 output size.
	MinPepperSize = sha256.Size
	// HashSize is the length in bytes of a blind index hash.
	HashSize = sha256.Size
)

// BlindIndex computes HMAC-SHA256(pepper, normalize(value)): a deterministic,
// keyed hash for equality lookups and unique indexes (rfc_hash, email_hash)
// that cannot be reversed by brute force without the pepper. Use a different
// pepper per purpose. It is safe for concurrent use.
type BlindIndex struct {
	pepper    []byte
	normalize func(string) string
}

// NewRFCIndex returns the rfc_hash index: the RFC is trimmed and upper-cased.
// Callers validate the RFC format before hashing.
func NewRFCIndex(pepper []byte) (*BlindIndex, error) {
	return newBlindIndex(pepper, normalizeRFC)
}

// NewEmailIndex returns the email_hash index: the email is trimmed and
// lower-cased.
func NewEmailIndex(pepper []byte) (*BlindIndex, error) {
	return newBlindIndex(pepper, normalizeEmail)
}

func newBlindIndex(pepper []byte, normalize func(string) string) (*BlindIndex, error) {
	if len(pepper) < MinPepperSize {
		return nil, fmt.Errorf("%w: %d bytes, want at least %d", ErrInvalidPepper, len(pepper), MinPepperSize)
	}
	return &BlindIndex{pepper: bytes.Clone(pepper), normalize: normalize}, nil
}

// Hash returns the HashSize-byte blind index of value. A value that is empty
// after normalization is ErrEmptyInput, so no two blank inputs ever collide.
func (b *BlindIndex) Hash(value string) ([]byte, error) {
	normalized := b.normalize(value)
	if normalized == "" {
		return nil, ErrEmptyInput
	}
	mac := hmac.New(sha256.New, b.pepper)
	mac.Write([]byte(normalized)) // hash.Hash.Write never returns an error
	return mac.Sum(nil), nil
}

// Format redacts the index for every fmt verb (%v, %+v, %#v, %s), so
// secret material never reaches a log or an error message.
func (b *BlindIndex) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the index in slog.
func (b *BlindIndex) LogValue() slog.Value { return slog.StringValue(redacted) }

func normalizeRFC(rfc string) string { return strings.ToUpper(strings.TrimSpace(rfc)) }

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
