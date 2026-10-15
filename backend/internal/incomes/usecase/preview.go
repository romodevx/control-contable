package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// PreviewQuery is previewIncomeWithholdings: the facts of an income being
// typed, by query parameter. On is "YYYY-MM-DD" ("" is today); an empty
// Operation, a nil ViaPlatform and an empty PlatformActivity take the
// defaults a create would store.
type PreviewQuery struct {
	TaxpayerID       string
	Regime           fiscal.RegimeCode
	On               string
	Amount           money.Cents
	IVARate          fiscal.IVARate
	IVA              money.Cents
	Counterparty     fiscal.ReceptorKind
	Operation        fiscal.OperationKind
	ResidentialRent  bool
	ViaPlatform      *bool
	PlatformActivity domain.PlatformActivity
}

// Preview is what createIncome would derive for the same facts (contract
// WithholdingPreview): the record, the facts its explanation needs and the
// withholding warnings only.
type Preview struct {
	Withholding      fiscal.WithholdingRecord
	Amount           money.Cents
	IVA              money.Cents
	Counterparty     fiscal.ReceptorKind
	PlatformActivity domain.PlatformActivity
	Warnings         []domain.Warning
}

// Preview answers previewIncomeWithholdings (ADR-0008 §4): the same rules
// and derivation as a create on the day On, in a read-only unit of work that
// stores nothing and takes no lock. A read: allowed for a read-only user and
// an archived taxpayer. The regime must be held on the day (422
// regime_not_active on regime).
func (uc *Incomes) Preview(ctx context.Context, userID string, q PreviewQuery) (Preview, error) {
	taxpayerID, ok := canonicalID(q.TaxpayerID)
	if !ok {
		return Preview{}, taxpayerNotFound()
	}
	today, err := uc.today()
	if err != nil {
		return Preview{}, err
	}
	on, err := uc.previewDay(q, today)
	if err != nil {
		return Preview{}, present("preview income withholdings", err)
	}
	var out Preview
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		out, err = uc.preview(ctx, userID, taxpayerID, q, on, today)
		return err
	})
	if err != nil {
		return Preview{}, present("preview income withholdings", err)
	}
	return out, nil
}

// previewDay validates the regime and the day ("" is today): within
// [fiscal.MinRecordDate, today] and covered by a parameter set.
func (uc *Incomes) previewDay(q PreviewQuery, today fiscal.Date) (fiscal.Date, error) {
	var vs fiscal.Violations
	if q.Regime == "" {
		vs.Add(fieldRegime, fiscal.ErrRequired)
	}
	on, refusal, ok := uc.parseDefaultsDay(q.On, today)
	if !ok {
		vs.Add(fieldOn, refusal)
	}
	if len(vs) > 0 {
		return fiscal.Date{}, vs
	}
	return on, nil
}

func (uc *Incomes) preview(ctx context.Context, userID, taxpayerID string, q PreviewQuery, on, today fiscal.Date) (Preview, error) {
	profile, err := uc.deps.Taxpayers.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return Preview{}, err
	}
	held, ok := holding(profile, q.Regime, on.Period())
	if !ok {
		return Preview{}, regimeNotHeldOn(fieldRegime)
	}
	payee, err := uc.payee(ctx, profile.PersonType, held)
	if err != nil {
		return Preview{}, err
	}
	f, err := uc.resolveDefaults(ctx, taxpayerID, profile.PersonType, request{
		fields: Fields{
			Regime: q.Regime, Counterparty: q.Counterparty, Operation: q.Operation, ResidentialRent: q.ResidentialRent,
			Amount: q.Amount, IVARate: q.IVARate, IVA: q.IVA, PlatformActivity: q.PlatformActivity,
			// The day is the withholding date and, for either basis, the
			// booking date: both dates the rules need are set to it.
			Dates: domain.Dates{InvoicedOn: on, CollectedOn: on},
		},
		viaPlatform: q.ViaPlatform,
	})
	if err != nil {
		return Preview{}, err
	}
	assessment, err := domain.Assess(draftOf(f), payee, today, uc.deps.Withholdings)
	if err != nil {
		return Preview{}, err
	}
	return Preview{
		Withholding: assessment.Withholding, Amount: f.Amount, IVA: f.IVA, Counterparty: f.Counterparty,
		PlatformActivity: assessment.PlatformActivity, Warnings: assessment.Warnings,
	}, nil
}
