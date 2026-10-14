package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

var (
	key612PF  = domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}
	key626PF  = domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonFisica}
	key626PM  = domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}
	key605PF  = domain.StrategyKey{Regime: "605", PersonType: fiscal.PersonFisica}
	errStrat  = errors.New("strategy failed")
	isrResult = domain.NewTaxResult(415_096_000)
)

// stubStrategy is a scripted Strategy: it records the Facts and Input it was
// given and returns its fixed requirements and estimate.
type stubStrategy struct {
	key       domain.StrategyKey
	required  []domain.RequiredInput
	estimate  domain.RegimeEstimate
	err       error
	gotFacts  *domain.Facts
	gotInput  *domain.Input
	estimated *bool
}

func newStub(key domain.StrategyKey, required ...domain.RequiredInput) stubStrategy {
	return stubStrategy{
		key: key, required: required,
		estimate:  domain.RegimeEstimate{Key: key, Coverage: domain.TreatmentFull, ISR: &isrResult},
		gotFacts:  new(domain.Facts),
		gotInput:  new(domain.Input),
		estimated: new(bool),
	}
}

func (s stubStrategy) Key() domain.StrategyKey     { return s.key }
func (s stubStrategy) Treatment() domain.Treatment { return domain.TreatmentFull }
func (s stubStrategy) RequiredInputs(_ domain.Options, _ domain.Period, f domain.Facts) []domain.RequiredInput {
	*s.gotFacts = f
	return s.required
}

func (s stubStrategy) Estimate(in domain.Input, _ *taxparams.Params) (domain.RegimeEstimate, error) {
	*s.estimated = true
	*s.gotInput = in
	return s.estimate, s.err
}

func period(t *testing.T, raw string) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

func inputs(t *testing.T, values ...domain.TaxInputValue) domain.TaxInputs {
	t.Helper()
	in, err := domain.NewTaxInputs(values...)
	require.NoError(t, err)
	return in
}

func TestNewRegistryRejectsInvalidRegistrations(t *testing.T) {
	tests := []struct {
		name       string
		strategies []domain.Strategy
		want       error
	}{
		{name: "same key twice", strategies: []domain.Strategy{newStub(key612PF), newStub(key626PF), newStub(key612PF)}, want: domain.ErrDuplicateStrategy},
		{name: "nil strategy", strategies: []domain.Strategy{nil}, want: domain.ErrInvalidStrategy},
		{name: "blank regime", strategies: []domain.Strategy{newStub(domain.StrategyKey{Regime: " ", PersonType: fiscal.PersonFisica})}, want: domain.ErrInvalidStrategy},
		{name: "unknown person type", strategies: []domain.Strategy{newStub(domain.StrategyKey{Regime: "612", PersonType: "juridica"})}, want: domain.ErrInvalidStrategy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewRegistry(tt.strategies...)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestRegistryDuplicateErrorNamesTheKey(t *testing.T) {
	_, err := domain.NewRegistry(newStub(key626PM), newStub(key626PM))
	require.ErrorIs(t, err, domain.ErrDuplicateStrategy)
	assert.Contains(t, err.Error(), "626/moral")
}

func TestRegistryLookup(t *testing.T) {
	pf612, pm626 := newStub(key612PF), newStub(key626PM)
	reg, err := domain.NewRegistry(pf612, pm626)
	require.NoError(t, err)

	t.Run("registered key returns its strategy", func(t *testing.T) {
		got, ok := reg.Lookup(key612PF)
		assert.True(t, ok)
		assert.Equal(t, key612PF, got.Key())
		assert.Equal(t, domain.TreatmentFull, got.Treatment())
	})
	t.Run("same code, other person type is not the same strategy", func(t *testing.T) {
		got, ok := reg.Lookup(key626PF)
		assert.False(t, ok)
		assert.Equal(t, key626PF, got.Key())
		assert.Equal(t, domain.TreatmentInformational, got.Treatment())
	})
	t.Run("unregistered key falls back to informational", func(t *testing.T) {
		got, ok := reg.Lookup(key605PF)
		assert.False(t, ok)
		assert.Equal(t, domain.NewInformationalStrategy(key605PF), got)
	})
}

func TestZeroRegistryIsAllInformational(t *testing.T) {
	var reg domain.Registry
	got, ok := reg.Lookup(key612PF)
	assert.False(t, ok)
	assert.Equal(t, domain.TreatmentInformational, got.Treatment())
}

func TestRegistryIsImmutableAfterBuild(t *testing.T) {
	strategies := []domain.Strategy{newStub(key612PF)}
	reg, err := domain.NewRegistry(strategies...)
	require.NoError(t, err)

	strategies[0] = newStub(key626PM) // the caller's slice is not the registry
	got, ok := reg.Lookup(key612PF)
	assert.True(t, ok)
	assert.Equal(t, key612PF, got.Key())
	_, ok = reg.Lookup(key626PM)
	assert.False(t, ok)
}

func TestInformationalStrategy(t *testing.T) {
	s := domain.NewInformationalStrategy(key605PF)
	assert.Equal(t, key605PF, s.Key())
	assert.Equal(t, domain.TreatmentInformational, s.Treatment())
	assert.Empty(t, s.RequiredInputs(domain.Options{}, period(t, "2026-03"), domain.Facts{}))

	est, err := s.Estimate(domain.Input{Key: key605PF, Period: period(t, "2026-03")}, nil)
	require.NoError(t, err)
	assert.Equal(t, domain.RegimeEstimate{
		Key:         key605PF,
		Coverage:    domain.TreatmentInformational,
		Explanation: domain.RegimeExplanation(key605PF),
	}, est)
	assert.Equal(t, domain.StatusInformational, est.Status())
	assert.Nil(t, est.ISR, "an informational regime has no number")
	assert.Nil(t, est.IVA)
}

func TestRegimeExplanationIsTheCatalogKey(t *testing.T) {
	assert.Equal(t, domain.LabelKey("regime_explanation:605/fisica"), domain.RegimeExplanation(key605PF))
	assert.Equal(t, domain.LabelKey("regime_explanation:626/moral"), domain.RegimeExplanation(key626PM))
}

func TestMissingInputs(t *testing.T) {
	prior := func(months ...time.Month) domain.RequiredInput {
		return domain.RequiredInput{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: months}
	}
	coef := domain.RequiredInput{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate}
	priorIncome := domain.RequiredInput{Key: domain.InputPriorYearIncome, Kind: domain.InputMoney}
	firstYear := domain.RequiredInput{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean}
	paidJan := domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Money: 1_000_00}
	paidFeb := domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.February}

	tests := []struct {
		name     string
		required []domain.RequiredInput
		values   []domain.TaxInputValue
		want     []domain.MissingInput
	}{
		{name: "nothing required", want: nil},
		{name: "january: no prior months", required: []domain.RequiredInput{prior()}, want: nil},
		{
			name: "every prior month present, a zero amount counts as entered", required: []domain.RequiredInput{prior(time.January, time.February)},
			values: []domain.TaxInputValue{paidJan, paidFeb}, want: nil,
		},
		{
			name: "missing months named in ascending order", required: []domain.RequiredInput{prior(time.March, time.January, time.February)},
			values: []domain.TaxInputValue{paidFeb},
			want:   []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January, time.March}}},
		},
		{
			name: "scalar kinds", required: []domain.RequiredInput{coef, priorIncome, firstYear},
			values: []domain.TaxInputValue{{Key: domain.InputPriorYearIncome, Kind: domain.InputMoney}},
			want:   []domain.MissingInput{{Key: domain.InputCoeficienteUtilidad}, {Key: domain.InputFirstFiscalYear}},
		},
		{
			name: "a value stored under another kind is missing", required: []domain.RequiredInput{coef},
			values: []domain.TaxInputValue{{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputMoney, Money: 1}},
			want:   []domain.MissingInput{{Key: domain.InputCoeficienteUtilidad}},
		},
		{
			name: "strategy order kept, the first missing key leads the CTA", required: []domain.RequiredInput{coef, prior(time.January)},
			want: []domain.MissingInput{
				{Key: domain.InputCoeficienteUtilidad},
				{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January}},
			},
		},
		{
			name: "a key required twice is reported once, months merged", required: []domain.RequiredInput{prior(time.February), coef, prior(time.January, time.February)},
			want: []domain.MissingInput{
				{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January, time.February}},
				{Key: domain.InputCoeficienteUtilidad},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.MissingInputs(tt.required, inputs(t, tt.values...))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMissingInputsRejectsMalformedRequirements(t *testing.T) {
	tests := []struct {
		name     string
		required domain.RequiredInput
		want     error
	}{
		{name: "unknown kind", required: domain.RequiredInput{Key: domain.InputPTUPaid, Kind: "text"}, want: domain.ErrInvalidTaxInput},
		{name: "blank key", required: domain.RequiredInput{Key: "", Kind: domain.InputMoney}, want: domain.ErrInvalidTaxInput},
		{name: "month out of range", required: domain.RequiredInput{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: []time.Month{13}}, want: domain.ErrInvalidMonth},
		{name: "months on a scalar input", required: domain.RequiredInput{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Months: []time.Month{time.January}}, want: domain.ErrInvalidMonth},
		{name: "one key under two kinds", required: domain.RequiredInput{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoney}, want: domain.ErrInvalidTaxInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			required := []domain.RequiredInput{
				{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: []time.Month{time.January}},
				tt.required,
			}
			_, err := domain.MissingInputs(required, domain.TaxInputs{})
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestRegistryEstimate(t *testing.T) {
	march := period(t, "2026-03")
	active := []time.Month{time.January, time.February, time.March}
	priorReq := domain.RequiredInput{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: []time.Month{time.January, time.February}}
	paidJan := domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Money: 3_000_00}

	t.Run("all inputs present: complete, the strategy's numbers", func(t *testing.T) {
		paidFeb := paidJan
		paidFeb.Month = time.February
		stub := newStub(key612PF, priorReq)
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)
		in := domain.Input{Key: key612PF, Period: march, Inputs: inputs(t, paidJan, paidFeb)}

		est, err := reg.Estimate(in, active, nil)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusComplete, est.Status())
		assert.Empty(t, est.MissingInputs)
		assert.Equal(t, &isrResult, est.ISR)
		assert.Equal(t, domain.Facts{ActiveMonths: active, Inputs: in.Inputs}, *stub.gotFacts,
			"RequiredInputs sees the input's own tax inputs")
		assert.Equal(t, in, *stub.gotInput)
	})

	t.Run("missing input: incomplete with the missing keys, computed with zero (spec §2.2)", func(t *testing.T) {
		stub := newStub(key612PF, priorReq)
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)

		est, err := reg.Estimate(domain.Input{Key: key612PF, Period: march, Inputs: inputs(t, paidJan)}, active, nil)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusIncomplete, est.Status())
		assert.Equal(t, []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.February}}}, est.MissingInputs)
		assert.True(t, *stub.estimated)
		assert.Equal(t, &isrResult, est.ISR, "TX-612-INC-01: the ISR is still shown, labelled incompleto")
	})

	t.Run("the registry's check replaces whatever the strategy reported", func(t *testing.T) {
		stub := newStub(key612PF)
		stub.estimate.MissingInputs = []domain.MissingInput{{Key: domain.InputPTUPaid}}
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)

		est, err := reg.Estimate(domain.Input{Key: key612PF, Period: march}, active, nil)
		require.NoError(t, err)
		assert.Empty(t, est.MissingInputs)
		assert.Equal(t, domain.StatusComplete, est.Status())
	})

	t.Run("partial strategy with a missing input is incomplete (precedence)", func(t *testing.T) {
		stub := newStub(key626PM, domain.RequiredInput{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate})
		stub.estimate.Coverage = domain.TreatmentPartial
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)

		est, err := reg.Estimate(domain.Input{Key: key626PM, Period: march}, active, nil)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusIncomplete, est.Status())
	})

	t.Run("unregistered regime: informational, no number, explanation key", func(t *testing.T) {
		reg, err := domain.NewRegistry(newStub(key612PF))
		require.NoError(t, err)

		est, err := reg.Estimate(domain.Input{Key: key605PF, Period: march}, active, nil)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusInformational, est.Status())
		assert.Nil(t, est.ISR)
		assert.Nil(t, est.IVA)
		assert.Equal(t, domain.RegimeExplanation(key605PF), est.Explanation)
		assert.Empty(t, est.MissingInputs)
	})

	t.Run("strategy error is wrapped with its key", func(t *testing.T) {
		stub := newStub(key612PF)
		stub.err = errStrat
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)

		_, err = reg.Estimate(domain.Input{Key: key612PF, Period: march}, active, nil)
		require.ErrorIs(t, err, errStrat)
		assert.Contains(t, err.Error(), "612/fisica")
	})

	t.Run("malformed requirement fails before the strategy runs", func(t *testing.T) {
		stub := newStub(key612PF, domain.RequiredInput{Key: domain.InputPTUPaid, Kind: "text"})
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)

		_, err = reg.Estimate(domain.Input{Key: key612PF, Period: march}, active, nil)
		require.ErrorIs(t, err, domain.ErrInvalidTaxInput)
		assert.False(t, *stub.estimated)
	})

	t.Run("estimate for another key is refused", func(t *testing.T) {
		stub := newStub(key612PF)
		stub.estimate.Key = key626PF
		reg, err := domain.NewRegistry(stub)
		require.NoError(t, err)

		_, err = reg.Estimate(domain.Input{Key: key612PF, Period: march}, active, nil)
		require.ErrorIs(t, err, domain.ErrStrategyMismatch)
	})
}

// An input the user entered as zero is present: only an absent key is
// missing (ADR-0002 §2).
func TestEnteredZeroIsPresent(t *testing.T) {
	in := inputs(t, domain.TaxInputValue{Key: domain.InputPriorYearIncome, Kind: domain.InputMoney, Money: money.Cents(0)})
	got, err := domain.MissingInputs([]domain.RequiredInput{{Key: domain.InputPriorYearIncome, Kind: domain.InputMoney}}, in)
	require.NoError(t, err)
	assert.Nil(t, got)
}
