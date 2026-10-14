package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Lock is the row lock a TaxInputStore read takes in the unit of work's
// transaction.
type Lock int

// The lock modes.
const (
	// LockNone is a plain read.
	LockNone Lock = iota
	// LockWrite is what a replace holds until commit: FOR SHARE on the
	// taxpayer (an archive waits, so no value lands on an archived
	// taxpayer) and FOR NO KEY UPDATE on the taxpayer_regimes row (two
	// replaces of one regime run one at a time; the lock does not block the
	// inputs' foreign-key checks).
	LockWrite
)

// AccessPolicy is the part of access.Policy (internal/platform/access) the
// tax-inputs writes consume. It returns a client-ready *apperr.Error (402
// entitlement_required) for a read-only user.
type AccessPolicy interface {
	EnsureWritable(ctx context.Context, userID string) error
}

// TaxpayerState is what the tax-inputs use cases read about a taxpayer.
type TaxpayerState struct {
	Archived bool
}

// HeldRegime is a regime the taxpayer holds, with the catalog definitions
// of its inputs (regimes.required_inputs for the taxpayer's person type).
type HeldRegime struct {
	Start domain.Period
	// End is the zero Period while the regime is open.
	End         domain.Period
	Definitions []domain.InputDef
}

// ActiveIn reports whether the regime is held in some month of fiscalYear.
func (r HeldRegime) ActiveIn(fiscalYear int) bool {
	return r.Start.Year() <= fiscalYear && (r.End.IsZero() || r.End.Year() >= fiscalYear)
}

// InputsKey names the inputs of one taxpayer, regime and fiscal year: the
// set a PUT replaces.
type InputsKey struct {
	TaxpayerID string
	Regime     fiscal.RegimeCode
	FiscalYear int
}

// StoredInput is one taxpayer_tax_inputs value. Only the field of Kind is
// meaningful; Month is set only for money_monthly.
type StoredInput struct {
	Key   domain.InputKey
	Kind  domain.InputKind
	Month time.Month
	Cents money.Cents
	// Rate is the decimal exactly as stored ("0.1500" keeps its scale).
	Rate string
	Bool bool
}

// TaxInputStore is taxpayer_tax_inputs, with the taxpayer and regime rows
// it hangs from, in the unit of work in ctx (RLS: only the context user's
// rows exist, so another user's taxpayer is ErrTaxpayerNotFound).
type TaxInputStore interface {
	// Taxpayer reads one taxpayer under lock; ErrTaxpayerNotFound.
	Taxpayer(ctx context.Context, taxpayerID string, lock Lock) (TaxpayerState, error)
	// Regime reads the taxpayer's regime code under lock, with its input
	// definitions; ErrRegimeNotHeld.
	Regime(ctx context.Context, taxpayerID string, code fiscal.RegimeCode, lock Lock) (HeldRegime, error)
	// Values reads the stored values of key, in no particular order.
	Values(ctx context.Context, key InputsKey) ([]StoredInput, error)
	// Replace deletes every value of key and stores values as userID's.
	Replace(ctx context.Context, userID string, key InputsKey, values []StoredInput) error
}
