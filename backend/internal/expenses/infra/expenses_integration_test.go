//go:build integration

package infra_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

const folio = "3A9C7E21-5B4D-4F6A-8E2C-9D1B7A3F5C60"

// expenseID is the n-th fixture expense id.
func expenseID(n int) string { return fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", n) }

// software is a 612 software expense of 500.00 + IVA 80.00 by card, paid on.
func software(t *testing.T, n int, on string) usecase.CreateInput {
	t.Helper()
	return usecase.CreateInput{
		UserID: userA, TaxpayerID: taxpayerA, ID: expenseID(n), Regime: "612", Description: "Suscripción",
		Draft: domain.Draft{
			Category: "software", Amount: 500_00, IVARate: fiscal.IVARate16, IVA: 80_00,
			PaymentMethod: domain.PaymentCardDebit, PaidOn: date(t, on), HasCFDI: true,
		},
	}
}

// requireCode asserts err is an *apperr.Error with code.
func requireCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	_ = appErrOf(t, err, code)
}

// appErrOf is requireCode returning the error for further assertions.
func appErrOf(t *testing.T, err error, code apperr.Code) *apperr.Error {
	t.Helper()
	require.Error(t, err)
	e, ok := err.(*apperr.Error) //nolint:errorlint // the use cases return the *apperr.Error itself
	require.Truef(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, e.Code, e.Error())
	return e
}

func TestCreate_StoresTheExpenseAsTheCaller(t *testing.T) {
	e := newEnv(t)
	in := software(t, 1, "2026-09-12")
	in.Draft.CFDIUUID, _ = fiscal.ParseCFDIUUID(folio)

	got, err := e.uc.Create(t.Context(), in)

	require.NoError(t, err)
	assert.True(t, got.Created)
	assert.True(t, got.Expense.Deductible)
	assert.Empty(t, got.Expense.Warnings)
	assert.False(t, got.Expense.CreatedAt.IsZero())
	var (
		owner, cfdi, description string
		deductible, overridden   bool
		paidOn                   time.Time
	)
	queryAs(t, e.app, userA, `SELECT user_id, cfdi_uuid, description, deductible, deductible_overridden, paid_on
		FROM expenses WHERE id = $1`, []any{expenseID(1)}, &owner, &cfdi, &description, &deductible, &overridden, &paidOn)
	assert.Equal(t, userA, owner)
	assert.Equal(t, folio, cfdi)
	assert.Equal(t, "Suscripción", description)
	assert.True(t, deductible)
	assert.False(t, overridden)
	assert.Equal(t, "2026-09-12", paidOn.Format("2006-01-02"))
	assert.Equal(t, 1, expensesCreated(t, e.app, userA))

	read, err := e.uc.Get(t.Context(), userA, taxpayerA, expenseID(1))
	require.NoError(t, err)
	got.Expense.Warnings = []domain.Warning{}
	assert.Equal(t, got.Expense, read)
}

func TestCreate_TrialCapIsFiveExpensesPerUserAcrossTaxpayers(t *testing.T) {
	e := newEnv(t)
	for n := 1; n <= access.TrialRecordCap; n++ {
		in := software(t, n, "2026-09-12")
		if n == access.TrialRecordCap {
			in.TaxpayerID, in.Regime = taxpayerR, "626"
		}
		_, err := e.uc.Create(t.Context(), in)
		require.NoError(t, err)
	}

	_, err := e.uc.Create(t.Context(), software(t, 6, "2026-09-12"))

	refused := appErrOf(t, err, apperr.CodeEntitlementRequired)
	require.NotEmpty(t, refused.Fields)
	assert.Equal(t, "expenses_remaining", refused.Fields[0].Field)
	assert.Equal(t, access.TrialRecordCap, expensesCreated(t, e.app, userA))
	_, err = e.uc.Get(t.Context(), userA, taxpayerA, expenseID(6))
	requireCode(t, err, apperr.CodeNotFound)
}

// US-E3: at 4/5, three concurrent creates give exactly one 201. The usage
// row is held until the three creates wait on it, so they really race for
// the last slot past every other check.
func TestCreate_ConcurrentCreatesNeverOvershootTheCap(t *testing.T) {
	e := newEnv(t)
	for n := 1; n <= access.TrialRecordCap-1; n++ {
		_, err := e.uc.Create(t.Context(), software(t, n, "2026-09-12"))
		require.NoError(t, err)
	}
	const racers = 3
	release := holdUsageRow(t, e.app, userA)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		created int
		capped  int
	)
	for i := range racers {
		wg.Go(func() {
			_, err := e.uc.Create(context.Background(), software(t, 10+i, "2026-09-12"))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				created++
			case isCode(err, apperr.CodeEntitlementRequired):
				capped++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
	waitForLockWaiters(t, e.app, racers)
	release()
	wg.Wait()

	assert.Equal(t, 1, created)
	assert.Equal(t, racers-1, capped)
	assert.Equal(t, access.TrialRecordCap, expensesCreated(t, e.app, userA))
}

func isCode(err error, code apperr.Code) bool {
	e, ok := err.(*apperr.Error) //nolint:errorlint // the use cases return the *apperr.Error itself
	return ok && e.Code == code
}

// holdUsageRow locks userID's user_usage row in an open transaction until
// the returned func ends it.
func holdUsageRow(t *testing.T, app *sql.DB, userID string) func() {
	t.Helper()
	holder, err := app.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, err = holder.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", userID)
	require.NoError(t, err)
	_, err = holder.ExecContext(t.Context(), "SELECT 1 FROM user_usage WHERE user_id = $1 FOR UPDATE", userID)
	require.NoError(t, err)
	var once sync.Once
	release := func() { once.Do(func() { require.NoError(t, holder.Rollback()) }) }
	t.Cleanup(release)
	return release
}

// waitForLockWaiters waits until n sessions of this database wait on a lock.
func waitForLockWaiters(t *testing.T, app *sql.DB, n int) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting int
		err := app.QueryRowContext(t.Context(), `SELECT count(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`).Scan(&waiting)
		return err == nil && waiting >= n
	}, 10*time.Second, 10*time.Millisecond)
}

func TestCreate_IdempotentRetries(t *testing.T) {
	e := newEnv(t)
	first, err := e.uc.Create(t.Context(), software(t, 1, "2026-09-12"))
	require.NoError(t, err)

	again, err := e.uc.Create(t.Context(), software(t, 1, "2026-09-12"))
	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Equal(t, first.Expense.CreatedAt, again.Expense.CreatedAt)

	changed := software(t, 1, "2026-09-12")
	changed.Description = "otra"
	_, err = e.uc.Create(t.Context(), changed)
	reused := appErrOf(t, err, apperr.CodeConflict)
	require.NotEmpty(t, reused.Fields)
	assert.Equal(t, usecase.DetailIDReused, reused.Fields[0].Code)

	assert.Equal(t, 1, expensesCreated(t, e.app, userA), "no retry spends a slot")
}

// A double submit (same id, concurrently) answers the original to the
// second request: the id lock serializes them and only one slot is spent.
func TestCreate_ConcurrentDoubleSubmitCreatesOnce(t *testing.T) {
	e := newEnv(t)
	const submits = 4
	results := make(chan usecase.CreateResult, submits)
	var wg sync.WaitGroup
	for range submits {
		wg.Go(func() {
			got, err := e.uc.Create(context.Background(), software(t, 1, "2026-09-12"))
			assert.NoError(t, err)
			results <- got
		})
	}
	wg.Wait()
	close(results)

	created := 0
	for r := range results {
		if r.Created {
			created++
		}
		assert.Equal(t, expenseID(1), r.Expense.ID)
	}
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, expensesCreated(t, e.app, userA))
}

func TestCrossUser_IsolationAnswers404And409(t *testing.T) {
	e := newEnv(t)
	_, err := e.uc.Create(t.Context(), software(t, 1, "2026-09-12"))
	require.NoError(t, err)

	t.Run("another user's id on create is a bare 409", func(t *testing.T) {
		in := software(t, 1, "2026-09-12")
		in.UserID, in.TaxpayerID = userB, taxpayerB
		_, err := e.uc.Create(t.Context(), in)
		conflict := appErrOf(t, err, apperr.CodeConflict)
		assert.Empty(t, conflict.Fields)
		assert.Equal(t, 0, expensesCreated(t, e.app, userB), "the refused insert rolled its slot back")
	})
	t.Run("reads and writes of another user's expense are 404", func(t *testing.T) {
		_, err := e.uc.Get(t.Context(), userB, taxpayerB, expenseID(1))
		requireCode(t, err, apperr.CodeNotFound)
		_, err = e.uc.Update(t.Context(), usecase.UpdateInput{UserID: userB, TaxpayerID: taxpayerB, ExpenseID: expenseID(1)})
		requireCode(t, err, apperr.CodeNotFound)
		requireCode(t, e.uc.Delete(t.Context(), userB, taxpayerB, expenseID(1)), apperr.CodeNotFound)
		_, err = e.uc.Restore(t.Context(), userB, taxpayerB, expenseID(1))
		requireCode(t, err, apperr.CodeNotFound)
	})
	t.Run("another user's taxpayer is 404", func(t *testing.T) {
		_, err := e.uc.List(t.Context(), usecase.ListInput{UserID: userB, TaxpayerID: taxpayerA, Period: "2026-09"})
		requireCode(t, err, apperr.CodeNotFound)
		_, err = e.uc.Get(t.Context(), userB, taxpayerA, expenseID(1))
		requireCode(t, err, apperr.CodeNotFound)
		_, err = e.uc.Categories(t.Context(), usecase.CategoriesInput{UserID: userB, TaxpayerID: taxpayerA, Regime: "612"})
		requireCode(t, err, apperr.CodeNotFound)
		in := software(t, 2, "2026-09-12")
		in.UserID = userB
		_, err = e.uc.Create(t.Context(), in)
		requireCode(t, err, apperr.CodeNotFound)
	})
	var stillDeleted sql.NullTime
	queryAs(t, e.app, userA, "SELECT deleted_at FROM expenses WHERE id = $1", []any{expenseID(1)}, &stillDeleted)
	assert.False(t, stillDeleted.Valid)
}

// DECIDED: restore never consumes a slot (no double count), even at 5/5,
// and a read-only user may restore their own expense.
func TestRestore_NeverCountsTwice(t *testing.T) {
	e := newEnv(t)
	for n := 1; n <= access.TrialRecordCap; n++ {
		_, err := e.uc.Create(t.Context(), software(t, n, "2026-09-12"))
		require.NoError(t, err)
	}
	require.NoError(t, e.uc.Delete(t.Context(), userA, taxpayerA, expenseID(1)))

	restored, err := e.uc.Restore(t.Context(), userA, taxpayerA, expenseID(1))
	require.NoError(t, err)
	assert.Equal(t, expenseID(1), restored.ID)
	assert.Equal(t, access.TrialRecordCap, expensesCreated(t, e.app, userA))

	_, err = e.uc.Create(t.Context(), software(t, 6, "2026-09-12"))
	requireCode(t, err, apperr.CodeEntitlementRequired)

	readOnly := e.wire(t, readOnlyAccess(e.txm))
	require.NoError(t, readOnly.uc.Delete(t.Context(), userA, taxpayerA, expenseID(2)))
	_, err = readOnly.uc.Restore(t.Context(), userA, taxpayerA, expenseID(2))
	require.NoError(t, err, "a read-only user restores their own expense")
	_, err = readOnly.uc.Update(t.Context(), usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: expenseID(2)})
	requireCode(t, err, apperr.CodeEntitlementRequired)
	assert.Equal(t, access.TrialRecordCap, expensesCreated(t, e.app, userA))
}

func TestDeleteRestore_LeavesAndReturnsToTheList(t *testing.T) {
	e := newEnv(t)
	for n := 1; n <= 3; n++ {
		_, err := e.uc.Create(t.Context(), software(t, n, fmt.Sprintf("2026-09-%02d", n)))
		require.NoError(t, err)
	}
	list := usecase.ListInput{UserID: userA, TaxpayerID: taxpayerA, Period: "2026-09"}

	require.NoError(t, e.uc.Delete(t.Context(), userA, taxpayerA, expenseID(2)))
	require.NoError(t, e.uc.Delete(t.Context(), userA, taxpayerA, expenseID(2)), "idempotent")
	page, err := e.uc.List(t.Context(), list)
	require.NoError(t, err)
	assert.Equal(t, 2, page.Totals.Count)
	_, err = e.uc.Get(t.Context(), userA, taxpayerA, expenseID(2))
	requireCode(t, err, apperr.CodeNotFound)

	_, err = e.uc.Restore(t.Context(), userA, taxpayerA, expenseID(2))
	require.NoError(t, err)
	page, err = e.uc.List(t.Context(), list)
	require.NoError(t, err)
	assert.Equal(t, 3, page.Totals.Count)
}

func TestArchivedTaxpayer_FreezesWritesButStaysReadable(t *testing.T) {
	e := newEnv(t)
	_, err := e.uc.Create(t.Context(), software(t, 1, "2026-09-12"))
	require.NoError(t, err)
	execAs(t, e.app, userA, "UPDATE taxpayers SET archived_at = now() WHERE id = $1", taxpayerA)

	_, err = e.uc.Create(t.Context(), software(t, 2, "2026-09-12"))
	requireCode(t, err, apperr.CodeBusinessRuleViolation)
	requireCode(t, e.uc.Delete(t.Context(), userA, taxpayerA, expenseID(1)), apperr.CodeBusinessRuleViolation)
	page, err := e.uc.List(t.Context(), usecase.ListInput{UserID: userA, TaxpayerID: taxpayerA, Period: "2026-09"})
	require.NoError(t, err)
	assert.Len(t, page.Items, 1)
	_, err = e.uc.Categories(t.Context(), usecase.CategoriesInput{UserID: userA, TaxpayerID: taxpayerA, Regime: "612"})
	require.NoError(t, err)
}

func TestCreate_RegimeWindowAndCategoryRulesPerRegime(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(*usecase.CreateInput)
		deductible bool
		warnings   []string
	}{
		{"612 software", func(*usecase.CreateInput) {}, true, nil},
		{"RESICO física never deducts", func(in *usecase.CreateInput) { in.TaxpayerID, in.Regime = taxpayerR, "626" }, false, nil},
		{"606 blind 35 %: maintenance", func(in *usecase.CreateInput) { in.Regime, in.Draft.Category = "606", "mantenimiento" }, false, nil},
		{"606 blind 35 %: predial", func(in *usecase.CreateInput) { in.Regime, in.Draft.Category = "606", "predial" }, true, nil},
		{"601 moral withholding fees", func(in *usecase.CreateInput) {
			in.TaxpayerID, in.Regime, in.Draft.Category = taxpayerM, "601", "servicios_profesionales"
			in.Draft.Amount, in.Draft.IVA = 10_000_00, 1_600_00
			in.Draft.Withheld = fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67}
		}, true, nil},
		{"cash over $2,000", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.PaymentMethod = "papeleria", domain.PaymentCash
			in.Draft.Amount, in.Draft.IVA = 2_000_00, 320_00
		}, false, []string{"cash_over_threshold"}},
		{"fuel in cash", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.PaymentMethod = "combustible", domain.PaymentCash
			in.Draft.Amount, in.Draft.IVA = 300_00, 48_00
		}, false, []string{"cash_over_threshold"}},
		{"cash claimed deductible", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.PaymentMethod = "papeleria", domain.PaymentCash
			in.Draft.Amount, in.Draft.IVA = 2_000_00, 320_00
			in.Draft.Deductible = domain.ClaimDeductible
		}, true, []string{"cash_over_threshold", "iva_not_creditable_cash_payment"}},
		{"personal expense claimed", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.Deductible = "gastos_personales", domain.ClaimDeductible
		}, true, []string{"category_not_deductible"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newEnv(t)
			in := software(t, 1, "2026-09-12")
			c.mutate(&in)

			got, err := e.uc.Create(t.Context(), in)

			require.NoError(t, err)
			assert.Equal(t, c.deductible, got.Expense.Deductible)
			var codes []string
			for _, w := range got.Expense.Warnings {
				codes = append(codes, string(w.Code))
			}
			assert.Equal(t, c.warnings, codes)
		})
	}
}

func TestCreate_Refusals(t *testing.T) {
	e := newEnv(t)
	cases := []struct {
		name   string
		mutate func(*usecase.CreateInput)
		code   apperr.Code
		detail string
	}{
		{"regime not held", func(in *usecase.CreateInput) { in.Regime = "626" }, apperr.CodeBusinessRuleViolation, "regime_not_active"},
		{"before the regime starts", func(in *usecase.CreateInput) { in.Draft.PaidOn = date(t, "2025-12-31") }, apperr.CodeBusinessRuleViolation, "regime_not_active"},
		{"unknown category", func(in *usecase.CreateInput) { in.Draft.Category = "viajes-luna" }, apperr.CodeValidationFailed, "unknown_category"},
		{"física withholds", func(in *usecase.CreateInput) { in.Draft.Withheld.ISR = 50_00 }, apperr.CodeBusinessRuleViolation, "withholding_not_applicable"},
		{"iva mismatch", func(in *usecase.CreateInput) { in.Draft.IVA = 70_00 }, apperr.CodeBusinessRuleViolation, "iva_amount_mismatch"},
		{"paid before the first parameter set", func(in *usecase.CreateInput) {
			in.TaxpayerID, in.Regime = taxpayerOld, "612"
			in.Draft.PaidOn = date(t, "2025-06-10")
		}, apperr.CodeValidationFailed, "out_of_range"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := software(t, 1, "2026-09-12")
			c.mutate(&in)

			_, err := e.uc.Create(t.Context(), in)

			refused := appErrOf(t, err, c.code)
			require.NotEmpty(t, refused.Fields)
			assert.Equal(t, c.detail, refused.Fields[0].Code)
		})
	}
	assert.Equal(t, 0, expensesCreated(t, e.app, userA))
}

func TestCreate_DuplicateFolioWarnsAcrossTheTaxpayer(t *testing.T) {
	e := newEnv(t)
	first := software(t, 1, "2026-09-12")
	first.Draft.CFDIUUID, _ = fiscal.ParseCFDIUUID(folio)
	_, err := e.uc.Create(t.Context(), first)
	require.NoError(t, err)
	second := software(t, 2, "2026-09-13")
	second.Draft.CFDIUUID = first.Draft.CFDIUUID

	got, err := e.uc.Create(t.Context(), second)
	require.NoError(t, err)
	assert.Equal(t, []domain.Warning{{Code: domain.WarningDuplicateCFDIUUID, Field: "cfdi_uuid"}}, got.Expense.Warnings)

	require.NoError(t, e.uc.Delete(t.Context(), userA, taxpayerA, expenseID(1)))
	updated, err := e.uc.Update(t.Context(), usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: expenseID(2),
		Patch: usecase.Patch{Description: ptr("")}})
	require.NoError(t, err)
	assert.Empty(t, updated.Warnings, "a deleted expense's folio no longer counts")
}

func ptr[T any](v T) *T { return &v }

func TestUpdate_PersistsTheMergedExpense(t *testing.T) {
	e := newEnv(t)
	_, err := e.uc.Create(t.Context(), software(t, 1, "2026-09-12"))
	require.NoError(t, err)

	got, err := e.uc.Update(t.Context(), usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: expenseID(1),
		Patch: usecase.Patch{Category: ptr(domain.CategoryCode("restaurantes")), Description: ptr("")}})
	require.NoError(t, err)
	assert.False(t, got.Deductible)

	read, err := e.uc.Get(t.Context(), userA, taxpayerA, expenseID(1))
	require.NoError(t, err)
	assert.Equal(t, domain.CategoryCode("restaurantes"), read.Category)
	assert.Empty(t, read.Description)
	assert.False(t, read.Deductible)
	assert.True(t, read.UpdatedAt.After(read.CreatedAt) || read.UpdatedAt.Equal(read.CreatedAt))

	overridden, err := e.uc.Update(t.Context(), usecase.UpdateInput{UserID: userA, TaxpayerID: taxpayerA, ExpenseID: expenseID(1),
		Patch: usecase.Patch{Deductible: ptr(domain.ClaimDeductible)}})
	require.NoError(t, err)
	assert.Equal(t, domain.SourceUser, overridden.Source)
	var flag bool
	queryAs(t, e.app, userA, "SELECT deductible_overridden FROM expenses WHERE id = $1", []any{expenseID(1)}, &flag)
	assert.True(t, flag)
}

// unlimitedAccess stands in for a pro plan: the list fixtures need more than
// a trial's five expenses.
type unlimitedAccess struct{}

func (unlimitedAccess) EnsureWritable(context.Context, string) error { return nil }

func (unlimitedAccess) ReserveRecord(context.Context, string, access.RecordKind) error { return nil }

func TestList_KeysetPagesTotalsAndRegimeFilter(t *testing.T) {
	e := newEnv(t, func(d *usecase.Deps) { d.Access = unlimitedAccess{} })
	// Two expenses share each day, so the id breaks the ties.
	for n := 1; n <= 6; n++ {
		_, err := e.uc.Create(t.Context(), software(t, n, fmt.Sprintf("2026-09-%02d", (n+1)/2)))
		require.NoError(t, err)
	}
	predial := software(t, 7, "2026-09-30")
	predial.Regime, predial.Draft.Category = "606", "predial"
	predial.Draft.Amount, predial.Draft.IVARate, predial.Draft.IVA = 1_234_00, fiscal.IVAExempt, 0
	_, err := e.uc.Create(t.Context(), predial)
	require.NoError(t, err)
	october := software(t, 8, "2026-10-01")
	_, err = e.uc.Create(t.Context(), october)
	require.NoError(t, err)

	var ids []string
	cursor := ""
	for {
		page, err := e.uc.List(t.Context(), usecase.ListInput{UserID: userA, TaxpayerID: taxpayerA, Period: "2026-09", Cursor: cursor, Limit: 3})
		require.NoError(t, err)
		assert.Equal(t, usecase.Totals{Count: 7, Amount: 6*500_00 + 1_234_00, IVA: 6 * 80_00}, page.Totals)
		for _, item := range page.Items {
			ids = append(ids, item.ID)
		}
		if cursor = page.NextCursor; cursor == "" {
			break
		}
	}
	assert.Equal(t, []string{
		expenseID(7), expenseID(6), expenseID(5), expenseID(4), expenseID(3), expenseID(2), expenseID(1),
	}, ids)

	only, err := e.uc.List(t.Context(), usecase.ListInput{UserID: userA, TaxpayerID: taxpayerA, Period: "2026-09", Regime: "606"})
	require.NoError(t, err)
	require.Len(t, only.Items, 1)
	assert.Equal(t, usecase.Totals{Count: 1, Amount: 1_234_00}, only.Totals)
}

func TestCategories_ResolvedPerRegimePersonTypeAndOptions(t *testing.T) {
	e := newEnv(t)
	resolve := func(taxpayerID, regime string) map[domain.CategoryCode]usecase.Category {
		t.Helper()
		list, err := e.uc.Categories(t.Context(), usecase.CategoriesInput{
			UserID: userA, TaxpayerID: taxpayerID, Regime: fiscal.RegimeCode(regime), On: "2026-09-12",
		})
		require.NoError(t, err)
		require.Len(t, list.Items, 26, "the whole catalog")
		assert.Equal(t, domain.CategoryCode("mercancia"), list.Items[0].Code, "catalog order")
		assert.Equal(t, money.Cents(2_000_00), list.CashPaymentLimit)
		assert.Equal(t, []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}, list.IVARatesAllowed)
		assert.Equal(t, domain.Legend, list.Legend)
		out := map[domain.CategoryCode]usecase.Category{}
		for _, c := range list.Items {
			out[c.Code] = c
		}
		return out
	}

	business := resolve(taxpayerA, "612")
	assert.Equal(t, domain.Deductible, business["software"].Deductibility)
	assert.True(t, business["software"].DefaultDeductible)
	assert.Equal(t, domain.PartiallyDeductible, business["restaurantes"].Deductibility)
	assert.False(t, business["restaurantes"].DefaultDeductible)
	assert.NotEmpty(t, business["combustible"].Note)

	blind := resolve(taxpayerA, "606")
	assert.Equal(t, domain.ISRNotApplicable, blind["mantenimiento"].Deductibility)
	assert.Equal(t, domain.Deductible, blind["predial"].Deductibility)

	resico := resolve(taxpayerR, "626")
	assert.Equal(t, domain.ISRNotApplicable, resico["software"].Deductibility)
	assert.False(t, resico["software"].DefaultDeductible)

	moral := resolve(taxpayerM, "601")
	assert.Equal(t, domain.Deductible, moral["software"].Deductibility)

	_, err := e.uc.Categories(t.Context(), usecase.CategoriesInput{UserID: userA, TaxpayerID: taxpayerA, Regime: "626", On: "2026-09-12"})
	requireCode(t, err, apperr.CodeBusinessRuleViolation)
}
