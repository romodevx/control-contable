package annual_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func (w *world) inputs() *annual.Inputs {
	return annual.NewInputs(annual.InputsDeps{
		Tx: w.tx, Access: w.access, Taxpayers: w.db, Store: w.db, Clock: clock.NewFixed(today), Logger: w.logger(),
	})
}

func TestInputs_GetDefaultsAndMirror(t *testing.T) {
	w := newWorld()
	w.db.taxInputs[factsKey{fisicaA, 2026}] = map[fiscal.RegimeCode]map[domain.InputKey]money.Cents{
		"612": {adomain.InputAnnualDepreciation: 1_250_000, domain.InputPTUPaid: 0, domain.InputPriorProvisionalPayments: 9},
	}

	got, err := w.inputs().Get(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	assert.Equal(t, annual.InputsView{
		FiscalYear: 2026,
		RegimeTaxInputs: []annual.RegimeTaxInput{
			{Regime: "612", Key: domain.InputPTUPaid, Label: "PTU pagada en el ejercicio", Value: ptr(money.Cents(0))},
			{Regime: "612", Key: domain.InputPriorYearLosses, Label: "Pérdidas fiscales de ejercicios anteriores"},
			{Regime: "612", Key: adomain.InputAnnualDepreciation, Label: "Depreciación del ejercicio", Value: ptr(money.Cents(1_250_000))},
		},
	}, got, "defaults, never-saved; the money annual inputs of every regime held, catalog order; 605 has none")
	assert.Equal(t, []string{"ro"}, w.tx.modes)
}

func TestInputs_GetMirrorsOnlyRegimesHeldInTheYear(t *testing.T) {
	w := newWorld()
	t2 := w.db.taxpayers[moralA]
	t2.regimes = append(t2.regimes, annual.HeldRegime{Code: "626", Start: mustPeriod("2022-01"), End: mustPeriod("2023-12"), Definitions: defs601})
	w.db.taxpayers[moralA] = t2

	got, err := w.inputs().Get(t.Context(), userA, moralA, 2026)
	require.NoError(t, err)
	var keys []string
	for _, in := range got.RegimeTaxInputs {
		keys = append(keys, string(in.Regime)+":"+string(in.Key))
	}
	assert.Equal(t, []string{"601:interest_withholding", "601:ptu_paid"}, keys, "the rate input is not mirrored; 626 ended in 2023")
}

func TestInputs_PutThenGet(t *testing.T) {
	w := newWorld()
	uc := w.inputs()
	facts := annual.Facts{PriorYearAccumulableIncome: ptr(money.Cents(68_000_000)), ToldEmployerWillFile: true}

	put, err := uc.Put(t.Context(), userA, fisicaA, 2026, facts)
	require.NoError(t, err)
	assert.Equal(t, facts, put.Facts)
	require.NotNil(t, put.UpdatedAt)
	assert.Equal(t, stamp, *put.UpdatedAt)
	assert.Equal(t, []string{"taxpayer:write", "upsert"}, w.db.calls)
	assert.Contains(t, w.logs.String(), `"msg":"annual_inputs.updated"`)
	assert.NotContains(t, w.logs.String(), "68000000")

	got, err := uc.Get(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	assert.Equal(t, put, got)

	cleared, err := uc.Put(t.Context(), userA, fisicaA, 2026, annual.Facts{})
	require.NoError(t, err)
	assert.Nil(t, cleared.Facts.PriorYearAccumulableIncome, "null is unknown, not zero")
}

func TestInputs_Refusals(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(w *world)
		id     string
		year   int
		facts  annual.Facts
		put    bool
		code   apperr.Code
		fields []string
	}{
		{"get: malformed id", nil, "x", 2026, annual.Facts{}, false, apperr.CodeNotFound, []string{}},
		{"get: future year", nil, fisicaA, 2027, annual.Facts{}, false, apperr.CodeValidationFailed, []string{"fiscalYear/out_of_range"}},
		{"get: another user's", nil, fisicaB, 2026, annual.Facts{}, false, apperr.CodeNotFound, []string{}},
		{"get: before 2026", nil, fisicaA, 2025, annual.Facts{}, false, apperr.CodeBusinessRuleViolation, []string{"fiscalYear/parameters_unavailable"}},
		{"put: future year", nil, fisicaA, 2027, annual.Facts{}, true, apperr.CodeValidationFailed, []string{"fiscalYear/out_of_range"}},
		{"put: read-only", func(w *world) { w.access.readOnly = true }, fisicaA, 2026, annual.Facts{}, true, apperr.CodeEntitlementRequired, []string{}},
		{"put: another user's", nil, fisicaB, 2026, annual.Facts{}, true, apperr.CodeNotFound, []string{}},
		{"put: archived", func(w *world) { w.archive(fisicaA) }, fisicaA, 2026, annual.Facts{}, true,
			apperr.CodeBusinessRuleViolation, []string{"taxpayerId/taxpayer_archived"}},
		{"put: persona moral", nil, moralA, 2026, annual.Facts{}, true, apperr.CodeBusinessRuleViolation, []string{"taxpayerId/person_type_not_allowed"}},
		{"put: before 2026", nil, fisicaA, 2025, annual.Facts{}, true, apperr.CodeBusinessRuleViolation, []string{"fiscalYear/parameters_unavailable"}},
		{"put: negative amounts", nil, fisicaA, 2026, annual.Facts{PriorYearAccumulableIncome: ptr(money.Cents(-1)), OtherExemptIncome: -1}, true,
			apperr.CodeValidationFailed, []string{"prior_year_accumulable_income_cents/invalid", "other_exempt_income_cents/invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newWorld()
			if tt.setup != nil {
				tt.setup(w)
			}
			var err error
			if tt.put {
				_, err = w.inputs().Put(t.Context(), userA, tt.id, tt.year, tt.facts)
			} else {
				_, err = w.inputs().Get(t.Context(), userA, tt.id, tt.year)
			}
			assert.Equal(t, tt.fields, requireCode(t, err, tt.code))
			assert.NotContains(t, w.db.calls, "upsert")
		})
	}
	t.Run("reads stay open to read-only users and archived taxpayers", func(t *testing.T) {
		w := newWorld()
		w.access.readOnly = true
		w.archive(moralA)
		_, err := w.inputs().Get(t.Context(), userA, moralA, 2026)
		require.NoError(t, err)
	})
}

func TestFactsDomain(t *testing.T) {
	f := annual.Facts{PriorYearAccumulableIncome: ptr(money.Cents(5)), OtherExemptIncome: 7, ToldEmployerWillFile: true}
	assert.Equal(t, adomain.PFAnnualInputs{PriorYearAccumulableIncome: ptr(money.Cents(5)), OtherExemptIncome: 7, ToldEmployerWillFile: true}, f.Domain())
	s := mainFields()
	assert.Equal(t, adomain.SalaryStatement{Accumulable: 30_000_000, ISRWithheld: 4_000_000, Simultaneous: true}, s.Domain())
}
