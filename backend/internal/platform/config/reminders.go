package config

import (
	"context"
	"errors"
	"fmt"
)

// Reminders module key material (M9-05, MODULE-reminders §2.3, OQ-R7): the
// push token is sealed with the same application envelope as the RFC, under
// its own keys. Hex in the secret source, decoded here.
const (
	// secretPushTokenKey is the AES-256-GCM key sealing devices.token_ciphertext.
	secretPushTokenKey = "PUSH_TOKEN_ENC_KEY" //nolint:gosec // G101: a secret's name, not its value
	// secretPushTokenIndexPepper keys the token_hmac blind index (crypto.NewTokenIndex).
	secretPushTokenIndexPepper = "PUSH_TOKEN_INDEX_PEPPER" //nolint:gosec // G101: a secret's name, not its value
	// keyPushTokenKeyID names PushTokenKey in every ciphertext (token_key_id).
	keyPushTokenKeyID = "PUSH_TOKEN_ENC_KEY_ID" //nolint:gosec // G101: a variable's name, not a credential

	defaultPushTokenKeyID = 1
)

// Reminders is the reminders module's key material, shared by the API
// (registerDevice seals) and the worker (the reminder job opens). The
// secrets hold decoded bytes: use RevealBytes at the point of use.
type Reminders struct {
	// PushTokenKey is zero only in dev when PUSH_TOKEN_ENC_KEY is not set:
	// cmd/api then uses an ephemeral key.
	PushTokenKey Secret
	// PushTokenKeyID names PushTokenKey in every ciphertext (crypto.Keyring).
	PushTokenKeyID int
	// PushTokenIndexPepper is zero only in dev when PUSH_TOKEN_INDEX_PEPPER
	// is not set.
	PushTokenIndexPepper Secret
	// PushTokenPrevious is the retired push token key while a rotation is in
	// flight (PUSH_TOKEN_ENC_PREVIOUS_KEY, ADR-0012); zero otherwise.
	PushTokenPrevious PreviousKey
}

// LoadReminders reads and validates the reminders keys: PUSH_TOKEN_ENC_KEY
// (exactly 32 bytes) and PUSH_TOKEN_INDEX_PEPPER (at least 32 bytes), both
// hex, and PUSH_TOKEN_ENC_KEY_ID (default 1). Outside dev both keys are
// required; in dev a missing one is left zero. Each key must differ from the
// other, from every auth key and from the RFC keys, in every environment
// (one secret, one purpose). The error joins every problem and never
// contains key material.
func LoadReminders(ctx context.Context, lookup Lookup, secrets SecretSource, env Environment, auth Auth, taxpayers Taxpayers) (Reminders, error) {
	var cfg Reminders
	errs := loadModuleKeys(ctx, secrets, env, []moduleKey{
		{secretPushTokenKey, keySize{bytes: keyBytes, exact: true}, &cfg.PushTokenKey},
		{secretPushTokenIndexPepper, keySize{bytes: keyBytes}, &cfg.PushTokenIndexPepper},
	})
	var err error
	if cfg.PushTokenKeyID, err = parseIntInRange(lookup, keyPushTokenKeyID, defaultPushTokenKeyID, 1, maxEmailKeyID); err != nil {
		errs = append(errs, err)
	}
	var previousErrs []error
	cfg.PushTokenPrevious, previousErrs = loadPreviousKey(ctx, lookup, secrets, secretPushTokenKey, keyPushTokenKeyID, cfg.PushTokenKeyID)
	errs = append(errs, previousErrs...)
	previous, _ := previousKeyNames(secretPushTokenKey)
	errs = append(errs, distinctSecrets(
		[]namedSecret{
			{secretPushTokenKey, cfg.PushTokenKey}, {secretPushTokenIndexPepper, cfg.PushTokenIndexPepper},
			{previous, cfg.PushTokenPrevious.Key},
		},
		append(authSecrets(auth), taxpayersSecrets(taxpayers)...),
	)...)
	if len(errs) > 0 {
		return Reminders{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}
