// Package pushtoken seals push tokens at rest (MODULE-reminders §2.3,
// OQ-R7) with the platform crypto, the same application envelope as the
// RFC under the reminders' own keys:
//
//   - token_ciphertext = Keyring.Encrypt(token, aad) where aad =
//     installation id (16 bytes) || user id (16 bytes); the keyring
//     prefixes and authenticates the key_id. A ciphertext moved to another
//     row or user does not open.
//   - token_hmac = HMAC-SHA256(k_token_index, token), unscoped: the claim
//     must find the token whoever holds it.
package pushtoken

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// redacted is how key-holding types render in fmt and slog.
const redacted = "REDACTED"

// errZeroToken reports a Seal of the zero token: a programming error.
var errZeroToken = errors.New("reminders pushtoken: cannot seal the zero token")

// Vault implements usecase.TokenSealer; Open serves the reminder job. It is
// safe for concurrent use.
type Vault struct {
	keyring *crypto.Keyring
	index   *crypto.BlindIndex
}

var _ usecase.TokenSealer = (*Vault)(nil)

// NewVault returns a vault sealing with keyring and hashing with index
// (crypto.NewTokenIndex).
func NewVault(keyring *crypto.Keyring, index *crypto.BlindIndex) *Vault {
	return &Vault{keyring: keyring, index: index}
}

// Seal implements usecase.TokenSealer.
func (v *Vault) Seal(token domain.PushToken, b usecase.TokenBinding) (usecase.SealedToken, error) {
	if token.IsZero() {
		return usecase.SealedToken{}, errZeroToken
	}
	aad, err := rowAAD(b)
	if err != nil {
		return usecase.SealedToken{}, err
	}
	ciphertext, err := v.keyring.Encrypt([]byte(token.Reveal()), aad)
	if err != nil {
		return usecase.SealedToken{}, fmt.Errorf("reminders pushtoken: seal: %w", err)
	}
	hmac, err := v.index.Hash(token.Reveal())
	if err != nil {
		return usecase.SealedToken{}, fmt.Errorf("reminders pushtoken: hash: %w", err)
	}
	return usecase.SealedToken{
		Ciphertext: ciphertext,
		KeyID:      int16(v.keyring.ActiveKeyID()), // #nosec G115 -- crypto.MaxKeyID fits int16
		HMAC:       hmac,
	}, nil
}

// Open decrypts a token_ciphertext of the row b names; a ciphertext of
// another row fails with crypto.ErrDecrypt.
func (v *Vault) Open(ciphertext []byte, b usecase.TokenBinding) (domain.PushToken, error) {
	aad, err := rowAAD(b)
	if err != nil {
		return domain.PushToken{}, err
	}
	plain, err := v.keyring.Decrypt(ciphertext, aad)
	if err != nil {
		return domain.PushToken{}, fmt.Errorf("reminders pushtoken: open: %w", err)
	}
	token, err := domain.NewPushToken(string(plain))
	if err != nil {
		return domain.PushToken{}, fmt.Errorf("reminders pushtoken: stored token: %w", err)
	}
	return token, nil
}

// Reseal moves a token_ciphertext of the row b names to the keyring's
// active key (ADR-0012 key re-seal job); the token never leaves the
// keyring. It fails like Open.
func (v *Vault) Reseal(ciphertext []byte, b usecase.TokenBinding) ([]byte, error) {
	aad, err := rowAAD(b)
	if err != nil {
		return nil, err
	}
	resealed, err := v.keyring.Reseal(ciphertext, aad)
	if err != nil {
		return nil, fmt.Errorf("reminders pushtoken: reseal: %w", err)
	}
	return resealed, nil
}

// Format redacts the vault for every fmt verb.
func (Vault) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the vault in slog.
func (Vault) LogValue() slog.Value { return slog.StringValue(redacted) }

// rowAAD is the caller part of the GCM additional data: installation id ||
// user id, 32 bytes. Fixed widths keep the concatenation unambiguous.
func rowAAD(b usecase.TokenBinding) ([]byte, error) {
	installation, err := uuid.Parse(b.InstallationID)
	if err != nil {
		return nil, fmt.Errorf("reminders pushtoken: binding installation id: %w", err)
	}
	user, err := uuid.Parse(b.UserID)
	if err != nil {
		return nil, fmt.Errorf("reminders pushtoken: binding user id: %w", err)
	}
	aad := make([]byte, 0, len(installation)+len(user))
	aad = append(aad, installation[:]...)
	return append(aad, user[:]...), nil
}
