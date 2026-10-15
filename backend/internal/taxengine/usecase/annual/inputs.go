package annual

import (
	"context"
	"log/slog"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// eventInputsUpdated is the audit event of putAnnualInputs (ids only).
const eventInputsUpdated = "annual_inputs.updated"

// Contract field names of the annual facts.
const (
	FieldPriorYearIncome   = "prior_year_accumulable_income_cents"
	FieldOtherExemptIncome = "other_exempt_income_cents"
)

// RegimeTaxInput is the read-only mirror of one annual-relevant M5 money
// tax input of a regime held in the year (contract AnnualRegimeTaxInput);
// a nil Value was not entered.
type RegimeTaxInput struct {
	Regime fiscal.RegimeCode
	Key    domain.InputKey
	Label  string
	Value  *money.Cents
}

// InputsView is getAnnualInputs: the year's facts (the defaults when never
// saved, UpdatedAt nil) and the mirrored tax inputs, by regime then the
// catalog order of keys.
type InputsView struct {
	FiscalYear      int
	Facts           Facts
	RegimeTaxInputs []RegimeTaxInput
	UpdatedAt       *time.Time
}

// InputsDeps are the ports the annual-facts use cases need.
type InputsDeps struct {
	Tx        tx.Manager
	Access    AccessPolicy
	Taxpayers TaxpayerStore
	Store     FactsStore
	Clock     clock.Clock
	// Logger receives the audit events; nil discards them.
	Logger *slog.Logger
}

// Inputs reads and replaces the per-year annual facts (getAnnualInputs,
// putAnnualInputs; MODULE-annual §2.6).
type Inputs struct {
	base
	deps InputsDeps
}

// NewInputs returns the use cases over deps.
func NewInputs(deps InputsDeps) *Inputs {
	return &Inputs{base: newBase(deps.Clock, deps.Logger), deps: deps}
}

// Get returns the facts of the year and the mirrored tax inputs. Order:
// taxpayer id (404) → future year (400) → taxpayer (404) → year floor
// (422). Open to any person type, a read-only user and an archived
// taxpayer.
func (uc *Inputs) Get(ctx context.Context, userID, rawTaxpayerID string, fiscalYear int) (InputsView, error) {
	tid, err := uc.resolve(rawTaxpayerID, fiscalYear)
	if err != nil {
		return InputsView{}, err
	}
	var out InputsView
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		if _, err := uc.deps.Taxpayers.Taxpayer(ctx, tid, LockNone); err != nil {
			return err
		}
		if err := checkYearFloor(fiscalYear, FieldFiscalYearPath); err != nil {
			return err
		}
		out, err = uc.view(ctx, tid, fiscalYear)
		return err
	})
	if err != nil {
		return InputsView{}, present("get annual inputs", err)
	}
	return out, nil
}

// Put replaces the facts of the year (putAnnualInputs, idempotent). Order:
// taxpayer id (404) → future year (400) → read-only (402) → taxpayer
// (404), archived, person type, year floor (422) → amounts (400) → upsert.
func (uc *Inputs) Put(ctx context.Context, userID, rawTaxpayerID string, fiscalYear int, f Facts) (InputsView, error) {
	tid, err := uc.resolve(rawTaxpayerID, fiscalYear)
	if err != nil {
		return InputsView{}, err
	}
	var out InputsView
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
			return err
		}
		if _, err := ensureWritable(ctx, uc.deps.Taxpayers, tid, true); err != nil {
			return err
		}
		if err := checkYearFloor(fiscalYear, FieldFiscalYearPath); err != nil {
			return err
		}
		if err := checkFacts(f); err != nil {
			return err
		}
		if err := uc.deps.Store.UpsertFacts(ctx, userID, tid, fiscalYear, f); err != nil {
			return err
		}
		out, err = uc.view(ctx, tid, fiscalYear)
		return err
	})
	if err != nil {
		return InputsView{}, present("put annual inputs", err)
	}
	uc.audit(ctx, eventInputsUpdated, slog.String("user_id", userID), slog.String("taxpayer_id", tid),
		slog.Int("fiscal_year", fiscalYear))
	return out, nil
}

func (uc *Inputs) resolve(rawTaxpayerID string, fiscalYear int) (string, error) {
	tid, err := taxpayerID(rawTaxpayerID)
	if err != nil {
		return "", err
	}
	return tid, uc.checkFutureYear(fiscalYear, FieldFiscalYearPath)
}

// checkFacts re-checks the schema's minimums (400).
func checkFacts(f Facts) error {
	var fields []apperr.FieldError
	if f.PriorYearAccumulableIncome != nil && *f.PriorYearAccumulableIncome < 0 {
		fields = append(fields, apperr.FieldError{Field: FieldPriorYearIncome, Code: DetailInvalid, Message: msgInvalidValue})
	}
	if f.OtherExemptIncome < 0 {
		fields = append(fields, apperr.FieldError{Field: FieldOtherExemptIncome, Code: DetailInvalid, Message: msgInvalidValue})
	}
	if len(fields) > 0 {
		return fieldErrors(fields)
	}
	return nil
}

// view reads the facts and the mirror of the year.
func (uc *Inputs) view(ctx context.Context, taxpayerID string, fiscalYear int) (InputsView, error) {
	out := InputsView{FiscalYear: fiscalYear}
	stored, found, err := uc.deps.Store.Facts(ctx, taxpayerID, fiscalYear)
	if err != nil {
		return InputsView{}, err
	}
	if found {
		out.Facts, out.UpdatedAt = stored.Facts, &stored.UpdatedAt
	}
	regimes, err := uc.deps.Taxpayers.Regimes(ctx, taxpayerID)
	if err != nil {
		return InputsView{}, err
	}
	values, err := uc.deps.Store.TaxInputMoney(ctx, taxpayerID, fiscalYear, adomain.TaxInputKeys())
	if err != nil {
		return InputsView{}, err
	}
	out.RegimeTaxInputs = mirror(regimes, fiscalYear, values)
	return out, nil
}

// mirror lists the money annual tax inputs each regime held in the year
// defines, in the catalog order, with the stored value when there is one.
func mirror(regimes []HeldRegime, fiscalYear int, values map[fiscal.RegimeCode]map[domain.InputKey]money.Cents) []RegimeTaxInput {
	out := []RegimeTaxInput{}
	for _, r := range regimes {
		if !r.ActiveIn(fiscalYear) {
			continue
		}
		for _, def := range r.Definitions {
			if def.Kind != domain.InputMoney || !adomain.IsTaxInput(def.Key) {
				continue
			}
			in := RegimeTaxInput{Regime: r.Code, Key: def.Key, Label: def.Label}
			if v, ok := values[r.Code][def.Key]; ok {
				in.Value = &v
			}
			out = append(out, in)
		}
	}
	return out
}
