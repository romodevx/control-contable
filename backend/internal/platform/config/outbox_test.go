package config_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

func loadOutbox(env map[string]string, auth config.Auth) (config.AuthOutbox, error) {
	lookup := lookupFrom(env)
	return config.LoadAuthOutbox(context.Background(), lookup, config.NewEnvSecrets(lookup), auth)
}

func TestLoadAuthOutbox(t *testing.T) {
	got, err := loadOutbox(map[string]string{"AUTH_OUTBOX_KEY": hexKey(0x09)}, config.Auth{})

	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0x09}, 32), got.Key.RevealBytes())
	assert.Equal(t, 1, got.KeyID)

	got, err = loadOutbox(map[string]string{"AUTH_OUTBOX_KEY": hexKey(0x09), "AUTH_OUTBOX_KEY_ID": "3"}, config.Auth{})
	require.NoError(t, err)
	assert.Equal(t, 3, got.KeyID)
}

// The API and the worker share this key, so it is required in every
// environment: an ephemeral dev key would leave the worker unable to open
// what the API sealed.
func TestLoadAuthOutbox_errors(t *testing.T) {
	auth, err := loadAuth(authEnv())
	require.NoError(t, err)
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"missing", map[string]string{}, "secret AUTH_OUTBOX_KEY"},
		{"not 32 bytes", map[string]string{"AUTH_OUTBOX_KEY": strings.Repeat("ab", 31)}, "secret AUTH_OUTBOX_KEY must be exactly 32 bytes"},
		{"key id out of range", map[string]string{"AUTH_OUTBOX_KEY": hexKey(0x09), "AUTH_OUTBOX_KEY_ID": "0"}, `AUTH_OUTBOX_KEY_ID "0"`},
		{"reuses an auth key", map[string]string{"AUTH_OUTBOX_KEY": hexKey(0x02)}, "AUTH_OUTBOX_KEY must differ from AUTH_EMAIL_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadOutbox(tt.env, auth)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.NotContains(t, err.Error(), hexKey(0x02), "a key never reaches the error")
		})
	}
}
