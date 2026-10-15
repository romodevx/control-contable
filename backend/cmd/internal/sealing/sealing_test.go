package sealing_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/cmd/internal/sealing"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
)

func randomKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

func TestKeyring_activeKeyOnly(t *testing.T) {
	kr, err := sealing.Keyring(3, randomKey(t), config.PreviousKey{})

	require.NoError(t, err)
	assert.Equal(t, crypto.KeyID(3), kr.ActiveKeyID())
}

// During a rotation the keyring seals with the new key and still opens what
// the previous key sealed (ADR-0012).
func TestKeyring_opensWhatThePreviousKeySealed(t *testing.T) {
	oldKey := randomKey(t)
	before, err := sealing.Keyring(1, oldKey, config.PreviousKey{})
	require.NoError(t, err)
	sealed, err := before.Encrypt([]byte("value"), []byte("row"))
	require.NoError(t, err)

	during, err := sealing.Keyring(2, randomKey(t), config.PreviousKey{Key: config.NewSecret(string(oldKey)), KeyID: 1})

	require.NoError(t, err)
	assert.Equal(t, crypto.KeyID(2), during.ActiveKeyID())
	plain, err := during.Decrypt(sealed, []byte("row"))
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), plain)
}

func TestKeyring_rejectsBadKeys(t *testing.T) {
	_, err := sealing.Keyring(2, randomKey(t), config.PreviousKey{Key: config.NewSecret(string(randomKey(t))), KeyID: 2})
	require.ErrorIs(t, err, crypto.ErrInvalidKey, "previous id equal to the active one")
	_, err = sealing.Keyring(2, randomKey(t)[:16], config.PreviousKey{})
	require.ErrorIs(t, err, crypto.ErrInvalidKey)
	_, err = sealing.Keyring(0, randomKey(t), config.PreviousKey{})
	require.ErrorIs(t, err, crypto.ErrInvalidKey)
}
