// Package infra holds the auth module's adapters: the PostgreSQL stores
// behind the use case ports (GORM over the unit of work in ctx, reaching the
// service tables only through the SECURITY DEFINER functions), the
// verification-code HMAC, the access-token signer and the dev email sender.
package infra

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// ErrWeakKey reports key material shorter than the algorithm needs.
var ErrWeakKey = errors.New("auth infra: key too short")

// minHMACKeyBytes is the shortest HMAC-SHA256 key: the hash output size.
const minHMACKeyBytes = sha256.Size

// redacted is how key-holding types render in fmt and slog.
const redacted = "REDACTED"

// CodeHasher computes code_hash = HMAC-SHA256(key, email_hash || code)
// (MODULE-auth §2). The email hash is a fixed 32 bytes and the code a fixed 6
// digits, so the concatenation is unambiguous; keying it makes the 10^6 code
// space useless to anyone holding the table but not the key.
type CodeHasher struct {
	key []byte
}

// NewCodeHasher returns a CodeHasher under key (at least 32 bytes).
func NewCodeHasher(key []byte) (*CodeHasher, error) {
	if len(key) < minHMACKeyBytes {
		return nil, fmt.Errorf("%w: code key has %d bytes, want at least %d", ErrWeakKey, len(key), minHMACKeyBytes)
	}
	return &CodeHasher{key: bytes.Clone(key)}, nil
}

// HashCode implements usecase.CodeHasher.
func (h *CodeHasher) HashCode(emailHash []byte, code domain.VerificationCode) []byte {
	mac := hmac.New(sha256.New, h.key)
	mac.Write(emailHash) // hash.Hash.Write never returns an error
	mac.Write([]byte(code.Reveal()))
	return mac.Sum(nil)
}

// Format redacts the key for every fmt verb.
func (CodeHasher) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the key in slog.
func (CodeHasher) LogValue() slog.Value { return slog.StringValue(redacted) }

// Access token claims (ADR-0001 O6). M1-07's authenticator validates exactly
// these: EdDSA only, this issuer and audience, expiry required.
const (
	AccessTokenIssuer   = "control-contable"
	AccessTokenAudience = "control-contable-api"
	AccessTokenTTL      = 15 * time.Minute
)

// AccessTokenSigner signs access tokens: EdDSA (Ed25519) JWTs with a kid
// header, sub = user id, 15-minute expiry and a random jti.
type AccessTokenSigner struct {
	key   ed25519.PrivateKey
	keyID string
}

// NewAccessTokenSigner returns a signer for the Ed25519 seed (32 bytes) named
// keyID in the kid header.
func NewAccessTokenSigner(seed []byte, keyID string) (*AccessTokenSigner, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("%w: access token seed has %d bytes, want %d", ErrWeakKey, len(seed), ed25519.SeedSize)
	}
	if keyID == "" {
		return nil, errors.New("auth infra: access token key id is empty")
	}
	return &AccessTokenSigner{key: ed25519.NewKeyFromSeed(seed), keyID: keyID}, nil
}

// Sign implements usecase.AccessTokens.
func (s *AccessTokenSigner) Sign(userID domain.UserID, issuedAt time.Time) (string, time.Duration, error) {
	jti, err := uuid.NewRandom()
	if err != nil {
		return "", 0, fmt.Errorf("auth infra: token id: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Issuer:    AccessTokenIssuer,
		Subject:   userID.String(),
		Audience:  jwt.ClaimStrings{AccessTokenAudience},
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(AccessTokenTTL)),
		ID:        jti.String(),
	})
	token.Header["kid"] = s.keyID
	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", 0, fmt.Errorf("auth infra: sign access token: %w", err)
	}
	return signed, AccessTokenTTL, nil
}

// Format redacts the key for every fmt verb.
func (AccessTokenSigner) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the key in slog.
func (AccessTokenSigner) LogValue() slog.Value { return slog.StringValue(redacted) }
