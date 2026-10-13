package domain

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// RefreshSecretBytes is the size of a refresh token secret: 256 bits from a
// CSPRNG (ADR-0001 O6).
const RefreshSecretBytes = 32

const refreshTokenSeparator = "."

// secretEncoding is unpadded base64url in strict mode, so every secret has
// exactly one accepted spelling (non-zero trailing bits are rejected).
var secretEncoding = base64.RawURLEncoding.Strict()

// UserID is a user's UUID. It has the memory layout of github.com/google/uuid's
// UUID, so adapters convert with uuid.UUID(id) and UserID(u).
type UserID [16]byte

// uuidTextLength is the length of the canonical 8-4-4-4-12 UUID text form.
const uuidTextLength = 36

// uuidHyphens are the byte offsets of the hyphens in the canonical form.
var uuidHyphens = [...]int{8, 13, 18, 23}

// ParseUserID parses the canonical 8-4-4-4-12 form, in either case. The nil
// UUID is rejected: no user has it, so it can only be a forged or empty value.
func ParseUserID(text string) (UserID, error) {
	if len(text) != uuidTextLength {
		return UserID{}, ErrInvalidUserID
	}
	var digits strings.Builder
	digits.Grow(2 * len(UserID{}))
	start := 0
	for _, hyphen := range uuidHyphens {
		if text[hyphen] != '-' {
			return UserID{}, ErrInvalidUserID
		}
		digits.WriteString(text[start:hyphen])
		start = hyphen + 1
	}
	digits.WriteString(text[start:])
	var id UserID
	if _, err := hex.Decode(id[:], []byte(digits.String())); err != nil {
		return UserID{}, ErrInvalidUserID
	}
	if id.IsZero() {
		return UserID{}, ErrInvalidUserID
	}
	return id, nil
}

// String returns the canonical lower-case 8-4-4-4-12 form.
func (id UserID) String() string {
	encoded := hex.EncodeToString(id[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}

// IsZero reports whether id is the nil UUID.
func (id UserID) IsZero() bool { return id == UserID{} }

// RefreshToken is an opaque refresh token "<user_uuid>.<secret>" (ADR-0001
// O6). The user id lets the server set the RLS user context before the
// lookup; the secret is RefreshSecretBytes random bytes in unpadded base64url.
// Only Hash() is persisted. It renders as REDACTED in fmt and slog; Reveal
// returns the text sent to the client.
type RefreshToken struct {
	redactor
	userID UserID
	text   string // canonical "<uuid>.<secret>"
}

// NewRefreshToken issues a token for userID with a secret read from entropy,
// which must be a CSPRNG (crypto/rand.Reader in production).
func NewRefreshToken(userID UserID, entropy io.Reader) (RefreshToken, error) {
	if userID.IsZero() {
		return RefreshToken{}, ErrInvalidUserID
	}
	secret := make([]byte, RefreshSecretBytes)
	if _, err := io.ReadFull(entropy, secret); err != nil {
		return RefreshToken{}, fmt.Errorf("%w: %w", ErrRandomSource, err)
	}
	return newRefreshToken(userID, secretEncoding.EncodeToString(secret)), nil
}

// ParseRefreshToken parses a token presented by a client. The user id may
// be in either case; the result is canonical, so its Hash matches the one
// stored at issue time.
func ParseRefreshToken(text string) (RefreshToken, error) {
	idText, secret, found := strings.Cut(text, refreshTokenSeparator)
	if !found {
		return RefreshToken{}, ErrMalformedRefreshToken
	}
	userID, err := ParseUserID(idText)
	if err != nil {
		return RefreshToken{}, ErrMalformedRefreshToken
	}
	decoded, err := secretEncoding.DecodeString(secret)
	if err != nil || len(decoded) != RefreshSecretBytes {
		return RefreshToken{}, ErrMalformedRefreshToken
	}
	return newRefreshToken(userID, secret), nil
}

func newRefreshToken(userID UserID, encodedSecret string) RefreshToken {
	return RefreshToken{userID: userID, text: userID.String() + refreshTokenSeparator + encodedSecret}
}

// UserID returns the user the token was issued to.
func (t RefreshToken) UserID() UserID { return t.userID }

// Reveal returns "<uuid>.<secret>", for the response body only.
func (t RefreshToken) Reveal() string { return t.text }

// Hash returns the value stored in refresh_tokens.token_hash: SHA-256 of the
// whole canonical token, so it also binds the user id. A 256-bit random
// secret needs no salt or slow hash.
func (t RefreshToken) Hash() TokenHash { return TokenHash(sha256.Sum256([]byte(t.text))) }

// TokenHash is the SHA-256 of a refresh token: the only form persisted.
type TokenHash [sha256.Size]byte

// TokenHashFromBytes rebuilds a TokenHash read from storage.
func TokenHashFromBytes(raw []byte) (TokenHash, error) {
	var hash TokenHash
	if len(raw) != len(hash) {
		return TokenHash{}, ErrMalformedTokenHash
	}
	copy(hash[:], raw)
	return hash, nil
}

// Bytes returns a copy of the hash, for storage.
func (h TokenHash) Bytes() []byte { return h[:] }

// Equal compares two hashes in constant time.
func (h TokenHash) Equal(other TokenHash) bool {
	return subtle.ConstantTimeCompare(h[:], other[:]) == 1
}
