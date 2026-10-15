//go:build integration

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The incomes use cases (M3-04, M3-05) over the real store, the real
// taxpayers module and access policy, wired by the composition root as
// cc_app under RLS. Fixture ids are this suite's own; every test gets its
// own cloned database.
const (
	incUserA = "0199d4a2-0000-7000-8000-0000000003a1"
	incUserB = "0199d4a2-0000-7000-8000-0000000003b1"

	incRFCFisica = "PEXJ900215HK8"
	incRFCOther  = "AAAA010101AA0"
	incRFCMoral  = "A&B020304XY2"
)

// incToday is the incomes clock: 13 October 2026, noon in Mexico City.
var incToday = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

// incRegisterSQL is sign-up as cc_app does it: the user, its counters and a
// 7-day trial starting at the database's now().
var incRegisterSQL = []string{
	`WITH v(id) AS (SELECT $1::uuid)
	 INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
	 SELECT id, sha256(convert_to(id::text, 'UTF8')),
	        '\x0001'::bytea || sha256(convert_to(id::text, 'UTF8')) || sha256(convert_to(id::text, 'UTF8')),
	        1, '$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$aGFzaGhhc2hoYXNo', 'Ana'
	 FROM v`,
	"INSERT INTO user_usage (user_id) VALUES ($1)",
	"INSERT INTO entitlements (user_id, kind, source, starts_at, ends_at) VALUES ($1, 'trial', 'registration', now(), now() + interval '168 hours')",
}

type incomesEnv struct {
	app       *sql.DB
	taxpayers *taxpayers.Taxpayers
	incomes   *incomes.Incomes
	// readOnly is the same wiring with a clock a month after the trial.
	readOnly *incomes.Incomes
}

func newIncomesEnv(t *testing.T, users ...string) *incomesEnv {
	t.Helper()
	app := pgtest.NewDatabase(t).Connect(t, pgtest.RoleApp)
	for _, u := range users {
		require.NoError(t, pgtest.AsUser(t.Context(), app, u, func(tx *sql.Tx) error {
			for _, statement := range incRegisterSQL {
				if _, err := tx.ExecContext(t.Context(), statement, u); err != nil {
					return err
				}
			}
			return nil
		}))
	}
	logger := slog.New(slog.DiscardHandler)
	deps, err := newTaxpayersDeps(taxpayersWiring{env: config.EnvDev, keys: config.Taxpayers{RFCKeyID: 1}, tx: newAppTx(t, app), logger: logger})
	require.NoError(t, err)
	rates, err := newRateCatalog()
	require.NoError(t, err)
	wire := func(accessClock clock.Clock) *incomes.Incomes {
		m, err := newIncomesModule(incomesWiring{
			tx: deps.Tx, taxpayers: taxpayers.NewQueries(deps), rates: rates,
			clock: clock.NewFixed(incToday), accessClock: accessClock, logger: logger,
		})
		require.NoError(t, err)
		return m.incomes
	}
	return &incomesEnv{
		app: app, taxpayers: taxpayers.NewTaxpayers(deps),
		incomes:  wire(clock.NewSystem()),
		readOnly: wire(clock.NewFixed(time.Now().Add(30 * 24 * time.Hour))),
	}
}

// taxpayer creates a taxpayer of user holding regime since January 2026.
func (e *incomesEnv) taxpayer(t *testing.T, user, personType, rfc, regime string) string {
	t.Helper()
	created, err := e.taxpayers.Create(t.Context(), user, taxpayers.CreateInput{
		PersonType: personType, RFC: rfc, DisplayName: "Prueba",
		Regimes: []taxpayers.RegimeInput{{Code: regime, Start: "2026-01"}},
	})
	require.NoError(t, err)
	return created.Taxpayer.ID
}

func incDate(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

// fees is a 612 income from the general public on day, for amount centavos
// (exempt, so any amount is consistent).
func fees(t *testing.T, day string, amount money.Cents) incomes.IncomeInput {
	return incomes.IncomeInput{
		Regime: "612", Counterparty: fiscal.ReceptorPFNoActivityOrPublic, Amount: amount, IVARate: fiscal.IVAExempt,
		Dates: domain.Dates{CollectedOn: incDate(t, day)},
	}
}

func incID(n int) string { return fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", n) }

func requireAppCode(t *testing.T, err error, code apperr.Code) []apperr.FieldError {
	t.Helper()
	require.Error(t, err)
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, app.Code, "error: %v", err)
	return app.Fields
}

func (e *incomesEnv) incomesCreated(t *testing.T, user string) int {
	t.Helper()
	var n int
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, user, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), "SELECT incomes_created FROM user_usage WHERE user_id = $1", user).Scan(&n)
	}))
	return n
}

func TestIncomes_CreateGetAndDerivedBookingDate(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	pm := e.taxpayer(t, incUserA, "moral", incRFCMoral, "601")
	in := incomes.IncomeInput{
		Regime: "601", Counterparty: fiscal.ReceptorPM, Amount: 45_000_00, IVARate: fiscal.IVARate16, IVA: 7_200_00,
		Dates: domain.Dates{InvoicedOn: incDate(t, "2026-09-05")}, CFDIUUID: "6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f",
		Description: "  Consultoría  ",
	}

	got, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(1), TaxpayerID: pm, Income: in})
	require.NoError(t, err)
	assert.True(t, got.Created)
	assert.Equal(t, "2026-09-05", got.Income.BookedOn.String(), "accrual: the trigger books on invoiced_on")
	assert.True(t, got.Income.PendingCollection)
	assert.Equal(t, "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F", got.Income.CFDIUUID.String())
	assert.Equal(t, "Consultoría", got.Income.Description)
	assert.False(t, got.Income.CreatedAt.IsZero())

	read, err := e.incomes.Get(t.Context(), incUserA, pm, incID(1))
	require.NoError(t, err)
	assert.Equal(t, got.Income, read)

	collected, err := e.incomes.Update(t.Context(), incUserA, pm, incID(1), incomes.Patch{CollectedOn: new(incDate(t, "2026-10-02"))})
	require.NoError(t, err)
	assert.False(t, collected.Income.PendingCollection)
	assert.Equal(t, "2026-09-05", collected.Income.BookedOn.String())
	assert.Equal(t, 1, e.incomesCreated(t, incUserA), "an edit takes no slot")

	second, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{TaxpayerID: pm, Income: in})
	require.NoError(t, err)
	assert.Equal(t, []domain.Warning{{Code: domain.WarningDuplicateCFDIUUID, Field: domain.FieldCFDIUUID}}, second.Warnings)
}

// US-I5: at 4/5, three concurrent creates admit exactly one; the counter
// stops at 5.
func TestIncomes_TrialCapRace(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	for n := 1; n <= access.TrialRecordCap-1; n++ {
		_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(n), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
		require.NoError(t, err)
	}
	const racers = 3
	// A blocker holds the counter row until every racer waits on it, so the
	// three reservations really overlap.
	blocker, err := e.app.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", incUserA)
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "SELECT 1 FROM user_usage WHERE user_id = $1 FOR UPDATE", incUserA)
	require.NoError(t, err)

	errs := make([]error, racers)
	var wg sync.WaitGroup
	for i := range racers {
		wg.Go(func() {
			_, errs[i] = e.incomes.Create(context.Background(), incUserA,
				incomes.CreateInput{ID: incID(100 + i), TaxpayerID: tp, Income: fees(t, "2026-09-11", 100_00)})
		})
	}
	waitForIncomeLockWaiters(t, e.app, racers)
	require.NoError(t, blocker.Commit())
	wg.Wait()

	created := 0
	for _, err := range errs {
		if err == nil {
			created++
			continue
		}
		fields := requireAppCode(t, err, apperr.CodeEntitlementRequired)
		require.NotEmpty(t, fields)
		assert.Equal(t, "incomes_remaining", fields[0].Field)
		assert.Equal(t, access.DetailTrialCapReached, fields[0].Code)
	}
	assert.Equal(t, 1, created, "exactly one 201")
	assert.Equal(t, access.TrialRecordCap, e.incomesCreated(t, incUserA))
	page, err := e.incomes.List(t.Context(), incUserA, incomes.ListQuery{TaxpayerID: tp, Period: "2026-09"})
	require.NoError(t, err)
	assert.Equal(t, int64(access.TrialRecordCap), page.Totals.Count, "refused creates rolled back")
}

// A double submit of the fifth income (same id, same body, both in flight)
// is one 201 and one 200 replay: the second never reads the counter the
// first just filled and answers 402 for a record that exists.
func TestIncomes_DoubleSubmitAtTheCapReplays(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	for n := 1; n <= access.TrialRecordCap-1; n++ {
		_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(n), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
		require.NoError(t, err)
	}
	blocker, err := e.app.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", incUserA)
	require.NoError(t, err)
	_, err = blocker.ExecContext(t.Context(), "SELECT 1 FROM user_usage WHERE user_id = $1 FOR UPDATE", incUserA)
	require.NoError(t, err)

	const submits = 2
	results := make([]incomes.WriteResult, submits)
	errs := make([]error, submits)
	var wg sync.WaitGroup
	for i := range submits {
		wg.Go(func() {
			results[i], errs[i] = e.incomes.Create(context.Background(), incUserA,
				incomes.CreateInput{ID: incID(100), TaxpayerID: tp, Income: fees(t, "2026-09-11", 100_00)})
		})
	}
	waitForIncomeLockWaiters(t, e.app, submits)
	require.NoError(t, blocker.Commit())
	wg.Wait()

	created := 0
	for i := range submits {
		require.NoError(t, errs[i], "both submits answer the stored income")
		assert.Equal(t, incID(100), results[i].Income.ID)
		if results[i].Created {
			created++
		}
	}
	assert.Equal(t, 1, created, "exactly one 201, the other a replay")
	assert.Equal(t, access.TrialRecordCap, e.incomesCreated(t, incUserA))
}

func waitForIncomeLockWaiters(t *testing.T, app *sql.DB, n int) {
	t.Helper()
	const query = "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock'"
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		require.NoError(t, app.QueryRowContext(t.Context(), query).Scan(&waiting))
		if waiting >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("fewer than %d racers ever waited on the counter lock", n)
}

// US-I1c: the same id and body replay the stored income (200) without a
// slot, even at 5/5; other fields are 409 id_reused; another user's id is a
// bare 409.
func TestIncomes_Idempotency(t *testing.T) {
	e := newIncomesEnv(t, incUserA, incUserB)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	tpB := e.taxpayer(t, incUserB, "fisica", incRFCFisica, "612")
	for n := 1; n <= access.TrialRecordCap; n++ {
		_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(n), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
		require.NoError(t, err)
	}

	again, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(5), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
	require.NoError(t, err, "a retry at 5/5 is never a 402")
	assert.False(t, again.Created)
	assert.Equal(t, incID(5), again.Income.ID)
	assert.Equal(t, access.TrialRecordCap, e.incomesCreated(t, incUserA))

	_, err = e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(5), TaxpayerID: tp, Income: fees(t, "2026-09-10", 200_00)})
	fields := requireAppCode(t, err, apperr.CodeConflict)
	require.Len(t, fields, 1)
	assert.Equal(t, "id_reused", fields[0].Code)

	_, err = e.incomes.Create(t.Context(), incUserB, incomes.CreateInput{ID: incID(5), TaxpayerID: tpB, Income: fees(t, "2026-09-10", 100_00)})
	assert.Empty(t, requireAppCode(t, err, apperr.CodeConflict), "a foreign id reveals no ownership")
	assert.Zero(t, e.incomesCreated(t, incUserB), "the refused insert gave its slot back")
}

// US-I6: user B never reaches user A's taxpayer or income: 404, never 403.
func TestIncomes_CrossUserIs404(t *testing.T) {
	e := newIncomesEnv(t, incUserA, incUserB)
	tpA := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	tpB := e.taxpayer(t, incUserB, "fisica", incRFCOther, "612")
	_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(1), TaxpayerID: tpA, Income: fees(t, "2026-09-10", 100_00)})
	require.NoError(t, err)

	for _, taxpayerID := range []string{tpA, tpB} {
		_, err = e.incomes.Get(t.Context(), incUserB, taxpayerID, incID(1))
		requireAppCode(t, err, apperr.CodeNotFound)
		_, err = e.incomes.Update(t.Context(), incUserB, taxpayerID, incID(1), incomes.Patch{Description: new("x")})
		requireAppCode(t, err, apperr.CodeNotFound)
		requireAppCode(t, e.incomes.Delete(t.Context(), incUserB, taxpayerID, incID(1)), apperr.CodeNotFound)
		_, err = e.incomes.Restore(t.Context(), incUserB, taxpayerID, incID(1))
		requireAppCode(t, err, apperr.CodeNotFound)
	}
	_, err = e.incomes.List(t.Context(), incUserB, incomes.ListQuery{TaxpayerID: tpA, Period: "2026-09"})
	requireAppCode(t, err, apperr.CodeNotFound)
	_, err = e.incomes.Create(t.Context(), incUserB, incomes.CreateInput{TaxpayerID: tpA, Income: fees(t, "2026-09-10", 100_00)})
	requireAppCode(t, err, apperr.CodeNotFound)

	got, err := e.incomes.Get(t.Context(), incUserA, tpA, incID(1))
	require.NoError(t, err, "untouched")
	assert.Empty(t, got.Description)
}

// US-I4: at 5/5, delete + restore is 200 and incomes_created stays 5; a
// deleted income never frees a slot.
func TestIncomes_RestoreNeverDoubleCounts(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	for n := 1; n <= access.TrialRecordCap; n++ {
		_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(n), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
		require.NoError(t, err)
	}
	require.NoError(t, e.incomes.Delete(t.Context(), incUserA, tp, incID(3)))
	require.NoError(t, e.incomes.Delete(t.Context(), incUserA, tp, incID(3)), "idempotent")
	_, err := e.incomes.Get(t.Context(), incUserA, tp, incID(3))
	requireAppCode(t, err, apperr.CodeNotFound)

	_, err = e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
	requireAppCode(t, err, apperr.CodeEntitlementRequired)

	restored, err := e.incomes.Restore(t.Context(), incUserA, tp, incID(3))
	require.NoError(t, err)
	assert.Equal(t, incID(3), restored.Income.ID)
	assert.Equal(t, access.TrialRecordCap, e.incomesCreated(t, incUserA))
	page, err := e.incomes.List(t.Context(), incUserA, incomes.ListQuery{TaxpayerID: tp, Period: "2026-09"})
	require.NoError(t, err)
	assert.Equal(t, int64(access.TrialRecordCap), page.Totals.Count)
}

// A read-only user lists, deletes and restores (coordinator decision);
// creates and edits are 402.
func TestIncomes_ReadOnlyUser(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(1), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
	require.NoError(t, err)

	_, err = e.readOnly.Create(t.Context(), incUserA, incomes.CreateInput{TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
	requireAppCode(t, err, apperr.CodeEntitlementRequired)
	_, err = e.readOnly.Update(t.Context(), incUserA, tp, incID(1), incomes.Patch{Description: new("x")})
	requireAppCode(t, err, apperr.CodeEntitlementRequired)

	require.NoError(t, e.readOnly.Delete(t.Context(), incUserA, tp, incID(1)))
	_, err = e.readOnly.Restore(t.Context(), incUserA, tp, incID(1))
	require.NoError(t, err)
	page, err := e.readOnly.List(t.Context(), incUserA, incomes.ListQuery{TaxpayerID: tp, Period: "2026-09"})
	require.NoError(t, err)
	assert.Len(t, page.Items, 1)
}

func TestIncomes_ArchivedTaxpayerFreezesTheData(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(1), TaxpayerID: tp, Income: fees(t, "2026-09-10", 100_00)})
	require.NoError(t, err)
	_, err = e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(2), TaxpayerID: tp, Income: fees(t, "2026-09-11", 100_00)})
	require.NoError(t, err)
	require.NoError(t, e.incomes.Delete(t.Context(), incUserA, tp, incID(2)))
	_, err = e.taxpayers.Archive(t.Context(), incUserA, tp)
	require.NoError(t, err)

	page, err := e.incomes.List(t.Context(), incUserA, incomes.ListQuery{TaxpayerID: tp, Period: "2026-09"})
	require.NoError(t, err, "archived taxpayers are listed")
	assert.Len(t, page.Items, 1)
	archived := func(err error) {
		t.Helper()
		fields := requireAppCode(t, err, apperr.CodeBusinessRuleViolation)
		require.NotEmpty(t, fields)
		assert.Equal(t, "taxpayer_archived", fields[0].Code)
	}
	archived(e.incomes.Delete(t.Context(), incUserA, tp, incID(1)))
	_, err = e.incomes.Restore(t.Context(), incUserA, tp, incID(2))
	archived(err)
	_, err = e.incomes.Update(t.Context(), incUserA, tp, incID(1), incomes.Patch{Description: new("x")})
	archived(err)
	_, err = e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{TaxpayerID: tp, Income: fees(t, "2026-09-12", 100_00)})
	archived(err)
}

// US-I2: keyset pages newest first, period and regime filters, totals of
// the whole filtered period, deleted incomes left out.
func TestIncomes_ListKeysetFiltersAndTotals(t *testing.T) {
	e := newIncomesEnv(t, incUserA)
	tp := e.taxpayer(t, incUserA, "fisica", incRFCFisica, "612")
	days := []string{"2026-09-03", "2026-09-03", "2026-09-17", "2026-09-30", "2026-08-31"} // the trial's 5
	for i, day := range days {
		_, err := e.incomes.Create(t.Context(), incUserA, incomes.CreateInput{ID: incID(i + 1), TaxpayerID: tp, Income: fees(t, day, money.Cents(100_00*(i+1)))})
		require.NoError(t, err)
	}
	require.NoError(t, e.incomes.Delete(t.Context(), incUserA, tp, incID(3)))

	var seen []string
	query := incomes.ListQuery{TaxpayerID: tp, Period: "2026-09", Regime: "612", Limit: 2}
	for page := 0; ; page++ {
		require.Less(t, page, 5)
		got, err := e.incomes.List(t.Context(), incUserA, query)
		require.NoError(t, err)
		assert.Equal(t, incomes.PeriodTotals{Count: 3, Amount: (100 + 200 + 400) * 100}, got.Totals)
		for _, item := range got.Items {
			seen = append(seen, item.ID)
		}
		if got.NextCursor == "" {
			break
		}
		query.Cursor = got.NextCursor
	}
	assert.Equal(t, []string{incID(4), incID(2), incID(1)}, seen, "(booked_on, id) descending, deleted and August left out")

	other, err := e.incomes.List(t.Context(), incUserA, incomes.ListQuery{TaxpayerID: tp, Period: "2026-09", Regime: "606"})
	require.NoError(t, err)
	assert.Empty(t, other.Items)
	assert.Zero(t, other.Totals.Count)
}
