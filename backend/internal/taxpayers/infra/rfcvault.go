package infra

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// redacted is how key-holding types render in fmt and slog.
const redacted = "REDACTED"

// errZeroRFC reports a Seal of the zero RFC: a programming error.
var errZeroRFC = errors.New("taxpayers infra: cannot seal the zero RFC")

// RFCVault implements usecase.RFCProtector over the platform crypto
// (MODULE-taxpayers OD-5, S-1, S-2):
//
//   - rfc_encrypted = Keyring.Encrypt(RFC, aad) where aad = taxpayer id (16
//     bytes) || user id (16 bytes); the keyring prefixes and authenticates
//     the key_id, so the whole additional data is key_id || taxpayer.id ||
//     user_id. Fixed widths keep the concatenation unambiguous.
//   - rfc_hash = HMAC-SHA256(k_rfc_index, user id (16 bytes) || RFC).
//
// The RFC key and the index pepper are separate secrets, and the pepper is
// not the email index's. It is safe for concurrent use.
type RFCVault struct {
	keyring *crypto.Keyring
	index   *crypto.BlindIndex
}

var _ usecase.RFCProtector = (*RFCVault)(nil)

// NewRFCVault returns a vault sealing with keyring and hashing with index
// (crypto.NewRFCIndex).
func NewRFCVault(keyring *crypto.Keyring, index *crypto.BlindIndex) *RFCVault {
	return &RFCVault{keyring: keyring, index: index}
}

// Seal implements usecase.RFCProtector.
func (v *RFCVault) Seal(rfc domain.RFC, binding usecase.RowBinding) (usecase.SealedRFC, error) {
	if rfc.IsZero() {
		return usecase.SealedRFC{}, errZeroRFC
	}
	taxpayerID, userID, err := bindingIDs(binding)
	if err != nil {
		return usecase.SealedRFC{}, err
	}
	ciphertext, err := v.keyring.Encrypt([]byte(rfc.String()), rowAAD(taxpayerID, userID))
	if err != nil {
		return usecase.SealedRFC{}, fmt.Errorf("taxpayers infra: seal rfc: %w", err)
	}
	hash, err := v.index.HashScoped(userID[:], rfc.String())
	if err != nil {
		return usecase.SealedRFC{}, fmt.Errorf("taxpayers infra: hash rfc: %w", err)
	}
	return usecase.SealedRFC{
		Ciphertext:      ciphertext,
		KeyID:           int16(v.keyring.ActiveKeyID()), // #nosec G115 -- crypto.MaxKeyID fits int16
		Hash:            hash,
		Masked:          rfc.Masked(),
		CheckDigitValid: rfc.CheckDigitValid(),
	}, nil
}

// Open decrypts an rfc_encrypted value of the row binding names. A
// ciphertext moved to another row or another user fails with
// crypto.ErrDecrypt. No response returns the full RFC (S-3): Open exists
// for key rotation and for the row-binding tests.
func (v *RFCVault) Open(ciphertext []byte, binding usecase.RowBinding) (domain.RFC, error) {
	taxpayerID, userID, err := bindingIDs(binding)
	if err != nil {
		return domain.RFC{}, err
	}
	plain, err := v.keyring.Decrypt(ciphertext, rowAAD(taxpayerID, userID))
	if err != nil {
		return domain.RFC{}, fmt.Errorf("taxpayers infra: open rfc: %w", err)
	}
	rfc, err := domain.NewRFC(string(plain))
	if err != nil {
		return domain.RFC{}, fmt.Errorf("taxpayers infra: stored rfc: %w", err)
	}
	return rfc, nil
}

// Reseal moves an rfc_encrypted value of the row binding names to the
// keyring's active key (ADR-0012 key re-seal job); the RFC never leaves the
// keyring. It fails like Open: crypto.ErrDecrypt for another row's value,
// crypto.ErrUnknownKeyID when the sealing key is no longer held.
func (v *RFCVault) Reseal(ciphertext []byte, binding usecase.RowBinding) ([]byte, error) {
	taxpayerID, userID, err := bindingIDs(binding)
	if err != nil {
		return nil, err
	}
	resealed, err := v.keyring.Reseal(ciphertext, rowAAD(taxpayerID, userID))
	if err != nil {
		return nil, fmt.Errorf("taxpayers infra: reseal rfc: %w", err)
	}
	return resealed, nil
}

// Format redacts the vault for every fmt verb.
func (RFCVault) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the vault in slog.
func (RFCVault) LogValue() slog.Value { return slog.StringValue(redacted) }

func bindingIDs(b usecase.RowBinding) (taxpayerID, userID uuid.UUID, err error) {
	if taxpayerID, err = uuid.Parse(b.TaxpayerID); err != nil {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("taxpayers infra: binding taxpayer id: %w", err)
	}
	if userID, err = uuid.Parse(b.UserID); err != nil {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("taxpayers infra: binding user id: %w", err)
	}
	return taxpayerID, userID, nil
}

// rowAAD is the caller part of the GCM additional data: taxpayer id ||
// user id, 32 bytes.
func rowAAD(taxpayerID, userID uuid.UUID) []byte {
	aad := make([]byte, 0, len(taxpayerID)+len(userID))
	aad = append(aad, taxpayerID[:]...)
	return append(aad, userID[:]...)
}
