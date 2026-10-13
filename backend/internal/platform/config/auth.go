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
	// secretPreviousAccessTokenKey is the optional Ed25519 seed of the
	// retired signing key: it only verifies, until its tokens expire.
	secretPreviousAccessTokenKey = "AUTH_ACCESS_TOKEN_PREVIOUS_KEY" //nolint:gosec // G101: a name or message, not a credential

	keyEmailKeyID       = "AUTH_EMAIL_KEY_ID"
	keyAccessTokenKeyID = "AUTH_ACCESS_TOKEN_KEY_ID" //nolint:gosec // G101: a name or message, not a credential
	// keyPreviousAccessTokenKeyID is the kid of the retired signing key.
	keyPreviousAccessTokenKeyID = "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID" //nolint:gosec // G101: a name or message, not a credential

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
	// PreviousAccessTokenKey is the Ed25519 seed of the key AccessTokenKey
	// replaced (zero when there is none). Tokens it signed keep verifying
	// under PreviousAccessTokenKeyID until they expire; it never signs.
	PreviousAccessTokenKey Secret
	// PreviousAccessTokenKeyID is the kid of PreviousAccessTokenKey.
	PreviousAccessTokenKeyID string
}

// keySize says how long a decoded key must be.
type keySize struct {
	bytes int
	exact bool
}

// LoadAuth reads and validates the auth keys: AUTH_EMAIL_PEPPER and
// AUTH_CODE_KEY (at least 32 bytes), AUTH_EMAIL_KEY and AUTH_ACCESS_TOKEN_KEY
// (exactly 32 bytes), all hex; AUTH_EMAIL_KEY_ID (default 1) and
// AUTH_ACCESS_TOKEN_KEY_ID (default k1). For a key rotation,
// AUTH_ACCESS_TOKEN_PREVIOUS_KEY (32 bytes, hex) and
// AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID name the retired key, both or neither.
// The error joins every problem and never contains key material.
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
	if err := checkKeyID(keyAccessTokenKeyID, cfg.AccessTokenKeyID); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, loadPreviousAccessTokenKey(ctx, lookup, secrets, &cfg)...)
	if len(errs) > 0 {
		return Auth{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}

// loadPreviousAccessTokenKey reads the optional retired signing key into cfg.
// It runs after the active key is loaded, to compare the two.
func loadPreviousAccessTokenKey(ctx context.Context, lookup Lookup, secrets SecretSource, cfg *Auth) []error {
	keyID := valueOr(lookup, keyPreviousAccessTokenKeyID, "")
	encoded, err := secrets.Secret(ctx, secretPreviousAccessTokenKey)
	switch {
	case errors.Is(err, ErrSecretNotFound) && keyID == "":
		return nil
	case errors.Is(err, ErrSecretNotFound):
		return []error{fmt.Errorf("%s is set without %s", keyPreviousAccessTokenKeyID, secretPreviousAccessTokenKey)}
	case err != nil:
		return []error{fmt.Errorf("secret %s: %w", secretPreviousAccessTokenKey, err)}
	case encoded.IsZero():
		return nil
	}
	var errs []error
	key, err := decodeHexKey(secretPreviousAccessTokenKey, encoded, keySize{bytes: keyBytes, exact: true})
	if err != nil {
		errs = append(errs, err)
	}
	if !key.IsZero() && key.Reveal() == cfg.AccessTokenKey.Reveal() {
		errs = append(errs, fmt.Errorf("%s must differ from %s", secretPreviousAccessTokenKey, secretAccessTokenKey))
	}
	switch keyID {
	case "":
		errs = append(errs, fmt.Errorf("%s is required with %s", keyPreviousAccessTokenKeyID, secretPreviousAccessTokenKey))
	case cfg.AccessTokenKeyID:
		errs = append(errs, fmt.Errorf("%s must differ from %s", keyPreviousAccessTokenKeyID, keyAccessTokenKeyID))
	default:
		if err := checkKeyID(keyPreviousAccessTokenKeyID, keyID); err != nil {
			errs = append(errs, err)
		}
	}
	cfg.PreviousAccessTokenKey, cfg.PreviousAccessTokenKeyID = key, keyID
	return errs
}

// checkKeyID bounds a JWT kid value.
func checkKeyID(name, value string) error {
	if !accessTokenKeyIDPattern.MatchString(value) {
		return fmt.Errorf("%s %q is not 1-64 of [A-Za-z0-9._-]", name, value)
	}
	return nil
}

func loadHexKey(ctx context.Context, secrets SecretSource, name string, size keySize) (Secret, error) {
	encoded, err := secrets.Secret(ctx, name)
	if err != nil {
		return Secret{}, fmt.Errorf("secret %s: %w", name, err)
	}
	return decodeHexKey(name, encoded, size)
}

// decodeHexKey decodes a hex secret and checks its decoded size.
func decodeHexKey(name string, encoded Secret, size keySize) (Secret, error) {
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
