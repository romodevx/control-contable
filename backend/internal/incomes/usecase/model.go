package usecase

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// List page sizes (contract Limit).
const (
	DefaultLimit = 50
	MaxLimit     = 100
)

// IncomeInput is an income as the HTTP adapter decoded it. Zero values are
// absent or null fields; CFDIUUID and Description are raw text the use
// case normalizes (upper case; trimmed and NFC, blank = null).
type IncomeInput struct {
	Regime           fiscal.RegimeCode
	Counterparty     domain.Counterparty
	Amount           money.Cents
	IVARate          fiscal.IVARate
	IVA              money.Cents
	Withheld         fiscal.Withholdings
	Dates            domain.Dates
	PlatformActivity domain.PlatformActivity
	CFDIUUID         string
	Description      string
}

// CreateInput is createIncome: the client's idempotency id (empty: the
// server generates one), the path's taxpayer and the body.
type CreateInput struct {
	ID         string
	TaxpayerID string
	Income     IncomeInput
}

// Patch is updateIncome's merge-patch: a nil field is unchanged. A nullable
// field set to its zero value is cleared (a zero Date, an empty
// PlatformActivity, CFDIUUID or Description).
type Patch struct {
	Regime           *fiscal.RegimeCode
	Counterparty     *domain.Counterparty
	Amount           *money.Cents
	IVARate          *fiscal.IVARate
	IVA              *money.Cents
	ISRWithheld      *money.Cents
	IVAWithheld      *money.Cents
	InvoicedOn       *fiscal.Date
	CollectedOn      *fiscal.Date
	PlatformActivity *domain.PlatformActivity
	CFDIUUID         *string
	Description      *string
}

// apply returns stored's fields with p applied, as an input to normalize
// again.
func (p Patch) apply(stored Fields) IncomeInput {
	in := IncomeInput{
		Regime: stored.Regime, Counterparty: stored.Counterparty, Amount: stored.Amount,
		IVARate: stored.IVARate, IVA: stored.IVA, Withheld: stored.Withheld, Dates: stored.Dates,
		PlatformActivity: stored.PlatformActivity, CFDIUUID: stored.CFDIUUID.String(), Description: stored.Description,
	}
	set(&in.Regime, p.Regime)
	set(&in.Counterparty, p.Counterparty)
	set(&in.Amount, p.Amount)
	set(&in.IVARate, p.IVARate)
	set(&in.IVA, p.IVA)
	set(&in.Withheld.ISR, p.ISRWithheld)
	set(&in.Withheld.IVA, p.IVAWithheld)
	set(&in.Dates.InvoicedOn, p.InvoicedOn)
	set(&in.Dates.CollectedOn, p.CollectedOn)
	set(&in.PlatformActivity, p.PlatformActivity)
	set(&in.CFDIUUID, p.CFDIUUID)
	set(&in.Description, p.Description)
	return in
}

func set[T any](dst *T, v *T) {
	if v != nil {
		*dst = *v
	}
}

// Income is an income as the contract shows it (without warnings).
type Income struct {
	ID         string
	TaxpayerID string
	Fields
	BookedOn fiscal.Date
	// PendingCollection is "Por cobrar": no collection date yet. Only an
	// accrual income can lack it (a cash income books on it).
	PendingCollection bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func toIncome(s StoredIncome) Income {
	return Income{
		ID: s.ID, TaxpayerID: s.TaxpayerID, Fields: s.Fields, BookedOn: s.BookedOn,
		PendingCollection: s.Dates.CollectedOn.IsZero(), CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

// WriteResult is the answer of a write: the income, its non-blocking
// warnings (computed per write, never stored) and, on create, whether a
// record was created (201) or an idempotent retry replayed (200).
type WriteResult struct {
	Income   Income
	Warnings []domain.Warning
	Created  bool
}

// ListQuery is listIncomes: the path's taxpayer, the required period
// ("YYYY-MM", by booking date), an optional regime, the opaque cursor of
// the previous page and the page size (0: DefaultLimit).
type ListQuery struct {
	TaxpayerID string
	Period     string
	Regime     fiscal.RegimeCode
	Cursor     string
	Limit      int
}

// Page is one page of a list: newest first, the next page's cursor (empty
// on the last page) and the totals of the whole filtered period.
type Page struct {
	Items      []Income
	NextCursor string
	Totals     PeriodTotals
}
