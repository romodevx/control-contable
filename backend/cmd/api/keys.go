package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

// errMissingKey reports a module key missing outside dev, where config
// refuses it first; it guards the wiring against a zero key anyway.
var errMissingKey = errors.New("a module key is required outside dev")

// keyOrEphemeral is the configured secret, or, in dev only, an ephemeral
// random key of size bytes, logged as such by its variable name: values
// sealed or indexed with it cannot be opened or matched after a restart.
func keyOrEphemeral(env config.Environment, logger *slog.Logger, secret config.Secret, name string, size int) ([]byte, error) {
	if !secret.IsZero() {
		return secret.RevealBytes(), nil
	}
	if env != config.EnvDev {
		return nil, fmt.Errorf("%w: %s", errMissingKey, name)
	}
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("ephemeral %s: %w", name, err)
	}
	logger.Warn("secret not set; using an ephemeral key until restart (dev only)", slog.String("variable", name))
	return key, nil
}
