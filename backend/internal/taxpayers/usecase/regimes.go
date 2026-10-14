package usecase

import (
	"context"
	"errors"
	"log/slog"
	"maps"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Audit events of regime assignment (S-9).
const (
	eventRegimeAdded    = "taxpayer_regime.added"
	eventRegimeArchived = "taxpayer_regime.archived"
)

// Regimes assigns regimes to a taxpayer (US-T3, US-T3b): add or reopen,
// change options, end. Every write locks the taxpayer row first, so the
// regime writes of one taxpayer run one at a time and their compatibility
// checks see each other's results.
type Regimes struct {
	service
}

// NewRegimes returns the use case over deps.
func NewRegimes(deps Deps) *Regimes { return &Regimes{service: newService(deps)} }

// Add adds a regime to a taxpayer, or reopens one that ended
// (addTaxpayerRegime). Order: access policy (402) → taxpayer (404, 422
// archived) → catalog, options and start (400) → already active (409) →
// compatibility with every overlapping regime (422, or a warning for a
// pair under CPA review).
//
// Reopening (OQ-5) clears end_period and applies the body options; the
// row's original start_period is kept and the body's start is ignored, so
// compatibility is checked on the window the row will really have. Sending
// the body start to the database instead could raise a false overlap
// (23P01) for a start earlier than the stored one (PR #57 follow-up).
func (uc *Regimes) Add(ctx context.Context, userID, rawTaxpayerID string, in RegimeInput) (TaxpayerRegime, error) {
	id, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return TaxpayerRegime{}, notFound()
	}
	var out TaxpayerRegime
	err := uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		var err error
		out, err = uc.add(ctx, userID, id, in)
		return err
	})
	if err != nil {
		return TaxpayerRegime{}, present("add regime", err)
	}
	uc.audit(ctx, eventRegimeAdded, userID, id, slog.String("regime_code", out.Code))
	return out, nil
}

func (uc *Regimes) add(ctx context.Context, userID, id string, in RegimeInput) (TaxpayerRegime, error) {
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return TaxpayerRegime{}, err
	}
	t, err := uc.findWritable(ctx, id)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	fields := &fieldErrors{}
	regime, ok, err := uc.validateRegime(ctx, t.PersonType, in, "", fields)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	if !ok {
		return TaxpayerRegime{}, fields.err()
	}
	regime.TaxpayerID, regime.UserID = id, userID
	stored, err := uc.deps.Regimes.FindRegime(ctx, id, in.Code, LockUpdate)
	reopen := err == nil
	switch {
	case reopen && stored.Window.IsOpen():
		return TaxpayerRegime{}, ErrRegimeAlreadyActive
	case reopen:
		regime.Start = stored.Window.Start
	case !errors.Is(err, ErrRegimeNotHeld):
		return TaxpayerRegime{}, err
	}
	warnings, err := uc.checkAgainstHeld(ctx, t.PersonType, regime.Code, openFrom(regime.Start), id)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	regime.NeedsCPAReviewAck = len(warnings) > 0
	if reopen {
		err = uc.deps.Regimes.ReopenRegime(ctx, id, regime.Code, regime.Options, regime.NeedsCPAReviewAck)
	} else {
		err = uc.deps.Regimes.InsertRegime(ctx, regime)
	}
	if err != nil {
		return TaxpayerRegime{}, err
	}
	return uc.view(ctx, id, regime.Code, warnings)
}

// checkAgainstHeld checks code over w against every other regime the
// taxpayer holds whose window overlaps w.
func (uc *Regimes) checkAgainstHeld(ctx context.Context, personType domain.PersonType, code string, w domain.Window, taxpayerID string) ([]Warning, error) {
	held, err := uc.deps.Regimes.Regimes(ctx, taxpayerID)
	if err != nil {
		return nil, err
	}
	codes := []string{code}
	for _, h := range held {
		codes = append(codes, h.Code)
	}
	matrix, err := uc.deps.Catalog.Compatibility(ctx, personType, codes)
	if err != nil {
		return nil, err
	}
	var warnings []Warning
	for _, h := range held {
		if h.Code == code || !w.Overlaps(h.Window) {
			continue
		}
		warning, err := verdict(matrix, fieldRegimeCode, code, h.Code)
		if err != nil {
			return nil, err
		}
		if warning != nil {
			warnings = append(warnings, *warning)
		}
	}
	return warnings, nil
}

// view reads one regime for a response.
func (uc *Regimes) view(ctx context.Context, taxpayerID, code string, warnings []Warning) (TaxpayerRegime, error) {
	held, err := uc.deps.Regimes.FindRegime(ctx, taxpayerID, code, LockNone)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	out := toRegime(held)
	out.Warnings = warnings
	return out, nil
}

// UpdateOptions changes a regime's options with merge semantics
// (updateTaxpayerRegime): each key sent replaces its value, the others keep
// theirs, and the merged set is validated against the catalog. Not
// versioned in time (OQ-7): the change applies to every period.
func (uc *Regimes) UpdateOptions(ctx context.Context, userID, rawTaxpayerID, code string, patch map[string]domain.OptionValue) (TaxpayerRegime, error) {
	id, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return TaxpayerRegime{}, notFound()
	}
	var out TaxpayerRegime
	err := uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		var err error
		out, err = uc.updateOptions(ctx, userID, id, code, patch)
		return err
	})
	if err != nil {
		return TaxpayerRegime{}, present("update regime options", err)
	}
	return out, nil
}

func (uc *Regimes) updateOptions(ctx context.Context, userID, id, code string, patch map[string]domain.OptionValue) (TaxpayerRegime, error) {
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return TaxpayerRegime{}, err
	}
	t, err := uc.findWritable(ctx, id)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	held, err := uc.deps.Regimes.FindRegime(ctx, id, code, LockUpdate)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	catalog, err := uc.deps.Catalog.Regime(ctx, t.PersonType, code)
	if err != nil {
		// The composite foreign key keeps a held regime in the catalog.
		return TaxpayerRegime{}, errors.Join(errors.New("taxpayers: held regime missing from the catalog"), err)
	}
	merged := maps.Clone(held.Options)
	if merged == nil {
		merged = make(map[string]domain.OptionValue, len(patch))
	}
	maps.Copy(merged, patch)
	options, err := catalog.Options.Validate(merged)
	if err != nil {
		fields := &fieldErrors{}
		if err := fields.addOptions(fieldOptions, err); err != nil {
			return TaxpayerRegime{}, err
		}
		return TaxpayerRegime{}, fields.err()
	}
	if err := uc.deps.Regimes.UpdateOptions(ctx, id, code, options); err != nil {
		return TaxpayerRegime{}, err
	}
	return uc.view(ctx, id, code, nil)
}

// End ends a regime at in.End (archiveTaxpayerRegime): records are kept.
// Allowed for a read-only user (an archive). Order: taxpayer (404, 422
// archived) → regime (404, 409 already ended) → end period range (400) →
// last active regime (422) → records after the end (422) → records without
// confirm (409).
func (uc *Regimes) End(ctx context.Context, userID, rawTaxpayerID string, in EndRegimeInput) (TaxpayerRegime, error) {
	id, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return TaxpayerRegime{}, notFound()
	}
	end, err := domain.ParsePeriod(in.End)
	if err != nil {
		fields := &fieldErrors{}
		fields.addPeriod(fieldEndPeriod, DetailInvalid, err)
		return TaxpayerRegime{}, fields.err()
	}
	var out TaxpayerRegime
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		var err error
		out, err = uc.end(ctx, id, in.Code, end, in.Confirm)
		return err
	})
	if err != nil {
		return TaxpayerRegime{}, present("end regime", err)
	}
	uc.audit(ctx, eventRegimeArchived, userID, id, slog.String("regime_code", out.Code))
	return out, nil
}

func (uc *Regimes) end(ctx context.Context, id, code string, end domain.Period, confirm bool) (TaxpayerRegime, error) {
	if _, err := uc.findWritable(ctx, id); err != nil {
		return TaxpayerRegime{}, err
	}
	held, err := uc.deps.Regimes.FindRegime(ctx, id, code, LockUpdate)
	if err != nil {
		return TaxpayerRegime{}, err
	}
	if !held.Window.IsOpen() {
		return TaxpayerRegime{}, ErrRegimeAlreadyEnded
	}
	if end.Before(held.Window.Start) || end.After(uc.currentPeriod()) {
		fields := &fieldErrors{}
		fields.addPeriod(fieldEndPeriod, DetailOutOfRange, ErrPeriodOutOfRange)
		return TaxpayerRegime{}, fields.err()
	}
	if err := uc.requireAnotherOpenRegime(ctx, id, code); err != nil {
		return TaxpayerRegime{}, err
	}
	if err := uc.checkRecords(ctx, id, code, end, confirm); err != nil {
		return TaxpayerRegime{}, err
	}
	if err := uc.deps.Regimes.EndRegime(ctx, id, code, end); err != nil {
		return TaxpayerRegime{}, err
	}
	return uc.view(ctx, id, code, nil)
}

// requireAnotherOpenRegime refuses to end the taxpayer's last open regime:
// archive the taxpayer instead (§2.7).
func (uc *Regimes) requireAnotherOpenRegime(ctx context.Context, id, code string) error {
	held, err := uc.deps.Regimes.Regimes(ctx, id)
	if err != nil {
		return err
	}
	for _, h := range held {
		if h.Code != code && h.Window.IsOpen() {
			return nil
		}
	}
	return businessRule(ErrLastActiveRegime, msgLastActive, fieldRegimeCodePath, DetailLastActiveRegime, msgLastActiveField)
}

// checkRecords keeps every record inside the window (§2.6) and makes the
// user confirm ending a regime that has records.
func (uc *Regimes) checkRecords(ctx context.Context, id, code string, end domain.Period, confirm bool) error {
	latest, err := uc.deps.Records.LatestRecordPeriod(ctx, id, code)
	switch {
	case err != nil:
		return err
	case latest.IsZero():
		return nil
	case latest.After(end):
		return businessRule(ErrRecordsAfterEndPeriod, msgRecordsAfter, fieldEndPeriod, DetailRecordsAfterEndPeriod, msgRecordsAfterField)
	case !confirm:
		return ErrRegimeHasRecords
	}
	return nil
}
