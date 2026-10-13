package infra

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

const (
	// AccessTokenLeeway is the clock skew tolerated on exp, nbf and iat:
	// small, so an expired token stops working within seconds.
	AccessTokenLeeway = 30 * time.Second
	// MaxAccessTokenBytes bounds a bearer token before any decoding: the
	// tokens this API signs are about 400 bytes.
	MaxAccessTokenBytes = 4096
)

// errKeyID reports a missing or unknown kid header.
var errKeyID = errors.New("unknown or missing kid")

// PublicKeyFromSeed returns the Ed25519 public key of a 32-byte seed: what a
// verifier needs of a signing key, and all a retired key keeps doing.
func PublicKeyFromSeed(seed []byte) (ed25519.PublicKey, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("%w: access token seed has %d bytes, want %d", ErrWeakKey, len(seed), ed25519.SeedSize)
	}
	public, ok := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("auth infra: ed25519 public key has an unexpected type")
	}
	return public, nil
}

// AccessTokenVerifier implements usecase.AccessTokenVerifier over a JWKS-style
// key set: one Ed25519 public key per kid (the active one and, during a
// rotation, the retired one). It accepts only what AccessTokenSigner issues:
// alg EdDSA, a known kid, this issuer and audience, sub a canonical user id,
// a jti, and exp, nbf and iat all present and within AccessTokenLeeway of now,
// with a lifetime no longer than AccessTokenTTL.
type AccessTokenVerifier struct {
	keys   map[string]ed25519.PublicKey
	parser *jwt.Parser
}

var _ usecase.AccessTokenVerifier = (*AccessTokenVerifier)(nil)

// NewAccessTokenVerifier returns a verifier for keys (kid → public key) that
// reads the time from now.
func NewAccessTokenVerifier(now func() time.Time, keys map[string]ed25519.PublicKey) (*AccessTokenVerifier, error) {
	if len(keys) == 0 {
		return nil, errors.New("auth infra: no access token verification key")
	}
	for kid, key := range keys {
		if kid == "" {
			return nil, errors.New("auth infra: access token key id is empty")
		}
		if len(key) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("%w: access token public key %q has %d bytes, want %d", ErrWeakKey, kid, len(key), ed25519.PublicKeySize)
		}
	}
	return &AccessTokenVerifier{
		keys: maps.Clone(keys),
		parser: jwt.NewParser(
			// The algorithm allowlist is what defeats alg confusion: "none",
			// HS256 keyed with the public key and RS256 never reach a key.
			jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
			jwt.WithIssuer(AccessTokenIssuer),
			jwt.WithAudience(AccessTokenAudience),
			jwt.WithExpirationRequired(),
			jwt.WithNotBeforeRequired(),
			jwt.WithIssuedAt(),
			jwt.WithLeeway(AccessTokenLeeway),
			jwt.WithTimeFunc(now),
		),
	}, nil
}

// Verify implements usecase.AccessTokenVerifier.
func (v *AccessTokenVerifier) Verify(token string) (usecase.AccessClaims, error) {
	if token == "" || len(token) > MaxAccessTokenBytes {
		return usecase.AccessClaims{}, fmt.Errorf("%w: empty or oversized", usecase.ErrInvalidAccessToken)
	}
	var claims accessClaims
	if _, err := v.parser.ParseWithClaims(token, &claims, v.key); err != nil {
		return usecase.AccessClaims{}, fmt.Errorf("%w: %w", usecase.ErrInvalidAccessToken, err)
	}
	return usecase.AccessClaims{UserID: claims.userID, TokenID: claims.ID, ExpiresAt: claims.ExpiresAt.Time}, nil
}

// key is the jwt.Keyfunc: the public key named by the kid header. The
// parser has already refused every algorithm but EdDSA.
func (v *AccessTokenVerifier) key(token *jwt.Token) (any, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errKeyID
	}
	key, ok := v.keys[kid]
	if !ok {
		return nil, errKeyID
	}
	return key, nil
}

// accessClaims are the registered claims plus the checks the parser options
// do not cover. Validate runs only after the signature verified.
type accessClaims struct {
	jwt.RegisteredClaims
	userID domain.UserID
}

// Validate implements jwt.ClaimsValidator.
func (c *accessClaims) Validate() error {
	id, err := domain.ParseUserID(c.Subject)
	if err != nil || id.String() != c.Subject {
		return errors.New("sub is not a canonical user id")
	}
	if c.ID == "" {
		return errors.New("jti is missing")
	}
	if c.IssuedAt == nil || c.ExpiresAt == nil || c.ExpiresAt.Sub(c.IssuedAt.Time) > AccessTokenTTL {
		return errors.New("lifetime exceeds the access token TTL")
	}
	c.userID = id
	return nil
}
