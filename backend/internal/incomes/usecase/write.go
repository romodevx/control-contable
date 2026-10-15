package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// createAttempts bounds the units of work of one create: a primary-key
// collision is retried once, so a create that still meets a committed row of
// its id (the LockID wait is the primary guard) sees it and replays it.
const createAttempts = 2

// Create records an income (createIncome). Created is false for an
// idempotent retry (200), which never takes a trial slot nor checks access.
func (uc *Incomes) Create(ctx context.Context, userID string, in CreateInput) (WriteResult, error) {
	taxpayerID, ok := canonicalID(in.TaxpayerID)
	if !ok {
		return WriteResult{}, taxpayerNotFound()
	}
	id, err := uc.clientID(in.ID)
	if err != nil {
		return WriteResult{}, present("create income", err)
	}
	req, err := normalize(in.Income)
	if err != nil {
		return WriteResult{}, present("create income", err)
	}
	fromClient := in.ID != ""
	var res WriteResult
	for attempt := 1; ; attempt++ {
		err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
			var err error
			res, err = uc.create(ctx, userID, id, taxpayerID, req, fromClient)
			return err
		})
		if !errors.Is(err, ErrIDTaken) || attempt == createAttempts {
			break
		}
	}
	if err != nil {
		return WriteResult{}, present("create income", err)
	}
	return res, nil
}

// clientID returns the canonical client id, or a new one when it is absent.
func (uc *Incomes) clientID(raw string) (string, error) {
	if raw == "" {
		return uc.deps.NewID()
	}
	id, ok := canonicalID(raw)
	if !ok {
		return "", fieldError(ErrInvalidID, fieldID, fiscal.ErrInvalid, msgInvalidField)
	}
	return id, nil
}

// create is one attempt, in the contract's order of checks. A client id is
// locked first: a double submit in flight waits for the first and replays
// it, never reading the trial counter the first just filled.
func (uc *Incomes) create(ctx context.Context, userID, id, taxpayerID string, req request, fromClient bool) (WriteResult, error) {
	if fromClient {
		if err := uc.deps.Store.LockID(ctx, id); err != nil {
			return WriteResult{}, err
		}
	}
	existing, err := uc.deps.Store.Find(ctx, id, LockNone)
	if err == nil {
		return uc.replay(ctx, userID, existing, taxpayerID, req)
	}
	if !errors.Is(err, ErrIncomeNotFound) {
		return WriteResult{}, err
	}
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return WriteResult{}, err
	}
	ev, err := uc.evaluate(ctx, userID, taxpayerID, id, req, true)
	if err != nil {
		return WriteResult{}, err
	}
	if err := uc.deps.Access.ReserveRecord(ctx, userID, access.RecordIncome); err != nil {
		return WriteResult{}, err
	}
	income := NewIncome{ID: id, UserID: userID, TaxpayerID: taxpayerID, Fields: ev.fields, Withholding: ev.withholding}
	if err := uc.deps.Store.Insert(ctx, income); err != nil {
		return WriteResult{}, err
	}
	stored, err := uc.deps.Store.Find(ctx, id, LockNone)
	if err != nil {
		return WriteResult{}, err
	}
	uc.audit(ctx, "income.created", userID, taxpayerID, id)
	return WriteResult{Income: toIncome(stored), Warnings: ev.warnings, Created: true}, nil
}

// replay answers a create whose id the caller already holds: the stored
// income with the warnings of its write when the retry carries the same
// facts for the same taxpayer and the income is live, else ErrIDReused.
func (uc *Incomes) replay(ctx context.Context, userID string, existing StoredIncome, taxpayerID string, req request) (WriteResult, error) {
	if existing.TaxpayerID != taxpayerID || existing.Deleted || !sameFields(existing.Fields, req) {
		return WriteResult{}, ErrIDReused
	}
	ev, err := uc.evaluate(ctx, userID, taxpayerID, existing.ID, stored(existing.Fields), false)
	switch {
	case isClientError(err):
		// The write happened; a later change (an archive) only drops the
		// warnings of the replay, never the answer.
		ev.warnings = nil
	case err != nil:
		return WriteResult{}, err
	}
	return WriteResult{Income: toIncome(existing), Warnings: ev.warnings}, nil
}

// Update edits a live income (updateIncome, merge-patch). It re-runs every
// rule of a create but the trial slot.
func (uc *Incomes) Update(ctx context.Context, userID, rawTaxpayerID, rawIncomeID string, p Patch) (WriteResult, error) {
	taxpayerID, id, err := pathIDs(rawTaxpayerID, rawIncomeID)
	if err != nil {
		return WriteResult{}, err
	}
	var res WriteResult
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.findLive(ctx, taxpayerID, id, LockUpdate)
		if err != nil {
			return err
		}
		if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
			return err
		}
		req, err := normalize(p.apply(stored.Fields))
		if err != nil {
			return err
		}
		ev, err := uc.evaluate(ctx, userID, taxpayerID, id, req, true)
		if err != nil {
			return err
		}
		if err := uc.deps.Store.Update(ctx, id, ev.fields, ev.withholding); err != nil {
			return err
		}
		updated, err := uc.deps.Store.Find(ctx, id, LockNone)
		if err != nil {
			return err
		}
		uc.audit(ctx, "income.updated", userID, taxpayerID, id)
		res = WriteResult{Income: toIncome(updated), Warnings: ev.warnings}
		return nil
	})
	if err != nil {
		return WriteResult{}, present("update income", err)
	}
	return res, nil
}

// evaluation is a checked income: the facts to store (the defaults and the
// 625 activity resolved), the derived withholding and the write's warnings.
type evaluation struct {
	fields      Fields
	withholding fiscal.WithholdingRecord
	warnings    []domain.Warning
}

// evaluate applies MODULE-incomes §2.4–§2.8 and MODULE-withholdings §3 to
// req for the taxpayer, inside the unit of work: archive, regime window
// (when window, CheckRegime per distinct period), the defaults of the hidden
// questions, domain rules and derivation, parameter coverage of the dates,
// and the warnings. incomeID is left out of the duplicate-folio lookup.
func (uc *Incomes) evaluate(ctx context.Context, userID, taxpayerID, incomeID string, req request, window bool) (evaluation, error) {
	f := req.fields
	today, err := uc.today()
	if err != nil {
		return evaluation{}, err
	}
	profile, err := uc.deps.Taxpayers.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return evaluation{}, err
	}
	if profile.Archived {
		return evaluation{}, archivedError()
	}
	held, ok := findRegime(profile, f.Regime)
	if !ok {
		return evaluation{}, regimeNotHeldError()
	}
	if window {
		if err := uc.checkWindow(ctx, userID, taxpayerID, f, today); err != nil {
			return evaluation{}, err
		}
	}
	payee, err := uc.payee(ctx, profile.PersonType, held)
	if err != nil {
		return evaluation{}, err
	}
	if f, err = uc.resolveDefaults(ctx, taxpayerID, profile.PersonType, req); err != nil {
		return evaluation{}, err
	}
	return uc.assess(ctx, taxpayerID, incomeID, f, payee, today)
}

// payee is the domain's view of the taxpayer under a held regime.
func (uc *Incomes) payee(ctx context.Context, person fiscal.PersonType, held RegimeProfile) (domain.Payee, error) {
	basis, err := uc.deps.Store.IncomeBasis(ctx, person, held.Code)
	if err != nil {
		return domain.Payee{}, err
	}
	return newPayee(person, held, basis), nil
}

// newPayee is the domain's payee: the person under a held regime with its
// income basis and 625 options.
func newPayee(person fiscal.PersonType, held RegimeProfile, basis domain.IncomeBasis) domain.Payee {
	return domain.Payee{
		PersonType: person, Regime: held.Code, Basis: basis,
		PlatformActivities: held.PlatformActivities, RFCGivenToPlatform: held.RFCGivenToPlatform,
	}
}

// resolveDefaults answers the questions req left unanswered
// (MODULE-withholdings §2.1): via a platform in 625 only; the operation last
// used in the regime, else the regime's default.
func (uc *Incomes) resolveDefaults(ctx context.Context, taxpayerID string, person fiscal.PersonType, req request) (Fields, error) {
	f := req.fields
	f.ViaPlatform = domain.DefaultViaPlatform(f.Regime)
	if req.viaPlatform != nil {
		f.ViaPlatform = *req.viaPlatform
	}
	if f.Operation != "" {
		return f, nil
	}
	last, ok, err := uc.deps.Store.LastFacts(ctx, taxpayerID, f.Regime)
	if err != nil {
		return Fields{}, err
	}
	f.Operation = domain.DefaultOperation(person, f.Regime)
	if ok && last.Operation != "" {
		f.Operation = last.Operation
	}
	return f, nil
}

// assess runs the domain rules, then the parameter coverage: the domain's
// validation (400) comes first, then a date no parameter set covers (400
// out_of_range: before 2026 no rate can be checked), before the business
// rules that read its rates (422).
func (uc *Incomes) assess(ctx context.Context, taxpayerID, incomeID string, f Fields, payee domain.Payee, today fiscal.Date) (evaluation, error) {
	assessment, assessErr := domain.Assess(draftOf(f), payee, today, uc.deps.Withholdings)
	if assessErr != nil && !isBusinessRule(assessErr) {
		return evaluation{}, assessErr
	}
	provisional, err := uc.checkParameters(payee.Basis, f.Dates)
	if err != nil {
		return evaluation{}, err
	}
	if assessErr != nil {
		return evaluation{}, assessErr
	}
	f.PlatformActivity = assessment.PlatformActivity
	duplicate, err := uc.duplicateFolio(ctx, taxpayerID, f.CFDIUUID, incomeID)
	if err != nil {
		return evaluation{}, err
	}
	return evaluation{
		fields: f, withholding: assessment.Withholding,
		warnings: slices.Concat(assessment.Warnings, provisional, duplicate),
	}, nil
}

// checkWindow asks CheckRegime once per distinct period among the
// income's dates within [fiscal.MinRecordDate, today]; the others are the
// domain's out_of_range.
func (uc *Incomes) checkWindow(ctx context.Context, userID, taxpayerID string, f Fields, today fiscal.Date) error {
	var checked []string
	for _, on := range []fiscal.Date{f.Dates.InvoicedOn, f.Dates.CollectedOn} {
		if on.IsZero() || on.Before(fiscal.MinRecordDate) || on.After(today) || slices.Contains(checked, on.Period()) {
			continue
		}
		checked = append(checked, on.Period())
		if err := uc.deps.Taxpayers.CheckRegime(ctx, userID, taxpayerID, string(f.Regime), on.Period()); err != nil {
			return err
		}
	}
	return nil
}

// checkParameters checks that a parameter set covers the dates the rules
// read rates on (the booking date and the IVA date): an uncovered date is
// out_of_range, as for expenses (the contract's date code), a provisional
// one a warning.
func (uc *Incomes) checkParameters(basis domain.IncomeBasis, dates domain.Dates) ([]domain.Warning, error) {
	var vs fiscal.Violations
	var warnings []domain.Warning
	var seen []string
	for _, d := range ratedDates(basis, dates) {
		if d.on.IsZero() || slices.Contains(seen, d.field) {
			continue
		}
		seen = append(seen, d.field)
		provisional, err := uc.deps.Rates.Provisional(d.on)
		switch {
		case err != nil:
			vs.Add(d.field, fiscal.ErrOutOfRange)
		case provisional:
			warnings = append(warnings, domain.Warning{Code: domain.WarningProvisionalParameters, Field: d.field})
		}
	}
	if len(vs) > 0 {
		return nil, vs
	}
	return warnings, nil
}

// datedField is a date with the contract field that holds it.
type datedField struct {
	field string
	on    fiscal.Date
}

// ratedDates are the booking date (ISR suggestions) and the IVA date.
func ratedDates(basis domain.IncomeBasis, dates domain.Dates) []datedField {
	ivaField := domain.FieldCollectedOn
	if dates.CollectedOn.IsZero() {
		ivaField = domain.FieldInvoicedOn
	}
	return []datedField{
		{field: basis.RequiredDateField(), on: basis.BookingDate(dates)},
		{field: ivaField, on: dates.RateDate()},
	}
}

// duplicateFolio is the duplicate_cfdi_uuid warning when another live
// income of the taxpayer has the folio (OQ-9: never an error).
func (uc *Incomes) duplicateFolio(ctx context.Context, taxpayerID string, folio fiscal.CFDIUUID, incomeID string) ([]domain.Warning, error) {
	if folio.IsZero() {
		return nil, nil
	}
	inUse, err := uc.deps.Store.CFDIUUIDInUse(ctx, taxpayerID, folio, incomeID)
	if err != nil || !inUse {
		return nil, err
	}
	return []domain.Warning{{Code: domain.WarningDuplicateCFDIUUID, Field: domain.FieldCFDIUUID}}, nil
}

func findRegime(p TaxpayerProfile, code fiscal.RegimeCode) (RegimeProfile, bool) {
	for _, r := range p.Regimes {
		if r.Code == code {
			return r, true
		}
	}
	return RegimeProfile{}, false
}

func draftOf(f Fields) domain.Draft {
	return domain.Draft{
		Amount: f.Amount, IVARate: f.IVARate, IVA: f.IVA, Counterparty: f.Counterparty, Operation: f.Operation,
		ResidentialRent: f.ResidentialRent, ViaPlatform: f.ViaPlatform, Dates: f.Dates, PlatformActivity: f.PlatformActivity,
	}
}

// isBusinessRule reports business-rule violations (422), the only domain
// failures the parameter check may precede.
func isBusinessRule(err error) bool {
	vs, ok := errors.AsType[fiscal.Violations](err)
	return ok && vs.Class() == fiscal.ClassBusinessRule
}

// isClientError reports a refusal the client caused (an *apperr.Error or
// domain violations), as opposed to an infrastructure failure.
func isClientError(err error) bool {
	if _, ok := errors.AsType[*apperr.Error](err); ok {
		return true
	}
	_, ok := errors.AsType[fiscal.Violations](err)
	return ok
}
