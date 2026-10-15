package pushtoken_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const (
	userA   = "0199d4a2-0000-7000-8000-00000000f00a"
	userB   = "0199d4a2-0000-7000-8000-00000000f00b"
	deviceA = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23"
	keyID   = crypto.KeyID(7)
)

// newVault builds a vault over random keys generated per test.
func newVault(t *testing.T) (*pushtoken.Vault, *crypto.BlindIndex) {
	t.Helper()
	key, pepper := make([]byte, crypto.KeySize), make([]byte, crypto.MinPepperSize)
	_, _ = rand.Read(key)
	_, _ = rand.Read(pepper)
	keyring, err := crypto.NewKeyring(keyID, map[crypto.KeyID][]byte{keyID: key})
	require.NoError(t, err)
	index, err := crypto.NewTokenIndex(pepper)
	require.NoError(t, err)
	return pushtoken.NewVault(keyring, index), index
}

func token(t *testing.T) domain.PushToken {
	t.Helper()
	tok, err := domain.NewPushToken("fcm:" + rand.Text())
	require.NoError(t, err)
	return tok
}

// MODULE-reminders §2.3: sealed with the RFC's envelope (key id framed),
// HMAC-indexed by the token alone (the claim matches it across users), and
// opened only for the row it was bound to.
func TestVault_sealsBoundToTheRowAndIndexesTheToken(t *testing.T) {
	vault, index := newVault(t)
	tok := token(t)
	binding := usecase.TokenBinding{InstallationID: deviceA, UserID: userA}

	sealed, err := vault.Seal(tok, binding)

	require.NoError(t, err)
	assert.Equal(t, int16(keyID), sealed.KeyID)
	framed, err := crypto.KeyIDOf(sealed.Ciphertext)
	require.NoError(t, err)
	assert.Equal(t, keyID, framed)
	assert.False(t, bytes.Contains(sealed.Ciphertext, []byte(tok.Reveal())), "never stored in clear")
	want, err := index.Hash(tok.Reveal())
	require.NoError(t, err)
	assert.Equal(t, want, sealed.HMAC)

	opened, err := vault.Open(sealed.Ciphertext, binding)
	require.NoError(t, err)
	assert.Equal(t, tok.Reveal(), opened.Reveal())
}

func TestVault_theHMACIsTheSameForEveryUser(t *testing.T) {
	vault, _ := newVault(t)
	tok := token(t)

	forA, err := vault.Seal(tok, usecase.TokenBinding{InstallationID: deviceA, UserID: userA})
	require.NoError(t, err)
	forB, err := vault.Seal(tok, usecase.TokenBinding{InstallationID: deviceA, UserID: userB})
	require.NoError(t, err)

	assert.Equal(t, forA.HMAC, forB.HMAC, "the claim finds A's row from B's registration")
	assert.NotEqual(t, forA.Ciphertext, forB.Ciphertext)
}

func TestVault_aCiphertextMovedToAnotherRowDoesNotOpen(t *testing.T) {
	vault, _ := newVault(t)
	sealed, err := vault.Seal(token(t), usecase.TokenBinding{InstallationID: deviceA, UserID: userA})
	require.NoError(t, err)

	for name, other := range map[string]usecase.TokenBinding{
		"another user":         {InstallationID: deviceA, UserID: userB},
		"another installation": {InstallationID: userB, UserID: userA},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := vault.Open(sealed.Ciphertext, other)

			require.ErrorIs(t, err, crypto.ErrDecrypt)
		})
	}
}

func TestVault_refusesABadBindingOrTheZeroToken(t *testing.T) {
	vault, _ := newVault(t)

	_, err := vault.Seal(token(t), usecase.TokenBinding{InstallationID: "nope", UserID: userA})
	require.Error(t, err)
	_, err = vault.Seal(token(t), usecase.TokenBinding{InstallationID: deviceA, UserID: "nope"})
	require.Error(t, err)
	_, err = vault.Seal(domain.PushToken{}, usecase.TokenBinding{InstallationID: deviceA, UserID: userA})
	require.Error(t, err)
	_, err = vault.Open([]byte("x"), usecase.TokenBinding{InstallationID: "nope", UserID: userA})
	require.Error(t, err)
}

func TestVault_neverPrintsItsKeys(t *testing.T) {
	vault, _ := newVault(t)

	assert.Equal(t, "REDACTED", fmt.Sprintf("%v", vault))
	assert.Equal(t, "REDACTED", vault.LogValue().String())
}
