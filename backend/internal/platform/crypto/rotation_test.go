package crypto

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRotatingKeyringValidation(t *testing.T) {
	tests := []struct {
		name    string
		active  KeyVersion
		retired []KeyVersion
	}{
		{"retired id equals the active id", KeyVersion{keyIDA, testKeyA}, []KeyVersion{{keyIDA, testKeyB}}},
		{"two retired keys share an id", KeyVersion{keyIDA, testKeyA}, []KeyVersion{{keyIDB, testKeyB}, {keyIDB, testKeyA}}},
		{"bad active key", KeyVersion{keyIDA, testKeyA[:16]}, nil},
		{"bad retired key", KeyVersion{keyIDA, testKeyA}, []KeyVersion{{keyIDB, testKeyB[:31]}}},
		{"retired id out of range", KeyVersion{keyIDA, testKeyA}, []KeyVersion{{0, testKeyB}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kr, err := NewRotatingKeyring(tt.active, tt.retired...)

			require.ErrorIs(t, err, ErrInvalidKey)
			assert.Nil(t, kr)
		})
	}
}

// A keyring with only the active key is today's single-key keyring: the
// ciphertexts it seals open in a keyring that later adds a retired key.
func TestRotatingKeyringOpensRowsSealedBeforeTheRotation(t *testing.T) {
	before, err := NewRotatingKeyring(KeyVersion{keyIDA, testKeyA})
	require.NoError(t, err)
	stored, err := before.Encrypt([]byte("GOMA800101AB3"), []byte("row-1"))
	require.NoError(t, err)

	during, err := NewRotatingKeyring(KeyVersion{keyIDB, testKeyB}, KeyVersion{keyIDA, testKeyA})
	require.NoError(t, err)

	plain, err := during.Decrypt(stored, []byte("row-1"))
	require.NoError(t, err)
	assert.Equal(t, []byte("GOMA800101AB3"), plain)
	assert.Equal(t, keyIDB, during.ActiveKeyID())
}

func TestKeyVersionNeverPrintsKeyMaterial(t *testing.T) {
	v := KeyVersion{ID: keyIDA, Key: testKeyA}

	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%x"} {
		assert.Equal(t, redacted, fmt.Sprintf(verb, v), verb)
	}
	assert.Equal(t, redacted, v.LogValue().String())
}

func TestNeedsReseal(t *testing.T) {
	old := newTestKeyring(t, keyIDA)
	current := newTestKeyring(t, keyIDB)
	onOld, err := old.Encrypt([]byte("x"), nil)
	require.NoError(t, err)
	onCurrent, err := current.Encrypt([]byte("x"), nil)
	require.NoError(t, err)

	needs, err := current.NeedsReseal(onOld)
	require.NoError(t, err)
	assert.True(t, needs)
	needs, err = current.NeedsReseal(onCurrent)
	require.NoError(t, err)
	assert.False(t, needs)
	_, err = current.NeedsReseal([]byte{0x01})
	require.ErrorIs(t, err, ErrMalformedCiphertext)
}

func TestResealMovesACiphertextToTheActiveKey(t *testing.T) {
	old := newTestKeyring(t, keyIDA)
	sealed, err := old.Encrypt([]byte("GOMA800101AB3"), []byte("row-1"))
	require.NoError(t, err)
	rotating, err := NewRotatingKeyring(KeyVersion{keyIDB, testKeyB}, KeyVersion{keyIDA, testKeyA})
	require.NoError(t, err)

	resealed, err := rotating.Reseal(sealed, []byte("row-1"))
	require.NoError(t, err)

	id, err := KeyIDOf(resealed)
	require.NoError(t, err)
	assert.Equal(t, keyIDB, id)
	assert.NotEqual(t, sealed[HeaderSize:], resealed[HeaderSize:])
	// The new key alone opens it: the old key can be dropped afterwards.
	onlyNew, err := NewRotatingKeyring(KeyVersion{keyIDB, testKeyB})
	require.NoError(t, err)
	plain, err := onlyNew.Decrypt(resealed, []byte("row-1"))
	require.NoError(t, err)
	assert.Equal(t, []byte("GOMA800101AB3"), plain)
	// The aad still binds it to its row.
	_, err = onlyNew.Decrypt(resealed, []byte("row-2"))
	require.ErrorIs(t, err, ErrDecrypt)
}

func TestResealFailsClosed(t *testing.T) {
	old := newTestKeyring(t, keyIDA)
	sealed, err := old.Encrypt([]byte("GOMA800101AB3"), []byte("row-1"))
	require.NoError(t, err)
	withoutOld, err := NewRotatingKeyring(KeyVersion{keyIDB, testKeyB})
	require.NoError(t, err)
	rotating, err := NewRotatingKeyring(KeyVersion{keyIDB, testKeyB}, KeyVersion{keyIDA, testKeyA})
	require.NoError(t, err)

	tests := []struct {
		name       string
		keyring    *Keyring
		ciphertext []byte
		aad        []byte
		want       error
	}{
		{"old key not in the keyring", withoutOld, sealed, []byte("row-1"), ErrUnknownKeyID},
		{"ciphertext of another row", rotating, sealed, []byte("row-2"), ErrDecrypt},
		{"truncated", rotating, sealed[:HeaderSize], []byte("row-1"), ErrMalformedCiphertext},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.keyring.Reseal(bytes.Clone(tt.ciphertext), tt.aad)

			require.ErrorIs(t, err, tt.want)
			assert.Nil(t, out)
		})
	}
}

// Rolling back is the same operation with the roles swapped: the old key
// active again, the new one retired.
func TestResealRollsBack(t *testing.T) {
	forward, err := NewRotatingKeyring(KeyVersion{keyIDB, testKeyB}, KeyVersion{keyIDA, testKeyA})
	require.NoError(t, err)
	back, err := NewRotatingKeyring(KeyVersion{keyIDA, testKeyA}, KeyVersion{keyIDB, testKeyB})
	require.NoError(t, err)
	onNew, err := forward.Encrypt([]byte("GOMA800101AB3"), nil)
	require.NoError(t, err)

	onOld, err := back.Reseal(onNew, nil)
	require.NoError(t, err)

	needs, err := back.NeedsReseal(onOld)
	require.NoError(t, err)
	assert.False(t, needs)
	plain, err := newTestKeyring(t, keyIDA).Decrypt(onOld, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte("GOMA800101AB3"), plain)
}
