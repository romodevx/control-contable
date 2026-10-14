package usecase_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

func request(year int) usecase.InputsRequest {
	return usecase.InputsRequest{TaxpayerID: taxpayerA, RegimeCode: "612", FiscalYear: year}
}

func key612(year int) usecase.InputsKey {
	return usecase.InputsKey{TaxpayerID: taxpayerA, Regime: "612", FiscalYear: year}
}

func ptr[T any](v T) *T { return &v }

func monthlyPayment(month int, c int64) domain.SubmittedInput {
	return domain.SubmittedInput{Key: domain.InputPriorProvisionalPayments, Month: &month, Cents: &c}
}

func storedPayment(month time.Month, c int64) usecase.StoredInput {
	return usecase.StoredInput{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: month, Cents: money.Cents(c)}
}

// appError asserts err is a client-ready *apperr.Error with code and returns
// its "field/code" details.
func appError(t *testing.T, err error, code apperr.Code) []string {
	t.Helper()
	e, ok := errors.AsType[*apperr.Error](err)
	require.Truef(t, ok, "want *apperr.Error, got %v", err)
	assert.Equal(t, code, e.Code, e.Error())
	var out []string
	for _, f := range e.Fields {
		out = append(out, f.Field+"/"+f.Code)
	}
	return out
}

func TestInputsGet_definitionsAndOrderedValues(t *testing.T) {
	f := newFixture()
	ptu := usecase.StoredInput{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Cents: 0}
	f.store.values[key612(2026)] = []usecase.StoredInput{ptu, storedPayment(time.February, 2_000_00), storedPayment(time.January, 1_000_00)}

	got, err := f.uc.Get(t.Context(), userA, request(2026))

	require.NoError(t, err)
	assert.Equal(t, usecase.TaxInputsView{
		Regime: "612", FiscalYear: 2026, Definitions: defs612,
		Values: []usecase.StoredInput{storedPayment(time.January, 1_000_00), storedPayment(time.February, 2_000_00), ptu},
	}, got, "values by the key's position in definitions, then month")
	assert.Equal(t, []string{"ro"}, f.tx.modes)
	assert.Equal(t, []string{"taxpayer:none", "regime:none"}, f.store.locks)
	assert.Zero(t, f.access.calls, "reads stay open to a read-only user")
}

func TestInputsGet_emptyValuesAreAnEmptyList(t *testing.T) {
	f := newFixture()

	got, err := f.uc.Get(t.Context(), userA, request(2026))

	require.NoError(t, err)
	assert.NotNil(t, got.Values)
	assert.Empty(t, got.Values)
}

func TestInputsGet_readOnlyUserAndArchivedTaxpayerRead(t *testing.T) {
	f := newFixture()
	f.access.readOnly = true
	f.store.taxpayers[taxpayerA] = usecase.TaxpayerState{Archived: true}

	_, err := f.uc.Get(t.Context(), userA, request(2026))

	assert.NoError(t, err)
}

func TestInputsGet_canonicalizesTheTaxpayerID(t *testing.T) {
	f := newFixture()
	req := request(2026)
	req.TaxpayerID = strings.ToUpper(taxpayerA)

	_, err := f.uc.Get(t.Context(), userA, req)

	assert.NoError(t, err)
}

func TestInputs_refusals(t *testing.T) {
	ended := usecase.HeldRegime{Start: mustPeriod("2022-01"), End: mustPeriod("2025-12"), Definitions: defs612}
	lateStart := usecase.HeldRegime{Start: mustPeriod("2027-01"), Definitions: defs612}
	for _, tc := range []struct {
		name    string
		mutate  func(*fixture)
		req     usecase.InputsRequest
		code    apperr.Code
		details []string
		cause   error
	}{
		{
			name: "a malformed taxpayer id", req: usecase.InputsRequest{TaxpayerID: "nope", RegimeCode: "612", FiscalYear: 2026},
			code: apperr.CodeNotFound, cause: usecase.ErrTaxpayerNotFound,
		},
		{
			name: "another user's or an unknown taxpayer", req: usecase.InputsRequest{TaxpayerID: "0199d4a2-1111-7000-8000-0000000000ff", RegimeCode: "612", FiscalYear: 2026},
			code: apperr.CodeNotFound, cause: usecase.ErrTaxpayerNotFound,
		},
		{
			name: "a future fiscal year", req: request(2027),
			code: apperr.CodeValidationFailed, details: []string{"fiscal_year/out_of_range"},
		},
		{
			name: "a fiscal year before the first with parameters", req: request(2025),
			code: apperr.CodeBusinessRuleViolation, details: []string{"fiscal_year/parameters_unavailable"},
		},
		{
			name: "a regime the taxpayer never held", req: usecase.InputsRequest{TaxpayerID: taxpayerA, RegimeCode: "626", FiscalYear: 2026},
			code: apperr.CodeBusinessRuleViolation, details: []string{"regimeCode/regime_not_active"}, cause: usecase.ErrRegimeNotActive,
		},
		{
			name: "a regime that ended before the year", req: request(2026),
			mutate: func(f *fixture) { f.store.hold(taxpayerA, "612", ended) },
			code:   apperr.CodeBusinessRuleViolation, details: []string{"regimeCode/regime_not_active"}, cause: usecase.ErrRegimeNotActive,
		},
		{
			name: "a regime that starts after the year", req: request(2026),
			mutate: func(f *fixture) { f.store.hold(taxpayerA, "612", lateStart) },
			code:   apperr.CodeBusinessRuleViolation, details: []string{"regimeCode/regime_not_active"}, cause: usecase.ErrRegimeNotActive,
		},
	} {
		for _, op := range []string{"get", "put"} {
			t.Run(op+": "+tc.name, func(t *testing.T) {
				f := newFixture()
				if tc.mutate != nil {
					tc.mutate(f)
				}
				var err error
				if op == "get" {
					_, err = f.uc.Get(t.Context(), userA, tc.req)
				} else {
					_, err = f.uc.Put(t.Context(), userA, tc.req, nil)
				}
				assert.Equal(t, tc.details, appError(t, err, tc.code))
				if tc.cause != nil {
					assert.ErrorIs(t, err, tc.cause)
				}
				assert.Empty(t, f.store.replaced)
			})
		}
	}
}

func TestInputsGet_regimeHeldInSomeMonthOfTheYear(t *testing.T) {
	for _, window := range []usecase.HeldRegime{
		{Start: mustPeriod("2026-10")},
		{Start: mustPeriod("2022-01"), End: mustPeriod("2026-01")},
		{Start: mustPeriod("2026-03"), End: mustPeriod("2026-03")},
	} {
		f := newFixture()
		window.Definitions = defs612
		f.store.hold(taxpayerA, "612", window)

		_, err := f.uc.Get(t.Context(), userA, request(2026))

		assert.NoError(t, err)
	}
}

func TestInputsPut_replacesAndReturnsTheStoredInputs(t *testing.T) {
	f := newFixture()
	f.store.values[key612(2026)] = []usecase.StoredInput{storedPayment(time.March, 9)}

	got, err := f.uc.Put(t.Context(), userA, request(2026), []domain.SubmittedInput{
		{Key: domain.InputPTUPaid, Cents: ptr[int64](0)},
		monthlyPayment(2, 2_000_00),
		monthlyPayment(1, 2_000_00),
	})

	require.NoError(t, err)
	ptu := usecase.StoredInput{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Cents: 0}
	require.Len(t, f.store.replaced, 1)
	assert.Equal(t, replaceCall{
		userID: userA, key: key612(2026),
		values: []usecase.StoredInput{ptu, storedPayment(time.February, 2_000_00), storedPayment(time.January, 2_000_00)},
	}, f.store.replaced[0])
	assert.Equal(t, []usecase.StoredInput{storedPayment(time.January, 2_000_00), storedPayment(time.February, 2_000_00), ptu}, got.Values)
	assert.Equal(t, defs612, got.Definitions)
	assert.Equal(t, []string{"rw"}, f.tx.modes)
	assert.Equal(t, []string{"taxpayer:write", "regime:write"}, f.store.locks, "the regime row lock serializes concurrent replaces")
}

func TestInputsPut_keepsTheRateAsGiven(t *testing.T) {
	f := newFixture()
	f.store.taxpayers[taxpayerA] = usecase.TaxpayerState{}
	f.store.hold(taxpayerA, "601", usecase.HeldRegime{Start: mustPeriod("2022-01"), Definitions: defs601})
	req := usecase.InputsRequest{TaxpayerID: taxpayerA, RegimeCode: "601", FiscalYear: 2026}

	got, err := f.uc.Put(t.Context(), userA, req, []domain.SubmittedInput{
		{Key: domain.InputCoeficienteUtilidad, Rate: ptr("0.1500")},
		{Key: domain.InputFirstFiscalYear, Bool: ptr(true)},
	})

	require.NoError(t, err)
	assert.Equal(t, []usecase.StoredInput{
		{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: "0.1500"},
		{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: true},
	}, got.Values)
}

func TestInputsPut_emptyValuesClear(t *testing.T) {
	f := newFixture()
	f.store.values[key612(2026)] = []usecase.StoredInput{storedPayment(time.March, 9)}

	got, err := f.uc.Put(t.Context(), userA, request(2026), []domain.SubmittedInput{})

	require.NoError(t, err)
	assert.Empty(t, got.Values)
	require.Len(t, f.store.replaced, 1)
	assert.Empty(t, f.store.replaced[0].values)
}

func TestInputsPut_auditsIdsOnly(t *testing.T) {
	f := newFixture()

	_, err := f.uc.Put(t.Context(), userA, request(2026), []domain.SubmittedInput{monthlyPayment(1, 123_456_78)})

	require.NoError(t, err)
	logs := f.logs.String()
	assert.Contains(t, logs, `"msg":"tax_inputs.updated"`)
	assert.Contains(t, logs, `"user_id":"`+userA+`"`)
	assert.Contains(t, logs, `"taxpayer_id":"`+taxpayerA+`"`)
	assert.Contains(t, logs, `"regime_code":"612"`)
	assert.Contains(t, logs, `"fiscal_year":2026`)
	assert.NotContains(t, logs, "12345678", "S-3/S-8: values are never logged")
	assert.NotContains(t, logs, "prior_provisional_payments")
}

func TestInputsPut_valueRules(t *testing.T) {
	f := newFixture()

	_, err := f.uc.Put(t.Context(), userA, request(2026), []domain.SubmittedInput{
		{Key: "coeficiente_utilidad", Rate: ptr("0.5")},
		{Key: domain.InputPTUPaid, Rate: ptr("0.5")},
		monthlyPayment(1, 1),
		{Key: domain.InputPriorProvisionalPayments, Cents: ptr[int64](1)},
		monthlyPayment(1, 2),
	})

	assert.Equal(t, []string{
		"values.0.key/unknown_input",
		"values.1.value_rate/invalid",
		"values.1.value_cents/invalid",
		"values.3.month/invalid",
		"values.4.month/invalid",
	}, appError(t, err, apperr.CodeValidationFailed))
	assert.ErrorIs(t, err, domain.ErrUnknownInput)
	assert.Empty(t, f.store.replaced)
	assert.NotContains(t, err.Error(), "0.5", "messages never echo values")
}

func TestInputsPut_rateOutOfRange(t *testing.T) {
	f := newFixture()
	f.store.hold(taxpayerA, "601", usecase.HeldRegime{Start: mustPeriod("2022-01"), Definitions: defs601})
	req := usecase.InputsRequest{TaxpayerID: taxpayerA, RegimeCode: "601", FiscalYear: 2026}

	_, err := f.uc.Put(t.Context(), userA, req, []domain.SubmittedInput{{Key: domain.InputCoeficienteUtilidad, Rate: ptr("1.5")}})

	assert.Equal(t, []string{"values.0.value_rate/out_of_range"}, appError(t, err, apperr.CodeValidationFailed))
}

func TestInputsPut_orderOfChecks(t *testing.T) {
	bad := []domain.SubmittedInput{{Key: "nope", Cents: ptr[int64](1)}}

	t.Run("read-only before archived", func(t *testing.T) {
		f := newFixture()
		f.access.readOnly = true
		f.store.taxpayers[taxpayerA] = usecase.TaxpayerState{Archived: true}

		_, err := f.uc.Put(t.Context(), userA, request(2026), bad)

		assert.ErrorIs(t, err, errReadOnly)
		assert.Empty(t, f.store.locks, "no row is read for a read-only user")
	})
	t.Run("archived before the regime and the values", func(t *testing.T) {
		f := newFixture()
		f.store.taxpayers[taxpayerA] = usecase.TaxpayerState{Archived: true}

		_, err := f.uc.Put(t.Context(), userA, usecase.InputsRequest{TaxpayerID: taxpayerA, RegimeCode: "626", FiscalYear: 2026}, bad)

		assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, appError(t, err, apperr.CodeBusinessRuleViolation))
		assert.ErrorIs(t, err, usecase.ErrTaxpayerArchived)
	})
	t.Run("the regime window before the values", func(t *testing.T) {
		f := newFixture()

		_, err := f.uc.Put(t.Context(), userA, usecase.InputsRequest{TaxpayerID: taxpayerA, RegimeCode: "626", FiscalYear: 2026}, bad)

		assert.Equal(t, []string{"regimeCode/regime_not_active"}, appError(t, err, apperr.CodeBusinessRuleViolation))
	})
	t.Run("the fiscal year floor before the values", func(t *testing.T) {
		f := newFixture()

		_, err := f.uc.Put(t.Context(), userA, request(2024), bad)

		assert.Equal(t, []string{"fiscal_year/parameters_unavailable"}, appError(t, err, apperr.CodeBusinessRuleViolation))
	})
	t.Run("a future year before access", func(t *testing.T) {
		f := newFixture()
		f.access.readOnly = true

		_, err := f.uc.Put(t.Context(), userA, request(2027), bad)

		assert.Equal(t, []string{"fiscal_year/out_of_range"}, appError(t, err, apperr.CodeValidationFailed))
		assert.Zero(t, f.access.calls)
	})
}

func TestInputs_infrastructureFailureIsNotAClientError(t *testing.T) {
	boom := errors.New("connection reset")
	f := newFixture()
	f.store.failWith = boom

	_, getErr := f.uc.Get(t.Context(), userA, request(2026))
	_, putErr := f.uc.Put(t.Context(), userA, request(2026), nil)

	for _, err := range []error{getErr, putErr} {
		assert.ErrorIs(t, err, boom)
		_, isClient := errors.AsType[*apperr.Error](err)
		assert.False(t, isClient, "a 500, never a client error")
	}
}

func TestHeldRegime_ActiveIn(t *testing.T) {
	r := usecase.HeldRegime{Start: mustPeriod("2023-07"), End: mustPeriod("2025-02")}

	assert.False(t, r.ActiveIn(2022))
	assert.True(t, r.ActiveIn(2023))
	assert.True(t, r.ActiveIn(2025))
	assert.False(t, r.ActiveIn(2026))
	assert.True(t, usecase.HeldRegime{Start: mustPeriod("2023-07")}.ActiveIn(2100), "open-ended")
}
