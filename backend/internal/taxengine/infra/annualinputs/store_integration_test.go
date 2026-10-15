//go:build integration

package annualinputs_test

import (
	"context"
	"database/sql"
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
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const (
	deductionID  = "0199d4b1-6666-7000-8000-000000000001"
	deductionID2 = "0199d4b1-6666-7000-8000-000000000002"
	statementID  = "0199d4b0-6666-7000-8000-000000000001"
	statementID2 = "0199d4b0-6666-7000-8000-000000000002"
)

func ptr[T any](v T) *T { return &v }

func mustDate(raw string) fiscal.Date {
	d, err := fiscal.ParseDate(raw)
	if err != nil {
		panic(err)
	}
	return d
}

func requireCode(t *testing.T, err error, code apperr.Code) []string {
	t.Helper()
	e, ok := errors.AsType[*apperr.Error](err)
	require.Truef(t, ok, "want *apperr.Error, got %v", err)
	require.Equal(t, code, e.Code, e.Error())
	out := []string{}
	for _, f := range e.Fields {
		out = append(out, f.Field+"/"+f.Code)
	}
	return out
}

func medical(id, taxpayerID string, method adomain.PaymentMethod) annual.DeductionInput {
	return annual.DeductionInput{
		ID: id, TaxpayerID: taxpayerID, Category: adomain.CategoryMedical, Amount: 4_000_000, PaidOn: mustDate("2026-06-12"),
		PaymentMethod: method, Beneficiary: adomain.BeneficiarySelf,
	}
}

func tuition(id string) annual.DeductionInput {
	return annual.DeductionInput{
		ID: id, TaxpayerID: fisicaA, Category: adomain.CategoryTuition, Amount: 3_000_000, PaidOn: mustDate("2026-08-20"),
		PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiaryDescendant,
		StudentLabel: ptr("María"), TuitionLevel: ptr(adomain.LevelPrimaria), Scholarship: ptr(money.Cents(0)),
		CFDIUUID: ptr("6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f"),
	}
}

func TestStore_TaxpayerAndRegimesUnderRLS(t *testing.T) {
	e := newEnv(t)
	e.archive(t, userA, moralA)

	err := e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
		got, err := e.store.Taxpayer(ctx, fisicaA, annual.LockWrite)
		require.NoError(t, err, "cc_app may take FOR SHARE")
		assert.Equal(t, annual.Taxpayer{PersonType: fiscal.PersonFisica}, got)
		got, err = e.store.Taxpayer(ctx, moralA, annual.LockNone)
		require.NoError(t, err)
		assert.Equal(t, annual.Taxpayer{PersonType: fiscal.PersonMoral, Archived: true}, got)
		_, err = e.store.Taxpayer(ctx, fisicaB, annual.LockNone)
		assert.ErrorIs(t, err, annual.ErrTaxpayerNotFound, "another user's taxpayer does not exist")

		regimes, err := e.store.Regimes(ctx, fisicaA)
		require.NoError(t, err)
		require.Len(t, regimes, 3)
		assert.Equal(t, []fiscal.RegimeCode{"605", "612", "625"}, []fiscal.RegimeCode{regimes[0].Code, regimes[1].Code, regimes[2].Code})
		assert.True(t, regimes[2].DefinitiveOption)
		assert.False(t, regimes[1].DefinitiveOption, "absent option = catalog default")
		assert.Equal(t, "2022-01", regimes[1].Start.String())
		assert.True(t, regimes[1].End.IsZero())
		var keys []domain.InputKey
		for _, d := range regimes[1].Definitions {
			keys = append(keys, d.Key)
		}
		assert.Contains(t, keys, adomain.InputAnnualDepreciation, "ANN-03 catalog key on 612 PF")

		none, err := e.store.Regimes(ctx, fisicaB)
		require.NoError(t, err)
		assert.Empty(t, none)

		require.NoError(t, e.store.LockRegime(ctx, fisicaA, "612"), "cc_app may take FOR NO KEY UPDATE")
		assert.ErrorIs(t, e.store.LockRegime(ctx, fisicaA, "606"), annual.ErrRegimeNotHeld)
		assert.ErrorIs(t, e.store.LockRegime(ctx, fisicaB, "612"), annual.ErrRegimeNotHeld)
		return e.store.LockList(ctx, fisicaA, annual.ListDeductions, 2026)
	})
	require.NoError(t, err)
}

func TestDeductions_RoundTripAsAppRole(t *testing.T) {
	e := newEnv(t)
	uc := e.deductions(today)

	created, err := uc.Create(t.Context(), userA, tuition(deductionID))
	require.NoError(t, err)
	assert.True(t, created.Created)
	d := created.Deduction
	assert.Equal(t, 2026, d.FiscalYear, "generated from paid_on")
	assert.Equal(t, "María", d.StudentLabel, "stored NFC")
	assert.Equal(t, "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F", d.CFDIUUID)
	assert.Equal(t, annual.NullCents{Valid: true}, d.Scholarship)
	assert.False(t, d.CreatedAt.IsZero())

	replay, err := uc.Create(t.Context(), userA, tuition(deductionID))
	require.NoError(t, err)
	assert.False(t, replay.Created)
	assert.Equal(t, d, replay.Deduction, "the stored row round-trips into equal fields")

	cash, err := uc.Create(t.Context(), userA, medical(deductionID2, fisicaA, adomain.PaymentCash))
	require.NoError(t, err)
	assert.Equal(t, annual.ExclusionCashNotDeductible, cash.Deduction.Excluded)

	list, err := uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026})
	require.NoError(t, err)
	require.Len(t, list.Items, 2)
	assert.Equal(t, deductionID2, list.Items[0].ID, "medical before tuition")
	assert.Equal(t, []annual.CategoryTotal{
		{Category: adomain.CategoryMedical, Registered: 4_000_000, Excluded: 4_000_000, Count: 1},
		{Category: adomain.CategoryTuition, Registered: 3_000_000, Count: 1},
	}, list.Totals)

	updated, err := uc.Update(t.Context(), userA, fisicaA, deductionID, annual.DeductionPatch{
		Category: ptr(adomain.CategoryDonation), DoneeType: annual.Value(adomain.DoneeGovernment),
		StudentLabel: annual.Null[string](), TuitionLevel: annual.Null[adomain.TuitionLevel](), Scholarship: annual.Null[money.Cents](),
		CFDIUUID: annual.Null[string](), PaidOn: ptr(mustDate("2026-07-03")),
	})
	require.NoError(t, err, "cc_app may update every editable column")
	assert.Equal(t, annual.DeductionFields{
		Category: adomain.CategoryDonation, Amount: 3_000_000, PaidOn: mustDate("2026-07-03"), PaymentMethod: adomain.PaymentElectronic,
		Beneficiary: adomain.BeneficiaryDescendant, DoneeType: adomain.DoneeGovernment,
	}, updated.Deduction.DeductionFields)
	assert.True(t, updated.Deduction.UpdatedAt.After(d.UpdatedAt) || updated.Deduction.UpdatedAt.Equal(d.UpdatedAt))

	require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, deductionID))
	list, err = uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026})
	require.NoError(t, err)
	assert.Len(t, list.Items, 1, "a deleted row leaves the list at once")

	restored, err := uc.Restore(t.Context(), userA, fisicaA, deductionID)
	require.NoError(t, err)
	assert.Equal(t, deductionID, restored.Deduction.ID)

	years, err := uc.List(t.Context(), userA, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026, Category: adomain.CategoryDonation})
	require.NoError(t, err)
	assert.Len(t, years.Items, 1)
}

func TestDeductions_IsolationAndIDs(t *testing.T) {
	e := newEnv(t)
	uc := e.deductions(today)
	_, err := uc.Create(t.Context(), userA, medical(deductionID, fisicaA, adomain.PaymentElectronic))
	require.NoError(t, err)

	_, err = uc.Create(t.Context(), userB, medical(deductionID, fisicaB, adomain.PaymentElectronic))
	assert.Empty(t, requireCode(t, err, apperr.CodeConflict), "another user's id: 409 without details")

	_, err = uc.Update(t.Context(), userB, fisicaA, deductionID, annual.DeductionPatch{Amount: ptr(money.Cents(1))})
	requireCode(t, err, apperr.CodeNotFound)
	_, err = uc.Update(t.Context(), userB, fisicaB, deductionID, annual.DeductionPatch{Amount: ptr(money.Cents(1))})
	requireCode(t, err, apperr.CodeNotFound)
	requireCode(t, uc.Delete(t.Context(), userB, fisicaB, deductionID), apperr.CodeNotFound)
	_, err = uc.Restore(t.Context(), userB, fisicaB, deductionID)
	requireCode(t, err, apperr.CodeNotFound)
	_, err = uc.List(t.Context(), userB, annual.DeductionQuery{TaxpayerID: fisicaA, FiscalYear: 2026})
	requireCode(t, err, apperr.CodeNotFound)

	err = e.txm.ReadOnly(t.Context(), userB, func(ctx context.Context) error {
		_, err := e.store.FindDeduction(ctx, deductionID, annual.LockNone)
		assert.ErrorIs(t, err, annual.ErrDeductionNotFound)
		rows, err := e.store.ListDeductions(ctx, fisicaA, 2026, "")
		assert.Empty(t, rows)
		n, countErr := e.store.CountLiveDeductions(ctx, fisicaA, 2026)
		assert.Zero(t, n)
		return errors.Join(err, countErr)
	})
	require.NoError(t, err)

	_, err = uc.Create(t.Context(), userA, medical(deductionID2, moralA, adomain.PaymentElectronic))
	assert.Equal(t, []string{"taxpayerId/person_type_not_allowed"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
}

func TestDeductions_ReadOnlyUser(t *testing.T) {
	e := newEnv(t)
	uc := e.deductions(today)
	_, err := uc.Create(t.Context(), userRO, medical(deductionID, fisicaRO, adomain.PaymentElectronic))
	requireCode(t, err, apperr.CodeEntitlementRequired)

	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userRO, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), `INSERT INTO personal_deductions
			(id, user_id, taxpayer_id, category, amount_cents, paid_on, payment_method, beneficiary)
			VALUES ($1, $2, $3, 'medical', 100, '2026-05-01', 'electronic', 'self')`, deductionID, userRO, fisicaRO)
		return err
	}))
	require.NoError(t, uc.Delete(t.Context(), userRO, fisicaRO, deductionID), "a read-only user may delete")
	_, err = uc.Restore(t.Context(), userRO, fisicaRO, deductionID)
	require.NoError(t, err, "and undo it")
	list, err := uc.List(t.Context(), userRO, annual.DeductionQuery{TaxpayerID: fisicaRO, FiscalYear: 2026})
	require.NoError(t, err)
	assert.Len(t, list.Items, 1)
	_, err = uc.Update(t.Context(), userRO, fisicaRO, deductionID, annual.DeductionPatch{Amount: ptr(money.Cents(5))})
	requireCode(t, err, apperr.CodeEntitlementRequired)
}

// The list lock serializes two writers of one taxpayer and year: the
// second waits (an ungranted advisory lock) until the first commits.
func TestStore_LockListSerializesWriters(t *testing.T) {
	e := newEnv(t)
	held, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseFirst) // a failed assertion must not leave the first writer blocked
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- e.txm.ReadWrite(context.WithoutCancel(t.Context()), userA, func(ctx context.Context) error {
			if err := e.store.LockList(ctx, fisicaA, annual.ListDeductions, 2026); err != nil {
				return err
			}
			close(held)
			<-release
			return nil
		})
	}()
	<-held
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- e.txm.ReadWrite(t.Context(), userA, func(ctx context.Context) error {
			return e.store.LockList(ctx, fisicaA, annual.ListDeductions, 2026)
		})
	}()
	require.Eventually(t, func() bool {
		var waiting int
		err := e.app.QueryRowContext(t.Context(),
			"SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND NOT granted AND classid = 1095650865::oid").Scan(&waiting)
		return err == nil && waiting == 1
	}, 10*time.Second, 20*time.Millisecond, "the second writer waits on the first's list lock")
	select {
	case err := <-secondDone:
		t.Fatalf("the second writer did not wait: %v", err)
	default:
	}
	releaseFirst()
	require.NoError(t, <-firstDone)
	require.NoError(t, <-secondDone)
}

func TestSalary_RoundTripAndMoveYear(t *testing.T) {
	e := newEnv(t)
	uc := e.salary(today)

	created, err := uc.Create(t.Context(), userA, annual.SalaryInput{
		ID: statementID, TaxpayerID: fisicaA, FiscalYear: 2026, EmployerLabel: ptr("Empleo principal"),
		Accumulable: 30_000_000, ISRWithheld: 4_000_000, Simultaneous: true,
	})
	require.NoError(t, err)
	assert.Empty(t, created.Warnings, "605 held in 2026")
	_, err = uc.Create(t.Context(), userA, annual.SalaryInput{
		ID: statementID2, TaxpayerID: fisicaA, FiscalYear: 2026, Accumulable: 15_000_000, ISRWithheld: 800_000, Simultaneous: true,
	})
	require.NoError(t, err)

	list, err := uc.List(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	require.Len(t, list.Items, 2)
	assert.Equal(t, statementID, list.Items[0].ID)
	assert.Empty(t, list.Items[1].EmployerLabel, "an absent label is NULL")
	assert.Equal(t, annual.SalaryTotals{Accumulable: 45_000_000, ISRWithheld: 4_800_000}, list.Totals)

	next := e.salary(time.Date(2027, time.March, 1, 12, 0, 0, 0, clock.Location()))
	moved, err := next.Update(t.Context(), userA, fisicaA, statementID2,
		annual.SalaryPatch{FiscalYear: ptr(2027), EmployerLabel: annual.Value("Segundo empleo"), LeftBeforeYearEnd: ptr(true)})
	require.NoError(t, err, "cc_app may move fiscal_year (20261017102000)")
	assert.Equal(t, 2027, moved.Statement.FiscalYear)
	assert.Equal(t, "Segundo empleo", moved.Statement.EmployerLabel)

	_, err = uc.List(t.Context(), userB, fisicaA, 2026)
	requireCode(t, err, apperr.CodeNotFound)
	requireCode(t, uc.Delete(t.Context(), userB, fisicaB, statementID), apperr.CodeNotFound)

	require.NoError(t, uc.Delete(t.Context(), userA, fisicaA, statementID))
	requireCode(t, uc.Delete(t.Context(), userA, fisicaA, statementID), apperr.CodeNotFound)

	err = e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		n, err := e.store.CountStatements(ctx, fisicaA, 2026)
		assert.Zero(t, n)
		return err
	})
	require.NoError(t, err)
}

func TestInputs_UpsertMirrorAndIdempotence(t *testing.T) {
	e := newEnv(t)
	uc := e.inputs()
	e.exec(t, userA, `INSERT INTO taxpayer_tax_inputs (user_id, taxpayer_id, regime_code, fiscal_year, input_key, value_cents)
		VALUES ($1, '`+fisicaA+`', '612', 2026, 'annual_depreciation', 1250000)`)

	empty, err := uc.Get(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	assert.Nil(t, empty.UpdatedAt)
	var mirrored []string
	for _, in := range empty.RegimeTaxInputs {
		value := "null"
		if in.Value != nil {
			value = in.Value.String()
		}
		mirrored = append(mirrored, string(in.Regime)+":"+string(in.Key)+"="+value)
	}
	assert.Contains(t, mirrored, "612:annual_depreciation=12500.00")

	facts := annual.Facts{PriorYearAccumulableIncome: ptr(money.Cents(68_000_000)), ToldEmployerWillFile: true}
	put, err := uc.Put(t.Context(), userA, fisicaA, 2026, facts)
	require.NoError(t, err)
	require.NotNil(t, put.UpdatedAt)
	assert.Equal(t, facts, put.Facts)

	again, err := uc.Put(t.Context(), userA, fisicaA, 2026, facts)
	require.NoError(t, err)
	assert.Equal(t, put.UpdatedAt, again.UpdatedAt, "the same body twice leaves the same state")

	cleared, err := uc.Put(t.Context(), userA, fisicaA, 2026, annual.Facts{})
	require.NoError(t, err)
	assert.Nil(t, cleared.Facts.PriorYearAccumulableIncome, "NULL is unknown")

	_, err = uc.Get(t.Context(), userB, fisicaA, 2026)
	requireCode(t, err, apperr.CodeNotFound)
	err = e.txm.ReadOnly(t.Context(), userB, func(ctx context.Context) error {
		_, found, err := e.store.Facts(ctx, fisicaA, 2026)
		assert.False(t, found, "RLS")
		return err
	})
	require.NoError(t, err)
}

func TestFiled_ReplaceReadAndRules(t *testing.T) {
	e := newEnv(t)
	uc := e.filed(today)
	req := annual.FiledRequest{TaxpayerID: fisicaA, RegimeCode: "612", FiscalYear: 2026}

	put, err := uc.Put(t.Context(), userA, req, []annual.FiledEntry{{Month: 1, Amount: 500_000}, {Month: 2, Amount: 0}})
	require.NoError(t, err)
	require.NotNil(t, put.UpdatedAt)
	assert.Equal(t, money.Cents(500_000), *put.Months[0].Filed)
	assert.Equal(t, money.Cents(0), *put.Months[1].Filed)
	assert.Nil(t, put.Months[2].Filed)

	put, err = uc.Put(t.Context(), userA, req, []annual.FiledEntry{{Month: 3, Amount: 7}})
	require.NoError(t, err)
	assert.Nil(t, put.Months[0].Filed, "replace, not merge")

	got, err := uc.Get(t.Context(), userA, req)
	require.NoError(t, err)
	assert.Equal(t, put, got)

	err = e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		all, err := e.store.FiledPayments(ctx, fisicaA, 2026)
		assert.Equal(t, map[fiscal.RegimeCode]map[time.Month]money.Cents{"612": {time.March: 7}}, all)
		return err
	})
	require.NoError(t, err)
	err = e.txm.ReadOnly(t.Context(), userB, func(ctx context.Context) error {
		all, err := e.store.FiledPayments(ctx, fisicaA, 2026)
		assert.Empty(t, all, "RLS")
		return err
	})
	require.NoError(t, err)

	_, err = uc.Get(t.Context(), userA, annual.FiledRequest{TaxpayerID: fisicaA, RegimeCode: "625", FiscalYear: 2026})
	assert.Equal(t, []string{"regimeCode/payments_not_applicable"}, requireCode(t, err, apperr.CodeBusinessRuleViolation), "625 definitive")
	_, err = uc.Put(t.Context(), userB, req, nil)
	requireCode(t, err, apperr.CodeNotFound)

	cleared, err := uc.Put(t.Context(), userA, req, nil)
	require.NoError(t, err)
	assert.Nil(t, cleared.UpdatedAt)
}

func TestYearInputsReader(t *testing.T) {
	e := newEnv(t)
	_, err := e.deductions(today).Create(t.Context(), userA, medical(deductionID, fisicaA, adomain.PaymentElectronic))
	require.NoError(t, err)
	_, err = e.deductions(today).Create(t.Context(), userA, medical(deductionID2, fisicaA, adomain.PaymentElectronic))
	require.NoError(t, err)
	require.NoError(t, e.deductions(today).Delete(t.Context(), userA, fisicaA, deductionID2))
	_, err = e.salary(today).Create(t.Context(), userA, annual.SalaryInput{ID: statementID, TaxpayerID: fisicaA, FiscalYear: 2026, Accumulable: 1})
	require.NoError(t, err)

	err = e.txm.ReadOnly(t.Context(), userA, func(ctx context.Context) error {
		got, err := e.store.YearInputs(ctx, fisicaA, 2026)
		require.NoError(t, err)
		assert.Len(t, got.Deductions, 1, "live rows only")
		assert.NoError(t, got.Deductions[0].Validate())
		assert.Equal(t, []adomain.SalaryStatement{{Accumulable: 1}}, got.SalaryStatements)
		assert.Equal(t, adomain.PFAnnualInputs{}, got.Facts, "defaults when never saved")
		return nil
	})
	require.NoError(t, err)

	err = e.txm.ReadOnly(t.Context(), userB, func(ctx context.Context) error {
		got, err := e.store.YearInputs(ctx, fisicaA, 2026)
		assert.Empty(t, got.Deductions)
		assert.Empty(t, got.SalaryStatements)
		return err
	})
	require.NoError(t, err)
}
