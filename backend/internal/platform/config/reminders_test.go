package config_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

func remindersEnv() map[string]string {
	return map[string]string{
		"PUSH_TOKEN_ENC_KEY":      hexKey(0x0c),
		"PUSH_TOKEN_INDEX_PEPPER": hexKey(0x0d),
	}
}

func loadReminders(t *testing.T, env map[string]string, appEnv config.Environment) (config.Reminders, error) {
	t.Helper()
	auth, err := loadAuth(authEnv())
	require.NoError(t, err)
	taxpayers, err := loadTaxpayers(t, taxpayersEnv(), config.EnvProduction)
	require.NoError(t, err)
	lookup := lookupFrom(env)
	return config.LoadReminders(context.Background(), lookup, config.NewEnvSecrets(lookup), appEnv, auth, taxpayers)
}

func TestLoadReminders_decodesKeysAndAppliesDefaults(t *testing.T) {
	got, err := loadReminders(t, remindersEnv(), config.EnvProduction)

	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0x0c}, 32), got.PushTokenKey.RevealBytes())
	assert.Equal(t, bytes.Repeat([]byte{0x0d}, 32), got.PushTokenIndexPepper.RevealBytes())
	assert.Equal(t, 1, got.PushTokenKeyID)
}

func TestLoadReminders_keyIDOverride(t *testing.T) {
	env := remindersEnv()
	env["PUSH_TOKEN_ENC_KEY_ID"] = "4"

	got, err := loadReminders(t, env, config.EnvStaging)

	require.NoError(t, err)
	assert.Equal(t, 4, got.PushTokenKeyID)
}

// One secret, one purpose (PR #59 security gate): the push token keys
// differ from each other, from every auth key and from the RFC keys, in
// every environment.
func TestLoadReminders_refusesSharedSecrets(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"enc key equals index pepper", map[string]string{"PUSH_TOKEN_ENC_KEY": hexKey(0x0c), "PUSH_TOKEN_INDEX_PEPPER": hexKey(0x0c)},
			"PUSH_TOKEN_INDEX_PEPPER must differ from PUSH_TOKEN_ENC_KEY"},
		{"enc key equals the RFC key", map[string]string{"PUSH_TOKEN_ENC_KEY": hexKey(0x0a), "PUSH_TOKEN_INDEX_PEPPER": hexKey(0x0d)},
			"PUSH_TOKEN_ENC_KEY must differ from RFC_ENC_KEY"},
		{"index pepper equals the RFC pepper", map[string]string{"PUSH_TOKEN_ENC_KEY": hexKey(0x0c), "PUSH_TOKEN_INDEX_PEPPER": hexKey(0x0b)},
			"PUSH_TOKEN_INDEX_PEPPER must differ from RFC_INDEX_PEPPER"},
		{"enc key equals the email pepper", map[string]string{"PUSH_TOKEN_ENC_KEY": hexKey(0x01), "PUSH_TOKEN_INDEX_PEPPER": hexKey(0x0d)},
			"PUSH_TOKEN_ENC_KEY must differ from AUTH_EMAIL_PEPPER"},
		{"index pepper equals the code key", map[string]string{"PUSH_TOKEN_ENC_KEY": hexKey(0x0c), "PUSH_TOKEN_INDEX_PEPPER": hexKey(0x03)},
			"PUSH_TOKEN_INDEX_PEPPER must differ from AUTH_CODE_KEY"},
	}
	for _, tt := range tests {
		for _, appEnv := range []config.Environment{config.EnvDev, config.EnvProduction} {
			t.Run(tt.name+"/"+string(appEnv), func(t *testing.T) {
				_, err := loadReminders(t, tt.env, appEnv)

				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				for _, v := range tt.env {
					assert.NotContains(t, err.Error(), v, "a key never reaches the error")
				}
			})
		}
	}
}

func TestLoadReminders_missingKeysOutsideDevRefuseToStart(t *testing.T) {
	for _, appEnv := range []config.Environment{config.EnvStaging, config.EnvProduction} {
		t.Run(string(appEnv), func(t *testing.T) {
			_, err := loadReminders(t, map[string]string{}, appEnv)

			require.Error(t, err)
			assert.ErrorContains(t, err, "secret PUSH_TOKEN_ENC_KEY")
			assert.ErrorContains(t, err, "secret PUSH_TOKEN_INDEX_PEPPER")
		})
	}
}

// In dev a missing key is left zero: cmd/api replaces it with an ephemeral
// one and says so in the log.
func TestLoadReminders_missingKeysInDevAreZero(t *testing.T) {
	got, err := loadReminders(t, map[string]string{"PUSH_TOKEN_INDEX_PEPPER": hexKey(0x0d)}, config.EnvDev)

	require.NoError(t, err)
	assert.True(t, got.PushTokenKey.IsZero())
	assert.False(t, got.PushTokenIndexPepper.IsZero())
}

func TestLoadReminders_malformedKeys(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr string
	}{
		{"enc key not 32 bytes", func(e map[string]string) { e["PUSH_TOKEN_ENC_KEY"] = hexKey(0x0c) + "ab" },
			"secret PUSH_TOKEN_ENC_KEY must be exactly 32 bytes"},
		{"pepper too short", func(e map[string]string) { e["PUSH_TOKEN_INDEX_PEPPER"] = "abcd" },
			"secret PUSH_TOKEN_INDEX_PEPPER must be at least 32 bytes"},
		{"key id out of range", func(e map[string]string) { e["PUSH_TOKEN_ENC_KEY_ID"] = "0" },
			`PUSH_TOKEN_ENC_KEY_ID "0" is not an integer in [1, 32767]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := remindersEnv()
			tt.mutate(env)

			_, err := loadReminders(t, env, config.EnvDev)

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
