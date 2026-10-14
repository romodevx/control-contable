package usecase

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// firstParamsYear is the first fiscal year with transcribed tax parameters
// (OQ-T9): an earlier year is 422 parameters_unavailable.
const firstParamsYear = 2026

// eventInputsUpdated is the audit event of a replace (S-8): ids only.
const eventInputsUpdated = "tax_inputs.updated"

// InputsDeps are the ports the tax-inputs use cases need.
type InputsDeps struct {
	Tx     tx.Manager
	Access AccessPolicy
	Store  TaxInputStore
	Clock  clock.Clock
	// Logger receives the audit events (S-8); nil discards them.
	Logger *slog.Logger
}

// InputsRequest names the inputs of one taxpayer, regime and fiscal year as
// the client sent them (the ids are not validated yet).
type InputsRequest struct {
	TaxpayerID string
	RegimeCode string
	FiscalYear int
}

// TaxInputsView is a regime's tax inputs for a fiscal year (contract
// TaxInputs): the catalog definitions and the stored values, ordered by
// their key's position in Definitions, then month.
type TaxInputsView struct {
	Regime      fiscal.RegimeCode
	FiscalYear  int
	Definitions []domain.InputDef
	Values      []StoredInput
}

// Inputs reads and replaces a regime's per-fiscal-year tax inputs
// (getTaxInputs, putTaxInputs; MODULE-tax-engine §2.8, US-T4).
type Inputs struct {
	deps InputsDeps
}

// NewInputs returns the use cases over deps.
func NewInputs(deps InputsDeps) *Inputs {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &Inputs{deps: deps}
}

// Get returns the inputs of req (getTaxInputs). Reads stay open to a
// read-only user and to an archived taxpayer. Order: taxpayer id (404) →
// future year (400) → taxpayer (404) → regime held in the year (422) →
// fiscal year floor (422).
func (uc *Inputs) Get(ctx context.Context, userID string, req InputsRequest) (TaxInputsView, error) {
	key, err := uc.resolve(req)
	if err != nil {
		return TaxInputsView{}, err
	}
	var out TaxInputsView
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		if _, err := uc.deps.Store.Taxpayer(ctx, key.TaxpayerID, LockNone); err != nil {
			return err
		}
		regime, err := uc.activeRegime(ctx, key, LockNone)
		if err != nil {
			return err
		}
		out, err = uc.view(ctx, key, regime)
		return err
	})
	if err != nil {
		return TaxInputsView{}, present("get tax inputs", err)
	}
	return out, nil
}

// Put replaces every value of req with values (putTaxInputs) and returns the
// stored inputs. Order (contract): taxpayer id (404) → future year (400) →
// read-only user (402) → taxpayer (404), archived (422) → regime held in
// the year (422) → fiscal year floor (422) → value rules (400) → replace.
// The taxpayer and regime rows stay locked until commit (LockWrite), so a
// concurrent archive waits and two replaces of one regime serialize.
func (uc *Inputs) Put(ctx context.Context, userID string, req InputsRequest, values []domain.SubmittedInput) (TaxInputsView, error) {
	key, err := uc.resolve(req)
	if err != nil {
		return TaxInputsView{}, err
	}
	var out TaxInputsView
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		var err error
		out, err = uc.put(ctx, userID, key, values)
		return err
	})
	if err != nil {
		return TaxInputsView{}, present("put tax inputs", err)
	}
	uc.deps.Logger.LogAttrs(ctx, slog.LevelInfo, eventInputsUpdated,
		slog.String("user_id", userID), slog.String("taxpayer_id", key.TaxpayerID),
		slog.String("regime_code", string(key.Regime)), slog.Int("fiscal_year", key.FiscalYear))
	return out, nil
}

func (uc *Inputs) put(ctx context.Context, userID string, key InputsKey, values []domain.SubmittedInput) (TaxInputsView, error) {
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return TaxInputsView{}, err
	}
	t, err := uc.deps.Store.Taxpayer(ctx, key.TaxpayerID, LockWrite)
	if err != nil {
		return TaxInputsView{}, err
	}
	if t.Archived {
		return TaxInputsView{}, archivedError()
	}
	regime, err := uc.activeRegime(ctx, key, LockWrite)
	if err != nil {
		return TaxInputsView{}, err
	}
	checked, violations := domain.CheckSubmittedInputs(regime.Definitions, values)
	if len(violations) > 0 {
		return TaxInputsView{}, valuesError(violations)
	}
	if err := uc.deps.Store.Replace(ctx, userID, key, toStored(checked, values)); err != nil {
		return TaxInputsView{}, err
	}
	return uc.view(ctx, key, regime)
}

// resolve validates what needs no database: the taxpayer id (canonical
// form, so "ABC…" and "abc…" name one row) and a future fiscal year.
func (uc *Inputs) resolve(req InputsRequest) (InputsKey, error) {
	id, err := uuid.Parse(req.TaxpayerID)
	if err != nil {
		return InputsKey{}, notFound(errors.Join(ErrTaxpayerNotFound, err))
	}
	if req.FiscalYear > uc.deps.Clock.Now().Year() {
		return InputsKey{}, futureYearError()
	}
	return InputsKey{TaxpayerID: id.String(), Regime: fiscal.RegimeCode(req.RegimeCode), FiscalYear: req.FiscalYear}, nil
}

// activeRegime reads the regime of key under lock and refuses one not held
// in any month of the fiscal year, then a year before firstParamsYear.
func (uc *Inputs) activeRegime(ctx context.Context, key InputsKey, lock Lock) (HeldRegime, error) {
	regime, err := uc.deps.Store.Regime(ctx, key.TaxpayerID, key.Regime, lock)
	switch {
	case errors.Is(err, ErrRegimeNotHeld):
		return HeldRegime{}, regimeNotActiveError()
	case err != nil:
		return HeldRegime{}, err
	case !regime.ActiveIn(key.FiscalYear):
		return HeldRegime{}, regimeNotActiveError()
	case key.FiscalYear < firstParamsYear:
		return HeldRegime{}, parametersUnavailableError()
	}
	return regime, nil
}

// view reads the stored values of key and orders them for the response.
func (uc *Inputs) view(ctx context.Context, key InputsKey, regime HeldRegime) (TaxInputsView, error) {
	values, err := uc.deps.Store.Values(ctx, key)
	if err != nil {
		return TaxInputsView{}, err
	}
	sortValues(values, domain.InputDefPositions(regime.Definitions))
	if values == nil {
		values = []StoredInput{}
	}
	return TaxInputsView{Regime: key.Regime, FiscalYear: key.FiscalYear, Definitions: regime.Definitions, Values: values}, nil
}

// sortValues orders values by their key's position, then month. A stored
// key the definitions no longer name sorts last, by key. O(n log n).
func sortValues(values []StoredInput, position map[domain.InputKey]int) {
	rank := func(k domain.InputKey) int {
		if p, ok := position[k]; ok {
			return p
		}
		return len(position)
	}
	slices.SortFunc(values, func(a, b StoredInput) int {
		return cmp.Or(cmp.Compare(rank(a.Key), rank(b.Key)), cmp.Compare(a.Key, b.Key), cmp.Compare(a.Month, b.Month))
	})
}

// toStored pairs each checked value with the rate text it was sent with:
// the rate is stored as given (numeric keeps its scale, "0.1500").
func toStored(checked []domain.TaxInputValue, submitted []domain.SubmittedInput) []StoredInput {
	out := make([]StoredInput, 0, len(checked))
	for i, v := range checked {
		s := StoredInput{Key: v.Key, Kind: v.Kind, Month: v.Month, Cents: v.Money, Bool: v.Bool}
		if v.Kind == domain.InputRate {
			s.Rate = *submitted[i].Rate
		}
		out = append(out, s)
	}
	return out
}
