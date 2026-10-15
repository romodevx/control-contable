package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Lock is the row lock a store read takes in the unit of work's
// transaction. Writers lock before they check, so the check and the write
// see the same row (READ COMMITTED: the lock wait ends on a fresh snapshot).
type Lock int

// The lock modes.
const (
	// LockNone is a plain read.
	LockNone Lock = iota
	// LockShare is FOR SHARE: other readers proceed, writers of the row wait.
	// TaxpayerRegimeChecker takes it so a concurrent end-regime or archive
	// waits for the record write it vouched for.
	LockShare
	// LockUpdate is FOR NO KEY UPDATE: serializes the writers of one row
	// without blocking the foreign-key checks of its children.
	LockUpdate
)

// AccessPolicy is the part of access.Policy (internal/platform/access) the
// taxpayers use cases consume. Both methods return a client-ready
// *apperr.Error when they refuse.
type AccessPolicy interface {
	// EnsureWritable refuses a read-only user (402 entitlement_required).
	EnsureWritable(ctx context.Context, userID string) error
	// CheckTaxpayerLimit refuses a create at existing >= TaxpayerCap (422
	// taxpayer_limit_reached).
	CheckTaxpayerLimit(ctx context.Context, userID string, existing int) error
}

// RowBinding is what an RFC ciphertext is bound to (MODULE-taxpayers S-1):
// the taxpayer row and its owner. Both are canonical UUID strings.
type RowBinding struct {
	TaxpayerID string
	UserID     string
}

// SealedRFC is an RFC as the taxpayers table stores it: never in clear.
type SealedRFC struct {
	// Ciphertext is AES-256-GCM in the crypto.Keyring frame, sealed with
	// additional data key_id || taxpayer id || user id.
	Ciphertext []byte
	// KeyID is the key_id prefix of Ciphertext (rfc_key_id).
	KeyID int16
	// Hash is the per-user blind index HMAC(k_rfc_index, user id || RFC).
	Hash []byte
	// Masked is the only form ever returned (rfc_masked).
	Masked string
	// CheckDigitValid is the SAT check digit verdict (D-RFC: a mismatch
	// is stored, not refused).
	CheckDigitValid bool
}

// RFCProtector seals an RFC for one taxpayer row. The use case generates
// the row id before it calls Seal (S-1): the ciphertext authenticates the
// id, so it never decrypts on another row or for another user.
type RFCProtector interface {
	Seal(rfc domain.RFC, binding RowBinding) (SealedRFC, error)
}

// NewTaxpayer is the row Create inserts.
type NewTaxpayer struct {
	ID          string
	UserID      string
	PersonType  domain.PersonType
	DisplayName domain.DisplayName
	RFC         SealedRFC
}

// StoredTaxpayer is a taxpayers row as the use cases read it: no RFC
// ciphertext, hash or key id (no response needs the full RFC, S-3).
type StoredTaxpayer struct {
	ID          string
	PersonType  domain.PersonType
	DisplayName string
	RFCMasked   string
	// ArchivedAt is the zero time while the taxpayer is active.
	ArchivedAt time.Time
	CreatedAt  time.Time
}

// Archived reports whether the taxpayer is archived.
func (t StoredTaxpayer) Archived() bool { return !t.ArchivedAt.IsZero() }

// TaxpayerStore is the taxpayers table in the unit of work in ctx (RLS:
// only the context user's rows exist). A row of another user is
// ErrTaxpayerNotFound, never a distinct error.
type TaxpayerStore interface {
	// CountForCreate locks the context user (FOR NO KEY UPDATE) and counts
	// their taxpayers, archived included: the cap check of Create (OD-2).
	CountForCreate(ctx context.Context) (int, error)
	// Insert writes t with its explicit id. ErrRFCAlreadyRegistered for an
	// RFC the user already holds; ErrIDConflict for an id taken outside
	// the user's scope.
	Insert(ctx context.Context, t NewTaxpayer) error
	// Find reads one taxpayer under lock; ErrTaxpayerNotFound.
	Find(ctx context.Context, id string, lock Lock) (StoredTaxpayer, error)
	// List reads the user's taxpayers by created_at, then id; archived
	// ones only when includeArchived.
	List(ctx context.Context, includeArchived bool) ([]StoredTaxpayer, error)
	// Rename sets display_name; ErrTaxpayerNotFound.
	Rename(ctx context.Context, id string, name domain.DisplayName) error
	// ReplaceRFC rewrites the five rfc_* columns together (OQ-1);
	// ErrRFCAlreadyRegistered, ErrTaxpayerNotFound.
	ReplaceRFC(ctx context.Context, id string, rfc SealedRFC) error
	// SetArchivedAt archives (non-zero at) or unarchives (zero at);
	// ErrTaxpayerNotFound.
	SetArchivedAt(ctx context.Context, id string, at time.Time) error
	// Delete hard-deletes the taxpayer; its regimes (and, from M3 on, its
	// records) cascade. It reports whether a row was deleted.
	Delete(ctx context.Context, id string) (bool, error)
}

// HeldRegime is a taxpayer_regimes row with its catalog name, treatment
// and income basis.
type HeldRegime struct {
	TaxpayerID string
	Code       string
	Name       string
	Treatment  string
	// IncomeBasis is regimes.income_basis ("cash" or "accrual").
	IncomeBasis string
	// Options are canonical (validated when written).
	Options map[string]domain.OptionValue
	Window  domain.Window
}

// NewRegime is the row AddRegime and Create insert.
type NewRegime struct {
	TaxpayerID string
	UserID     string
	PersonType domain.PersonType
	Code       string
	Options    map[string]domain.OptionValue
	Start      domain.Period
	// NeedsCPAReviewAck records that the write answered the
	// regime_combination_needs_cpa_review warning.
	NeedsCPAReviewAck bool
}

// RegimeStore is the taxpayer_regimes table in the unit of work in ctx.
type RegimeStore interface {
	// Regimes reads the regimes of the given taxpayers, ordered by
	// taxpayer, start_period, then code. One query, whatever the count.
	Regimes(ctx context.Context, taxpayerIDs ...string) ([]HeldRegime, error)
	// FindRegime reads one regime under lock; ErrRegimeNotHeld.
	FindRegime(ctx context.Context, taxpayerID, code string, lock Lock) (HeldRegime, error)
	// InsertRegime adds a regime. ErrRegimeAlreadyActive (primary key),
	// ErrIncompatibleRegimes (the overlap trigger, 23P01),
	// ErrRegimeNotForPersonType (catalog foreign key).
	InsertRegime(ctx context.Context, r NewRegime) error
	// ReopenRegime clears end_period of an ended regime and stores
	// options (OQ-5). It never writes start_period: the reopened window
	// starts where the row always did, so the overlap trigger checks the
	// real window. ErrRegimeNotHeld when no ended row matched.
	ReopenRegime(ctx context.Context, taxpayerID, code string, options map[string]domain.OptionValue, ack bool) error
	// UpdateOptions replaces options; ErrRegimeNotHeld.
	UpdateOptions(ctx context.Context, taxpayerID, code string, options map[string]domain.OptionValue) error
	// EndRegime sets end_period; ErrRegimeNotHeld.
	EndRegime(ctx context.Context, taxpayerID, code string, end domain.Period) error
}

// CatalogRegime is one regimes catalog row for a person type.
type CatalogRegime struct {
	Code       string
	PersonType domain.PersonType
	Name       string
	Treatment  string
	Options    domain.OptionSet
}

// RegimePair is an unordered pair of regime codes, stored with A < B.
type RegimePair struct {
	A, B string
}

// PairOf returns the pair of two codes in canonical order.
func PairOf(x, y string) RegimePair {
	if y < x {
		x, y = y, x
	}
	return RegimePair{A: x, B: y}
}

// Compatibility is a regime_compatibility verdict.
type Compatibility struct {
	Allowed        bool
	NeedsCPAReview bool
	// Reason is the user-facing (es-MX) explanation.
	Reason string
}

// RegimeCatalog reads the global regime catalog (M2-02).
type RegimeCatalog interface {
	// Regime returns code for personType; ErrRegimeNotForPersonType when
	// the catalog has no such key.
	Regime(ctx context.Context, personType domain.PersonType, code string) (CatalogRegime, error)
	// Compatibility returns the verdicts for every pair among codes. A
	// pair the matrix lacks is allowed (contract: an absent pair is
	// allowed).
	Compatibility(ctx context.Context, personType domain.PersonType, codes []string) (map[RegimePair]Compatibility, error)
}

// RegimeRecords reports the records (incomes, expenses) a regime holds.
// M3 and M4 own those tables; until they exist the infra adapter reports
// none.
type RegimeRecords interface {
	// LatestRecordPeriod returns the latest period holding a record of the
	// regime, or the zero Period when it has none.
	LatestRecordPeriod(ctx context.Context, taxpayerID, code string) (domain.Period, error)
}
