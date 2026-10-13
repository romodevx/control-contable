package config

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

// Auth module key material (M1-05). Every secret is hex-encoded in the secret
// source and decoded here, so a malformed key fails at startup, never at the
// first sign-up. Each key has one purpose; none is ever reused for another.
const (
	// secretEmailPepper keys the email_hash blind index (crypto.NewEmailIndex).
	secretEmailPepper = "AUTH_EMAIL_PEPPER" //nolint:gosec // G101: a secret's name, not its value
	// secretEmailKey is the AES-256-GCM key sealing users.email_encrypted.
	secretEmailKey = "AUTH_EMAIL_KEY" //nolint:gosec // G101: a secret's name, not its value
	// secretCodeKey keys the verification-code HMAC (code_hash).
	secretCodeKey = "AUTH_CODE_KEY" //nolint:gosec // G101: a secret's name, not its value
	// secretAccessTokenKey is the Ed25519 seed signing access tokens (EdDSA JWT).
	secretAccessTokenKey = "AUTH_ACCESS_TOKEN_KEY" //nolint:gosec // G101: a name or message, not a credential

	keyEmailKeyID       = "AUTH_EMAIL_KEY_ID"
	keyAccessTokenKeyID = "AUTH_ACCESS_TOKEN_KEY_ID" //nolint:gosec // G101: a name or message, not a credential

	// keyBytes is the size of the AES-256 key, the Ed25519 seed and the minimum
	// HMAC key (the SHA-256 output size).
	keyBytes = 32

	defaultEmailKeyID       = 1
	maxEmailKeyID           = 32767 // crypto.MaxKeyID
	defaultAccessTokenKeyID = "k1"
)

// accessTokenKeyIDPattern bounds the JWT "kid" header value.
var accessTokenKeyIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// Auth is the auth module's key material. The secrets hold decoded bytes:
// use RevealBytes at the point of use.
type Auth struct {
	EmailPepper Secret
	EmailKey    Secret
	// EmailKeyID names EmailKey in every ciphertext (crypto.Keyring).
	EmailKeyID int
	CodeKey    Secret
	// AccessTokenKey is the Ed25519 seed of the access-token signing key.
	AccessTokenKey Secret
	// AccessTokenKeyID is the JWT "kid" of AccessTokenKey.
	AccessTokenKeyID string
}

// keySize says how long a decoded key must be.
type keySize struct {
	bytes int
	exact bool
}

// LoadAuth reads and validates the auth keys: AUTH_EMAIL_PEPPER and
// AUTH_CODE_KEY (at least 32 bytes), AUTH_EMAIL_KEY and AUTH_ACCESS_TOKEN_KEY
// (exactly 32 bytes), all hex; AUTH_EMAIL_KEY_ID (default 1) and
// AUTH_ACCESS_TOKEN_KEY_ID (default k1). The error joins every problem and
// never contains key material.
func LoadAuth(ctx context.Context, lookup Lookup, secrets SecretSource) (Auth, error) {
	var (
		cfg  Auth
		errs []error
		err  error
	)
	keys := []struct {
		name string
		size keySize
		dst  *Secret
	}{
		{secretEmailPepper, keySize{bytes: keyBytes}, &cfg.EmailPepper},
		{secretEmailKey, keySize{bytes: keyBytes, exact: true}, &cfg.EmailKey},
		{secretCodeKey, keySize{bytes: keyBytes}, &cfg.CodeKey},
		{secretAccessTokenKey, keySize{bytes: keyBytes, exact: true}, &cfg.AccessTokenKey},
	}
	for _, k := range keys {
		if *k.dst, err = loadHexKey(ctx, secrets, k.name, k.size); err != nil {
			errs = append(errs, err)
		}
	}
	if cfg.EmailKeyID, err = parseIntInRange(lookup, keyEmailKeyID, defaultEmailKeyID, 1, maxEmailKeyID); err != nil {
		errs = append(errs, err)
	}
	cfg.AccessTokenKeyID = valueOr(lookup, keyAccessTokenKeyID, defaultAccessTokenKeyID)
	if !accessTokenKeyIDPattern.MatchString(cfg.AccessTokenKeyID) {
		errs = append(errs, fmt.Errorf("%s %q is not 1-64 of [A-Za-z0-9._-]", keyAccessTokenKeyID, cfg.AccessTokenKeyID))
	}
	if len(errs) > 0 {
		return Auth{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}

func loadHexKey(ctx context.Context, secrets SecretSource, name string, size keySize) (Secret, error) {
	encoded, err := secrets.Secret(ctx, name)
	if err != nil {
		return Secret{}, fmt.Errorf("secret %s: %w", name, err)
	}
	raw, err := hex.DecodeString(encoded.Reveal())
	if err != nil {
		// The decode error quotes the offending byte: never wrap it.
		return Secret{}, fmt.Errorf("secret %s is not hex", name)
	}
	switch {
	case size.exact && len(raw) != size.bytes:
		return Secret{}, fmt.Errorf("secret %s must be exactly %d bytes, got %d", name, size.bytes, len(raw))
	case len(raw) < size.bytes:
		return Secret{}, fmt.Errorf("secret %s must be at least %d bytes, got %d", name, size.bytes, len(raw))
	}
	return NewSecret(string(raw)), nil
}
