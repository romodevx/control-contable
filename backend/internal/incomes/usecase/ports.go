package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Lock is the row lock a store read takes in the unit of work's
// transaction.
type Lock int

// The lock modes.
const (
	// LockNone is a plain read.
	LockNone Lock = iota
	// LockUpdate is FOR NO KEY UPDATE: serializes the writers of one income.
	LockUpdate
)

// Fields are an income's persisted facts after normalization (the derived
// withholding is apart: it is a function of them). The zero
// value of a nullable field is null: a zero Date, an empty
// PlatformActivity, a zero CFDIUUID, an empty Description. Fields is
// comparable, which the idempotent replay relies on.
type Fields struct {
	Regime           fiscal.RegimeCode
	Counterparty     fiscal.ReceptorKind  // "" on a record stored before WH-06
	Operation        fiscal.OperationKind // "" on a record stored before WH-06
	ResidentialRent  bool
	ViaPlatform      bool
	Amount           money.Cents // subtotal before IVA
	IVARate          fiscal.IVARate
	IVA              money.Cents
	Dates            domain.Dates
	PlatformActivity domain.PlatformActivity
	CFDIUUID         fiscal.CFDIUUID
	Description      string
}

// StoredIncome is an income row as the caller may see it (RLS), deleted or
// not.
type StoredIncome struct {
	ID         string
	TaxpayerID string
	Fields
	Withholding fiscal.WithholdingRecord
	// CounterpartyType is the deprecated counterparty_type column.
	CounterpartyType domain.Counterparty
	// BookedOn is the ISR date the database derives (incomes.booked_on).
	BookedOn  fiscal.Date
	Deleted   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewIncome is an income to insert with its derived withholding. The
// database derives booked_on.
type NewIncome struct {
	ID         string
	UserID     string
	TaxpayerID string
	Fields
	Withholding fiscal.WithholdingRecord
}

// LastFacts are the withholding facts of the taxpayer's newest live income
// in a regime (MODULE-withholdings §2.1 "last used").
type LastFacts struct {
	Counterparty    fiscal.ReceptorKind
	Operation       fiscal.OperationKind
	ResidentialRent bool
	ViaPlatform     bool
}

// ListFilter selects the live incomes of a taxpayer booked in
// [From, Until), optionally of one regime (empty: all).
type ListFilter struct {
	TaxpayerID string
	From       fiscal.Date
	Until      fiscal.Date
	Regime     fiscal.RegimeCode
}

// Keyset is the position after which a page starts, in (booked_on DESC,
// id DESC) order. The zero Keyset is the first page.
type Keyset struct {
	BookedOn fiscal.Date
	ID       string
}

// PeriodTotals are the sums of every live income of a filter (contract
// PeriodTotals), not just a page.
type PeriodTotals struct {
	Count       int64
	Amount      money.Cents
	IVA         money.Cents
	ISRWithheld money.Cents
	IVAWithheld money.Cents
}

// Store is the incomes persistence port (infra). It runs on the unit of
// work in ctx as the caller (RLS): another user's incomes are invisible.
type Store interface {
	// LockID serializes the creates of one client id until the unit of
	// work ends: a double submit waits for the first, then finds it.
	LockID(ctx context.Context, id string) error
	// Find returns the income id, deleted or not, or ErrIncomeNotFound.
	Find(ctx context.Context, id string, lock Lock) (StoredIncome, error)
	// Insert stores a new income, or ErrIDTaken when the id exists (for
	// any user: the primary key sees across RLS).
	Insert(ctx context.Context, in NewIncome) error
	// Update replaces the facts and the derived withholding of income id.
	Update(ctx context.Context, id string, f Fields, w fiscal.WithholdingRecord) error
	// SetDeleted soft-deletes (true) or restores (false) income id.
	SetDeleted(ctx context.Context, id string, deleted bool) error
	// CFDIUUIDInUse reports another live income of taxpayerID (not
	// exceptID) with the folio.
	CFDIUUIDInUse(ctx context.Context, taxpayerID string, uuid fiscal.CFDIUUID, exceptID string) (bool, error)
	// List returns at most limit incomes of f after the keyset.
	List(ctx context.Context, f ListFilter, after Keyset, limit int) ([]StoredIncome, error)
	// Totals sums every income of f.
	Totals(ctx context.Context, f ListFilter) (PeriodTotals, error)
	// LastFacts returns the facts of the newest live income of taxpayerID
	// in regime that has them (a record stored before WH-06 has none), and
	// false when there is none.
	LastFacts(ctx context.Context, taxpayerID string, regime fiscal.RegimeCode) (LastFacts, bool, error)
	// IncomeBasis returns the catalog's income basis of a regime
	// (regimes.income_basis).
	IncomeBasis(ctx context.Context, person fiscal.PersonType, regime fiscal.RegimeCode) (domain.IncomeBasis, error)
}

// TaxpayerProfile is what incomes reads about the payee (M2's profile,
// adapted in cmd/api).
type TaxpayerProfile struct {
	PersonType fiscal.PersonType
	Archived   bool
	Regimes    []RegimeProfile
}

// RegimeProfile is one regime the taxpayer holds, with its window and the
// 625 options the withholding rules read.
type RegimeProfile struct {
	Code fiscal.RegimeCode
	// StartPeriod and EndPeriod bound the window ("YYYY-MM", inclusive;
	// an empty EndPeriod is open). Writes check the window through
	// CheckRegime (locked); only reads use these.
	StartPeriod        string
	EndPeriod          string
	PlatformActivities []domain.PlatformActivity // 625 option "activities"
	RFCGivenToPlatform bool                      // 625 option "rfc_given_to_platform"
}

// Taxpayers is the taxpayers module (M2) as incomes consumes it. Both
// methods return a client-ready *apperr.Error when they refuse (404
// not_found; 422 taxpayer_archived or regime_not_active), passed on as is.
type Taxpayers interface {
	// CheckRegime: does the caller's taxpayer hold the regime in the
	// period ("YYYY-MM"), and is it not archived? Called inside the
	// income's unit of work, its FOR SHARE locks hold until commit.
	CheckRegime(ctx context.Context, userID, taxpayerID, regimeCode, period string) error
	// Profile returns the caller's taxpayer, archived ones included.
	Profile(ctx context.Context, userID, taxpayerID string) (TaxpayerProfile, error)
}

// AccessPolicy is the part of access.Policy (P0-16) incomes consumes. Both
// methods return a 402 *apperr.Error when they refuse.
type AccessPolicy interface {
	// EnsureWritable refuses a read-only user (create, edit).
	EnsureWritable(ctx context.Context, userID string) error
	// ReserveRecord takes one trial slot, on create only.
	ReserveRecord(ctx context.Context, userID string, kind access.RecordKind) error
}

// RateCatalog is the tax engine's dated parameters (M5-08) as incomes
// consumes them: the rates the domain checks, and whether a date reads a
// provisional set.
type RateCatalog interface {
	fiscal.RateCatalog
	// Provisional reports whether the values in force on day on come
	// from a provisional parameter set. Any error means no parameter set
	// covers the day.
	Provisional(on fiscal.Date) (bool, error)
}
