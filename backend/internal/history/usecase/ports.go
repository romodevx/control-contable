// Package usecase holds the history's application layer (M8-03,
// MODULE-history): getHistory (one fiscal year of monthly estimates with
// the "changed since viewed" marker) and markHistoryViewed (an idempotent
// upsert of what the user saw). The ports are declared here, on the
// consumer side (ADR-0001); cmd/api adapts the tax engine's range estimate
// and the estimate_views store to them. Every port runs nested in the use
// case's unit of work, so the composition root must give them all the same
// tx.Manager.
package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
)

// Flags reports whether a module is effectively ON (its prerequisites
// included); *flags.Service implements it, failing closed.
type Flags interface {
	Enabled(ctx context.Context, module string) bool
}

// Estimator is the tax engine's range estimate (Estimator.EstimateRange)
// as History reads it: one aggregate pass for [from, to], never one
// estimate per month (§2.6). It reads the taxpayer's profile first, so an
// unknown or another user's taxpayer is its 404 *apperr.Error.
type Estimator interface {
	EstimateMonths(ctx context.Context, userID, taxpayerID string, from, to domain.Month) (Estimates, error)
}

// Estimates is the engine's answer for [from, to].
type Estimates struct {
	// Months has one MonthEstimate per month, oldest first.
	Months []MonthEstimate
	// Legend is the mandatory legend under every tax figure.
	Legend string
}

// MonthEstimate is one month as the engine estimated it, reduced to what
// History shows and compares.
type MonthEstimate struct {
	Month domain.Month
	// Regimes are the statuses of the regimes held in the month.
	Regimes []domain.Status
	Tax     domain.Tax
	// DueDate is the payment date, "YYYY-MM-DD".
	DueDate          string
	CPAReviewPending bool
	// Seen is the estimate's fingerprints, engine version and parameter
	// set label: what a view of it stores.
	Seen         domain.Seen
	ParameterSet domain.ParameterSet
	Disclaimers  []domain.Disclaimer
	// HasIncomes and HasExpenses: a live record of the table dated in the
	// month.
	HasIncomes  bool
	HasExpenses bool
}

// ViewStore is the estimate_views table (M8-02), under the caller's RLS.
type ViewStore interface {
	// Views returns the caller's views of the taxpayer's months in
	// [from, to], in one statement (the primary key's range scan).
	Views(ctx context.Context, taxpayerID string, from, to domain.Month) ([]domain.View, error)
	// SaveView upserts the caller's view of one month (INSERT … ON
	// CONFLICT DO UPDATE: the last view wins). A taxpayer that is not the
	// caller's is ErrTaxpayerNotFound (the composite foreign key).
	SaveView(ctx context.Context, userID, taxpayerID string, view domain.View) error
}
