package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// BackfillStore is what the withholdings backfill (WH-08, MODULE-withholdings
// §5.2, §5.3) needs from persistence, on the unit of work in ctx as the
// caller (RLS).
type BackfillStore interface {
	// Underived returns the caller's expenses, soft-deleted ones included
	// (a restore must find their supplier facts), without a complete stored
	// derivation or derived under another table version than version,
	// ordered by id. LockUpdate locks them.
	Underived(ctx context.Context, version string, lock Lock) ([]StoredExpense, error)
	// SaveDerivation writes only e's supplier facts, derived withholding
	// and creditable IVA split: never an amount, a date, the deduction
	// verdict or the description.
	SaveDerivation(ctx context.Context, e StoredExpense) error
}

// BackfillDeps are the ports of the backfill. Tx must be the manager the
// caller's per-user unit of work runs on, so the backfill joins it.
type BackfillDeps struct {
	Tx           tx.Manager
	Store        BackfillStore
	Taxpayers    TaxpayerProfiles
	Categories   CategoryCatalog
	Parameters   domain.RateCatalog
	Withholdings fiscal.WithholdingCatalog
	Clock        clock.Clock
}

// Backfill re-derives the stored withholdings of expenses written before
// automatic derivation (WH-07) or under an older table version, with the
// same domain rules a write runs (ADR-0008).
type Backfill struct {
	deps BackfillDeps
}

// NewBackfill returns the backfill over deps.
func NewBackfill(deps BackfillDeps) *Backfill { return &Backfill{deps: deps} }

// Rederived is one expense the backfill derives (or would, in a dry run).
type Rederived struct {
	ID string
	// AmountsChanged: the derived ISR or IVA withheld differs from the
	// stored one.
	AmountsChanged bool
	// FactsAssumed: a persona moral's expense had no supplier facts; they
	// were assumed per §5.2, so the user should review them.
	FactsAssumed bool
}

// BackfillResult is the outcome of one user's backfill: ids only, never an
// amount or a description.
type BackfillResult struct {
	Rederived []Rederived
	// NeedsInput are the expenses the rules refuse with the facts they
	// hold (an unknown category, a date out of range): left as stored for
	// the user to fix.
	NeedsInput []string
}

// Run backfills userID's expenses. A dry run (apply false) reads in a
// read-only unit of work and writes nothing; apply locks the rows and
// stores each derivation. Run is idempotent: a derived row is not
// underived any more, so a second run finds nothing to do.
func (b *Backfill) Run(ctx context.Context, userID string, apply bool) (BackfillResult, error) {
	var out BackfillResult
	work := func(ctx context.Context) error {
		var err error
		out, err = b.run(ctx, userID, apply)
		return err
	}
	var err error
	if apply {
		err = b.deps.Tx.ReadWrite(ctx, userID, work)
	} else {
		err = b.deps.Tx.ReadOnly(ctx, userID, work)
	}
	if err != nil {
		return BackfillResult{}, fmt.Errorf("expenses backfill: %w", err)
	}
	return out, nil
}

func (b *Backfill) run(ctx context.Context, userID string, apply bool) (BackfillResult, error) {
	lock := LockNone
	if apply {
		lock = LockUpdate
	}
	rows, err := b.deps.Store.Underived(ctx, fiscal.WithholdingTableVersion, lock)
	if err != nil {
		return BackfillResult{}, err
	}
	today, err := fiscal.ParseDate(b.deps.Clock.Now().Format(dayLayout))
	if err != nil {
		return BackfillResult{}, err
	}
	profiles := map[string]TaxpayerProfile{}
	var out BackfillResult
	for _, row := range rows {
		profile, err := b.profile(ctx, userID, row.TaxpayerID, profiles)
		if err != nil {
			return BackfillResult{}, err
		}
		derived, assumed, ok, err := b.rederive(ctx, row, profile, today)
		if err != nil {
			return BackfillResult{}, err
		}
		if !ok {
			out.NeedsInput = append(out.NeedsInput, row.ID)
			continue
		}
		if apply {
			if err := b.deps.Store.SaveDerivation(ctx, derived); err != nil {
				return BackfillResult{}, err
			}
		}
		w := derived.Withholding
		out.Rederived = append(out.Rederived, Rederived{
			ID: row.ID, AmountsChanged: w.ISR != row.Withholding.ISR || w.IVA != row.Withholding.IVA, FactsAssumed: assumed,
		})
	}
	return out, nil
}

// profile reads a taxpayer once per run.
func (b *Backfill) profile(ctx context.Context, userID, taxpayerID string, cache map[string]TaxpayerProfile) (TaxpayerProfile, error) {
	if p, ok := cache[taxpayerID]; ok {
		return p, nil
	}
	p, err := b.deps.Taxpayers.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return TaxpayerProfile{}, err
	}
	cache[taxpayerID] = p
	return p, nil
}

// rederive assumes the supplier facts a persona moral's legacy expense
// lacks and derives its withholding with the category's rule. Only the
// supplier, the withholding and the creditable IVA split change: the
// deduction verdict stored at the write is kept, so a past month is not
// rewritten. ok is false when the rules refuse the facts; err aborts.
func (b *Backfill) rederive(ctx context.Context, e StoredExpense, profile TaxpayerProfile, today fiscal.Date) (
	derived StoredExpense, assumed, ok bool, err error,
) {
	draft := draftOf(e)
	if profile.PersonType == fiscal.PersonMoral && draft.Supplier.Kind == "" {
		draft.Supplier, assumed = domain.LegacySupplier(e.Withholding.ISR+e.Withholding.IVA), true
	}
	entry, err := b.deps.Categories.Rule(ctx, e.Regime, profile.PersonType, e.Category)
	if errors.Is(err, ErrCategoryNotFound) {
		return StoredExpense{}, false, false, nil
	}
	if err != nil {
		return StoredExpense{}, false, false, err
	}
	payer := domain.Payer{PersonType: profile.PersonType, Regime: e.Regime}
	if hold, held := profile.holding(e.Regime, e.PaidOn.Period()); held {
		payer = profile.payer(hold)
	}
	a, err := domain.Assess(draft, payer, entry.Rule, today, b.deps.Parameters, b.deps.Withholdings)
	if _, refused := errors.AsType[fiscal.Violations](err); refused {
		return StoredExpense{}, false, false, nil
	}
	if err != nil {
		return StoredExpense{}, false, false, err
	}
	creditable := e.CreditableIVA+e.CreditableIVAWithheld > 0
	e.Supplier, e.Withholding = a.Supplier, a.Withholding
	e.CreditableIVA, e.CreditableIVAWithheld = domain.CreditableSplit(creditable, e.IVA, a.Withholding.IVA)
	return e, assumed, true, nil
}
