package main

import (
	"crypto/rand"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authdomain "github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	remdomain "github.com/romodevx/control-contable/backend/internal/reminders/domain"
	remusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// randomSecret is a fresh 32-byte key: no key material is ever committed.
func randomSecret(t *testing.T) config.Secret {
	t.Helper()
	b := make([]byte, crypto.KeySize)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return config.NewSecret(string(b))
}

func requireKeyID(t *testing.T, ciphertext []byte, want crypto.KeyID) {
	t.Helper()
	got, err := crypto.KeyIDOf(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// ADR-0012: with X_PREVIOUS_KEY configured, the API still opens what the
// previous key sealed and seals with the new one only.
func TestKeyRotation_RFCVaultOpensWhatThePreviousKeySealed(t *testing.T) {
	oldKey, pepper := randomSecret(t), randomSecret(t)
	binding := tpusecase.RowBinding{TaxpayerID: "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80", UserID: "0199d4a2-0000-7000-8000-00000000000a"}
	rfc, err := tpdomain.NewRFCFor(tpdomain.PersonTypeFisica, "PEXJ900215HK8")
	require.NoError(t, err)
	wiring := func(keys config.Taxpayers) taxpayersWiring {
		keys.RFCIndexPepper = pepper
		return taxpayersWiring{env: config.EnvProduction, keys: keys, logger: slog.New(slog.DiscardHandler)}
	}
	before, err := newRFCVault(wiring(config.Taxpayers{RFCKey: oldKey, RFCKeyID: 1}))
	require.NoError(t, err)
	sealed, err := before.Seal(rfc, binding)
	require.NoError(t, err)

	during, err := newRFCVault(wiring(config.Taxpayers{
		RFCKey: randomSecret(t), RFCKeyID: 2, RFCPrevious: config.PreviousKey{Key: oldKey, KeyID: 1},
	}))

	require.NoError(t, err)
	opened, err := during.Open(sealed.Ciphertext, binding)
	require.NoError(t, err)
	assert.Equal(t, rfc.String(), opened.String())
	fresh, err := during.Seal(rfc, binding)
	require.NoError(t, err)
	assert.Equal(t, int16(2), fresh.KeyID)
	requireKeyID(t, fresh.Ciphertext, 2)
}

func TestKeyRotation_PushTokenVaultOpensWhatThePreviousKeySealed(t *testing.T) {
	oldKey, pepper := randomSecret(t), randomSecret(t)
	binding := remusecase.TokenBinding{InstallationID: "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23", UserID: "0199d4a2-0000-7000-8000-00000000000a"}
	tok, err := remdomain.NewPushToken("fcm:" + t.Name())
	require.NoError(t, err)
	wiring := func(keys config.Reminders) remindersWiring {
		w := remindersTestWiring(t)
		keys.PushTokenIndexPepper = pepper
		w.env, w.keys = config.EnvProduction, keys
		return w
	}
	before, err := newPushTokenVault(wiring(config.Reminders{PushTokenKey: oldKey, PushTokenKeyID: 1}))
	require.NoError(t, err)
	sealed, err := before.Seal(tok, binding)
	require.NoError(t, err)

	during, err := newPushTokenVault(wiring(config.Reminders{
		PushTokenKey: randomSecret(t), PushTokenKeyID: 2, PushTokenPrevious: config.PreviousKey{Key: oldKey, KeyID: 1},
	}))

	require.NoError(t, err)
	opened, err := during.Open(sealed.Ciphertext, binding)
	require.NoError(t, err)
	assert.Equal(t, tok.Reveal(), opened.Reveal())
	fresh, err := during.Seal(tok, binding)
	require.NoError(t, err)
	requireKeyID(t, fresh.Ciphertext, 2)
}

func TestKeyRotation_EmailKeyringOpensWhatThePreviousKeySealed(t *testing.T) {
	keys := config.Auth{
		EmailPepper: randomSecret(t), EmailKey: randomSecret(t), EmailKeyID: 1, CodeKey: randomSecret(t),
		AccessTokenKey: randomSecret(t), AccessTokenKeyID: "k1",
	}
	before, err := newAuthCrypto(authWiring{keys: keys, argon2: fastArgon2})
	require.NoError(t, err)
	user, err := authdomain.ParseUserID(sessionUser)
	require.NoError(t, err)
	aad := []byte("users.email_encrypted:" + user.String())
	sealed, err := before.emails.Encrypt([]byte("ana@example.com"), aad)
	require.NoError(t, err)

	keys.EmailPrevious = config.PreviousKey{Key: keys.EmailKey, KeyID: 1}
	keys.EmailKey, keys.EmailKeyID = randomSecret(t), 2
	during, err := newAuthCrypto(authWiring{keys: keys, argon2: fastArgon2})

	require.NoError(t, err)
	plain, err := during.emails.Decrypt(sealed, aad)
	require.NoError(t, err)
	assert.Equal(t, "ana@example.com", string(plain))
	assert.Equal(t, crypto.KeyID(2), during.emails.ActiveKeyID())
}

func TestKeyRotation_OutboxOpensWhatThePreviousKeySealed(t *testing.T) {
	oldKey := randomSecret(t)
	before, err := outboxKeyring(config.AuthOutbox{Key: oldKey, KeyID: 1})
	require.NoError(t, err)
	sealed, err := before.Encrypt([]byte("payload"), []byte("row"))
	require.NoError(t, err)

	during, err := outboxKeyring(config.AuthOutbox{Key: randomSecret(t), KeyID: 2, Previous: config.PreviousKey{Key: oldKey, KeyID: 1}})

	require.NoError(t, err)
	plain, err := during.Decrypt(sealed, []byte("row"))
	require.NoError(t, err)
	assert.Equal(t, []byte("payload"), plain)
	assert.Equal(t, crypto.KeyID(2), during.ActiveKeyID())
}
