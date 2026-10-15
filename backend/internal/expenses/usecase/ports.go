package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// AccessPolicy is the part of access.Policy (internal/platform/access) the
// expense use cases consume. Both methods return a client-ready
// *apperr.Error (402 entitlement_required) when they refuse.
type AccessPolicy interface {
	// EnsureWritable refuses an edit by a read-only user.
	EnsureWritable(ctx context.Context, userID string) error
	// ReserveRecord takes one trial slot of kind in the caller's unit of
	// work: on create only, never on restore (MODULE-incomes §2.7).
	ReserveRecord(ctx context.Context, userID string, kind access.RecordKind) error
}

// TaxpayerRegimeChecker is this module's copy of the taxpayers port: does
// the caller's taxpayer hold regimeCode in period ("YYYY-MM"), and is it not
// archived? It answers nil or a client-ready *apperr.Error (404, 422
// taxpayer_archived / regime_not_active). Called inside the expense's unit
// of work, its FOR SHARE locks hold until the expense commits.
type TaxpayerRegimeChecker interface {
	CheckRegime(ctx context.Context, userID, taxpayerID, regimeCode, period string) error
}

// TaxpayerProfiles reads what the expense rules need about a taxpayer of
// the caller: its person type, archived state and regime windows with their
// options. An unknown or foreign taxpayer is a client-ready 404
// *apperr.Error. cmd/api adapts the taxpayers module's profile reader.
type TaxpayerProfiles interface {
	Profile(ctx context.Context, userID, taxpayerID string) (TaxpayerProfile, error)
}

// TaxpayerProfile is a taxpayer as the expense rules see it.
type TaxpayerProfile struct {
	PersonType fiscal.PersonType
	Archived   bool
	Regimes    []RegimeHold
}

// RegimeHold is one window of a held regime and the options that change a
// category's hint (606 deduction_mode, 625 definitive_option).
type RegimeHold struct {
	Code fiscal.RegimeCode
	// StartPeriod and EndPeriod are "YYYY-MM"; an empty EndPeriod is open.
	StartPeriod, EndPeriod string
	Options                domain.Options
}

// covers reports whether the window holds period ("YYYY-MM"). Fixed-width
// periods order as strings.
func (h RegimeHold) covers(period string) bool {
	return period >= h.StartPeriod && (h.EndPeriod == "" || period <= h.EndPeriod)
}

// holding returns the window of code that holds period.
func (p TaxpayerProfile) holding(code fiscal.RegimeCode, period string) (RegimeHold, bool) {
	for _, h := range p.Regimes {
		if h.Code == code && h.covers(period) {
			return h, true
		}
	}
	return RegimeHold{}, false
}

// payer is the domain's view of the taxpayer paying under hold.
func (p TaxpayerProfile) payer(hold RegimeHold) domain.Payer {
	return domain.Payer{PersonType: p.PersonType, Regime: hold.Code, Options: hold.Options}
}

// Lock is the row lock ExpenseStore.Find takes in the unit of work.
type Lock int

// The lock modes.
const (
	// LockNone is a plain read.
	LockNone Lock = iota
	// LockUpdate is FOR NO KEY UPDATE: serializes the writers of one expense.
	LockUpdate
)

// StoredExpense is an expenses row.
type StoredExpense struct {
	ID            string
	TaxpayerID    string
	Regime        fiscal.RegimeCode
	Category      domain.CategoryCode
	Amount        money.Cents
	IVARate       fiscal.IVARate
	IVA           money.Cents
	PaymentMethod domain.PaymentMethod
	PaidOn        fiscal.Date
	HasCFDI       bool
	CFDIUUID      fiscal.CFDIUUID
	// Deductible is the effective verdict; Overridden says the user set it.
	Deductible bool
	Overridden bool
	// CashOverLimit and CreditableIVA are the assessed tax facts the tax
	// engine's aggregate sums (domain.Assessment): stored on every write,
	// like Deductible, so a later rule or parameter change never rewrites a
	// past month.
	CashOverLimit bool
	CreditableIVA money.Cents
	// CreditableIVAWithheld is the creditable share of the IVA withheld
	// from the supplier; CreditableIVA is then the share paid
	// (MODULE-withholdings §4.2).
	CreditableIVAWithheld money.Cents
	// Supplier is the persona moral's description of the supplier, defaults
	// resolved; zero for a persona física.
	Supplier domain.Supplier
	// Withholding is the stored derivation (ADR-0008); its zero value is a
	// record written before WH-07.
	Withholding fiscal.WithholdingRecord
	Description string // "" is NULL
	Deleted     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ExpenseStore persists expenses on the unit of work in ctx, as the ctx
// user (RLS): another user's rows are invisible. It never begins, commits
// or sets session state.
type ExpenseStore interface {
	// LockID serializes the creates of one client id until the unit of
	// work ends (a double submit waits for the first, then finds it).
	LockID(ctx context.Context, id string) error
	// Find returns the expense id, live or soft-deleted, or
	// ErrExpenseNotFound.
	Find(ctx context.Context, id string, lock Lock) (StoredExpense, error)
	// Insert stores e for userID and returns the stored row; an id already
	// taken (by anyone) is ErrIDConflict.
	Insert(ctx context.Context, userID string, e StoredExpense) (StoredExpense, error)
	// Update writes every editable column of e and returns the stored row.
	Update(ctx context.Context, e StoredExpense) (StoredExpense, error)
	// SetDeleted soft-deletes (true) or restores (false) the expense.
	SetDeleted(ctx context.Context, id string, deleted bool) (StoredExpense, error)
	// CFDIInUse reports another live expense of the taxpayer, other than
	// exceptID, with the folio fiscal.
	CFDIInUse(ctx context.Context, taxpayerID, exceptID string, uuid fiscal.CFDIUUID) (bool, error)
	// List returns up to q.Limit live expenses of the period after q.After,
	// ordered by paid_on, then id, both descending.
	List(ctx context.Context, q ListQuery) ([]StoredExpense, error)
	// Totals sums every live expense of the period.
	Totals(ctx context.Context, q PeriodQuery) (Totals, error)
}

// PeriodQuery selects a taxpayer's live expenses paid in [From, Until),
// optionally of one regime.
type PeriodQuery struct {
	TaxpayerID string
	From       fiscal.Date
	Until      fiscal.Date
	Regime     fiscal.RegimeCode // "" = every regime
}

// ListQuery is one keyset page of a PeriodQuery.
type ListQuery struct {
	PeriodQuery
	After *Cursor // nil = the first page
	Limit int
}

// Cursor is the keyset position of the last expense of a page.
type Cursor struct {
	PaidOn fiscal.Date
	ID     string
}

// Totals are the sums of a period's live expenses (contract PeriodTotals).
type Totals struct {
	Count       int
	Amount      money.Cents
	IVA         money.Cents
	ISRWithheld money.Cents
	IVAWithheld money.Cents
}

// CategoryEntry is one category with its rule for a regime and person type
// (expense_categories joined with category_regime_rules).
type CategoryEntry struct {
	Rule       domain.CategoryRule
	Name       string
	Note       string // "" = no note
	LegalBasis string
}

// CategoryCatalog reads the global category catalog.
type CategoryCatalog interface {
	// Rule returns category's rule for the regime and person type, or
	// ErrCategoryNotFound.
	Rule(ctx context.Context, regime fiscal.RegimeCode, person fiscal.PersonType, category domain.CategoryCode) (CategoryEntry, error)
	// Rules returns every category's rule for the regime and person type,
	// in the catalog's sort order.
	Rules(ctx context.Context, regime fiscal.RegimeCode, person fiscal.PersonType) ([]CategoryEntry, error)
}

// Parameters is the domain's RateCatalog plus the dated facts the form and
// the provisional warning need. cmd/api adapts the tax engine's
// ratecatalog.Catalog[domain.AmountKey] (M5-08).
type Parameters interface {
	domain.RateCatalog
	// On returns the parameters in force on day on, or
	// ErrParametersUnavailable when no parameter set covers the day.
	On(on fiscal.Date) (DatedParameters, error)
}

// DatedParameters are the parameter facts of one day.
type DatedParameters struct {
	IVARatesAllowed  []fiscal.IVARate
	CashPaymentLimit money.Cents
	// Provisional: the day's fiscal year has no published set yet, so the
	// values are the previous year's (provisional_parameters).
	Provisional bool
}
