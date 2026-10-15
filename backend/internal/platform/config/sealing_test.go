package config_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
)

// previousKeyByte is the retired key of every rotation case: distinct from
// every key the base environments hold.
const previousKeyByte = 0x0e

// sealingKeyCase is one AES-256-GCM sealing key and how to load its retired
// key (ADR-0012).
type sealingKeyCase struct {
	name      string
	activeKey string // the active key's secret name
	activeHex string // the active key's value in base
	activeID  string // the active key id's variable
	base      func() map[string]string
	load      func(t *testing.T, env map[string]string) (config.PreviousKey, error)
}

func sealingKeyCases() []sealingKeyCase {
	return []sealingKeyCase{
		{
			name: "rfc", activeKey: "RFC_ENC_KEY", activeHex: hexKey(0x0a), activeID: "RFC_ENC_KEY_ID", base: taxpayersEnv,
			load: func(t *testing.T, env map[string]string) (config.PreviousKey, error) {
				cfg, err := loadTaxpayers(t, env, config.EnvProduction)
				return cfg.RFCPrevious, err
			},
		},
		{
			name: "email", activeKey: "AUTH_EMAIL_KEY", activeHex: hexKey(0x02), activeID: "AUTH_EMAIL_KEY_ID", base: authEnv,
			load: func(_ *testing.T, env map[string]string) (config.PreviousKey, error) {
				cfg, err := loadAuth(env)
				return cfg.EmailPrevious, err
			},
		},
		{
			name: "push token", activeKey: "PUSH_TOKEN_ENC_KEY", activeHex: hexKey(0x0c), activeID: "PUSH_TOKEN_ENC_KEY_ID",
			base: remindersEnv,
			load: func(t *testing.T, env map[string]string) (config.PreviousKey, error) {
				cfg, err := loadReminders(t, env, config.EnvProduction)
				return cfg.PushTokenPrevious, err
			},
		},
		{
			name: "outbox", activeKey: "AUTH_OUTBOX_KEY", activeHex: hexKey(0x09), activeID: "AUTH_OUTBOX_KEY_ID",
			base: func() map[string]string { return map[string]string{"AUTH_OUTBOX_KEY": hexKey(0x09)} },
			load: func(_ *testing.T, env map[string]string) (config.PreviousKey, error) {
				cfg, err := loadOutbox(env, config.Auth{})
				return cfg.Previous, err
			},
		},
	}
}

// previousNames are the retired key's variables: <ACTIVE>_PREVIOUS... with
// the active key's prefix (RFC_ENC_KEY -> RFC_ENC_PREVIOUS_KEY).
func previousNames(activeKey string) (key, id string) {
	prefix := activeKey[:len(activeKey)-len("KEY")]
	return prefix + "PREVIOUS_KEY", prefix + "PREVIOUS_KEY_ID"
}

func TestSealingKeys_noPreviousKeyByDefault(t *testing.T) {
	for _, c := range sealingKeyCases() {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.load(t, c.base())

			require.NoError(t, err)
			assert.True(t, got.IsZero())
		})
	}
}

func TestSealingKeys_loadThePreviousKey(t *testing.T) {
	for _, c := range sealingKeyCases() {
		t.Run(c.name, func(t *testing.T) {
			key, id := previousNames(c.activeKey)
			env := c.base()
			env[key], env[id] = hexKey(previousKeyByte), "7"

			got, err := c.load(t, env)

			require.NoError(t, err)
			assert.False(t, got.IsZero())
			assert.Equal(t, bytes.Repeat([]byte{previousKeyByte}, 32), got.Key.RevealBytes())
			assert.Equal(t, 7, got.KeyID)
		})
	}
}

func TestSealingKeys_previousKeyErrors(t *testing.T) {
	for _, c := range sealingKeyCases() {
		key, id := previousNames(c.activeKey)
		tests := []struct {
			name    string
			set     map[string]string
			wantErr string
		}{
			{"key without id", map[string]string{key: hexKey(previousKeyByte)}, id + " is required with " + key},
			{"id without key", map[string]string{id: "7"}, id + " is set without " + key},
			{"same id as the active key", map[string]string{key: hexKey(previousKeyByte), id: "1"},
				id + " must differ from " + c.activeID},
			{"same key as the active one", map[string]string{key: c.activeHex, id: "7"}, key + " must differ from " + c.activeKey},
			{"id out of range", map[string]string{key: hexKey(previousKeyByte), id: "32768"}, id},
			{"not 32 bytes", map[string]string{key: hexKey(previousKeyByte)[:62], id: "7"}, "secret " + key + " must be exactly 32 bytes"},
			{"not hex", map[string]string{key: "zz" + hexKey(previousKeyByte)[2:], id: "7"}, "secret " + key + " is not hex"},
		}
		for _, tt := range tests {
			t.Run(c.name+"/"+tt.name, func(t *testing.T) {
				env := c.base()
				for k, v := range tt.set {
					env[k] = v
				}

				_, err := c.load(t, env)

				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				assert.NotContains(t, err.Error(), hexKey(previousKeyByte), "a key never reaches the error")
				assert.NotContains(t, err.Error(), c.activeHex, "a key never reaches the error")
			})
		}
	}
}

func loadReseal(env map[string]string) (config.ResealKeys, error) {
	lookup := lookupFrom(env)
	return config.LoadResealKeys(context.Background(), lookup, config.NewEnvSecrets(lookup))
}

// The worker re-seals the RFC and email columns only when it is given their
// keys (ADR-0012): none is required, in any environment.
func TestLoadResealKeys_optional(t *testing.T) {
	got, err := loadReseal(map[string]string{})

	require.NoError(t, err)
	assert.True(t, got.RFC.IsZero())
	assert.True(t, got.Email.IsZero())
}

func TestLoadResealKeys_loadsBothColumnsWithTheirPreviousKeys(t *testing.T) {
	got, err := loadReseal(map[string]string{
		"RFC_ENC_KEY": hexKey(0x0a), "RFC_ENC_KEY_ID": "2",
		"RFC_ENC_PREVIOUS_KEY": hexKey(0x1a), "RFC_ENC_PREVIOUS_KEY_ID": "1",
		"AUTH_EMAIL_KEY": hexKey(0x02),
	})

	require.NoError(t, err)
	assert.Equal(t, bytes.Repeat([]byte{0x0a}, 32), got.RFC.Key.RevealBytes())
	assert.Equal(t, 2, got.RFC.KeyID)
	assert.Equal(t, bytes.Repeat([]byte{0x1a}, 32), got.RFC.Previous.Key.RevealBytes())
	assert.Equal(t, 1, got.RFC.Previous.KeyID)
	assert.Equal(t, bytes.Repeat([]byte{0x02}, 32), got.Email.Key.RevealBytes())
	assert.Equal(t, 1, got.Email.KeyID)
	assert.True(t, got.Email.Previous.IsZero())
}

func TestLoadResealKeys_errors(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"previous key without the active one", map[string]string{
			"RFC_ENC_PREVIOUS_KEY": hexKey(0x1a), "RFC_ENC_PREVIOUS_KEY_ID": "1",
		}, "RFC_ENC_PREVIOUS_KEY is set without RFC_ENC_KEY"},
		{"rfc key equals email key", map[string]string{
			"RFC_ENC_KEY": hexKey(0x0a), "AUTH_EMAIL_KEY": hexKey(0x0a),
		}, "AUTH_EMAIL_KEY must differ from RFC_ENC_KEY"},
		{"bad active key", map[string]string{"AUTH_EMAIL_KEY": hexKey(0x02)[:10]}, "secret AUTH_EMAIL_KEY must be exactly 32 bytes"},
		{"previous equals active", map[string]string{
			"AUTH_EMAIL_KEY": hexKey(0x02), "AUTH_EMAIL_PREVIOUS_KEY": hexKey(0x02), "AUTH_EMAIL_PREVIOUS_KEY_ID": "2",
		}, "AUTH_EMAIL_PREVIOUS_KEY must differ from AUTH_EMAIL_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadReseal(tt.env)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
			for _, v := range tt.env {
				if len(v) > 2 {
					assert.NotContains(t, err.Error(), v, "a key never reaches the error")
				}
			}
		})
	}
}
