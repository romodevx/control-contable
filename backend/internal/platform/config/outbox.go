package config

import (
	"context"
	"errors"
	"fmt"
)

// Auth email outbox key material: the API seals each queued auth email
// (register, resend) and the worker opens it to deliver it, so both load it.
const (
	// secretOutboxKey is the AES-256-GCM key sealing auth_email_outbox.payload.
	secretOutboxKey = "AUTH_OUTBOX_KEY" //nolint:gosec // G101: a secret's name, not its value
	// keyOutboxKeyID names the key in every ciphertext (crypto.Keyring).
	keyOutboxKeyID = "AUTH_OUTBOX_KEY_ID"

	defaultOutboxKeyID = 1
)

// AuthOutbox is the outbox key shared by cmd/api and cmd/worker. Use
// RevealBytes at the point of use.
type AuthOutbox struct {
	Key Secret
	// KeyID names Key in every ciphertext (crypto.Keyring).
	KeyID int
}

// LoadAuthOutbox reads AUTH_OUTBOX_KEY (exactly 32 bytes, hex) and
// AUTH_OUTBOX_KEY_ID (default 1). It is required in every environment, dev
// included: the worker must open what the API sealed, so neither may make
// up an ephemeral key. The key must differ from every key of auth (one
// secret, one purpose); the worker, which loads no auth keys, passes a zero
// Auth. The error joins every problem and never contains key material.
func LoadAuthOutbox(ctx context.Context, lookup Lookup, secrets SecretSource, auth Auth) (AuthOutbox, error) {
	var (
		cfg  AuthOutbox
		errs []error
		err  error
	)
	if cfg.Key, err = loadHexKey(ctx, secrets, secretOutboxKey, keySize{bytes: keyBytes, exact: true}); err != nil {
		errs = append(errs, err)
	}
	if cfg.KeyID, err = parseIntInRange(lookup, keyOutboxKeyID, defaultOutboxKeyID, 1, maxEmailKeyID); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, distinctSecrets([]namedSecret{{secretOutboxKey, cfg.Key}}, authSecrets(auth))...)
	if len(errs) > 0 {
		return AuthOutbox{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}
