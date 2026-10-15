package pushtoken_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// rotationVaults are the vaults of a push token key rotation over random
// keys: before (old key only), during (new active, old retired), after (new only).
func rotationVaults(t *testing.T) (before, during, after *pushtoken.Vault) {
	t.Helper()
	random := func(n int) []byte {
		b := make([]byte, n)
		_, err := rand.Read(b)
		require.NoError(t, err)
		return b
	}
	old := crypto.KeyVersion{ID: keyID, Key: random(crypto.KeySize)}
	next := crypto.KeyVersion{ID: keyID + 1, Key: random(crypto.KeySize)}
	index, err := crypto.NewTokenIndex(random(crypto.MinPepperSize))
	require.NoError(t, err)
	vault := func(active crypto.KeyVersion, retired ...crypto.KeyVersion) *pushtoken.Vault {
		kr, err := crypto.NewRotatingKeyring(active, retired...)
		require.NoError(t, err)
		return pushtoken.NewVault(kr, index)
	}
	return vault(old), vault(next, old), vault(next)
}

// ADR-0012: the re-seal job moves a token to the active key under the same
// row binding; the new key alone then opens it.
func TestVault_resealMovesTheTokenToTheActiveKey(t *testing.T) {
	before, during, after := rotationVaults(t)
	tok := token(t)
	own := usecase.TokenBinding{InstallationID: deviceA, UserID: userA}
	sealed, err := before.Seal(tok, own)
	require.NoError(t, err)

	resealed, err := during.Reseal(sealed.Ciphertext, own)

	require.NoError(t, err)
	id, err := crypto.KeyIDOf(resealed)
	require.NoError(t, err)
	assert.Equal(t, keyID+1, id)
	opened, err := after.Open(resealed, own)
	require.NoError(t, err)
	assert.Equal(t, tok.Reveal(), opened.Reveal())

	_, err = during.Reseal(sealed.Ciphertext, usecase.TokenBinding{InstallationID: deviceA, UserID: userB})
	require.ErrorIs(t, err, crypto.ErrDecrypt, "another user's binding")
	_, err = after.Reseal(sealed.Ciphertext, own)
	require.ErrorIs(t, err, crypto.ErrUnknownKeyID)
	_, err = during.Reseal(sealed.Ciphertext, usecase.TokenBinding{InstallationID: "x", UserID: userA})
	require.Error(t, err)
}
