package infra_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const (
	vaultUserA     = "0199d4a2-0000-7000-8000-00000000000a"
	vaultUserB     = "0199d4a2-0000-7000-8000-00000000000b"
	vaultTaxpayer1 = "0199d4a2-0000-7000-8000-0000000000a1"
	vaultTaxpayer2 = "0199d4a2-0000-7000-8000-0000000000a2"
)

func newVault(t *testing.T) *infra.RFCVault {
	t.Helper()
	keyring, err := crypto.NewKeyring(3, map[crypto.KeyID][]byte{3: bytes.Repeat([]byte{0x5a}, crypto.KeySize)})
	require.NoError(t, err)
	index, err := crypto.NewRFCIndex(bytes.Repeat([]byte{0x6b}, crypto.MinPepperSize))
	require.NoError(t, err)
	return infra.NewRFCVault(keyring, index)
}

func mustRFC(t *testing.T, raw string) domain.RFC {
	t.Helper()
	rfc, err := domain.NewRFC(raw)
	require.NoError(t, err)
	return rfc
}

func TestRFCVault_SealsBoundToTheRowAndOwner(t *testing.T) {
	vault := newVault(t)
	own := usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA}

	sealed, err := vault.Seal(mustRFC(t, "PEXJ900215HK9"), own)

	require.NoError(t, err)
	assert.Equal(t, int16(3), sealed.KeyID)
	assert.Equal(t, []byte{0, 3}, sealed.Ciphertext[:2], "keyring frame: key_id prefix")
	assert.Len(t, sealed.Hash, crypto.HashSize)
	assert.Equal(t, "PEXJ******HK9", sealed.Masked)
	assert.False(t, sealed.CheckDigitValid)
	assert.NotContains(t, string(sealed.Ciphertext), "PEXJ")

	opened, err := vault.Open(sealed.Ciphertext, own)
	require.NoError(t, err)
	assert.Equal(t, "PEXJ900215HK9", opened.String())

	for name, other := range map[string]usecase.RowBinding{
		"another row of the owner":     {TaxpayerID: vaultTaxpayer2, UserID: vaultUserA},
		"the same row id, other owner": {TaxpayerID: vaultTaxpayer1, UserID: vaultUserB},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := vault.Open(sealed.Ciphertext, other)
			require.ErrorIs(t, err, crypto.ErrDecrypt)
		})
	}
}

func TestRFCVault_BlindIndexIsPerOwnerNotPerRow(t *testing.T) {
	vault := newVault(t)
	rfc := mustRFC(t, "PEXJ900215HK8")

	a1, err := vault.Seal(rfc, usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA})
	require.NoError(t, err)
	a2, err := vault.Seal(rfc, usecase.RowBinding{TaxpayerID: vaultTaxpayer2, UserID: vaultUserA})
	require.NoError(t, err)
	b1, err := vault.Seal(rfc, usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserB})
	require.NoError(t, err)

	assert.Equal(t, a1.Hash, a2.Hash, "one user, one RFC, one hash: the unique key works")
	assert.NotEqual(t, a1.Hash, b1.Hash, "S-2: two users are unlinkable")
	assert.NotEqual(t, a1.Ciphertext, a2.Ciphertext, "fresh nonce per seal")
}

func TestRFCVault_RefusesBadInput(t *testing.T) {
	vault := newVault(t)

	_, err := vault.Seal(domain.RFC{}, usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA})
	require.Error(t, err)
	_, err = vault.Seal(mustRFC(t, "PEXJ900215HK8"), usecase.RowBinding{TaxpayerID: "nope", UserID: vaultUserA})
	require.Error(t, err)
	_, err = vault.Open([]byte{0, 3}, usecase.RowBinding{TaxpayerID: vaultTaxpayer1, UserID: vaultUserA})
	require.ErrorIs(t, err, crypto.ErrMalformedCiphertext)
}

func TestRFCVault_NeverPrintsItsKeys(t *testing.T) {
	vault := newVault(t)
	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		assert.Equal(t, "REDACTED", fmt.Sprintf(verb, vault))
	}
}
