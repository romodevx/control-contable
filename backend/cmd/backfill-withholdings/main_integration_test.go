//go:build integration

package main

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/cmd/internal/profiles"
	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	expenses "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	tpinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// Fixture ids: this suite's own.
const (
	userA = "0199d4a2-0000-7000-8000-0000000bf00a"
	userB = "0199d4a2-0000-7000-8000-0000000bf00b"

	resicoA = "0199d4a2-1111-7000-8000-0000000bf0a1" // userA's física in RESICO (626)
	moralA  = "0199d4a2-1111-7000-8000-0000000bf0a2" // userA's persona moral (601)
	fisicaB = "0199d4a2-1111-7000-8000-0000000bf0b1" // userB's física in 612

	ownerIncome   = "0199d4a3-0000-7000-8000-0000000bf001" // US-W1, typed with no withholding
	feesIncome    = "0199d4a3-0000-7000-8000-0000000bf002" // US-I1, typed right
	olderIncome   = "0199d4a3-0000-7000-8000-0000000bf003" // derived by table 2025.9
	consultantFee = "0199d4a3-0000-7000-8000-0000000bf004" // US-W6, typed right
	deletedBuy    = "0199d4a3-0000-7000-8000-0000000bf005" // PM expense deleted before WH-07
)

// today is the derivation's clock: 13 October 2026 in Mexico City.
var today = time.Date(2026, time.October, 13, 12, 0, 0, 0, clock.Location())

// insertUserSQL registers user $1 as cc_app does at sign-up. The credential
// column holds an opaque PHC-shaped value derived from the id.
const insertUserSQL = `
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
VALUES ($1::text::uuid, sha256(convert_to($1::text, 'UTF8')),
        '\x0001'::bytea || sha256(convert_to($1::text, 'UTF8')) || sha256(convert_to($1::text, 'UTF8')),
        1, '$' || 'argon2id$v=19$m=1,t=1,p=1$' || md5($1::text) || '$' || md5($1::text || '/k'), 'Ana')`

const insertTaxpayerSQL = `
INSERT INTO taxpayers (id, user_id, person_type, display_name, rfc_encrypted, rfc_key_id, rfc_hash, rfc_masked,
                       rfc_check_digit_valid)
VALUES ($2::text::uuid, $1, $3, 'Ana (respaldo)',
        '\x0001'::bytea || sha256(convert_to($2::text, 'UTF8')) || sha256(convert_to($2::text, 'UTF8')),
        1, sha256(convert_to($2::text || '/rfc', 'UTF8')),
        CASE $3 WHEN 'fisica' THEN 'GOMA******AB6' ELSE 'GOM******AB6' END, true)`

const insertRegimeSQL = `
INSERT INTO taxpayer_regimes (user_id, taxpayer_id, regime_code, person_type, start_period)
VALUES ($1, $2, $3, $4, '2026-01')`

// insertLegacyIncomeSQL is an income as M3 wrote it before WH-06: the
// counterparty_type and the typed withholdings only.
const insertLegacyIncomeSQL = `
INSERT INTO incomes (id, user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
                     isr_withheld_cents, iva_withheld_cents, invoiced_on, collected_on)
VALUES ($1, $2, $3, $4, 'persona_moral', $5, 'rate_16', $6, $7, $8, $9, $10)`

// insertOlderIncomeSQL is US-I1 derived by an older table (2025.9) with
// other amounts.
const insertOlderIncomeSQL = `
INSERT INTO incomes (id, user_id, taxpayer_id, regime_code, counterparty_type, counterparty_kind, operation_kind,
                     amount_cents, iva_rate, iva_cents, isr_withheld_cents, iva_withheld_cents, invoiced_on,
                     collected_on, withholding_rule_ids, withholding_table_version, withholding_isr_rate,
                     withholding_iva_basis, withholding_withholder)
VALUES ($1, $2, $3, '612', 'persona_moral', 'pm', 'professional_services', 1000000, 'rate_16', 160000, 1, 1,
        '2026-09-05', '2026-09-10', '{ISR-5,IVA-6}', '2025.9', 0.10, 'two_thirds_of_iva', 'receptor')`

// insertLegacyExpenseSQL is a persona moral's expense as M4 wrote it before
// WH-07: no supplier facts, typed withholdings, the whole IVA creditable.
const insertLegacyExpenseSQL = `
INSERT INTO expenses (id, user_id, taxpayer_id, regime_code, category_code, amount_cents, iva_rate, iva_cents,
                      payment_method, has_cfdi, deductible, paid_on, creditable_iva_cents, isr_withheld_cents,
                      iva_withheld_cents)
VALUES ($1, $2, $3, '601', $4, $5, 'rate_16', $6, 'transfer', true, true, '2026-06-15', $6, $7, $8)`

type env struct {
	db     *pgtest.Database
	app    *sql.DB
	worker *sql.DB
}

// newEnv seeds both users' legacy records through cc_app, as M3/M4 stored
// them, and opens cc_worker's pool.
func newEnv(t *testing.T) *env {
	t.Helper()
	db := pgtest.NewDatabase(t)
	e := &env{db: db, app: db.Connect(t, pgtest.RoleApp), worker: db.Connect(t, pgtest.RoleWorker)}
	for _, user := range []string{userA, userB} {
		e.exec(t, user, insertUserSQL, user)
	}
	for _, tp := range []struct{ user, id, person, regime string }{
		{userA, resicoA, "fisica", "626"}, {userA, moralA, "moral", "601"}, {userB, fisicaB, "fisica", "612"},
	} {
		e.exec(t, tp.user, insertTaxpayerSQL, tp.user, tp.id, tp.person)
		e.exec(t, tp.user, insertRegimeSQL, tp.user, tp.id, tp.regime, tp.person)
	}
	e.exec(t, userA, insertLegacyIncomeSQL, ownerIncome, userA, resicoA, "626", 10_000_000, 1_600_000, 0, 0,
		"2026-07-08", "2026-07-10")
	e.exec(t, userB, insertLegacyIncomeSQL, feesIncome, userB, fisicaB, "612", 1_000_000, 160_000, 100_000, 106_667,
		"2026-09-05", "2026-09-10")
	e.exec(t, userB, insertOlderIncomeSQL, olderIncome, userB, fisicaB)
	e.exec(t, userA, insertLegacyExpenseSQL, consultantFee, userA, moralA, "servicios_profesionales", 4_000_000,
		640_000, 400_000, 426_667)
	e.exec(t, userA, insertLegacyExpenseSQL, deletedBuy, userA, moralA, "software", 50_000, 8_000, 0, 0)
	e.exec(t, userA, "UPDATE expenses SET deleted_at = now() WHERE id = $1", deletedBuy)
	return e
}

func (e *env) exec(t *testing.T, userID, statement string, args ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), statement, args...)
		return err
	}))
}

func (e *env) query(t *testing.T, userID, query string, args []any, dest ...any) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, userID, func(tx *sql.Tx) error {
		return tx.QueryRowContext(t.Context(), query, args...).Scan(dest...)
	}))
}

// snapshot is every row of both tables as each user sees them, audit
// columns included: equal snapshots mean nothing was written.
func (e *env) snapshot(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	for _, user := range []string{userA, userB} {
		for _, table := range []string{"incomes", "expenses"} {
			var rows sql.NullString
			e.query(t, user, "SELECT string_agg(r::text, E'\\n' ORDER BY r.id) FROM "+table+" r", nil, &rows)
			b.WriteString(rows.String)
		}
	}
	return b.String()
}

// backfill runs the command's core as cc_worker and returns its report.
func (e *env) backfill(t *testing.T, apply bool) string {
	t.Helper()
	var out bytes.Buffer
	require.NoError(t, backfill(t.Context(), e.worker, clock.NewFixed(today), apply, &out, logging.New(io.Discard, nil)))
	return out.String()
}

func TestBackfill_DryRunReportsIDsAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	before := e.snapshot(t)

	out := e.backfill(t, false)

	assert.Equal(t, before, e.snapshot(t), "a dry run writes nothing")
	assert.Contains(t, out, "DRY RUN")
	assert.Contains(t, out, "users visited: 2")
	assert.Contains(t, out, "incomes: 3 to re-derive, 2 with different amounts, 2 with assumed facts, 0 need user input")
	assert.Contains(t, out, "expenses: 2 to re-derive, 0 with different amounts, 2 with assumed facts, 0 need user input")
	for _, id := range []string{ownerIncome, olderIncome, consultantFee, deletedBuy} {
		assert.Contains(t, out, id)
	}
	for _, leak := range []string{userA, userB, "100000", "1250", "Ana", "426667"} {
		assert.NotContains(t, out, leak, "ids and counts only")
	}
}

func TestBackfill_ApplyDerivesExactValuesAndASecondRunChangesNothing(t *testing.T) {
	e := newEnv(t)

	e.backfill(t, true)

	var kind, operation, rules, version string
	var isr, iva int64
	e.query(t, userA, `SELECT counterparty_kind, operation_kind, isr_withheld_cents, iva_withheld_cents,
		array_to_string(withholding_rule_ids, ','), withholding_table_version FROM incomes WHERE id = $1`,
		[]any{ownerIncome}, &kind, &operation, &isr, &iva, &rules, &version)
	assert.Equal(t, []any{"pm", "professional_services", int64(125_000), int64(1_066_667), "ISR-4,IVA-6", fiscal.WithholdingTableVersion},
		[]any{kind, operation, isr, iva, rules, version}, "US-W1, the owner's case")

	e.query(t, userB, `SELECT isr_withheld_cents, iva_withheld_cents, withholding_table_version FROM incomes WHERE id = $1`,
		[]any{olderIncome}, &isr, &iva, &version)
	assert.Equal(t, []any{int64(100_000), int64(106_667), fiscal.WithholdingTableVersion}, []any{isr, iva, version},
		"an older table version is re-derived")

	var supplier, regime string
	var paid, withheld int64
	e.query(t, userA, `SELECT supplier_kind, supplier_regime, operation_kind, isr_withheld_cents, iva_withheld_cents,
		creditable_iva_cents, creditable_iva_withheld_cents FROM expenses WHERE id = $1`,
		[]any{consultantFee}, &supplier, &regime, &operation, &isr, &iva, &paid, &withheld)
	assert.Equal(t, []any{"pf_with_activity", "612", "professional_services", int64(400_000), int64(426_667), int64(213_333), int64(426_667)},
		[]any{supplier, regime, operation, isr, iva, paid, withheld}, "US-W6")

	after := e.snapshot(t)
	out := e.backfill(t, true)
	assert.Equal(t, after, e.snapshot(t), "a second run writes nothing")
	assert.Contains(t, out, "incomes: 0 to re-derive")
	assert.Contains(t, out, "expenses: 0 to re-derive")
}

// Security review of #139: a persona moral's expense deleted before WH-07
// had no supplier facts, so restoring it answered 400. After the backfill
// the restore succeeds.
func TestBackfill_DeletedLegacyExpenseCanBeRestored(t *testing.T) {
	e := newEnv(t)
	uc := e.expensesAsApp(t)

	_, err := uc.Restore(t.Context(), userA, moralA, deletedBuy)
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "before the backfill: %v", err)
	assert.Equal(t, apperr.CodeValidationFailed, app.Code)

	e.backfill(t, true)

	restored, err := uc.Restore(t.Context(), userA, moralA, deletedBuy)
	require.NoError(t, err)
	assert.Equal(t, expdomain.SupplierPM, restored.Supplier.Kind)
	assert.Equal(t, fiscal.WithholdingTableVersion, restored.Withholding.TableVersion)
}

// expensesAsApp wires the expense use cases as cmd/api does, as cc_app.
func (e *env) expensesAsApp(t *testing.T) *expenses.Expenses {
	t.Helper()
	txm, err := pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	sets, err := params.Embedded()
	require.NoError(t, err)
	rates, err := ratecatalog.New[expdomain.AmountKey](sets)
	require.NoError(t, err)
	queries := taxpayers.NewQueries(taxpayers.Deps{
		Tx: txm, Taxpayers: tpinfra.NewStore(), Regimes: tpinfra.NewStore(), Catalog: tpinfra.NewCatalog(),
		Directory: tpinfra.NewCatalog(), Records: tpinfra.NoRecords{}, Clock: clock.NewFixed(today),
	})
	return expenses.NewExpenses(expenses.Deps{
		Tx: txm, Regimes: queries, Taxpayers: profiles.ExpensesReader{Reader: queries}, Expenses: expinfra.NewStore(),
		Categories: expinfra.NewCategoryCatalog(), Parameters: parameters{Catalog: rates},
		Withholdings: rates.Withholdings(), Clock: clock.NewFixed(today),
	})
}

// parameters mirrors cmd/api's expenseParameters (package main of another
// command cannot be imported).
type parameters struct {
	*ratecatalog.Catalog[expdomain.AmountKey]
}

func (a parameters) On(on fiscal.Date) (expenses.DatedParameters, error) {
	limit, err := a.CashPaymentLimit(on)
	if err != nil {
		return expenses.DatedParameters{}, err
	}
	rates, err := a.AllowedIVARates(on)
	if err != nil {
		return expenses.DatedParameters{}, err
	}
	return expenses.DatedParameters{IVARatesAllowed: rates.Value, CashPaymentLimit: limit.Value}, nil
}

// cc_worker's grants (migration 20261018100000) stay inside the user's
// context: it cannot see or write another user's rows, sees nothing without
// a context, reads only the users' ids, and writes only the derived columns.
func TestWorkerRole_RLSHolds(t *testing.T) {
	e := newEnv(t)
	ctx := t.Context()
	for _, table := range []pgtest.Table{
		{Name: "incomes", OwnerColumn: "user_id"}, {Name: "expenses", OwnerColumn: "user_id"},
		{Name: "taxpayers", OwnerColumn: "user_id"}, {Name: "taxpayer_regimes", OwnerColumn: "user_id"},
	} {
		t.Run(table.Name, func(t *testing.T) {
			owner, viewer := userB, userA
			if table.Name == "expenses" {
				owner, viewer = userA, userB
			}
			require.NoError(t, pgtest.CheckRLSForced(ctx, e.worker, table))
			require.NoError(t, pgtest.CheckCannotSee(ctx, e.worker, table, viewer, owner))
			require.NoError(t, pgtest.CheckUnsetContextFailsClosed(ctx, e.worker, table, owner))
			require.NoError(t, pgtest.CheckNoBypass(ctx, e.worker, table))
		})
	}
	t.Run("cross-user writes of the granted columns match no row", func(t *testing.T) {
		for _, statement := range []string{
			"UPDATE incomes SET operation_kind = 'other' WHERE user_id = $1",
			"UPDATE expenses SET supplier_kind = 'pm' WHERE user_id = $1",
		} {
			var affected int64
			require.NoError(t, pgtest.AsUser(ctx, e.worker, userB, func(tx *sql.Tx) error {
				res, err := tx.ExecContext(ctx, statement, userA)
				if err != nil {
					return err
				}
				affected, err = res.RowsAffected()
				return err
			}))
			assert.Zero(t, affected, statement)
		}
	})
	t.Run("the user list comes from the definer function only", func(t *testing.T) {
		var listed int
		require.NoError(t, pgtest.WithoutUser(ctx, e.worker, func(tx *sql.Tx) error {
			return tx.QueryRowContext(ctx, "SELECT count(*) FROM withholding_backfill_users()").Scan(&listed)
		}))
		assert.Equal(t, 2, listed)
		err := pgtest.WithoutUser(ctx, e.worker, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "SELECT email_hash FROM users")
			return err
		})
		require.Error(t, err, "cc_worker reads no users column but the ones the key re-seal rewrites (ADR-0012)")
		assert.Contains(t, err.Error(), "42501")
		err = pgtest.WithoutUser(ctx, e.app, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "SELECT withholding_backfill_users()")
			return err
		})
		require.Error(t, err, "cc_app may not list users")
		assert.Contains(t, err.Error(), "42501")
	})
	t.Run("only the derived columns", func(t *testing.T) {
		for _, statement := range []string{
			"SELECT name FROM users",
			"SELECT rfc_hash FROM taxpayers", // rfc_encrypted is the key re-seal's (ADR-0012)
			"UPDATE incomes SET amount_cents = 1",
			"UPDATE expenses SET deductible = false",
			"UPDATE incomes SET deleted_at = NULL",
			"INSERT INTO incomes (id) VALUES ('0199d4a3-0000-7000-8000-0000000bf0ff')",
			"DELETE FROM expenses",
		} {
			err := pgtest.AsUser(ctx, e.worker, userA, func(tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, statement)
				return err
			})
			require.Error(t, err, statement)
			assert.Contains(t, err.Error(), "42501", statement)
		}
	})
}
