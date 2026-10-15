package config

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Key rotation of the AES-256-GCM sealing keys (ADR-0012). Each sealing key
// X_KEY with id X_KEY_ID may have one retired key, X_PREVIOUS_KEY with id
// X_PREVIOUS_KEY_ID, both or neither: the keyring seals with X_KEY and still
// opens what X_PREVIOUS_KEY sealed until the re-seal job has moved every row.
const (
	activeKeySuffix   = "KEY"
	previousKeySuffix = "PREVIOUS_KEY"
	keyIDSuffix       = "_ID"
)

// PreviousKey is the retired key of a sealing key rotation. The zero value
// means no rotation is in flight.
type PreviousKey struct {
	Key   Secret
	KeyID int
}

// IsZero reports whether no retired key is configured.
func (p PreviousKey) IsZero() bool { return p.Key.IsZero() }

// SealingKey is one sealing key with its optional retired key. The zero
// value means the key is not configured.
type SealingKey struct {
	Key      Secret
	KeyID    int
	Previous PreviousKey
}

// IsZero reports whether the key is not configured.
func (s SealingKey) IsZero() bool { return s.Key.IsZero() }

// previousKeyNames derives the retired key's variable names from the active
// key's: RFC_ENC_KEY -> RFC_ENC_PREVIOUS_KEY and RFC_ENC_PREVIOUS_KEY_ID.
func previousKeyNames(activeKey string) (key, id string) {
	key = strings.TrimSuffix(activeKey, activeKeySuffix) + previousKeySuffix
	return key, key + keyIDSuffix
}

// loadPreviousKey reads the optional retired key of the sealing key named
// activeKey, whose id is activeID (read from activeIDName). The retired id
// must differ from the active one; that the key differs from every other key
// is the caller's distinctSecrets check. It returns every problem found.
func loadPreviousKey(ctx context.Context, lookup Lookup, secrets SecretSource,
	activeKey, activeIDName string, activeID int,
) (PreviousKey, []error) {
	keyName, idName := previousKeyNames(activeKey)
	idSet := valueOr(lookup, idName, "") != ""
	encoded, err := secrets.Secret(ctx, keyName)
	switch {
	case errors.Is(err, ErrSecretNotFound) && !idSet:
		return PreviousKey{}, nil
	case errors.Is(err, ErrSecretNotFound):
		return PreviousKey{}, []error{fmt.Errorf("%s is set without %s", idName, keyName)}
	case err != nil:
		return PreviousKey{}, []error{fmt.Errorf("secret %s: %w", keyName, err)}
	case !idSet:
		return PreviousKey{}, []error{fmt.Errorf("%s is required with %s", idName, keyName)}
	}
	var errs []error
	key, err := decodeHexKey(keyName, encoded, keySize{bytes: keyBytes, exact: true})
	if err != nil {
		errs = append(errs, err)
	}
	id, err := parseIntInRange(lookup, idName, 0, 1, maxEmailKeyID)
	switch {
	case err != nil:
		errs = append(errs, err)
	case id == activeID:
		errs = append(errs, fmt.Errorf("%s must differ from %s", idName, activeIDName))
	}
	if len(errs) > 0 {
		return PreviousKey{}, errs
	}
	return PreviousKey{Key: key, KeyID: id}, nil
}

// ResealKeys are the sealing keys cmd/worker needs only to re-seal
// (ADR-0012): the API's RFC and email keys. Each is optional in every
// environment; a zero one means the worker does not re-seal that column.
// The worker already loads the push token and outbox keys for its jobs.
type ResealKeys struct {
	RFC   SealingKey
	Email SealingKey
}

// LoadResealKeys reads the optional RFC_ENC_KEY and AUTH_EMAIL_KEY (exactly
// 32 bytes, hex) with their ids and retired keys. A retired key without its
// active key is an error, and the keys must all differ. The error joins
// every problem and never contains key material.
func LoadResealKeys(ctx context.Context, lookup Lookup, secrets SecretSource) (ResealKeys, error) {
	var (
		cfg  ResealKeys
		errs []error
	)
	errs = append(errs, loadOptionalSealingKey(ctx, lookup, secrets, secretRFCKey, keyRFCKeyID, defaultRFCKeyID, &cfg.RFC)...)
	errs = append(errs, loadOptionalSealingKey(ctx, lookup, secrets, secretEmailKey, keyEmailKeyID, defaultEmailKeyID, &cfg.Email)...)
	rfcPrevious, _ := previousKeyNames(secretRFCKey)
	emailPrevious, _ := previousKeyNames(secretEmailKey)
	errs = append(errs, distinctSecrets([]namedSecret{
		{secretRFCKey, cfg.RFC.Key}, {rfcPrevious, cfg.RFC.Previous.Key},
		{secretEmailKey, cfg.Email.Key}, {emailPrevious, cfg.Email.Previous.Key},
	}, nil)...)
	if len(errs) > 0 {
		return ResealKeys{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}

// loadOptionalSealingKey reads the sealing key name, its id and its retired
// key into dst. A missing key leaves dst zero, unless a retired key is set.
func loadOptionalSealingKey(ctx context.Context, lookup Lookup, secrets SecretSource,
	name, idName string, defaultID int, dst *SealingKey,
) []error {
	key, err := loadHexKey(ctx, secrets, name, keySize{bytes: keyBytes, exact: true})
	if errors.Is(err, ErrSecretNotFound) {
		previousName, _ := previousKeyNames(name)
		if _, perr := secrets.Secret(ctx, previousName); perr == nil {
			return []error{fmt.Errorf("%s is set without %s", previousName, name)}
		}
		return nil
	}
	if err != nil {
		return []error{err}
	}
	id, err := parseIntInRange(lookup, idName, defaultID, 1, maxEmailKeyID)
	if err != nil {
		return []error{err}
	}
	previous, errs := loadPreviousKey(ctx, lookup, secrets, name, idName, id)
	*dst = SealingKey{Key: key, KeyID: id, Previous: previous}
	return errs
}
