package config_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

// hexKey is a 32-byte fixture key in hex whose bytes are all b.
func hexKey(b byte) string { return strings.Repeat(fmt.Sprintf("%02x", b), 32) }

func authEnv() map[string]string {
	return map[string]string{
		"AUTH_EMAIL_PEPPER":     hexKey(0x01),
		"AUTH_EMAIL_KEY":        hexKey(0x02),
		"AUTH_CODE_KEY":         hexKey(0x03),
		"AUTH_ACCESS_TOKEN_KEY": hexKey(0x04),
	}
}

func loadAuth(env map[string]string) (config.Auth, error) {
	lookup := lookupFrom(env)
	return config.LoadAuth(context.Background(), lookup, config.NewEnvSecrets(lookup))
}

func TestLoadAuth_decodesKeysAndAppliesDefaults(t *testing.T) {
	got, err := loadAuth(authEnv())

	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0x01}, 32), got.EmailPepper.RevealBytes())
	assert.Equal(t, bytes.Repeat([]byte{0x02}, 32), got.EmailKey.RevealBytes())
	assert.Equal(t, bytes.Repeat([]byte{0x03}, 32), got.CodeKey.RevealBytes())
	assert.Equal(t, bytes.Repeat([]byte{0x04}, 32), got.AccessTokenKey.RevealBytes())
	assert.Equal(t, 1, got.EmailKeyID)
	assert.Equal(t, "k1", got.AccessTokenKeyID)
}

func TestLoadAuth_overrides(t *testing.T) {
	env := authEnv()
	env["AUTH_EMAIL_PEPPER"] = strings.Repeat("ab", 48) // longer peppers are fine
	env["AUTH_EMAIL_KEY_ID"] = "7"
	env["AUTH_ACCESS_TOKEN_KEY_ID"] = "2026-10.a"

	got, err := loadAuth(env)

	require.NoError(t, err)
	assert.Len(t, got.EmailPepper.RevealBytes(), 48)
	assert.Equal(t, 7, got.EmailKeyID)
	assert.Equal(t, "2026-10.a", got.AccessTokenKeyID)
}

func TestLoadAuth_errors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr string
	}{
		{"missing pepper", func(e map[string]string) { delete(e, "AUTH_EMAIL_PEPPER") }, "secret AUTH_EMAIL_PEPPER"},
		{"pepper not hex", func(e map[string]string) { e["AUTH_EMAIL_PEPPER"] = strings.Repeat("zz", 32) }, "secret AUTH_EMAIL_PEPPER is not hex"},
		{"pepper too short", func(e map[string]string) { e["AUTH_EMAIL_PEPPER"] = strings.Repeat("ab", 31) }, "secret AUTH_EMAIL_PEPPER must be at least 32 bytes"},
		{"email key not 32 bytes", func(e map[string]string) { e["AUTH_EMAIL_KEY"] = strings.Repeat("ab", 33) }, "secret AUTH_EMAIL_KEY must be exactly 32 bytes"},
		{"code key too short", func(e map[string]string) { e["AUTH_CODE_KEY"] = "abcd" }, "secret AUTH_CODE_KEY must be at least 32 bytes"},
		{"missing access token key", func(e map[string]string) { delete(e, "AUTH_ACCESS_TOKEN_KEY") }, "secret AUTH_ACCESS_TOKEN_KEY"},
		{"access token key not a seed", func(e map[string]string) { e["AUTH_ACCESS_TOKEN_KEY"] = strings.Repeat("ab", 64) }, "secret AUTH_ACCESS_TOKEN_KEY must be exactly 32 bytes"},
		{"key id out of range", func(e map[string]string) { e["AUTH_EMAIL_KEY_ID"] = "0" }, `AUTH_EMAIL_KEY_ID "0" is not an integer in [1, 32767]`},
		{"bad token key id", func(e map[string]string) { e["AUTH_ACCESS_TOKEN_KEY_ID"] = "has space" }, `AUTH_ACCESS_TOKEN_KEY_ID "has space"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := authEnv()
			tt.mutate(env)

			_, err := loadAuth(env)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			for _, v := range authEnv() {
				assert.NotContains(t, err.Error(), v, "a key never reaches the error")
			}
		})
	}
}

func TestLoadAuth_noPreviousAccessTokenKeyByDefault(t *testing.T) {
	got, err := loadAuth(authEnv())

	require.NoError(t, err)
	assert.True(t, got.PreviousAccessTokenKey.IsZero())
	assert.Empty(t, got.PreviousAccessTokenKeyID)
}

// Key rotation (M1-07): the retired key keeps verifying the tokens it signed
// until they expire, under its own kid.
func TestLoadAuth_previousAccessTokenKey(t *testing.T) {
	env := authEnv()
	env["AUTH_ACCESS_TOKEN_KEY_ID"] = "k2"
	env["AUTH_ACCESS_TOKEN_PREVIOUS_KEY"] = hexKey(0x05)
	env["AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID"] = "k1"

	got, err := loadAuth(env)

	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0x05}, 32), got.PreviousAccessTokenKey.RevealBytes())
	assert.Equal(t, "k1", got.PreviousAccessTokenKeyID)
	assert.Equal(t, "k2", got.AccessTokenKeyID)
}

func TestLoadAuth_previousAccessTokenKeyErrors(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"previous key without its kid", map[string]string{"AUTH_ACCESS_TOKEN_PREVIOUS_KEY": hexKey(0x05)},
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID is required with AUTH_ACCESS_TOKEN_PREVIOUS_KEY"},
		{"kid without the previous key", map[string]string{"AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID": "k0"},
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID is set without AUTH_ACCESS_TOKEN_PREVIOUS_KEY"},
		{"previous kid equals the active kid", map[string]string{
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY": hexKey(0x05), "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID": "k1",
		}, "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID must differ from AUTH_ACCESS_TOKEN_KEY_ID"},
		{"previous kid malformed", map[string]string{
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY": hexKey(0x05), "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID": "bad kid",
		}, `AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID "bad kid"`},
		{"previous key not a seed", map[string]string{
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY": strings.Repeat("ab", 31), "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID": "k0",
		}, "secret AUTH_ACCESS_TOKEN_PREVIOUS_KEY must be exactly 32 bytes"},
		{"previous key not hex", map[string]string{
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY": strings.Repeat("zz", 32), "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID": "k0",
		}, "secret AUTH_ACCESS_TOKEN_PREVIOUS_KEY is not hex"},
		{"previous key equals the active key", map[string]string{
			"AUTH_ACCESS_TOKEN_PREVIOUS_KEY": hexKey(0x04), "AUTH_ACCESS_TOKEN_PREVIOUS_KEY_ID": "k0",
		}, "AUTH_ACCESS_TOKEN_PREVIOUS_KEY must differ from AUTH_ACCESS_TOKEN_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := authEnv()
			for k, v := range tt.env {
				env[k] = v
			}

			_, err := loadAuth(env)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.NotContains(t, err.Error(), hexKey(0x05), "a key never reaches the error")
			assert.NotContains(t, err.Error(), hexKey(0x04), "a key never reaches the error")
		})
	}
}

func TestLoadAuth_reportsEveryProblemAtOnce(t *testing.T) {
	_, err := loadAuth(map[string]string{})

	require.Error(t, err)
	for _, name := range []string{"AUTH_EMAIL_PEPPER", "AUTH_EMAIL_KEY", "AUTH_CODE_KEY", "AUTH_ACCESS_TOKEN_KEY"} {
		assert.ErrorContains(t, err, name)
	}
}

func TestSecret_RevealBytes(t *testing.T) {
	assert.Equal(t, []byte("x"), config.NewSecret("x").RevealBytes())
}
