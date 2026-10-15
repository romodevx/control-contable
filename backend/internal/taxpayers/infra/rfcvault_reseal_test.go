package infra_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// rfcRotation is an RFC key rotation over random keys.
type rfcRotation struct {
	before *infra.RFCVault    // old key only
	during *infra.RFCResealer // new key active, old retired; no pepper needed
	after  *infra.RFCVault    // new key only
	// tooEarly re-seals with the new key only: the old one was dropped first.
	tooEarly *infra.RFCResealer
}

func newRFCRotation(t *testing.T) rfcRotation {
	t.Helper()
	random := func(n int) []byte {
		b := make([]byte, n)
		_, err := rand.Read(b)
		require.NoError(t, err)
		return b
	}
	old := crypto.KeyVersion{ID: 3, Key: random(crypto.KeySize)}
	next := crypto.KeyVersion{ID: 4, Key: random(crypto.KeySize)}
	index, err := crypto.NewRFCIndex(random(crypto.MinPepperSize))
	require.NoError(t, err)
	keyring := func(active crypto.KeyVersion, retired ...crypto.KeyVersion) *crypto.Keyring {
		kr, err := crypto.NewRotatingKeyring(active, retired...)
		require.NoError(t, err)
		return kr
	}
	return rfcRotation{
		before:   infra.NewRFCVault(keyring(old), index),
		during:   infra.NewRFCResealer(keyring(next, old)),
		after:    infra.NewRFCVault(keyring(next), index),
		tooEarly: infra.NewRFCResealer(keyring(next)),
	}
}

// ADR-0012: the re-seal job moves an RFC to the active key under the same
// row binding; the new key alone then opens it.
func TestRFCResealer_MovesTheRFCToTheActiveKey(t *testing.T) {
	r := newRFCRotation(t)
	own := usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA}
	sealed, err := r.before.Seal(mustRFC(t, "PEXJ900215HK9"), own)
	require.NoError(t, err)

	resealed, err := r.during.Reseal(sealed.Ciphertext, own)

	require.NoError(t, err)
	keyID, err := crypto.KeyIDOf(resealed)
	require.NoError(t, err)
	assert.Equal(t, crypto.KeyID(4), keyID)
	assert.Equal(t, crypto.KeyID(4), r.during.ActiveKeyID())
	opened, err := r.after.Open(resealed, own)
	require.NoError(t, err)
	assert.Equal(t, "PEXJ900215HK9", opened.String())
}

func TestRFCResealer_FailsClosed(t *testing.T) {
	r := newRFCRotation(t)
	own := usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA}
	sealed, err := r.before.Seal(mustRFC(t, "PEXJ900215HK9"), own)
	require.NoError(t, err)

	_, err = r.during.Reseal(sealed.Ciphertext, usecase.RowBinding{TaxpayerID: vaultTaxpayer2, UserID: vaultUserA})
	require.ErrorIs(t, err, crypto.ErrDecrypt, "another row's binding")
	_, err = r.tooEarly.Reseal(sealed.Ciphertext, own)
	require.ErrorIs(t, err, crypto.ErrUnknownKeyID, "the old key was dropped too early")
	_, err = r.during.Reseal(sealed.Ciphertext, usecase.RowBinding{TaxpayerID: "not-a-uuid", UserID: vaultUserA})
	require.Error(t, err)
}

func TestRFCResealer_NeverPrintsKeyMaterial(t *testing.T) {
	r := newRFCRotation(t)

	assert.Equal(t, "REDACTED", fmt.Sprintf("%v", r.during))
	assert.Equal(t, "REDACTED", fmt.Sprintf("%+v", *r.during))
}
