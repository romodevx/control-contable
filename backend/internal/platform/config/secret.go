package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

// Redacted replaces a secret wherever it would be printed.
const Redacted = "[REDACTED]"

// ErrSecretNotFound reports that a SecretSource has no value for a name.
var ErrSecretNotFound = errors.New("config: secret not found")

// SecretSource is the port to wherever secrets live: environment variables in
// dev and tests (EnvSecrets), the cloud secret manager in shared environments
// (an infra adapter wired in cmd/).
type SecretSource interface {
	// Secret returns the value of name, or an error wrapping ErrSecretNotFound.
	Secret(ctx context.Context, name string) (Secret, error)
}

// Secret holds a sensitive value. Every printing path (fmt verbs, JSON, slog)
// renders Redacted; only Reveal returns the value.
type Secret struct {
	value string
}

// NewSecret wraps value.
func NewSecret(value string) Secret { return Secret{value: value} }

// Reveal returns the secret value. Call it only at the point of use.
func (s Secret) Reveal() string { return s.value }

// RevealBytes returns the secret value as a new byte slice (decoded key
// material). Call it only at the point of use.
func (s Secret) RevealBytes() []byte { return []byte(s.value) }

// IsZero reports whether the secret is empty.
func (s Secret) IsZero() bool { return s.value == "" }

// String implements fmt.Stringer, for callers outside fmt (templates, etc.).
func (Secret) String() string { return Redacted }

// Format implements fmt.Formatter, so no verb (%v, %+v, %#v, %x, …) can print
// the value.
func (Secret) Format(f fmt.State, _ rune) { _, _ = f.Write([]byte(Redacted)) }

// LogValue implements slog.LogValuer.
func (Secret) LogValue() slog.Value { return slog.StringValue(Redacted) }

// MarshalJSON implements json.Marshaler.
func (Secret) MarshalJSON() ([]byte, error) { return []byte(`"` + Redacted + `"`), nil }

// EnvSecrets is a SecretSource backed by environment variables, for local
// development and tests. Shared environments use the secret manager adapter.
type EnvSecrets struct {
	lookup Lookup
}

// NewEnvSecrets returns an EnvSecrets reading through lookup.
func NewEnvSecrets(lookup Lookup) EnvSecrets { return EnvSecrets{lookup: lookup} }

// Secret implements SecretSource. An empty variable counts as absent.
func (e EnvSecrets) Secret(_ context.Context, name string) (Secret, error) {
	value, ok := e.lookup(name)
	if !ok || value == "" {
		return Secret{}, fmt.Errorf("env %s: %w", name, ErrSecretNotFound)
	}
	return NewSecret(value), nil
}
