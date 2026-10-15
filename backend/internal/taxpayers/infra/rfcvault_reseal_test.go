package infra_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// rotationVaults are the vaults of an RFC key rotation over random keys:
// before (old key only), during (new active, old retired), after (new only).
func rotationVaults(t *testing.T) (before, during, after *infra.RFCVault) {
	t.Helper()
	key := func() crypto.KeyVersion {
		k := make([]byte, crypto.KeySize)
		_, err := rand.Read(k)
		require.NoError(t, err)
		return crypto.KeyVersion{Key: k}
	}
	old, next := key(), key()
	old.ID, next.ID = 3, 4
	pepper := make([]byte, crypto.MinPepperSize)
	_, err := rand.Read(pepper)
	require.NoError(t, err)
	index, err := crypto.NewRFCIndex(pepper)
	require.NoError(t, err)
	vault := func(active crypto.KeyVersion, retired ...crypto.KeyVersion) *infra.RFCVault {
		kr, err := crypto.NewRotatingKeyring(active, retired...)
		require.NoError(t, err)
		return infra.NewRFCVault(kr, index)
	}
	return vault(old), vault(next, old), vault(next)
}

// ADR-0012: the re-seal job moves an RFC to the active key under the same
// row binding; the new key alone then opens it.
func TestRFCVault_ResealMovesTheRFCToTheActiveKey(t *testing.T) {
	before, during, after := rotationVaults(t)
	own := usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA}
	sealed, err := before.Seal(mustRFC(t, "PEXJ900215HK9"), own)
	require.NoError(t, err)

	resealed, err := during.Reseal(sealed.Ciphertext, own)

	require.NoError(t, err)
	keyID, err := crypto.KeyIDOf(resealed)
	require.NoError(t, err)
	assert.Equal(t, crypto.KeyID(4), keyID)
	opened, err := after.Open(resealed, own)
	require.NoError(t, err)
	assert.Equal(t, "PEXJ900215HK9", opened.String())
}

func TestRFCVault_ResealFailsClosed(t *testing.T) {
	before, during, after := rotationVaults(t)
	own := usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA}
	sealed, err := before.Seal(mustRFC(t, "PEXJ900215HK9"), own)
	require.NoError(t, err)

	_, err = during.Reseal(sealed.Ciphertext, usecase.RowBinding{TaxpayerID: vaultTaxpayer2, UserID: vaultUserA})
	require.ErrorIs(t, err, crypto.ErrDecrypt, "another row's binding")
	_, err = after.Reseal(sealed.Ciphertext, own)
	require.ErrorIs(t, err, crypto.ErrUnknownKeyID, "the old key was dropped too early")
	_, err = during.Reseal(sealed.Ciphertext, usecase.RowBinding{TaxpayerID: "not-a-uuid", UserID: vaultUserA})
	require.Error(t, err)
}
