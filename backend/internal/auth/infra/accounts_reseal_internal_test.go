package infra

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
)

func randomKeyVersion(t *testing.T, id crypto.KeyID) crypto.KeyVersion {
	t.Helper()
	key := make([]byte, crypto.KeySize)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return crypto.KeyVersion{ID: id, Key: key}
}

func rotatingKeyring(t *testing.T, active crypto.KeyVersion, retired ...crypto.KeyVersion) *crypto.Keyring {
	t.Helper()
	kr, err := crypto.NewRotatingKeyring(active, retired...)
	require.NoError(t, err)
	return kr
}

func mustUserID(t *testing.T, raw string) domain.UserID {
	t.Helper()
	id, err := domain.ParseUserID(raw)
	require.NoError(t, err)
	return id
}

// ADR-0012: the re-seal job moves an email to the active key under the same
// user binding (emailAAD); the new key alone then opens it.
func TestAccountStore_ResealEmailMovesItToTheActiveKey(t *testing.T) {
	old, next := randomKeyVersion(t, 1), randomKeyVersion(t, 2)
	user := mustUserID(t, "0199d4a2-0000-7000-8000-0000000e0001")
	other := mustUserID(t, "0199d4a2-0000-7000-8000-0000000e0002")
	sealed, err := rotatingKeyring(t, old).Encrypt([]byte("ana@example.com"), emailAAD(user))
	require.NoError(t, err)
	during := NewAccountStore(rotatingKeyring(t, next, old))

	resealed, err := during.ResealEmail(sealed, user)

	require.NoError(t, err)
	id, err := crypto.KeyIDOf(resealed)
	require.NoError(t, err)
	assert.Equal(t, next.ID, id)
	plain, err := rotatingKeyring(t, next).Decrypt(resealed, emailAAD(user))
	require.NoError(t, err)
	assert.Equal(t, "ana@example.com", string(plain))

	_, err = during.ResealEmail(sealed, other)
	require.ErrorIs(t, err, crypto.ErrDecrypt, "another user's binding")
	_, err = NewAccountStore(rotatingKeyring(t, next)).ResealEmail(sealed, user)
	require.ErrorIs(t, err, crypto.ErrUnknownKeyID)
}
