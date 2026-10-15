package crypto

import (
	"fmt"
	"io"
	"log/slog"
)

// KeyVersion is one AES-256 key and the id that names it in every ciphertext
// it seals (ADR-0012).
type KeyVersion struct {
	ID  KeyID
	Key []byte
}

// Format redacts the key version for every fmt verb.
func (KeyVersion) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the key version in slog.
func (KeyVersion) LogValue() slog.Value { return slog.StringValue(redacted) }

// NewRotatingKeyring builds the keyring of a key rotation (ADR-0012): it
// seals with active and still opens what each retired key sealed, so rows
// keep opening while the re-seal job moves them to the active key. Every id
// must be distinct: a retired key may never shadow the active one.
func NewRotatingKeyring(active KeyVersion, retired ...KeyVersion) (*Keyring, error) {
	keys := make(map[KeyID][]byte, 1+len(retired))
	keys[active.ID] = active.Key
	for _, r := range retired {
		if _, taken := keys[r.ID]; taken {
			return nil, fmt.Errorf("%w: key id %d is given twice", ErrInvalidKey, r.ID)
		}
		keys[r.ID] = r.Key
	}
	return NewKeyring(active.ID, keys)
}

// NeedsReseal reports whether ciphertext was sealed with a key other than
// the active one, without decrypting it.
func (k *Keyring) NeedsReseal(ciphertext []byte) (bool, error) {
	id, err := KeyIDOf(ciphertext)
	if err != nil {
		return false, err
	}
	return id != k.active, nil
}

// Reseal opens ciphertext with the key its key_id names and seals the
// plaintext again with the active key under the same aad, which must be the
// owning row's. The plaintext never leaves the keyring and is zeroed before
// Reseal returns. It fails as Decrypt does, and then returns no ciphertext.
func (k *Keyring) Reseal(ciphertext, aad []byte) ([]byte, error) {
	plain, err := k.Decrypt(ciphertext, aad)
	if err != nil {
		return nil, err
	}
	defer clear(plain)
	return k.Encrypt(plain, aad)
}
