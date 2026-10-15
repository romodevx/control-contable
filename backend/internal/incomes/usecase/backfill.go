package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// BackfillStore is what the withholdings backfill (WH-08, MODULE-withholdings
// §5.2, §5.3) needs from persistence, on the unit of work in ctx as the
// caller (RLS).
type BackfillStore interface {
	// Underived returns the caller's incomes, soft-deleted ones included
	// (a restore must find their facts), whose facts or stored derivation
	// are missing, or whose derivation comes from another table version
	// than version, ordered by id. LockUpdate locks them.
	Underived(ctx context.Context, version string, lock Lock) ([]StoredIncome, error)
	// SaveDerivation writes only the facts and the derived withholding of
	// income id: never an amount, a date or a description.
	SaveDerivation(ctx context.Context, id string, f Fields, w fiscal.WithholdingRecord) error
	// IncomeBasis returns the catalog's income basis of a regime.
	IncomeBasis(ctx context.Context, person fiscal.PersonType, regime fiscal.RegimeCode) (domain.IncomeBasis, error)
}

// TaxpayerProfiles reads the caller's taxpayer, archived ones included.
type TaxpayerProfiles interface {
	Profile(ctx context.Context, userID, taxpayerID string) (TaxpayerProfile, error)
}

// BackfillDeps are the ports of the backfill. Tx must be the manager the
// caller's per-user unit of work runs on, so the backfill joins it.
type BackfillDeps struct {
	Tx           tx.Manager
	Store        BackfillStore
	Taxpayers    TaxpayerProfiles
	Withholdings fiscal.WithholdingCatalog
	Clock        clock.Clock
}

// Backfill re-derives the stored withholdings of incomes written before
// automatic derivation (WH-06) or under an older table version, with the
// same domain rules a write runs (ADR-0008).
type Backfill struct {
	deps BackfillDeps
}

// NewBackfill returns the backfill over deps.
func NewBackfill(deps BackfillDeps) *Backfill { return &Backfill{deps: deps} }

// Rederived is one income the backfill derives (or would, in a dry run).
type Rederived struct {
	ID string
	// AmountsChanged: the derived ISR or IVA withheld differs from the
	// stored one.
	AmountsChanged bool
	// FactsAssumed: the income had no facts; they were read from its
	// deprecated counterparty_type and the regime's defaults (§5.2), so the
	// user should review them.
	FactsAssumed bool
}

// BackfillResult is the outcome of one user's backfill: ids only, never an
// amount or a description.
type BackfillResult struct {
	Rederived []Rederived
	// NeedsInput are the incomes the rules refuse with the facts they
	// hold (a regime no longer held, a date out of range, no
	// counterparty): left as stored for the user to fix.
	NeedsInput []string
}

// Run backfills userID's incomes. A dry run (apply false) reads in a
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
		return BackfillResult{}, fmt.Errorf("incomes backfill: %w", err)
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
	today, err := fiscal.ParseDate(b.deps.Clock.Now().In(clock.Location()).Format(time.DateOnly))
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
		f, w, assumed, ok, err := b.rederive(ctx, row, profile, today)
		if err != nil {
			return BackfillResult{}, err
		}
		if !ok {
			out.NeedsInput = append(out.NeedsInput, row.ID)
			continue
		}
		if apply {
			if err := b.deps.Store.SaveDerivation(ctx, row.ID, f, w); err != nil {
				return BackfillResult{}, err
			}
		}
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

// rederive resolves row's facts and derives its withholding. ok is false
// when the rules refuse the facts; err is an infrastructure or parameter
// fault, which aborts the backfill.
func (b *Backfill) rederive(ctx context.Context, row StoredIncome, profile TaxpayerProfile, today fiscal.Date) (
	f Fields, w fiscal.WithholdingRecord, assumed, ok bool, err error,
) {
	f, assumed, ok = legacyFacts(row, profile.PersonType)
	held, holds := findRegime(profile, f.Regime)
	if !ok || !holds {
		return Fields{}, fiscal.WithholdingRecord{}, false, false, nil
	}
	basis, err := b.deps.Store.IncomeBasis(ctx, profile.PersonType, held.Code)
	if err != nil {
		return Fields{}, fiscal.WithholdingRecord{}, false, false, err
	}
	a, err := domain.Assess(draftOf(f), newPayee(profile.PersonType, held, basis), today, b.deps.Withholdings)
	if _, refused := errors.AsType[fiscal.Violations](err); refused {
		return Fields{}, fiscal.WithholdingRecord{}, false, false, nil
	}
	if err != nil {
		return Fields{}, fiscal.WithholdingRecord{}, false, false, err
	}
	return f, a.Withholding, assumed, true, nil
}

// legacyFacts are row's facts with the ones an income written before
// WH-06 lacks filled per MODULE-withholdings §5.2: the counterparty from
// counterparty_type, the operation from the regime's default, and a 606
// exempt rent as housing. assumed reports any fill; ok is false when
// counterparty_type cannot be read either.
func legacyFacts(row StoredIncome, person fiscal.PersonType) (f Fields, assumed, ok bool) {
	f = row.Fields
	if f.Counterparty == "" {
		kind, via, known := domain.LegacyFacts(row.CounterpartyType)
		if !known {
			return Fields{}, false, false
		}
		f.Counterparty, f.ViaPlatform, assumed = kind, via, true
	}
	if f.Operation == "" {
		f.Operation, assumed = domain.DefaultOperation(person, f.Regime), true
		f.ResidentialRent = f.Operation == fiscal.OperationRentRealEstate && f.Regime == fiscal.RegimeRent &&
			f.IVARate == fiscal.IVAExempt
	}
	return f, assumed, true
}
