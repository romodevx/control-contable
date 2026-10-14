package config_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

func taxpayersEnv() map[string]string {
	return map[string]string{
		"RFC_ENC_KEY":      hexKey(0x0a),
		"RFC_INDEX_PEPPER": hexKey(0x0b),
	}
}

func loadTaxpayers(t *testing.T, env map[string]string, appEnv config.Environment) (config.Taxpayers, error) {
	t.Helper()
	auth, err := loadAuth(authEnv())
	require.NoError(t, err)
	lookup := lookupFrom(env)
	return config.LoadTaxpayers(context.Background(), lookup, config.NewEnvSecrets(lookup), appEnv, auth)
}

func TestLoadTaxpayers_decodesKeysAndAppliesDefaults(t *testing.T) {
	got, err := loadTaxpayers(t, taxpayersEnv(), config.EnvProduction)

	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0x0a}, 32), got.RFCKey.RevealBytes())
	assert.Equal(t, bytes.Repeat([]byte{0x0b}, 32), got.RFCIndexPepper.RevealBytes())
	assert.Equal(t, 1, got.RFCKeyID)
}

func TestLoadTaxpayers_keyIDOverride(t *testing.T) {
	env := taxpayersEnv()
	env["RFC_ENC_KEY_ID"] = "3"

	got, err := loadTaxpayers(t, env, config.EnvStaging)

	require.NoError(t, err)
	assert.Equal(t, 3, got.RFCKeyID)
}

// PR #59 security gate, condition 1: the RFC encryption key, the RFC index
// pepper and the email index pepper are distinct secrets. Equal values are a
// startup error in every environment, dev included; so is reusing any other
// auth key (each key has one purpose).
func TestLoadTaxpayers_refusesSharedSecrets(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"enc key equals index pepper", map[string]string{"RFC_ENC_KEY": hexKey(0x0a), "RFC_INDEX_PEPPER": hexKey(0x0a)},
			"RFC_INDEX_PEPPER must differ from RFC_ENC_KEY"},
		{"enc key equals email pepper", map[string]string{"RFC_ENC_KEY": hexKey(0x01), "RFC_INDEX_PEPPER": hexKey(0x0b)},
			"RFC_ENC_KEY must differ from AUTH_EMAIL_PEPPER"},
		{"index pepper equals email pepper", map[string]string{"RFC_ENC_KEY": hexKey(0x0a), "RFC_INDEX_PEPPER": hexKey(0x01)},
			"RFC_INDEX_PEPPER must differ from AUTH_EMAIL_PEPPER"},
		{"enc key equals email key", map[string]string{"RFC_ENC_KEY": hexKey(0x02), "RFC_INDEX_PEPPER": hexKey(0x0b)},
			"RFC_ENC_KEY must differ from AUTH_EMAIL_KEY"},
		{"index pepper equals code key", map[string]string{"RFC_ENC_KEY": hexKey(0x0a), "RFC_INDEX_PEPPER": hexKey(0x03)},
			"RFC_INDEX_PEPPER must differ from AUTH_CODE_KEY"},
		{"upper-case hex of the same bytes", map[string]string{"RFC_ENC_KEY": hexKey(0x0a), "RFC_INDEX_PEPPER": "0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A0A"},
			"RFC_INDEX_PEPPER must differ from RFC_ENC_KEY"},
	}
	for _, tt := range tests {
		for _, appEnv := range []config.Environment{config.EnvDev, config.EnvProduction} {
			t.Run(tt.name+"/"+string(appEnv), func(t *testing.T) {
				_, err := loadTaxpayers(t, tt.env, appEnv)

				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				for _, v := range tt.env {
					assert.NotContains(t, err.Error(), v, "a key never reaches the error")
				}
			})
		}
	}
}

func TestLoadTaxpayers_missingKeysOutsideDevRefuseToStart(t *testing.T) {
	for _, appEnv := range []config.Environment{config.EnvStaging, config.EnvProduction} {
		t.Run(string(appEnv), func(t *testing.T) {
			_, err := loadTaxpayers(t, map[string]string{}, appEnv)

			require.Error(t, err)
			assert.ErrorContains(t, err, "secret RFC_ENC_KEY")
			assert.ErrorContains(t, err, "secret RFC_INDEX_PEPPER")
		})
	}
}

// In dev a missing key is left zero: cmd/api replaces it with an ephemeral
// one and says so in the log.
func TestLoadTaxpayers_missingKeysInDevAreZero(t *testing.T) {
	got, err := loadTaxpayers(t, map[string]string{"RFC_INDEX_PEPPER": hexKey(0x0b)}, config.EnvDev)

	require.NoError(t, err)
	assert.True(t, got.RFCKey.IsZero())
	assert.False(t, got.RFCIndexPepper.IsZero())
	assert.Equal(t, 1, got.RFCKeyID)
}

func TestLoadTaxpayers_malformedKeys(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr string
	}{
		{"enc key not 32 bytes", func(e map[string]string) { e["RFC_ENC_KEY"] = hexKey(0x0a) + "ab" }, "secret RFC_ENC_KEY must be exactly 32 bytes"},
		{"pepper too short", func(e map[string]string) { e["RFC_INDEX_PEPPER"] = "abcd" }, "secret RFC_INDEX_PEPPER must be at least 32 bytes"},
		{"enc key not hex", func(e map[string]string) { e["RFC_ENC_KEY"] = "zz" }, "secret RFC_ENC_KEY is not hex"},
		{"key id out of range", func(e map[string]string) { e["RFC_ENC_KEY_ID"] = "0" }, `RFC_ENC_KEY_ID "0" is not an integer in [1, 32767]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := taxpayersEnv()
			tt.mutate(env)

			_, err := loadTaxpayers(t, env, config.EnvDev)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
