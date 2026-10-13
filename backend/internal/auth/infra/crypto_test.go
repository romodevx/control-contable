package infra_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
)

var (
	codeKey   = bytes.Repeat([]byte{0x03}, 32)
	tokenSeed = bytes.Repeat([]byte{0x04}, ed25519.SeedSize)
	issuedAt  = time.Date(2026, 10, 13, 18, 30, 0, 0, time.UTC)
)

func code(t *testing.T, digits string) domain.VerificationCode {
	t.Helper()
	c, err := domain.ParseVerificationCode(digits)
	require.NoError(t, err)
	return c
}

// code_hash = HMAC-SHA256(server key, email_hash || code): MODULE-auth §2.
func TestCodeHasher_isTheKeyedHMACOfEmailHashAndCode(t *testing.T) {
	h, err := infra.NewCodeHasher(codeKey)
	require.NoError(t, err)
	emailHash := bytes.Repeat([]byte{0xAA}, 32)

	got := h.HashCode(emailHash, code(t, "482913"))

	mac := hmac.New(sha256.New, codeKey)
	mac.Write(emailHash)
	mac.Write([]byte("482913"))
	assert.Equal(t, mac.Sum(nil), got)
	assert.Len(t, got, 32)
}

func TestCodeHasher_isNeverAPlainDigest(t *testing.T) {
	h, err := infra.NewCodeHasher(codeKey)
	require.NoError(t, err)
	emailHash := bytes.Repeat([]byte{0xAA}, 32)

	plain := sha256.Sum256(append(bytes.Clone(emailHash), "482913"...))
	assert.NotEqual(t, plain[:], h.HashCode(emailHash, code(t, "482913")))
}

func TestCodeHasher_bindsTheEmailAndTheKey(t *testing.T) {
	h, err := infra.NewCodeHasher(codeKey)
	require.NoError(t, err)
	other, err := infra.NewCodeHasher(bytes.Repeat([]byte{0x09}, 32))
	require.NoError(t, err)
	a, b := bytes.Repeat([]byte{0xAA}, 32), bytes.Repeat([]byte{0xBB}, 32)

	assert.NotEqual(t, h.HashCode(a, code(t, "111111")), h.HashCode(b, code(t, "111111")), "same code, other email")
	assert.NotEqual(t, h.HashCode(a, code(t, "111111")), other.HashCode(a, code(t, "111111")), "same input, other key")
}

func TestNewCodeHasher_rejectsAShortKey(t *testing.T) {
	_, err := infra.NewCodeHasher(make([]byte, 31))
	assert.ErrorIs(t, err, infra.ErrWeakKey)
}

func TestCodeHasher_neverPrintsTheKey(t *testing.T) {
	h, err := infra.NewCodeHasher(codeKey)
	require.NoError(t, err)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		assert.Equal(t, "REDACTED", fmt.Sprintf(verb, h))
		assert.Equal(t, "REDACTED", fmt.Sprintf(verb, *h))
	}
}

func signer(t *testing.T) *infra.AccessTokenSigner {
	t.Helper()
	s, err := infra.NewAccessTokenSigner(tokenSeed, "k1")
	require.NoError(t, err)
	return s
}

func TestAccessTokenSigner_signsAnEdDSAJWTThatVerifies(t *testing.T) {
	userID, err := domain.ParseUserID("0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01")
	require.NoError(t, err)

	token, expiresIn, err := signer(t).Sign(userID, issuedAt)

	require.NoError(t, err)
	assert.Equal(t, infra.AccessTokenTTL, expiresIn)
	assert.Equal(t, 15*time.Minute, expiresIn, "ADR-0001 O6")
	public := ed25519.NewKeyFromSeed(tokenSeed).Public()
	parsed, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) { return public, nil },
		jwt.WithValidMethods([]string{"EdDSA"}),
		jwt.WithIssuer(infra.AccessTokenIssuer),
		jwt.WithAudience(infra.AccessTokenAudience),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return issuedAt.Add(time.Minute) }),
	)
	require.NoError(t, err)
	assert.Equal(t, "k1", parsed.Header["kid"])
	claims := parsed.Claims.(*jwt.RegisteredClaims)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, issuedAt, claims.IssuedAt.UTC())
	require.NotNil(t, claims.NotBefore, "nbf is set (M1-07)")
	assert.Equal(t, issuedAt, claims.NotBefore.UTC())
	assert.Equal(t, issuedAt.Add(15*time.Minute), claims.ExpiresAt.UTC())
	assert.NotEmpty(t, claims.ID, "a unique jti")
}

func TestAccessTokenSigner_expiresAfterFifteenMinutes(t *testing.T) {
	userID, err := domain.ParseUserID("0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01")
	require.NoError(t, err)
	token, _, err := signer(t).Sign(userID, issuedAt)
	require.NoError(t, err)

	_, err = jwt.Parse(token, func(*jwt.Token) (any, error) { return ed25519.NewKeyFromSeed(tokenSeed).Public(), nil },
		jwt.WithTimeFunc(func() time.Time { return issuedAt.Add(16 * time.Minute) }))

	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestNewAccessTokenSigner_validatesItsKey(t *testing.T) {
	_, err := infra.NewAccessTokenSigner(make([]byte, 31), "k1")
	assert.ErrorIs(t, err, infra.ErrWeakKey)
	_, err = infra.NewAccessTokenSigner(tokenSeed, "")
	assert.Error(t, err, "a kid is required")
}

func TestAccessTokenSigner_neverPrintsTheKey(t *testing.T) {
	s := signer(t)
	assert.Equal(t, "REDACTED", fmt.Sprintf("%+v", s))
	assert.Equal(t, "REDACTED", fmt.Sprintf("%#v", *s))
}
