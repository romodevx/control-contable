//go:build integration

package taxinputs_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

func ptr[T any](v T) *T { return &v }

func payment(month int, c int64) domain.SubmittedInput {
	return domain.SubmittedInput{Key: domain.InputPriorProvisionalPayments, Month: &month, Cents: &c}
}

func req(taxpayerID, code string, year int) usecase.InputsRequest {
	return usecase.InputsRequest{TaxpayerID: taxpayerID, RegimeCode: code, FiscalYear: year}
}

func requireCode(t *testing.T, err error, code apperr.Code) []string {
	t.Helper()
	e, ok := errors.AsType[*apperr.Error](err)
	require.Truef(t, ok, "want *apperr.Error, got %v", err)
	require.Equal(t, code, e.Code, e.Error())
	var out []string
	for _, f := range e.Fields {
		out = append(out, f.Field+"/"+f.Code)
	}
	return out
}

func keysOf(defs []domain.InputDef) map[domain.InputKey]domain.InputKind {
	out := map[domain.InputKey]domain.InputKind{}
	for _, d := range defs {
		out[d.Key] = d.Kind
	}
	return out
}

func TestStore_TaxpayerAndRegimeUnderRLS(t *testing.T) {
	e := newEnv(t)
	e.archive(t, userA, moralA)

	err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		got, err := e.store.Taxpayer(ctx, fisicaA, usecase.LockNone)
		require.NoError(t, err)
		assert.False(t, got.Archived)
		got, err = e.store.Taxpayer(ctx, moralA, usecase.LockNone)
		require.NoError(t, err)
		assert.True(t, got.Archived)
		_, err = e.store.Taxpayer(ctx, fisicaB, usecase.LockNone)
		assert.ErrorIs(t, err, usecase.ErrTaxpayerNotFound, "another user's taxpayer does not exist")

		held, err := e.store.Regime(ctx, fisicaA, "612", usecase.LockNone)
		require.NoError(t, err)
		assert.Equal(t, "2022-01", held.Start.String())
		assert.True(t, held.End.IsZero())
		assert.Equal(t, domain.InputMoneyMonthly, keysOf(held.Definitions)[domain.InputPriorProvisionalPayments])
		assert.Equal(t, domain.InputPriorProvisionalPayments, held.Definitions[0].Key, "catalog order")
		assert.NotEmpty(t, held.Definitions[0].Label)
		assert.True(t, held.Definitions[0].Required)

		ended, err := e.store.Regime(ctx, moralA, "626", usecase.LockNone)
		require.NoError(t, err)
		assert.Equal(t, "2023-12", ended.End.String())
		assert.Equal(t, domain.InputMoneyMonthly, keysOf(ended.Definitions)[domain.InputPriorProvisionalPayments], "626 PM, not 626 PF")

		moral, err := e.store.Regime(ctx, moralA, "601", usecase.LockNone)
		require.NoError(t, err)
		assert.Equal(t, domain.InputRate, keysOf(moral.Definitions)[domain.InputCoeficienteUtilidad])
		assert.Equal(t, domain.InputBoolean, keysOf(moral.Definitions)[domain.InputFirstFiscalYear])

		_, err = e.store.Regime(ctx, fisicaA, "626", usecase.LockNone)
		assert.ErrorIs(t, err, usecase.ErrRegimeNotHeld)
		_, err = e.store.Regime(ctx, fisicaB, "612", usecase.LockNone)
		assert.ErrorIs(t, err, usecase.ErrRegimeNotHeld, "another user's regime does not exist")
		return nil
	})
	require.NoError(t, err)

	err = e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
		_, err := e.store.Taxpayer(ctx, fisicaA, usecase.LockWrite)
		require.NoError(t, err)
		_, err = e.store.Regime(ctx, fisicaA, "612", usecase.LockWrite)
		return err
	})
	require.NoError(t, err, "cc_app may take the write locks")
}

func TestInputs_PutThenGetRoundTrip(t *testing.T) {
	e := newEnv(t)

	put, err := e.inputs.Put(t.Context(), userA, req(moralA, "601", 2026), []domain.SubmittedInput{
		{Key: domain.InputFirstFiscalYear, Bool: ptr(false)},
		payment(2, 2_000_00),
		{Key: domain.InputCoeficienteUtilidad, Rate: ptr("0.1500")},
		payment(1, 1_500_00),
	})
	require.NoError(t, err)

	got, err := e.inputs.Get(t.Context(), userA, req(moralA, "601", 2026))
	require.NoError(t, err)
	assert.Equal(t, put, got, "PUT answers the stored inputs")
	assert.Equal(t, []usecase.StoredInput{
		{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: "0.1500"},
		{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: false},
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Cents: 1_500_00},
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.February, Cents: 2_000_00},
	}, got.Values, "definition order then month; the rate keeps its scale")

	_, err = e.inputs.Put(t.Context(), userA, req(moralA, "601", 2026), []domain.SubmittedInput{payment(3, 7)})
	require.NoError(t, err)
	got, err = e.inputs.Get(t.Context(), userA, req(moralA, "601", 2026))
	require.NoError(t, err)
	assert.Equal(t, []usecase.StoredInput{
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.March, Cents: 7},
	}, got.Values, "replace, not merge")

	again, err := e.inputs.Put(t.Context(), userA, req(moralA, "601", 2026), []domain.SubmittedInput{payment(3, 7)})
	require.NoError(t, err)
	assert.Equal(t, got, again, "idempotent")

	cleared, err := e.inputs.Put(t.Context(), userA, req(moralA, "601", 2026), []domain.SubmittedInput{})
	require.NoError(t, err)
	assert.Empty(t, cleared.Values)
}

func TestInputs_CrossUserIsNotFound(t *testing.T) {
	e := newEnv(t)
	_, err := e.inputs.Put(t.Context(), userB, req(fisicaB, "612", 2026), []domain.SubmittedInput{payment(1, 5)})
	require.NoError(t, err)

	_, err = e.inputs.Get(t.Context(), userA, req(fisicaB, "612", 2026))
	requireCode(t, err, apperr.CodeNotFound)
	_, err = e.inputs.Put(t.Context(), userA, req(fisicaB, "612", 2026), nil)
	requireCode(t, err, apperr.CodeNotFound)

	got, err := e.inputs.Get(t.Context(), userB, req(fisicaB, "612", 2026))
	require.NoError(t, err)
	assert.Len(t, got.Values, 1, "B's value is untouched")
}

func TestInputs_BusinessRulesOverTheDatabase(t *testing.T) {
	e := newEnv(t)
	e.archive(t, userA, fisicaA)

	_, err := e.inputs.Get(t.Context(), userA, req(fisicaA, "612", 2026))
	assert.NoError(t, err, "an archived taxpayer stays readable")
	_, err = e.inputs.Put(t.Context(), userA, req(fisicaA, "612", 2026), nil)
	assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))

	_, err = e.inputs.Put(t.Context(), userA, req(moralA, "626", 2026), nil)
	assert.Equal(t, []string{"regimeCode/regime_not_active"}, requireCode(t, err, apperr.CodeBusinessRuleViolation),
		"626 ended in 2023")
	_, err = e.inputs.Get(t.Context(), userA, req(moralA, "601", 2025))
	assert.Equal(t, []string{"fiscal_year/parameters_unavailable"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))

	_, err = e.inputs.Put(t.Context(), userA, req(moralA, "601", 2026), []domain.SubmittedInput{
		{Key: "ytd_before_app_income", Cents: ptr[int64](1)},
		{Key: domain.InputCoeficienteUtilidad, Rate: ptr("1.5")},
	})
	assert.Equal(t, []string{"values.0.key/unknown_input", "values.1.value_rate/out_of_range"},
		requireCode(t, err, apperr.CodeValidationFailed), "601 PM takes no ytd_before_app_income")
}

func TestInputs_ReadOnlyUserReadsButCannotWrite(t *testing.T) {
	e := newEnv(t)
	readOnly := usecase.NewInputs(usecase.InputsDeps{Tx: e.txm, Access: refuseAll{}, Store: e.store, Clock: clock.NewFixed(today)})

	_, err := readOnly.Get(t.Context(), userA, req(fisicaA, "612", 2026))
	require.NoError(t, err)
	_, err = readOnly.Put(t.Context(), userA, req(fisicaA, "612", 2026), nil)
	requireCode(t, err, apperr.CodeEntitlementRequired)
}

// Two replaces of one regime serialize on the regime row: the second waits
// for the first to commit, then replaces its result whole (no 23505, no
// mixed set).
func TestInputs_ConcurrentReplacesSerialize(t *testing.T) {
	e := newEnv(t)
	holding, release := make(chan struct{}), make(chan struct{})
	var first error
	var wg sync.WaitGroup
	wg.Go(func() {
		first = e.txm.ReadWrite(context.Background(), userA, func(ctx context.Context) error {
			if _, err := e.store.Taxpayer(ctx, fisicaA, usecase.LockWrite); err != nil {
				return err
			}
			if _, err := e.store.Regime(ctx, fisicaA, "612", usecase.LockWrite); err != nil {
				return err
			}
			key := usecase.InputsKey{TaxpayerID: fisicaA, Regime: "612", FiscalYear: 2026}
			err := e.store.Replace(ctx, userA, key, []usecase.StoredInput{
				{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Cents: 1},
				{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.February, Cents: 2},
			})
			close(holding)
			<-release
			return err
		})
	})
	<-holding
	var second error
	wg.Go(func() {
		_, second = e.inputs.Put(context.Background(), userA, req(fisicaA, "612", 2026),
			[]domain.SubmittedInput{payment(1, 10), payment(3, 30)})
	})
	require.True(t, waitForLockWaiters(t, e, 1), "the second replace waits on the regime row lock")
	close(release)
	wg.Wait()

	require.NoError(t, first)
	require.NoError(t, second)
	got, err := e.inputs.Get(t.Context(), userA, req(fisicaA, "612", 2026))
	require.NoError(t, err)
	assert.Equal(t, []usecase.StoredInput{
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Cents: 10},
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.March, Cents: 30},
	}, got.Values)
}

// The TaxInputReader adapter (M5-02) reads every regime's inputs of a fiscal
// year in one query, as domain.TaxInputs.
func TestStore_TaxInputsReader(t *testing.T) {
	e := newEnv(t)
	_, err := e.inputs.Put(t.Context(), userA, req(moralA, "601", 2026), []domain.SubmittedInput{
		{Key: domain.InputCoeficienteUtilidad, Rate: ptr("0.1500")},
		{Key: domain.InputFirstFiscalYear, Bool: ptr(true)},
		payment(1, 2_000_00), payment(2, 3_000_00),
		{Key: domain.InputPTUPaid, Cents: ptr[int64](0)},
	})
	require.NoError(t, err)
	e.exec(t, userA, `INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, value_cents)
		VALUES ($1, '`+moralA+`', '601', 2025, 'ptu_paid', 99)`)
	var reader usecase.TaxInputReader = e.store

	var got map[fiscal.RegimeCode]domain.TaxInputs
	err = e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		var err error
		got, err = reader.TaxInputs(ctx, moralA, 2026)
		return err
	})
	require.NoError(t, err)

	require.Len(t, got, 1, "regimes without inputs are absent")
	in := got["601"]
	cu, ok := in.Rate(domain.InputCoeficienteUtilidad)
	assert.True(t, ok)
	assert.Equal(t, money.Rate(150_000), cu)
	first, ok := in.Bool(domain.InputFirstFiscalYear)
	assert.True(t, ok)
	assert.True(t, first)
	ptu, ok := in.Money(domain.InputPTUPaid)
	assert.True(t, ok, "the 2026 value, not 2025's")
	assert.Zero(t, ptu)
	prior, err := in.SumBefore(domain.InputPriorProvisionalPayments, time.March)
	require.NoError(t, err)
	assert.Equal(t, money.Cents(5_000_00), prior)

	err = e.txm.ReadOnly(t.Context(), userB, func(ctx context.Context) error {
		var err error
		got, err = reader.TaxInputs(ctx, moralA, 2026)
		return err
	})
	require.NoError(t, err)
	assert.Empty(t, got, "RLS: another user reads nothing")
}

// The catalog texts of the regimes a taxpayer holds or held, from the row of
// its person type (626 PF and 626 PM differ), under RLS.
func TestStore_RegimeTexts(t *testing.T) {
	e := newEnv(t)
	var reader usecase.RegimeTextReader = e.store

	var got map[fiscal.RegimeCode]usecase.RegimeText
	err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		var err error
		got, err = reader.RegimeTexts(ctx, moralA)
		return err
	})
	require.NoError(t, err)

	require.Len(t, got, 2, "the ended 626 and the open 601")
	assert.Equal(t, "General de Ley Personas Morales", got["601"].Name)
	assert.Empty(t, got["601"].Explanation, "a full regime has no explanation")
	assert.NotEmpty(t, got["601"].InputLabels[domain.InputCoeficienteUtilidad])
	assert.NotEmpty(t, got["626"].InputLabels[domain.InputPriorProvisionalPayments], "626 PM takes prior payments")

	err = e.txm.ReadOnly(t.Context(), userB, func(ctx context.Context) error {
		var err error
		got, err = reader.RegimeTexts(ctx, moralA)
		return err
	})
	require.NoError(t, err)
	assert.Empty(t, got, "RLS: another user's taxpayer holds nothing")
}

// A partial regime's catalog explanation reaches the estimate.
func TestStore_RegimeTextsCarryTheExplanation(t *testing.T) {
	e := newEnv(t)
	const rif = "0199d4a2-5555-7000-8000-0000000000a3"
	e.seedTaxpayer(t, userA, rif, "fisica", regime{"621", "2022-01", nil})

	var got map[fiscal.RegimeCode]usecase.RegimeText
	err := e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		var err error
		got, err = e.store.RegimeTexts(ctx, rif)
		return err
	})
	require.NoError(t, err)

	assert.NotEmpty(t, got["621"].Name)
	assert.NotEmpty(t, got["621"].Explanation)
}

type refuseAll struct{}

func (refuseAll) EnsureWritable(context.Context, string) error {
	return apperr.New(apperr.CodeEntitlementRequired, "Read-only.")
}

// waitForLockWaiters polls until n sessions of this database wait on a lock;
// false after a few seconds (a missing lock never makes them wait).
func waitForLockWaiters(t *testing.T, e *env, n int) bool {
	t.Helper()
	const query = "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'"
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		require.NoError(t, e.app.QueryRowContext(t.Context(), query).Scan(&waiting)) // same role: wait events are visible
		if waiting >= n {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
