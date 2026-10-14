package config

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
)

// Taxpayers module key material (M2-07, MODULE-taxpayers S-1/S-2): hex in
// the secret source, decoded here.
const (
	// secretRFCKey is the AES-256-GCM key sealing taxpayers.rfc_encrypted.
	secretRFCKey = "RFC_ENC_KEY" //nolint:gosec // G101: a secret's name, not its value
	// secretRFCIndexPepper keys the rfc_hash blind index (crypto.NewRFCIndex).
	secretRFCIndexPepper = "RFC_INDEX_PEPPER" //nolint:gosec // G101: a secret's name, not its value
	// keyRFCKeyID names RFCKey in every ciphertext (rfc_key_id).
	keyRFCKeyID = "RFC_ENC_KEY_ID"

	defaultRFCKeyID = 1
)

// Taxpayers is the taxpayers module's key material. The secrets hold decoded
// bytes: use RevealBytes at the point of use.
type Taxpayers struct {
	// RFCKey is zero only in dev when RFC_ENC_KEY is not set: cmd/api then
	// uses an ephemeral key.
	RFCKey Secret
	// RFCKeyID names RFCKey in every ciphertext (crypto.Keyring).
	RFCKeyID int
	// RFCIndexPepper is zero only in dev when RFC_INDEX_PEPPER is not set.
	RFCIndexPepper Secret
}

// namedSecret is a decoded secret and the name it was loaded under.
type namedSecret struct {
	name  string
	value Secret
}

// LoadTaxpayers reads and validates the taxpayers keys: RFC_ENC_KEY (exactly
// 32 bytes) and RFC_INDEX_PEPPER (at least 32 bytes), both hex, and
// RFC_ENC_KEY_ID (default 1). Outside dev both keys are required; in dev a
// missing one is left zero. Each key must differ from the other and from
// every auth key (PR #59 security gate: one secret, one purpose), in every
// environment. The error joins every problem and never contains key material.
func LoadTaxpayers(ctx context.Context, lookup Lookup, secrets SecretSource, env Environment, auth Auth) (Taxpayers, error) {
	var (
		cfg  Taxpayers
		errs []error
		err  error
	)
	keys := []struct {
		name string
		size keySize
		dst  *Secret
	}{
		{secretRFCKey, keySize{bytes: keyBytes, exact: true}, &cfg.RFCKey},
		{secretRFCIndexPepper, keySize{bytes: keyBytes}, &cfg.RFCIndexPepper},
	}
	for _, k := range keys {
		*k.dst, err = loadHexKey(ctx, secrets, k.name, k.size)
		if err != nil && !(env == EnvDev && errors.Is(err, ErrSecretNotFound)) {
			errs = append(errs, err)
		}
	}
	if cfg.RFCKeyID, err = parseIntInRange(lookup, keyRFCKeyID, defaultRFCKeyID, 1, maxEmailKeyID); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, distinctSecrets(
		[]namedSecret{{secretRFCKey, cfg.RFCKey}, {secretRFCIndexPepper, cfg.RFCIndexPepper}},
		[]namedSecret{
			{secretEmailPepper, auth.EmailPepper}, {secretEmailKey, auth.EmailKey},
			{secretCodeKey, auth.CodeKey}, {secretAccessTokenKey, auth.AccessTokenKey},
			{secretPreviousAccessTokenKey, auth.PreviousAccessTokenKey},
		},
	)...)
	if len(errs) > 0 {
		return Taxpayers{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}

// distinctSecrets reports every secret of owned equal to an earlier one of
// owned or to any of others. Zero secrets are skipped. The errors name the
// secrets, never their values.
func distinctSecrets(owned, others []namedSecret) []error {
	var errs []error
	for i, s := range owned {
		if s.value.IsZero() {
			continue
		}
		for _, o := range append(owned[:i:i], others...) {
			if !o.value.IsZero() && subtle.ConstantTimeCompare(s.value.RevealBytes(), o.value.RevealBytes()) == 1 {
				errs = append(errs, fmt.Errorf("%s must differ from %s", s.name, o.name))
			}
		}
	}
	return errs
}
