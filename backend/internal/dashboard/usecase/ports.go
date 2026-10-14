// Package usecase holds the dashboard's application layer (M6-02,
// MODULE-dashboard): one read-only use case that composes other modules'
// ports into the Home of one taxpayer. The ports are declared here, on the
// consumer side (ADR-0001); cmd/api adapts the tax engine, the incomes and
// expenses readers, the taxpayers module and the access policy to them.
// Every port runs nested in the use case's unit of work, so the
// composition root must give them all the same tx.Manager.
package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Flags reports whether a module is effectively ON (its prerequisites
// included); *flags.Service implements it, failing closed.
type Flags interface {
	Enabled(ctx context.Context, module string) bool
}

// TaxpayerChecker answers whether the caller owns the taxpayer: nil, or the
// taxpayers module's 404 *apperr.Error for an unknown or another user's
// taxpayer. Archived taxpayers are readable.
type TaxpayerChecker interface {
	CheckTaxpayer(ctx context.Context, userID, taxpayerID string) error
}

// TaxEstimator is the tax engine's range estimate (Estimator.EstimateRange)
// as the dashboard reads it. D is the full estimate of the window's last
// month, which the dashboard passes through to its presenter untouched (the
// hero, contract TaxpayerEstimate). It reads the taxpayer's profile first,
// so an unknown or foreign taxpayer is its 404.
type TaxEstimator[D any] interface {
	EstimateWindow(ctx context.Context, userID, taxpayerID string, from, to domain.Month) (TaxWindow[D], error)
}

// TaxWindow is the engine's answer for [from, to].
type TaxWindow[D any] struct {
	// Months has one MonthTax per month, oldest first.
	Months []MonthTax
	// Current is the full estimate of the last month.
	Current D
	// Legend is the mandatory legend under every tax figure.
	Legend string
}

// Outcome is the sign of the month's consolidated IVA (contract
// TrendTax.iva_outcome).
type Outcome string

// The IVA outcomes.
const (
	OutcomePayable Outcome = "payable"
	OutcomeInFavor Outcome = "in_favor"
	OutcomeZero    Outcome = "zero"
)

// MonthTax is one month of the engine's range: its totals, the status of
// each regime held, and the records it read (incomes collected and
// expenses paid, before IVA).
type MonthTax struct {
	Month               domain.Month
	IVAPayable          money.Cents
	ISRPayable          money.Cents
	WithholdingsToRemit money.Cents
	TotalPayable        money.Cents
	IVAOutcome          Outcome
	Regimes             []domain.Status
	IncomeCollected     money.Cents
	ExpensesPaid        money.Cents
}

// RecordReader reads the dashboard's own record figures, used only when
// the engine is OFF (with it ON they come from the engine's rows), and the
// existence checks of is_empty. Totals are by month, a month without
// records absent.
type RecordReader interface {
	IncomeCollected(ctx context.Context, taxpayerID string, from, to domain.Month) (map[domain.Month]money.Cents, error)
	ExpensesPaid(ctx context.Context, taxpayerID string, from, to domain.Month) (map[domain.Month]money.Cents, error)
	HasIncomes(ctx context.Context, taxpayerID string) (bool, error)
	HasExpenses(ctx context.Context, taxpayerID string) (bool, error)
}

// AccessReader is the access policy's summary (access.Policy.Summary).
type AccessReader interface {
	Summary(ctx context.Context, userID string) (access.Summary, error)
}
