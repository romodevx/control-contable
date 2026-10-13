package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"math"
)

// Wire format of a Keyring ciphertext:
//
//	key_id (2 bytes, big endian) || nonce (12 bytes) || ciphertext || tag (16 bytes)
//
// The key_id is not part of the additional data: swapping it can only select
// a different key, and GCM authentication then fails.
const (
	// KeySize is the AES-256 key length in bytes.
	KeySize = 32
	// NonceSize is the GCM nonce length in bytes.
	NonceSize = 12
	// TagSize is the GCM authentication tag length in bytes.
	TagSize = 16
	// HeaderSize is the key_id plus nonce prefix of every ciphertext.
	HeaderSize = keyIDSize + NonceSize
	// MaxKeyID is the largest key id; it fits the smallint rfc_key_id column.
	MaxKeyID KeyID = math.MaxInt16

	keyIDSize = 2
)

// KeyID names one encryption key version (rfc_key_id). Valid ids are 1..MaxKeyID.
type KeyID uint16

// Keyring encrypts with its active key and decrypts with any key it holds,
// which is what lets keys rotate: add the new key, make it active, re-encrypt
// rows whose KeyIDOf differs from ActiveKeyID, then drop the old key.
//
// Nonces are random (crypto/rand); keep each key under 2^32 encryptions
// (NIST SP 800-38D section 8.3). It is safe for concurrent use.
type Keyring struct {
	active KeyID
	aeads  map[KeyID]cipher.AEAD
	random io.Reader
}

// NewKeyring builds a keyring from 32-byte AES-256 keys. The key bytes are
// copied, so the caller may scrub its buffers afterwards.
func NewKeyring(active KeyID, keys map[KeyID][]byte) (*Keyring, error) {
	if _, ok := keys[active]; !ok {
		return nil, fmt.Errorf("%w: active key id %d has no key", ErrInvalidKey, active)
	}
	aeads := make(map[KeyID]cipher.AEAD, len(keys))
	for id, key := range keys {
		aead, err := newAEAD(id, key)
		if err != nil {
			return nil, err
		}
		aeads[id] = aead
	}
	return &Keyring{active: active, aeads: aeads, random: rand.Reader}, nil
}

func newAEAD(id KeyID, key []byte) (cipher.AEAD, error) {
	if id < 1 || id > MaxKeyID {
		return nil, fmt.Errorf("%w: key id %d not in [1, %d]", ErrInvalidKey, id, MaxKeyID)
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: key id %d is %d bytes, want %d", ErrInvalidKey, id, len(key), KeySize)
	}
	block, err := aes.NewCipher(bytes.Clone(key))
	if err != nil {
		return nil, fmt.Errorf("%w: key id %d: %w", ErrInvalidKey, id, err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: key id %d: %w", ErrInvalidKey, id, err)
	}
	return aead, nil
}

// ActiveKeyID returns the id of the key that Encrypt uses.
func (k *Keyring) ActiveKeyID() KeyID { return k.active }

// Encrypt seals plaintext with the active key under a fresh random nonce.
// aad (may be nil) is authenticated but not stored; Decrypt must receive the
// same aad, e.g. the owning row id, to stop ciphertexts being moved between rows.
func (k *Keyring) Encrypt(plaintext, aad []byte) ([]byte, error) {
	out := make([]byte, HeaderSize, HeaderSize+len(plaintext)+TagSize)
	binary.BigEndian.PutUint16(out, uint16(k.active))
	nonce := out[keyIDSize:HeaderSize]
	if _, err := io.ReadFull(k.random, nonce); err != nil {
		return nil, fmt.Errorf("crypto: read nonce: %w", err)
	}
	return k.aeads[k.active].Seal(out, nonce, plaintext, aad), nil
}

// Decrypt opens a ciphertext made by Encrypt with the key its key_id names.
// It returns ErrMalformedCiphertext, ErrUnknownKeyID or ErrDecrypt.
func (k *Keyring) Decrypt(ciphertext, aad []byte) ([]byte, error) {
	if len(ciphertext) < HeaderSize+TagSize {
		return nil, fmt.Errorf("%w: %d bytes", ErrMalformedCiphertext, len(ciphertext))
	}
	id := KeyID(binary.BigEndian.Uint16(ciphertext))
	aead, ok := k.aeads[id]
	if !ok {
		return nil, fmt.Errorf("%w: %d", ErrUnknownKeyID, id)
	}
	nonce, sealed := ciphertext[keyIDSize:HeaderSize], ciphertext[HeaderSize:]
	plaintext, err := aead.Open(nil, nonce, sealed, aad)
	if err != nil {
		return nil, fmt.Errorf("%w: key id %d", ErrDecrypt, id)
	}
	return plaintext, nil
}

// Format redacts the keyring for every fmt verb (%v, %+v, %#v, %s), so
// secret material never reaches a log or an error message.
func (k *Keyring) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the keyring in slog.
func (k *Keyring) LogValue() slog.Value { return slog.StringValue(redacted) }

// KeyIDOf returns the key id that sealed ciphertext, without decrypting it.
// Use it to fill rfc_key_id and to find rows that still need re-encryption.
func KeyIDOf(ciphertext []byte) (KeyID, error) {
	if len(ciphertext) < keyIDSize {
		return 0, fmt.Errorf("%w: %d bytes", ErrMalformedCiphertext, len(ciphertext))
	}
	return KeyID(binary.BigEndian.Uint16(ciphertext)), nil
}
