package infra_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

const (
	subjectID = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"
	claimJTI  = "6f1c2a4e-1d1b-4c55-8a0e-5e3f1f9b7d21"
)

var (
	previousSeed = bytes.Repeat([]byte{0x05}, ed25519.SeedSize)
	strangerSeed = bytes.Repeat([]byte{0x06}, ed25519.SeedSize)
)

func publicKey(t *testing.T, seed []byte) ed25519.PublicKey {
	t.Helper()
	key, err := infra.PublicKeyFromSeed(seed)
	require.NoError(t, err)
	return key
}

// verifierAt verifies with k1 (active, tokenSeed) and k0 (previous) at now.
func verifierAt(t *testing.T, now time.Time) *infra.AccessTokenVerifier {
	t.Helper()
	v, err := infra.NewAccessTokenVerifier(func() time.Time { return now }, map[string]ed25519.PublicKey{
		"k1": publicKey(t, tokenSeed),
		"k0": publicKey(t, previousSeed),
	})
	require.NoError(t, err)
	return v
}

// validClaims are the claims the signer issues at issuedAt.
func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss": infra.AccessTokenIssuer,
		"aud": []string{infra.AccessTokenAudience},
		"sub": subjectID,
		"iat": issuedAt.Unix(),
		"nbf": issuedAt.Unix(),
		"exp": issuedAt.Add(infra.AccessTokenTTL).Unix(),
		"jti": claimJTI,
	}
}

// signed returns claims as an EdDSA JWT signed by seed under kid ("" = no kid).
func signed(t *testing.T, seed []byte, kid any, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	if kid != nil {
		token.Header["kid"] = kid
	}
	s, err := token.SignedString(ed25519.NewKeyFromSeed(seed))
	require.NoError(t, err)
	return s
}

func withClaim(name string, value any) jwt.MapClaims {
	c := validClaims()
	if value == nil {
		delete(c, name)
		return c
	}
	c[name] = value
	return c
}

func TestAccessTokenVerifier_acceptsWhatTheSignerIssues(t *testing.T) {
	userID, err := domain.ParseUserID(subjectID)
	require.NoError(t, err)
	token, _, err := signer(t).Sign(userID, issuedAt)
	require.NoError(t, err)

	claims, err := verifierAt(t, issuedAt.Add(time.Minute)).Verify(token)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.NotEmpty(t, claims.TokenID)
	assert.Equal(t, issuedAt.Add(infra.AccessTokenTTL), claims.ExpiresAt.UTC())
}

// Kid rotation: a token signed by the retired key verifies under its own kid
// until it expires.
func TestAccessTokenVerifier_acceptsThePreviousKeyUnderItsKid(t *testing.T) {
	claims, err := verifierAt(t, issuedAt.Add(time.Minute)).Verify(signed(t, previousSeed, "k0", validClaims()))

	require.NoError(t, err)
	assert.Equal(t, subjectID, claims.UserID.String())
	assert.Equal(t, claimJTI, claims.TokenID)
}

func TestAccessTokenVerifier_toleratesSmallClockSkew(t *testing.T) {
	v := verifierAt(t, issuedAt.Add(infra.AccessTokenTTL).Add(infra.AccessTokenLeeway-time.Second))
	_, err := v.Verify(signed(t, tokenSeed, "k1", validClaims()))
	require.NoError(t, err, "just expired, within the leeway")

	v = verifierAt(t, issuedAt.Add(-infra.AccessTokenLeeway+time.Second))
	_, err = v.Verify(signed(t, tokenSeed, "k1", validClaims()))
	require.NoError(t, err, "issued slightly in the verifier's future, within the leeway")
}

func TestAccessTokenVerifier_rejects(t *testing.T) {
	inTwoMinutes := issuedAt.Add(2 * time.Minute)
	tests := []struct {
		name  string
		token func(t *testing.T) string
		now   time.Time
	}{
		{"alg none", func(t *testing.T) string {
			token := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims())
			token.Header["kid"] = "k1"
			s, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
			require.NoError(t, err)
			return s
		}, inTwoMinutes},
		{"HS256 keyed with the public key (alg confusion)", func(t *testing.T) string {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
			token.Header["kid"] = "k1"
			s, err := token.SignedString([]byte(publicKey(t, tokenSeed)))
			require.NoError(t, err)
			return s
		}, inTwoMinutes},
		{"RS256", func(t *testing.T) string {
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, validClaims())
			token.Header["kid"] = "k1"
			s, err := token.SignedString(key)
			require.NoError(t, err)
			return s
		}, inTwoMinutes},
		{"unknown kid", func(t *testing.T) string { return signed(t, tokenSeed, "k9", validClaims()) }, inTwoMinutes},
		{"no kid", func(t *testing.T) string { return signed(t, tokenSeed, nil, validClaims()) }, inTwoMinutes},
		{"kid not a string", func(t *testing.T) string { return signed(t, tokenSeed, 1, validClaims()) }, inTwoMinutes},
		{"another key under a known kid", func(t *testing.T) string { return signed(t, strangerSeed, "k1", validClaims()) }, inTwoMinutes},
		{"the previous key under the active kid", func(t *testing.T) string { return signed(t, previousSeed, "k1", validClaims()) }, inTwoMinutes},
		{"expired", func(t *testing.T) string { return signed(t, tokenSeed, "k1", validClaims()) },
			issuedAt.Add(infra.AccessTokenTTL).Add(infra.AccessTokenLeeway + time.Second)},
		{"not yet valid (nbf)", func(t *testing.T) string {
			return signed(t, tokenSeed, "k1", withClaim("nbf", issuedAt.Add(5*time.Minute).Unix()))
		}, inTwoMinutes},
		{"issued in the future (iat)", func(t *testing.T) string { return signed(t, tokenSeed, "k1", validClaims()) },
			issuedAt.Add(-infra.AccessTokenLeeway - time.Second)},
		{"no exp", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("exp", nil)) }, inTwoMinutes},
		{"no nbf", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("nbf", nil)) }, inTwoMinutes},
		{"no iat", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("iat", nil)) }, inTwoMinutes},
		{"wrong issuer", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("iss", "someone-else")) }, inTwoMinutes},
		{"no issuer", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("iss", nil)) }, inTwoMinutes},
		{"wrong audience", func(t *testing.T) string {
			return signed(t, tokenSeed, "k1", withClaim("aud", []string{"another-api"}))
		}, inTwoMinutes},
		{"no audience", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("aud", nil)) }, inTwoMinutes},
		{"sub not a uuid", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("sub", "admin")) }, inTwoMinutes},
		{"sub upper-case uuid", func(t *testing.T) string {
			return signed(t, tokenSeed, "k1", withClaim("sub", strings.ToUpper(subjectID)))
		}, inTwoMinutes},
		{"sub nil uuid", func(t *testing.T) string {
			return signed(t, tokenSeed, "k1", withClaim("sub", "00000000-0000-0000-0000-000000000000"))
		}, inTwoMinutes},
		{"no sub", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("sub", nil)) }, inTwoMinutes},
		{"no jti", func(t *testing.T) string { return signed(t, tokenSeed, "k1", withClaim("jti", nil)) }, inTwoMinutes},
		{"lifetime above the TTL", func(t *testing.T) string {
			return signed(t, tokenSeed, "k1", withClaim("exp", issuedAt.Add(24*time.Hour).Unix()))
		}, inTwoMinutes},
		{"tampered payload", func(t *testing.T) string {
			good := strings.Split(signed(t, tokenSeed, "k1", validClaims()), ".")
			other := strings.Split(signed(t, tokenSeed, "k1", withClaim("sub", "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d02")), ".")
			return good[0] + "." + other[1] + "." + good[2]
		}, inTwoMinutes},
		{"garbage", func(*testing.T) string { return "not-a-jwt" }, inTwoMinutes},
		{"empty", func(*testing.T) string { return "" }, inTwoMinutes},
		{"oversized", func(t *testing.T) string {
			return signed(t, tokenSeed, "k1", withClaim("pad", strings.Repeat("x", infra.MaxAccessTokenBytes)))
		}, inTwoMinutes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := verifierAt(t, tt.now).Verify(tt.token(t))

			assert.ErrorIs(t, err, usecase.ErrInvalidAccessToken)
		})
	}
}

func TestNewAccessTokenVerifier_validatesItsKeys(t *testing.T) {
	now := func() time.Time { return issuedAt }
	_, err := infra.NewAccessTokenVerifier(now, nil)
	require.Error(t, err, "at least one key")

	_, err = infra.NewAccessTokenVerifier(now, map[string]ed25519.PublicKey{"k1": make([]byte, 31)})
	require.ErrorIs(t, err, infra.ErrWeakKey)

	_, err = infra.NewAccessTokenVerifier(now, map[string]ed25519.PublicKey{"": publicKey(t, tokenSeed)})
	require.Error(t, err, "a kid is required")

	_, err = infra.PublicKeyFromSeed(make([]byte, 31))
	require.ErrorIs(t, err, infra.ErrWeakKey)
}
