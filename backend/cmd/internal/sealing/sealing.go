// Package sealing builds the keyrings of the AES-256-GCM sealing keys from
// their configuration (ADR-0012): the active key and, while a rotation is in
// flight, the retired one, so cmd/api and cmd/worker open what either key
// sealed and seal with the active key only.
package sealing

import (
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
)

// Keyring seals with key under id and also opens what previous sealed, when
// one is configured. config bounds both ids to [1, crypto.MaxKeyID].
func Keyring(id int, key []byte, previous config.PreviousKey) (*crypto.Keyring, error) {
	active := crypto.KeyVersion{ID: crypto.KeyID(id), Key: key} // #nosec G115 -- config bounds it to [1, crypto.MaxKeyID]
	if previous.IsZero() {
		return crypto.NewRotatingKeyring(active)
	}
	retired := crypto.KeyVersion{ID: crypto.KeyID(previous.KeyID), Key: previous.Key.RevealBytes()} // #nosec G115 -- config bounds it to [1, crypto.MaxKeyID]
	return crypto.NewRotatingKeyring(active, retired)
}
