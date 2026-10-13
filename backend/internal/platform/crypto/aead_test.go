package crypto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed, obviously-fake test keys. Real keys come from the secret manager.
var (
	testKeyA = bytes.Repeat([]byte{0xA1}, KeySize)
	testKeyB = bytes.Repeat([]byte{0xB2}, KeySize)
)

const (
	keyIDA KeyID = 1
	keyIDB KeyID = 2
)

func newTestKeyring(t *testing.T, active KeyID) *Keyring {
	t.Helper()
	kr, err := NewKeyring(active, map[KeyID][]byte{keyIDA: testKeyA, keyIDB: testKeyB})
	require.NoError(t, err)
	return kr
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestNewKeyringValidation(t *testing.T) {
	tests := []struct {
		name   string
		active KeyID
		keys   map[KeyID][]byte
	}{
		{"no keys", keyIDA, map[KeyID][]byte{}},
		{"active key missing", keyIDB, map[KeyID][]byte{keyIDA: testKeyA}},
		{"key too short (AES-128)", keyIDA, map[KeyID][]byte{keyIDA: testKeyA[:16]}},
		{"key too long", keyIDA, map[KeyID][]byte{keyIDA: append(bytes.Clone(testKeyA), 0)}},
		{"zero key id", 0, map[KeyID][]byte{0: testKeyA}},
		{"key id beyond smallint", MaxKeyID + 1, map[KeyID][]byte{MaxKeyID + 1: testKeyA}},
		{"one bad key among good ones", keyIDA, map[KeyID][]byte{keyIDA: testKeyA, keyIDB: testKeyB[:31]}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kr, err := NewKeyring(tt.active, tt.keys)

			require.ErrorIs(t, err, ErrInvalidKey)
			assert.Nil(t, kr)
		})
	}
}

func TestKeyringCopiesKeyMaterial(t *testing.T) {
	key := bytes.Clone(testKeyA)
	kr, err := NewKeyring(keyIDA, map[KeyID][]byte{keyIDA: key})
	require.NoError(t, err)
	ct, err := kr.Encrypt([]byte("GOMA800101AB3"), nil)
	require.NoError(t, err)

	key[0] ^= 0xFF // caller scrubs or reuses its buffer

	pt, err := kr.Decrypt(ct, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte("GOMA800101AB3"), pt)
}

// TestEncryptKnownAnswer pins the wire format against NIST GCM test case 14
// (256-bit zero key, zero IV, 16 zero-byte plaintext): key_id(2, big endian)
// || nonce(12) || ciphertext || tag(16). The key_id is the GCM additional data
// (caller aad empty), so the tag differs from the NIST no-AAD tag; it was
// computed independently with Node's crypto.createCipheriv("aes-256-gcm")
// and setAAD(0x0007), which also reproduces the NIST tag without the AAD.
func TestEncryptKnownAnswer(t *testing.T) {
	kr, err := NewKeyring(7, map[KeyID][]byte{7: make([]byte, KeySize)})
	require.NoError(t, err)
	kr.random = bytes.NewReader(make([]byte, NonceSize))

	got, err := kr.Encrypt(make([]byte, 16), nil)

	require.NoError(t, err)
	want := mustHex(t, "0007"+
		"000000000000000000000000"+
		"cea7403d4d606b6e074ec5d3baf39d18"+
		"3b328dc6085fe6a6dc4031eba2669376")
	assert.Equal(t, want, got)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	kr := newTestKeyring(t, keyIDA)
	tests := []struct {
		name      string
		plaintext []byte
		aad       []byte
	}{
		{"rfc persona fisica with aad", []byte("GOMA800101AB3"), []byte("taxpayer:0b6c2e1e")},
		{"rfc persona moral no aad", []byte("ABC010101XY9"), nil},
		{"non-ascii", []byte("ÑAÑO800101AB3"), []byte("x")},
		{"empty plaintext", []byte{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := kr.Encrypt(tt.plaintext, tt.aad)
			require.NoError(t, err)

			id, err := KeyIDOf(ct)
			require.NoError(t, err)
			assert.Equal(t, keyIDA, id)
			assert.Len(t, ct, HeaderSize+len(tt.plaintext)+TagSize)
			if len(tt.plaintext) > 0 {
				assert.NotContains(t, string(ct), string(tt.plaintext))
			}

			pt, err := kr.Decrypt(ct, tt.aad)
			require.NoError(t, err)
			assert.True(t, bytes.Equal(tt.plaintext, pt), "round trip changed the plaintext")
		})
	}
}

func TestEncryptUsesFreshNonce(t *testing.T) {
	kr := newTestKeyring(t, keyIDA)

	first, err := kr.Encrypt([]byte("GOMA800101AB3"), nil)
	require.NoError(t, err)
	second, err := kr.Encrypt([]byte("GOMA800101AB3"), nil)
	require.NoError(t, err)

	assert.NotEqual(t, first[keyIDSize:HeaderSize], second[keyIDSize:HeaderSize])
	assert.NotEqual(t, first, second)
}

func TestEncryptFailsWhenRandomSourceFails(t *testing.T) {
	kr := newTestKeyring(t, keyIDA)
	kr.random = failingReader{}

	ct, err := kr.Encrypt([]byte("GOMA800101AB3"), nil)

	require.Error(t, err)
	assert.Nil(t, ct)
	assert.NotContains(t, err.Error(), "GOMA800101AB3")
}

func TestDecryptRejectsTamperingAndWrongKeys(t *testing.T) {
	kr := newTestKeyring(t, keyIDA)
	aad := []byte("taxpayer:1")
	ct, err := kr.Encrypt([]byte("GOMA800101AB3"), aad)
	require.NoError(t, err)

	tamper := func(index int) []byte {
		c := bytes.Clone(ct)
		c[index] ^= 0x01
		return c
	}
	withKeyID := func(id KeyID) []byte {
		c := bytes.Clone(ct)
		binary.BigEndian.PutUint16(c, uint16(id))
		return c
	}
	onlyB, err := NewKeyring(keyIDA, map[KeyID][]byte{keyIDA: testKeyB})
	require.NoError(t, err)
	// The key_id is authenticated: relabelling a ciphertext fails even when
	// the other id maps to identical key material.
	sameKeyTwice, err := NewKeyring(keyIDA, map[KeyID][]byte{keyIDA: testKeyA, keyIDB: testKeyA})
	require.NoError(t, err)

	tests := []struct {
		name    string
		keyring *Keyring
		ct      []byte
		aad     []byte
		wantErr error
	}{
		{"flipped ciphertext byte", kr, tamper(HeaderSize), aad, ErrDecrypt},
		{"flipped tag byte", kr, tamper(len(ct) - 1), aad, ErrDecrypt},
		{"flipped nonce byte", kr, tamper(keyIDSize), aad, ErrDecrypt},
		{"key id swapped to another known key", kr, withKeyID(keyIDB), aad, ErrDecrypt},
		{"key id relabelled to an id holding the same key", sameKeyTwice, withKeyID(keyIDB), aad, ErrDecrypt},
		{"wrong aad", kr, ct, []byte("taxpayer:2"), ErrDecrypt},
		{"missing aad", kr, ct, nil, ErrDecrypt},
		{"same key id, different key material", onlyB, ct, aad, ErrDecrypt},
		{"unknown key id", kr, withKeyID(99), aad, ErrUnknownKeyID},
		{"truncated tag", kr, ct[:len(ct)-1], aad, ErrDecrypt},
		{"shorter than header+tag", kr, ct[:HeaderSize+TagSize-1], aad, ErrMalformedCiphertext},
		{"only a key id", kr, ct[:keyIDSize], aad, ErrMalformedCiphertext},
		{"empty", kr, nil, aad, ErrMalformedCiphertext},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, err := tt.keyring.Decrypt(tt.ct, tt.aad)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, pt)
		})
	}
}

func TestKeyRotation(t *testing.T) {
	oldRing := newTestKeyring(t, keyIDA)
	legacy, err := oldRing.Encrypt([]byte("GOMA800101AB3"), nil)
	require.NoError(t, err)

	rotated := newTestKeyring(t, keyIDB)
	fresh, err := rotated.Encrypt([]byte("GOMA800101AB3"), nil)
	require.NoError(t, err)

	assert.Equal(t, keyIDB, rotated.ActiveKeyID())
	legacyID, err := KeyIDOf(legacy)
	require.NoError(t, err)
	assert.Equal(t, keyIDA, legacyID, "old rows keep the key that sealed them")
	freshID, err := KeyIDOf(fresh)
	require.NoError(t, err)
	assert.Equal(t, keyIDB, freshID, "new rows use the active key")

	pt, err := rotated.Decrypt(legacy, nil)
	require.NoError(t, err, "retired keys still decrypt during rotation")
	assert.Equal(t, []byte("GOMA800101AB3"), pt)
}

func TestKeyIDOfRejectsShortInput(t *testing.T) {
	_, err := KeyIDOf([]byte{0x01})

	require.ErrorIs(t, err, ErrMalformedCiphertext)
}

func TestKeyringNeverPrintsKeyMaterial(t *testing.T) {
	kr := newTestKeyring(t, keyIDA)
	var logged bytes.Buffer
	slog.New(slog.NewTextHandler(&logged, nil)).Info("boot", "keyring", kr)

	for _, rendered := range []string{
		fmt.Sprintf("%v", kr), fmt.Sprintf("%+v", kr), fmt.Sprintf("%#v", kr), fmt.Sprintf("%s", kr), logged.String(),
	} {
		assert.NotContains(t, rendered, "a1a1")
		assert.NotContains(t, strings.ToLower(rendered), "161 161")
		assert.Contains(t, rendered, "REDACTED")
	}
}
